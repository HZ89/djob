/*
 * Copyright (c) 2017.  Harrison Zhu <wcg6121@gmail.com>
 * This file is part of djob <https://github.com/HZ89/djob>.
 *
 * djob is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * djob is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with djob.  If not, see <http://www.gnu.org/licenses/>.
 */

package store

import (
	"fmt"
	"hash/fnv"
	"path"
	"reflect"
	"runtime"
	"strings"
	"time"

	"github.com/HZ89/libkv"
	libstore "github.com/HZ89/libkv/store"
	"github.com/HZ89/libkv/store/consul"
	"github.com/HZ89/libkv/store/etcd"
	"github.com/HZ89/libkv/store/zookeeper"
	"github.com/Sirupsen/logrus"

	"github.com/HZ89/djob/errors"
	"github.com/HZ89/djob/log"
	"github.com/HZ89/djob/util"
)

func init() {
	etcd.Register()
	consul.Register()
	zookeeper.Register()
}

// some const
const (
	lockTTL = 60
)

type lockType int

const (
	OWN lockType = 1 + iota
	RW
	W
)

var lockTypes = []string{
	"OWN",
	"READWRITE",
	"WRITE",
}

func (lt lockType) String() string {
	return lockTypes[lt-1]
}

// KVStore struct
type KVStore struct {
	client       libstore.Store // client of store backend
	keyspace     string         // perfix of key
	backend      string         // backend, support etcd, zk etc...
	maxWorkerNum int
	workerList   []*workerSpeker
}

type operation struct {
	key       string
	valueType interface{}
	ops       func(interface{}) error
}

type workerSpeker struct {
	input  chan *operation
	output chan error
	stop   chan struct{}
}

// NewKVStore used to init a KVStore struct
func NewKVStore(backend string, servers []string, keyspace string) (*KVStore, error) {
	s, err := libkv.NewStore(libstore.Backend(backend), servers, nil)
	if err != nil {
		return nil, err
	}
	log.FmdLoger.WithFields(logrus.Fields{
		"backend":  backend,
		"servers":  servers,
		"keyspace": keyspace,
	}).Debug("store: Backend Connected")

	_, err = s.List(keyspace)
	if err != libstore.ErrKeyNotFound && err != nil {
		return nil, err
	}
	k := &KVStore{
		client:       s,
		backend:      backend,
		keyspace:     keyspace,
		maxWorkerNum: runtime.NumCPU(),
	}

	k.startAtomicWorker()

	return k, nil
}

// build lock key path
func (k *KVStore) buildLockKey(obj interface{}, lockType lockType) (key string, err error) {

	name := util.GetFieldValue(obj, "Name")
	region := util.GetFieldValue(obj, "Region")
	if name == nil || region == nil {
		log.FmdLoger.WithFields(logrus.Fields{
			"ObjType":         util.GetType(obj),
			"NameFieldType":   util.GetFieldKind(obj, "Name"),
			"RegionFieldType": util.GetFieldKind(obj, "Region"),
		}).Debug("Store: build lock key failed")
		return "", errors.ErrType
	}
	key = fmt.Sprintf("%s/%s/%s_locks/%s###%s",
		k.keyspace,
		util.GenerateSlug(region.(string)),
		util.GetType(obj),
		util.GenerateSlug(name.(string)),
		lockType)
	return
}

// WatchLock watch lock path, put any change key into chan
func (k *KVStore) WatchLock(obj interface{}, stopCh chan struct{}) (chan string, error) {
	stopWatchCh := make(chan struct{})
	outCh := make(chan string, 7)
	region := util.GetFieldValue(obj, "Region")
	watchPath := fmt.Sprintf("%s/%s/%s_locks", k.keyspace, util.GenerateSlug(region.(string)), util.GetType(obj))

	kpCh, err := k.client.WatchTree(watchPath, stopWatchCh)
	if err == libstore.ErrKeyNotFound {
		return nil, errors.ErrNotExist
	}
	if err != nil {
		return nil, err
	}
	go func() {
		defer close(outCh)
		for {
			select {
			case kps := <-kpCh:
				if len(kps) == 0 {
					continue
				}
				for _, kp := range kps {
					v := strings.Split(kp.Key, "###")
					outCh <- path.Base(v[0])
				}
			case <-stopCh:
				stopWatchCh <- struct{}{}
				return
			}
		}
	}()
	return outCh, nil
}

// IsLocked return true if have a lock key in kv store
func (k *KVStore) IsLocked(obj interface{}, lockType lockType) (r bool) {
	v := k.WhoLocked(obj, lockType)
	if v != "" {
		return true
	}
	return false
}

// WhoLocked return the lock's value
func (k *KVStore) WhoLocked(obj interface{}, lockType lockType) (v string) {
	key, _ := k.buildLockKey(obj, lockType)
	locker, err := k.client.Get(key)
	if err != nil && err != libstore.ErrKeyNotFound {
		log.FmdLoger.WithField("key", key).WithError(err).Fatal("Store: get key failed")
	}
	if locker != nil {
		v = string(locker.Value)
	}
	return
}

