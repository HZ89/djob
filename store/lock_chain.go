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
	"sync"
	"time"

	libstore "github.com/HZ89/libkv/store"
	"github.com/Sirupsen/logrus"

	"github.com/HZ89/djob/errors"
	"github.com/HZ89/djob/log"
	"github.com/HZ89/djob/util"
)

type LockerChain struct {
	lockerOwner string    // the owner of this chain, node name
	chain       *sync.Map // store objKey, this is the chain
	lockStore   *KVStore  // a kv store, store lock
}

type objKey struct {
	locker   libstore.Locker // locker obj
	renewCh  chan struct{}   // lock renew chan
	bornTime time.Time       // the time to create this lock
}

type LockOption struct {
	Global   bool          // if true, the locker can access by any goroutine
	LockType lockType      // lock type
	TimeOut  time.Duration // if no lock in this time is successful return timeout
}

// new lockerchain
func NewLockerChain(owner string, store *KVStore) *LockerChain {
	chain := &LockerChain{
		chain:       new(sync.Map),
		lockerOwner: owner,
		lockStore:   store,
	}

	return chain
}

// release all lock and stop renew process
func (l *LockerChain) Stop() {
	l.ReleaseAll()
}

// HaveIt, return true if find the obj's Locker in LockerChain
func (l *LockerChain) HaveIt(obj interface{}, lockType lockType) bool {
	lockpath, _ := l.lockStore.buildLockKey(obj, lockType)
	_, exist := l.chain.Load(lockpath)
	return exist
}

// AddLocker, lock a obj and save locker into LockerChain
// if global is true, all goroutine can access this lock, if not only the goroutine created it can access it.
func (l *LockerChain) AddLocker(obj interface{}, lockoption *LockOption) error {

	var mapPath string
	lockOpt := &libstore.LockOptions{
		Value: []byte(l.lockerOwner),
		TTL:   lockTTL * time.Second,
	}

	kvPath, err := l.lockStore.buildLockKey(obj, lockoption.LockType)
	if err != nil {
		return err
	}

	mapPath = kvPath
	if !lockoption.Global {
		mapPath += fmt.Sprintf("###%d", util.GoroutineID())
	}

	// do not allow duplicate additions
	if l.HaveIt(obj, lockoption.LockType) {
		return errors.ErrRepetition
	}

	locker, err := l.lockStore.lock(kvPath, lockoption.TimeOut, lockOpt)
	if err != nil {
		log.FmdLoger.WithFields(logrus.Fields{
			"mapPath": mapPath,
			"kvPath":  kvPath,
		}).WithError(err).Debug("Store-Lock: store a lock into kv failed")
		return err
	}
	objlock := &objKey{
		locker:   locker,
		bornTime: time.Now(),
	}
	l.chain.Store(mapPath, objlock)
	log.FmdLoger.WithFields(logrus.Fields{
		"mapPath": mapPath,
		"kvPath":  kvPath,
		"value":   objlock,
	}).Debug("Store-Lock: store a lock into map succeed")
	return nil
}

// ReleaseLocker release a lock
// if global is true, all goroutine can access this lock, if not only the goroutine created it can access it.
func (l *LockerChain) ReleaseLocker(obj interface{}, global bool, lockType lockType) error {

	var mapPath string
	kvPath, err := l.lockStore.buildLockKey(obj, lockType)
	if err != nil {
		return err
	}

	mapPath = kvPath
	if !global {
		mapPath += fmt.Sprintf("###%d", util.GoroutineID())
	}

	v, ok := l.chain.Load(mapPath)
	if !ok {
		return errors.ErrNotExist
	}
	t, ok := v.(*objKey)
	if !ok {
		log.FmdLoger.Fatalf("Store-Lock: lockerChain got a unknown type: %v", v)
	}
	if err = t.locker.Unlock(); err != nil {
		return err
	}
	l.chain.Delete(mapPath)
	log.FmdLoger.WithFields(logrus.Fields{
		"mapPath": mapPath,
		"kvPath":  kvPath,
		"value":   t,
	}).Debug("Store-Lock: delete a lock from map")
	return nil
}

// ReleaseAll release all lock
func (l *LockerChain) ReleaseAll() {
	l.chain.Range(func(key, v interface{}) bool {
		t, ok := v.(*objKey)
		if !ok {
			log.FmdLoger.Fatalf("Store-Lock: lockerChain got a unknown type: %v", v)
		}
		t.locker.Unlock()
		l.chain.Delete(key)
		return true
	})
}
