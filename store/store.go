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
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/HZ89/libkv"
	libstore "github.com/HZ89/libkv/store"
	"github.com/HZ89/libkv/store/consul"
	"github.com/HZ89/libkv/store/etcd"
	"github.com/HZ89/libkv/store/zookeeper"
	"github.com/Sirupsen/logrus"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"

	"version.uuzu.com/zhuhuipeng/djob/errors"
	"version.uuzu.com/zhuhuipeng/djob/log"
	pb "version.uuzu.com/zhuhuipeng/djob/message"
	"version.uuzu.com/zhuhuipeng/djob/util"
)

func init() {
	etcd.Register()
	consul.Register()
	zookeeper.Register()
}

const (
	sqlMaxOpenConnect = 100
	sqlMaxIdleConnect = 20
	LockTimeOut       = 60 * time.Second
	defaultPageSize   = 10
	separator         = "###"
	lockTTL           = 60
)

type LockType int

const (
	OWN LockType = 1 + iota
	RW
	W
)

var lockTypes = []string{
	"OWN",
	"READWRITE",
	"WRITE",
}

func (lt LockType) String() string {
	return lockTypes[lt-1]
}

type LockerChain struct {
	lockerOwner string
	chain       *sync.Map
	lockStore   *KVStore
	stopCh      chan struct{}
}

type objKey struct {
	locker   libstore.Locker
	renewCh  chan struct{}
	bornTime time.Time
}

func NewLockerChain(owner string, store *KVStore) *LockerChain {
	chain := &LockerChain{
		chain:       new(sync.Map),
		lockerOwner: owner,
		lockStore:   store,
		stopCh:      make(chan struct{}),
	}
	chain.lockRenew()
	return chain
}

func (l *LockerChain) Stop() {
	l.ReleaseAll()
	l.stopCh <- struct{}{}
}

func (l *LockerChain) lockRenew() {
	go func() {
		log.FmdLoger.Debug("Store-Lock: start lockRenew goroutine")
		for {
			select {
			case <-time.After(lockTTL * 0.1 * time.Second):
				l.renew()
			case <-l.stopCh:
				return
			}
		}
	}()
}

func (l *LockerChain) renew() {
	now := time.Now()
	l.chain.Range(func(key, value interface{}) bool {
		t, ok := value.(*objKey)
		if !ok {
			return false
		}
		// remove expired lock
		// TODO: Need to throw out the key expired
		if now.After(t.bornTime.Add(lockTTL * time.Second)) {
			l.chain.Delete(key)
			return true
		}
		// if the remaining ttl less than lockTTL * 0.8, renew it
		if now.Sub(t.bornTime) < lockTTL*0.8*time.Second {
			t.renewCh <- struct{}{}
			t.bornTime = now
		}
		return true
	})
}

func (l *LockerChain) HaveIt(obj interface{}, lockType LockType) bool {
	lockpath, _ := l.lockStore.buildLockKey(obj, lockType)
	_, exist := l.chain.Load(lockpath)
	return exist
}

func (l *LockerChain) AddLocker(obj interface{}, lockType LockType) error {

	if l.HaveIt(obj, lockType) {
		return errors.ErrRepetition
	}

	renewCh := make(chan struct{})
	lockOpt := &libstore.LockOptions{
		Value:     []byte(l.lockerOwner),
		TTL:       lockTTL * time.Second,
		RenewLock: renewCh,
	}
	lockpath, err := l.lockStore.buildLockKey(obj, lockType)
	if err != nil {
		return err
	}
	locker, err := l.lockStore.lock(obj, lockType, lockOpt)
	if err != nil {
		return err
	}
	objlock := &objKey{
		renewCh:  renewCh,
		locker:   locker,
		bornTime: time.Now(),
	}
	l.chain.Store(lockpath, objlock)
	log.FmdLoger.WithFields(logrus.Fields{
		"key":   lockpath,
		"value": objlock,
	}).Debug("Store-Lock: store a lock into map")
	return nil
}

func (l *LockerChain) ReleaseLocker(obj interface{}, lockType LockType) error {

	lockpath, err := l.lockStore.buildLockKey(obj, lockType)
	if err != nil {
		return err
	}
	v, ok := l.chain.Load(lockpath)
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
	l.chain.Delete(lockpath)
	log.FmdLoger.WithFields(logrus.Fields{
		"key":   lockpath,
		"value": t,
	}).Debug("Store-Lock: delete a lock from map")
	return nil
}

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

type KVStore struct {
	client   libstore.Store
	keyspace string
	backend  string
}

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
	return &KVStore{
		client:   s,
		backend:  backend,
		keyspace: keyspace,
	}, nil
}

