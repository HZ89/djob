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
	"math"
	"path"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/HZ89/libkv"
	libstore "github.com/HZ89/libkv/store"
	"github.com/HZ89/libkv/store/consul"
	"github.com/HZ89/libkv/store/etcd"
	"github.com/HZ89/libkv/store/zookeeper"
	"github.com/Sirupsen/logrus"
	"github.com/fatih/structs"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"

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
	sqlMaxOpenConnect = 100
	sqlMaxIdleConnect = 20
	//	LockTimeOut       = 2 * time.Second
	defaultPageSize = 10
	separator       = "###"
	lockTTL         = 60
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

// KVStore struct
type KVStore struct {
	client   libstore.Store // client of store backend
	keyspace string         // perfix of key
	backend  string         // backend, support etcd, zk etc...
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
	return &KVStore{
		client:   s,
		backend:  backend,
		keyspace: keyspace,
	}, nil
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

// MemStore used to cache job in local memory
type MemStore struct {
	memBuf *memDbBuffer // leveldb memory buffer
	stopCh chan struct{}
}

// base struct save in MemStore
type entry struct {
	DeadTime time.Time // the dead time of this entry
	Data     []byte    // user save data
}

func (m *MemStore) Delete(key string) error {
	return m.memBuf.Delete(Key(key))
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

// timeing seek and clear expired key
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

// seek and clear
func (m *MemStore) releaseOverdueKey() {
	now := time.Now()
	iter, err := m.memBuf.Seek(nil)
	if err != nil {
		log.FmdLoger.Fatalf("Store: memBuf Seek failed: %v", err)
	}
	defer iter.Close()
	// seek
	for ; iter.Valid(); iter.Next() {
		key := iter.Key()
		var e *entry
		e, err = m.getEntry(key)
		if err != nil && err != errors.ErrNotExist {
			log.FmdLoger.WithError(err).Fatalf("Store: memBuf Seek get key filed: key: %s, value: %v", string(key), e)
		}
		// clear
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

// Release used to stop MemStore
func (m *MemStore) Release() {
	m.stopCh <- struct{}{}
}

// NewMemStore init a MemStore and start a goroutine handle expired key
func NewMemStore() *MemStore {
	ms := &MemStore{
		memBuf: NewMemDbBuffer(),
		stopCh: make(chan struct{}),
	}
	ms.handleOverdueKey()
	return ms
}

// SQLStore database obj
type SQLStore struct {
	db           *gorm.DB      // gorm db obj
	sqlCondition *sqlCondition // sql where contition
	pageSize     int           // page size
	pageNum      int           // page num
	count        int64         // max page num
	Err          error         // error
}

// NewSQLStore init a SQLStore struct
func NewSQLStore(backend, dsn string) (*SQLStore, error) {

	db, err := gorm.Open(backend, dsn)
	if err != nil {
		return nil, err
	}

	db.DB().SetMaxOpenConns(sqlMaxOpenConnect)
	db.DB().SetMaxIdleConns(sqlMaxIdleConnect)

	if log.FmdLoger.Logger.Level.String() == "debug" {
		db.LogMode(true)
	}

	return &SQLStore{
		db: db,
	}, nil
}

func (s *SQLStore) clone() *SQLStore {
	store := SQLStore{db: s.db, sqlCondition: s.sqlCondition, pageSize: s.pageSize, pageNum: s.pageNum, Err: s.Err, count: s.count}
	return &store
}

// Close database connect
func (s *SQLStore) Close() error {
	return s.db.Close()
}

// Migrate auto migrate table struct
func (s *SQLStore) Migrate(drop bool, valus ...interface{}) error {
	if !drop {
		if err := s.db.AutoMigrate(valus...).Error; err != nil {
			return err
		}
		return nil
	}
	return nil
}

// Model set the model to be manipulated
func (s *SQLStore) Model(obj interface{}) *SQLStore {
	s.db = s.db.Model(obj)
	return s
}

// Where build where condition, can be a obj or SearchContion
func (s *SQLStore) Where(obj interface{}) *SQLStore {
	n := s.clone()
	switch t := obj.(type) {
	case *SearchCondition:
		condition, err := newSQLCondition(t, n.db.Value)
		if err != nil {
			n.Err = err
			return n
		}
		n.db = n.db.Where(condition.condition, condition.values...)
	default:
		n.db = n.db.Where(obj)
	}
	return n
}

// PageSize set page size
func (s *SQLStore) PageSize(i int) *SQLStore {
	n := s.clone()
	n.pageSize = i
	return n
}

// PageNum which page need scan into result set
func (s *SQLStore) PageNum(i int) *SQLStore {
	n := s.clone()
	if n.pageSize == 0 {
		n.pageSize = defaultPageSize
	}
	if i == 0 {
		i = 1
	}
	offSet := (i - 1) * n.pageSize
	limit := n.pageSize
	n.db = n.db.Limit(limit).Offset(offSet)
	return n
}

// PageCount func cal max page num
func (s *SQLStore) PageCount(out interface{}) *SQLStore {
	n := s.clone()
	n.Err = n.db.Offset(0).Count(&s.count).Error

	if n.Err != nil {
		return n
	}

	rf := math.Ceil(float64(s.count) / float64(s.pageSize))
	typ := util.IndirectType(reflect.TypeOf(out))
	val := util.Indirect(reflect.ValueOf(out))
	kind := typ.Kind()
	if int(kind) > 1 && int(kind) < 6 {
		val.SetInt(int64(rf))
		return n
	}
	n.Err = errors.ErrType
	return n
}

// Find func perform find in database
func (s *SQLStore) Find(out interface{}) *SQLStore {
	n := s.clone()
	if n.Err != nil {
		return n
	}
	e := n.db.Find(out).Error
	if e == gorm.ErrRecordNotFound {
		n.Err = errors.ErrNotExist
	} else {
		n.Err = e
	}
	return n
}

// Create func perform create in database
func (s *SQLStore) Create(obj interface{}) *SQLStore {
	n := s.clone()
	old := reflect.New(util.IndirectType(reflect.TypeOf(obj))).Interface()
	// should use primary key build where condition but just Job can be modify, so just copy Name and Region
	util.CopyField(old, obj, "Name", "Region")
	err := n.db.Where(old).First(old).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		n.Err = err
		return n
	}
	if err == gorm.ErrRecordNotFound {
		n.Err = n.db.Create(obj).Error
		return n
	}
	n.Err = errors.ErrRepetition
	return n
}

// Modify func perform modify
func (s *SQLStore) Modify(obj interface{}) *SQLStore {
	n := s.clone()
	old := reflect.New(util.IndirectType(reflect.TypeOf(obj))).Interface()
	//oldslice := reflect.SliceOf(util.IndirectType(reflect.TypeOf(obj)))
	// should use primary key build where condition but just Job can be modify, so just copy Name and Region
	util.CopyField(old, obj, "Name", "Region")
	err := n.db.Where(old).Find(old).Error
	if err == gorm.ErrRecordNotFound {
		n.Err = errors.ErrNotExist
		return n
	}
	if err != nil {
		n.Err = err
		return n
	}
	// gorm only update non blank value in struct fields, so use a map to update.
	m := structs.Map(obj)
	n.Err = n.db.Model(obj).Updates(m).Error
	return n
}

// Delete func perform delete
func (s *SQLStore) Delete(obj interface{}) *SQLStore {
	n := s.clone()
	out := reflect.New(util.IndirectType(reflect.TypeOf(obj))).Interface()
	err := n.db.Where(obj).First(out).Error
	if err != nil {
		n.Err = err
		return n
	}
	n.Err = n.db.Delete(out).Error
	if n.Err == nil {
		obj = out
	}
	return n
}

// SearchCondition save search condition
type SearchCondition struct {
	conditions []map[string]string // user input search condition,eg. a = 1 , b <> abc
	linkSymbol []logicSymbol       // user input link symbol, eg. AND
}

// NewSearchCondition serialized user input generated SearchCondition
func NewSearchCondition(conditions, links []string) (*SearchCondition, error) {

	if len(conditions) != len(links)+1 {
		return nil, errors.ErrLinkNum
	}

	regex := `^(?P<column>\w+)\s*(?P<symbol>=|>=|>|<=|<|<>|=)\s*(?P<value>\w+)$`
	s := new(SearchCondition)

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
	return s, nil
}

type logicSymbol int

var logicSymbolNames = map[int]string{
	1: "OR",
	2: "AND",
	3: "=",
	4: ">=",
	5: ">",
	6: "<=",
	7: "<",
	8: "<>",
}

var logicSymbolValues = map[string]int{
	"OR":  1,
	"AND": 2,
	"=":   3,
	">=":  4,
	">":   5,
	"<=":  6,
	"<":   7,
	"<>":  8,
}

func (l logicSymbol) String() string {
	return logicSymbolNames[int(l)]
}

// StringToLogicSymbol used to conversion string to logicSymbol
func StringToLogicSymbol(s string) (logicSymbol, bool) {
	if v, ok := logicSymbolValues[s]; ok {
		return logicSymbol(v), true
	}
	return 0, false
}

type sqlCondition struct {
	condition string        // string of where condition eg. a = ? and b <> ?
	values    []interface{} // value of where condition eg. 1,"avc"
}

// use SearchCondition generated sql where condition
func newSQLCondition(search *SearchCondition, obj interface{}) (*sqlCondition, error) {
	s := new(sqlCondition)
	for n, condition := range search.conditions {
		s.condition += fmt.Sprintf("%s %s ?", condition["column"], condition["symbol"])
		fieldKind := util.GetFieldKind(obj, condition["column"])
		switch fieldKind {
		case reflect.Int32:
			v, err := strconv.ParseInt(condition["value"], 10, 32)
			if err != nil {
				return nil, errors.New(95236, err.Error())
			}
			s.values = append(s.values, v)
		case reflect.Int64:
			v, err := strconv.ParseInt(condition["value"], 10, 64)
			if err != nil {
				return nil, errors.New(95236, err.Error())
			}
			s.values = append(s.values, v)
		case reflect.String:
			s.values = append(s.values, condition["value"])
		case reflect.Bool:
			v, err := strconv.ParseBool(condition["value"])
			if err != nil {
				return nil, errors.New(95236, err.Error())
			}
			s.values = append(s.values, v)
		default:
			return nil, errors.New(95236, "not support feild type")
		}
		if n < len(search.linkSymbol) {
			s.condition += fmt.Sprintf(" %s ", search.linkSymbol[n])
		}
	}
	return s, nil
}