// lock create a obj's lock with timeout
func (k *KVStore) lock(key string, lockTimeOut time.Duration, lockOpt *libstore.LockOptions) (locker libstore.Locker, err error) {

	locker, err = k.client.NewLock(key, lockOpt)
	if err != nil {
		log.FmdLoger.WithField("key", key).WithError(err).Fatal("Store: New lock failed")
	}

	timeOutCh := time.After(lockTimeOut)
	freeCh := make(chan struct{})
	errorCh := make(chan error)
	stopLockCh := make(chan struct{})

	// start a goroutine perform locking
	go func() {
		_, err = locker.Lock(stopLockCh)
		if err != nil {
			errorCh <- err
			return
		}
		freeCh <- struct{}{}
		return
	}()

	for {
		select {
		case <-freeCh:
			return locker, nil
		case err := <-errorCh:
			return nil, err
		case <-timeOutCh:
			if lockTimeOut != 0 {
				stopLockCh <- struct{}{}
				return nil, errors.ErrLockTimeout
			}
		}
	}

}

func (k *KVStore) buildKey(obj interface{}) string {
	objRegion := util.GetFieldValue(obj, "Region")
	objName := util.GetFieldValue(obj, "Name")
	regionString := objRegion.(string)
	nameString := objName.(string)

	return fmt.Sprintf("%s/%s/%s/%s",
		k.keyspace, util.GenerateSlug(regionString), util.GetType(obj),
		util.GenerateSlug(nameString))
}

// Delete from kv store
func (k *KVStore) Delete(obj interface{}) (err error) {
	key := k.buildKey(obj)
	if err := k.client.Delete(key); err != nil {
		if err == libstore.ErrKeyNotFound {
			return errors.ErrNotExist
		}
		return err
	}
	return nil
}

// Get obj from kv store, fill found data into give obj and return lastIndex
func (k *KVStore) Get(in interface{}) (lastIndex uint64, err error) {

	key := k.buildKey(in)
	res, err := k.client.Get(key)
	if err != nil && err != libstore.ErrKeyNotFound {
		return 0, err
	}

	// fill data to status
	if res == nil {
		return 0, errors.ErrNotExist
	}
	if err = util.JsonToPb(res.Value, in); err != nil {
		return 0, err
	}
	return res.LastIndex, nil
}

// Set obj to kv store, if lastIndex and previousObj is not nil, it will use atomicPut
func (k *KVStore) Set(currentObj interface{}, lastIndex uint64, previousObj interface{}) error {

	e, err := util.PbToJSON(currentObj)
	if err != nil {
		return err
	}
	key := k.buildKey(currentObj)
	if lastIndex != 0 && previousObj != nil {
		pe, err := util.PbToJSON(previousObj)
		if err != nil {
			return nil
		}
		pkp := &libstore.KVPair{
			Key:       key,
			Value:     pe,
			LastIndex: lastIndex,
		}
		ok, _, err := k.client.AtomicPut(key, e, pkp, nil)
		if !ok {
			if err == libstore.ErrKeyExists {
				return errors.ErrNotExist
			}
			if err == libstore.ErrKeyModified {
				return errors.ErrKeyModified
			}
			return err
		}
	} else {
		return k.client.Put(key, e, nil)
	}
	return nil
}

// AtomicOps picks a goroutine order of execution f
// for a key operation assigned to a fixed goroutine operation to avoid
// generating concurrent operations on the kv backend
func (k *KVStore) AtomicOps(targetObj interface{}, f func(value interface{}) error) error {

	key := k.buildKey(targetObj)
	// pick goroutine
	h := fnv.New64a()
	h.Write([]byte(key))
	i := h.Sum64() % uint64(k.maxWorkerNum)
	// send operation to it
	k.workerList[i].input <- &operation{key: key, ops: f, valueType: targetObj}

	err := <-k.workerList[i].output
	if err != errors.ErrNil {
		return err
	}
	return nil
}

func (k *KVStore) startAtomicWorker() {
	worker := func(input chan *operation, output chan error, stop chan struct{}) {
		for {
			select {
			case <-stop:
				close(input)
				return
			case ops := <-input:
				// read
				kp, err := k.client.Get(ops.key)
				if err != nil {
					output <- err
					continue
				}
				obj := reflect.New(util.IndirectType(reflect.TypeOf(ops.valueType))).Interface()
				if err := util.JsonToPb(kp.Value, obj); err != nil {
					output <- err
					continue
				}
				// perform
				if err := ops.ops(obj); err != nil {
					output <- err
					continue
				}
				v, err := util.PbToJSON(obj)
				if err != nil {
					output <- err
					continue
				}
				// write back
				if _, _, err := k.client.AtomicPut(ops.key, v, kp, nil); err != nil {
					output <- err
					continue
				}
				// succeed return a nil error
				output <- errors.ErrNil
			}
		}
	}

	for i := 0; i < k.maxWorkerNum; i++ {
		speker := &workerSpeker{
			input:  make(chan *operation),
			output: make(chan error),
			stop:   make(chan struct{}),
		}
		go worker(speker.input, speker.output, speker.stop)
		k.workerList = append(k.workerList, speker)
	}
}

func (k *KVStore) stopWorkers() {
	for _, w := range k.workerList {
		w.stop <- struct{}{}
	}
}