func (k *KVStore) buildLockKey(obj interface{}, lockType LockType) (key string, err error) {

	name := util.GetFieldValue(obj, "Name")
	region := util.GetFieldValue(obj, "Region")
	if name == nil || region == nil {
		log.FmdLoger.WithFields(logrus.Fields{
			"ObjType":         util.GetType(obj),
			"NameFieldType":   util.GetFieldType(obj, "Name"),
			"RegionFieldType": util.GetFieldType(obj, "Region"),
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
				log.FmdLoger.Debugf("Store: watch from %s got %v", watchPath, kps)
				for _, kp := range kps {
					v := strings.Split(kp.Key, "###")
					outCh <- v[0]
				}
			case <-stopCh:
				stopWatchCh <- struct{}{}
				return
			}
		}
	}()
	return outCh, nil
}

func (k *KVStore) IsLocked(obj interface{}, lockType LockType) (r bool) {
	v := k.WhoLocked(obj, lockType)
	if v != "" {
		return true
	}
	return false
}

func (k *KVStore) WhoLocked(obj interface{}, lockType LockType) (v string) {
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

func (k *KVStore) lock(obj interface{}, lockType LockType, lockOpt *libstore.LockOptions) (locker libstore.Locker, err error) {
	var key string

	key, err = k.buildLockKey(obj, lockType)
	if err != nil {
		return
	}
	locker, err = k.client.NewLock(key, lockOpt)
	if err != nil {
		log.FmdLoger.WithField("key", key).WithError(err).Fatal("Store: New lock failed")
	}

	errCh := make(chan error)
	freeCh := make(chan struct{})
	timeoutCh := time.After(LockTimeOut)
	stoplockingCh := make(chan struct{})

	go func() {
		_, err = locker.Lock(stoplockingCh)
		if err != nil {
			errCh <- err
			return
		}
		freeCh <- struct{}{}
	}()

	select {
	case <-freeCh:
		return locker, nil
	case err = <-errCh:
		return nil, err
	case <-timeoutCh:
		stoplockingCh <- struct{}{}
		return nil, errors.ErrLockTimeout
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

func (k *KVStore) DeleteJobStatus(status *pb.JobStatus) (out *pb.JobStatus, err error) {
	out, err = k.GetJobStatus(status)
	if err != nil {
		return
	}
	if err = k.client.Delete(k.buildKey(out)); err != nil {
		return nil, err
	}
	return
}

// Get JobStatus from KV store
// if the job do not exists, return nil
func (k *KVStore) GetJobStatus(in *pb.JobStatus) (out *pb.JobStatus, err error) {

	out = &pb.JobStatus{
		Name:   in.Name,
		Region: in.Region,
	}
	key := k.buildKey(out)
	res, err := k.client.Get(key)
	if err != nil && err != libstore.ErrKeyNotFound {
		return nil, err
	}

	// fill data to status
	if res == nil {
		return nil, errors.ErrNotExist
	}
	if err = util.JsonToPb(res.Value, out); err != nil {
		return nil, err
	}
	log.FmdLoger.WithFields(logrus.Fields{
		"Name":         out.Name,
		"Region":       out.Region,
		"SuccessCount": out.SuccessCount,
		"ErrorCount":   out.ErrorCount,
		"LastError":    out.LastError,
		"OrginData":    string(res.Value),
	}).Debug("Store: get jobstatus from kv store")
	return out, nil
}

// set job status to kv store
// safely set jobstatus need lock it with RW locker first
func (k *KVStore) SetJobStatus(status *pb.JobStatus) (*pb.JobStatus, error) {

	e, err := util.PbToJSON(status)
	if err != nil {
		return nil, err
	}
	log.FmdLoger.WithFields(logrus.Fields{
		"Name":         status.Name,
		"Region":       status.Region,
		"SuccessCount": status.SuccessCount,
		"ErrorCount":   status.ErrorCount,
		"LastSuccess":  status.LastSuccess,
		"LastError":    status.LastError,
		"OriginData":   string(e),
	}).Debug("Store: save job status to kv store")

	if err = k.client.Put(k.buildKey(status), e, nil); err != nil {
		return nil, err
	}
	return status, nil
}

// used to cache job in local memory
type MemStore struct {
	memBuf *memDbBuffer
	stopCh chan struct{}
}

type entry struct {
	DeadTime time.Time
	Data     []byte
}

func (m *MemStore) Set(key string, in interface{}, ttl time.Duration) (err error) {
	var value []byte
	e := &entry{}
	if ttl != 0 {
		e.DeadTime = time.Now().Add(ttl)
	} else {
		e.DeadTime = time.Time{}
	}
	e.Data, err = util.GetBytes(in)
	if err != nil {
		return
	}
	value, err = util.GetBytes(e)
	if err != nil {
		return
	}
	return m.memBuf.Set(Key(key), value)
}

func (m *MemStore) Get(key string, out interface{}) (err error) {
	var e *entry
	now := time.Now()
	e, err = m.getEntry(Key(key))
	if err != nil {
		return err
	}

	if !m.isOverdue(e, now) {
		return util.GetInterface(e.Data, out)
	}
	return errors.ErrNotExist
}

func (m *MemStore) getEntry(key Key) (*entry, error) {
	var (
		v   []byte
		err error
		e   entry
	)
	v, err = m.memBuf.Get(key)
	if err != nil {
		return nil, err
	}
	if len(v) == 0 {
		return nil, errors.ErrNotExist
	}
	err = util.GetInterface(v, &e)
	if err != nil {
		return nil, err
	}
	return &e, nil
}

func (m *MemStore) handleOverdueKey() {
	go func() {
		for true {
			select {
			case <-time.After(time.Second):
				m.releaseOverdueKey()
			case <-m.stopCh:
				return
			}
		}
	}()
}

func (m *MemStore) releaseOverdueKey() {
	now := time.Now()
	iter, err := m.memBuf.Seek(nil)
	if err != nil {
		log.FmdLoger.Fatalf("Store: memBuf Seek failed: %v", err)
	}
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		key := iter.Key()
		var e *entry
		e, err = m.getEntry(key)
		if err != nil && err != errors.ErrNotExist {
			log.FmdLoger.WithError(err).Fatalf("Store: memBuf Seek get key filed: key: %s, value: %v", string(key), e)
		}

		if e != nil {
			log.FmdLoger.WithFields(logrus.Fields{
				"key":       string(key),
				"deadTime":  e.DeadTime.String(),
				"isOverdue": m.isOverdue(e, now),
			}).Debug("Store: seek a key")
		}

		if e != nil && m.isOverdue(e, now) {
			if err = m.memBuf.Delete(key); err != nil {
				log.FmdLoger.WithField("key", string(key)).WithError(err).Fatal("Store: purge overdue key failed")
			}
		}
	}
}

func (m *MemStore) isOverdue(e *entry, now time.Time) bool {
	if e.DeadTime.IsZero() {
		return false
	}
	return e.DeadTime.Before(now)
}

func (m *MemStore) buildKey(name string, ttl time.Duration) (Key, error) {
	a := strings.Split(name, separator)
	if len(a) > 1 {
		return nil, errors.ErrIllegalCharacter
	}
	if ttl == 0 {
		return Key(name), nil
	}
	overtime := time.Now().Add(ttl)
	key := fmt.Sprintf("%s%s%s", name, separator, overtime.Format(time.RFC3339))
	return Key(key), nil
}

func (m *MemStore) Release() {
	m.stopCh <- struct{}{}
}

func NewMemStore() *MemStore {
	ms := &MemStore{
		memBuf: NewMemDbBuffer(),
		stopCh: make(chan struct{}),
	}
	ms.handleOverdueKey()
	return ms
}

type SQLStore struct {
	db           *gorm.DB
	sqlCondition *sqlCondition
	pageSize     int
	pageNum      int
	Err          error
}

func NewSQLStore(backend, dsn string) (*SQLStore, error) {
	db, err := gorm.Open(backend, dsn)
	if err != nil {
		return nil, err
	}

	db.DB().SetMaxOpenConns(sqlMaxOpenConnect)
	db.DB().SetMaxIdleConns(sqlMaxIdleConnect)
	db.LogMode(false)
	return &SQLStore{
		db: db,
	}, nil
}

func (s *SQLStore) Close() error {
	return s.db.Close()
}

func (s *SQLStore) Migrate(drop bool, valus ...interface{}) error {
	if !drop {
		if err := s.db.AutoMigrate(valus...).Error; err != nil {
			return err
		}
		return nil
	}
	return nil
}

func (s *SQLStore) Model(obj interface{}) *SQLStore {
	n := s.clone()
	n.db = s.db.Model(obj)
	return n
}

func (s *SQLStore) Where(obj interface{}) *SQLStore {
	n := s.clone()
	switch t := obj.(type) {
	case *SearchCondition:
		condition := newSQLCondition(t)
		n.db = s.db.Where(condition.condition, condition.values)
	default:
		n.db = s.db.Where(obj)
	}
	return n
}

func (s *SQLStore) PageSize(i int) *SQLStore {
	n := s.clone()
	n.pageSize = i
	return n
}

func (s *SQLStore) PageNum(i int) *SQLStore {
	n := s.clone()
	if n.pageSize == 0 {
		n.pageSize = defaultPageSize
	}
	if i == 0 {
		i = 1
	}
	offSet := i * n.pageSize
	limit := n.pageSize
	n.db = s.db.Limit(limit).Offset(offSet)
	return n
}

func (s *SQLStore) PageCount(out int) *SQLStore {
	if s.Err != nil {
		return s
	}
	n := s.clone()
	n.Err = s.db.Count(out).Error
	return n
}

func (s *SQLStore) Find(out interface{}) *SQLStore {
	if s.Err != nil {
		return s
	}
	n := s.clone()
	e := s.db.Find(out).Error
	if e == gorm.ErrRecordNotFound {
		n.Err = errors.ErrNotExist
	} else {
		n.Err = e
	}
	return n
}

func (s *SQLStore) Create(obj interface{}) *SQLStore {
	n := s.clone()
	old := reflect.New(util.IndirectType(reflect.TypeOf(obj))).Interface()
	// should use primary key build where condition but just Job can be modify, so just copy Name and Region
	util.CopyField(old, obj, "Name", "Region")
	err := s.db.Where(old).First(old).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		n.Err = err
		return n
	}
	if err == gorm.ErrRecordNotFound {
		n.Err = s.db.Create(obj).Error
		return n
	}
	n.Err = errors.ErrRepetition
	return n
}

func (s *SQLStore) Modify(obj interface{}) *SQLStore {
	n := s.clone()
	old := reflect.New(util.IndirectType(reflect.TypeOf(obj))).Interface()
	//oldslice := reflect.SliceOf(util.IndirectType(reflect.TypeOf(obj)))
	// should use primary key build where condition but just Job can be modify, so just copy Name and Region
	util.CopyField(old, obj, "Name", "Region")
	err := s.db.Where(old).Find(old).Error
	if err == gorm.ErrRecordNotFound {
		n.Err = errors.ErrNotExist
		return n
	}
	if err != nil {
		n.Err = err
		return n
	}

	n.Err = s.db.Model(obj).Updates(obj).Error
	return n
}

func (s *SQLStore) Delete(obj interface{}) *SQLStore {
	n := s.clone()
	out := reflect.New(util.IndirectType(reflect.TypeOf(obj))).Interface()
	err := s.db.Where(obj).First(out).Error
	if err != nil {
		n.Err = err
		return n
	}
	n.Err = s.db.Delete(out).Error
	if n.Err == nil {
		obj = out
	}
	return n
}

func (s *SQLStore) clone() *SQLStore {
	store := SQLStore{db: s.db, sqlCondition: s.sqlCondition, pageSize: s.pageSize, pageNum: s.pageNum, Err: s.Err}
	return &store
}

type SearchCondition struct {
	conditions []map[string]string
	linkSymbol []LogicSymbol
}

func NewSearchCondition(conditions, links []string) (*SearchCondition, error) {

	if len(conditions) != len(links)+1 {
		return nil, errors.ErrLinkNum
	}

	regex := `$(?P<column>\w+)\s*(?P<symbol>=|>=|>|<=|<|<>)\s*(?P<value>\w+)^`
	s := SearchCondition{
		conditions: make([]map[string]string, len(conditions)),
		linkSymbol: make([]LogicSymbol, len(links)),
	}

	for _, condition := range conditions {
		if !util.RegexpMatch(regex, condition) {
			return nil, errors.ErrConditionFormat
		}
		s.conditions = append(s.conditions, util.RegexpGetParams(regex, condition))
	}

	for _, link := range links {
		if symbol, ok := StringToLogicSymbol(link); ok {
			s.linkSymbol = append(s.linkSymbol, symbol)
		} else {
			return nil, errors.ErrNotSupportSymbol
		}
	}
	return &s, nil
}

type LogicSymbol int

var LogicSymbolName = map[int]string{
	1: "OR",
	2: "AND",
	3: "=",
	4: ">=",
	5: ">",
	6: "<=",
	7: "<",
	8: "<>",
}

var LogicSymbolValue = map[string]int{
	"OR":  1,
	"AND": 2,
	"=":   3,
	">=":  4,
	">":   5,
	"<=":  6,
	"<":   7,
	"<>":  8,
}

func (l LogicSymbol) String() string {
	return LogicSymbolName[int(l)]
}

func StringToLogicSymbol(s string) (LogicSymbol, bool) {
	if v, ok := LogicSymbolValue[s]; ok {
		return LogicSymbol(v), true
	}
	return 0, false
}

type sqlCondition struct {
	condition string
	values    []string
}

func newSQLCondition(search *SearchCondition) *sqlCondition {
	s := sqlCondition{
		condition: "",
		values:    make([]string, len(search.linkSymbol)),
	}
	for n, condition := range search.conditions {
		if n == len(search.conditions)-1 {
			s.condition += fmt.Sprintf("%s %s ?", condition["column"], condition["symbol"])
			s.values = append(s.values, condition["value"])
			break
		}
		s.condition += fmt.Sprintf("%s %s ? %s ", condition["column"], condition["symbol"], search.linkSymbol[n])
		s.values = append(s.values, condition["value"])
	}
	return &s
}
