package store

import (
	"fmt"
	"sync"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/docker/libkv"
	libstore "github.com/docker/libkv/store"
	"github.com/docker/libkv/store/consul"
	"github.com/docker/libkv/store/etcd"
	"github.com/docker/libkv/store/zookeeper"
	"github.com/gogo/protobuf/proto"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"

	"version.uuzu.com/zhuhuipeng/djob/errors"
	"version.uuzu.com/zhuhuipeng/djob/log"
	pb "version.uuzu.com/zhuhuipeng/djob/message"
	"version.uuzu.com/zhuhuipeng/djob/until"
)

func init() {
	etcd.Register()
	consul.Register()
	zookeeper.Register()
}

const (
	sqlMaxOpenConnect = 100
	sqlMaxIdleConnect = 20
	LockTimeOut       = 2 * time.Second
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
	chain map[string]libstore.Locker
	mutex *sync.Mutex
}

func NewLockerChain() *LockerChain {
	return &LockerChain{
		chain: make(map[string]libstore.Locker),
		mutex: &sync.Mutex{},
	}
}

func (l *LockerChain) HaveIt(name string) bool {
	_, exist := l.chain[name]
	return exist
}

func (l *LockerChain) AddLocker(name string, locker libstore.Locker) error {
	if l.HaveIt(name) {
		return errors.ErrAlreadyHaveIt
	}
	l.mutex.Lock()
	defer l.mutex.Unlock()
	l.chain[name] = locker
	return nil
}

func (l *LockerChain) ReleaseLocker(name string) error {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	locker, exist := l.chain[name]
	if exist {
		if err := locker.Unlock(); err != nil {
			return err
		}
		delete(l.chain, name)
		return nil
	}
	return errors.ErrNotExist
}

func (l *LockerChain) ReleaseAll() {
	for name := range l.chain {
		err := l.ReleaseLocker(name)
		if err != nil {
			log.Loger.Warn(err)
		}
	}
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
	log.Loger.WithFields(logrus.Fields{
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
	name := until.GetFieldValue(obj, "Name")
	region := until.GetFieldValue(obj, "Region")
	if name == nil || region == nil {
		log.Loger.WithFields(logrus.Fields{
			"ObjType":         until.GetType(obj),
			"NameFieldType":   until.GetFieldType(obj, "Name"),
			"RegionFieldType": until.GetFieldType(obj, "Region"),
		}).Debug("Store: build lock key failed")
		return "", errors.ErrType
	}
	key = fmt.Sprintf("%s/%s/%s_locks/%s/%s",
		k.keyspace,
		until.GenerateSlug(region.(string)),
		until.GetType(obj),
		until.GenerateSlug(name.(string)),
		lockType)
	return
}

func (k *KVStore) IsLocked(obj interface{}, lockType LockType) (r bool) {
	v := k.WhoLocked(obj, lockType)
	if v != "" {
		r = true
	}
	return
}

func (k *KVStore) WhoLocked(obj interface{}, lockType LockType) (v string) {
	key, _ := k.buildLockKey(obj, lockType)
	locker, err := k.client.Get(key)
	if err != nil && err != libstore.ErrKeyNotFound {
		log.Loger.WithField("key", key).WithError(err).Fatal("Store: get key failed")
	}
	if locker != nil {
		v = string(locker.Value)
	}
	return
}

func (k *KVStore) Lock(obj interface{}, lockType LockType, v string) (locker libstore.Locker, err error) {
	var key string
	var lockOpt = &libstore.LockOptions{}
	if v != "" {
		lockOpt.Value = []byte(v)
	}
	key, err = k.buildLockKey(obj, lockType)
	if err != nil {
		return
	}
	locker, err = k.client.NewLock(key, lockOpt)
	if err != nil {
		log.Loger.WithField("key", key).WithError(err).Fatal("Store: New lock failed")
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
	switch obj.(type) {
	case pb.Job:
		if obj.(pb.Job).Name == "" {
			return fmt.Sprintf("%s/%s/Jobs/",
				k.keyspace, until.GenerateSlug(obj.(pb.Job).Region))
		}
		return fmt.Sprintf("%s/%s/Jobs/%s",
			k.keyspace, until.GenerateSlug(obj.(pb.Job).Region),
			until.GenerateSlug(obj.(pb.Job).Name))
	case pb.JobStatus:
		return fmt.Sprintf("%s/%s/Status/%s",
			k.keyspace, until.GenerateSlug(obj.(pb.JobStatus).Region),
			until.GenerateSlug(obj.(pb.JobStatus).Name))
	default:
		return ""
	}
}

func (k *KVStore) Watch(obj interface{}, resCh chan interface{}) (stopCh chan struct{}, err error) {
	v, ok := obj.(proto.Message)
	if !ok {
		return nil, errors.ErrArgs
	}
	stopCh = make(chan struct{})
	stopWatchCh := make(chan struct{})
	kpCh, err := k.client.Watch(k.buildKey(v), stopWatchCh)
	if err != nil {
		return nil, err
	}
	go func() {
		select {
		case kp := <-kpCh:
			err = until.JsonToPb(kp.Value, v)
			if err != nil {
				return
			}
			resCh <- v
		case <-stopCh:
			stopWatchCh <- struct{}{}
			return
		}
	}()
	return
}

// Get JobStatus from KV store
// if the job do not exists, return nil
// if job exist, return *pb.JobStatus
func (k *KVStore) GetJobStatus(name, region string) (status *pb.JobStatus, err error) {
	// if job do not exist just return status not exist
	_, err = k.GetJob(name, region)
	if err != nil {
		return
	}

	status = &pb.JobStatus{
		Name:   name,
		Region: region,
	}

	res, err := k.client.Get(k.buildKey(status))
	if err != nil && err != libstore.ErrKeyNotFound {
		return nil, err
	}

	// fill data to status
	if res != nil {
		if err = until.JsonToPb(res.Value, status); err != nil {
			return nil, err
		}
		log.Loger.WithFields(logrus.Fields{
			"Name":         status.Name,
			"Region":       status.Region,
			"SuccessCount": status.SuccessCount,
			"ErrorCount":   status.ErrorCount,
			"LastError":    status.LastError,
			"OrginData":    string(res.Value),
		}).Debug("Store: get jobstatus from kv store")
		return
	}

	return
}

// Calculate JobStatus and save to KV store
// For safely edit data, need to lock JobStatus first
func (k *KVStore) SetJobStatus(ex *pb.Execution) error {
	es, err := k.GetJobStatus(ex.Name, ex.Region)
	if err != nil {
		return err
	}
	if ex.Succeed {
		es.SuccessCount += 1
		es.LastSuccess = time.Unix(0, ex.FinishTime).String()
	} else {
		es.ErrorCount += 1
		es.LastError = time.Unix(0, ex.FinishTime).String()
	}
	es.LastHandleAgent = ex.RunNodeName

	entry, err := until.PbToJSON(es)
	if err != nil {
		return err
	}
	log.Loger.WithFields(logrus.Fields{
		"Name":         es.Name,
		"Region":       es.Region,
		"SuccessCount": es.SuccessCount,
		"ErrorCount":   es.ErrorCount,
		"LastSuccess":  es.LastSuccess,
		"LastError":    es.LastError,
		"OrginData":    string(entry),
	}).Debug("Store: save jobstatus to kv store")

	if err = k.client.Put(k.buildKey(es), entry, nil); err != nil {
		return err
	}
	return nil
}

func (k *KVStore) GetRegionList() (regions []string, err error) {
	res, err := k.client.List(k.keyspace)
	if err != nil {
		return nil, err
	}
	for _, re := range res {
		regions = append(regions, string(re.Value))
	}
	return
}

func (k *KVStore) jobs(name, region string) (jobs []*pb.Job, err error) {
	if region != "" {
		var res []*libstore.KVPair
		if name != "" {
			var r *libstore.KVPair
			key := k.buildKey(&pb.Job{Name: name, Region: region})
			r, err = k.client.Get(key)
			if err != nil && err != libstore.ErrKeyNotFound {
				log.Loger.WithField("key", key).WithError(err).Fatal("Store: get key failed")
			}
			if res == nil {
				err = libstore.ErrKeyNotFound
				return
			}
			res = append(res, r)
		} else {
			key := fmt.Sprintf("%s/%s/Jobs/", k.keyspace, until.GenerateSlug(region))
			res, err = k.client.List(key)
			if err != nil && err != libstore.ErrKeyNotFound {
				log.Loger.WithField("key", key).WithError(err).Fatal("Store: get key failed")
			}
			if len(res) == 0 {
				err = libstore.ErrKeyNotFound
				return
			}
		}
		for _, entry := range res {
			var job pb.Job
			if err = until.JsonToPb(entry.Value, job); err != nil {
				return nil, err
			}
			jobs = append(jobs, &job)
		}
		return
	}
	var regions []string
	regions, err = k.GetRegionList()
	if err != nil {
		return
	}
	for _, r := range regions {
		var rjobs []*pb.Job
		rjobs, err = k.jobs("", r)
		if err != nil {
			return nil, err
		}
		jobs = append(jobs, rjobs...)
	}
	return
}

func (k *KVStore) GetRegionJobs(region string) ([]*pb.Job, error) {
	return k.jobs("", region)
}

func (k *KVStore) GetJob(name, region string) (*pb.Job, error) {
	jobs, err := k.jobs(name, region)
	if err != nil {
		return nil, err
	}
	return jobs[0], nil
}

func (k *KVStore) GetJobs() ([]*pb.Job, error) {
	return k.jobs("", "")
}

func (k *KVStore) SetJob(job *pb.Job) error {
	if err := until.VerifyJob(job); err != nil {
		return err
	}

	_, err := k.GetJob(job.Name, job.Region)

	if err != nil && err != libstore.ErrKeyNotFound {
		return err
	}

	json, err := until.PbToJSON(job)
	if err != nil {
		return err
	}
	log.Loger.WithFields(logrus.Fields{
		"name":       job.Name,
		"region":     job.Region,
		"expression": job.Expression,
		"scheduler":  job.Schedule,
		"marshal":    string(json),
	}).Debug("Store: save job to kv store")
	if err = k.client.Put(k.buildKey(job), json, nil); err != nil {
		return err
	}
	return nil
}

func (k *KVStore) DeleteJob(name, region string) (*pb.Job, error) {
	job, err := k.GetJob(name, region)
	if err != nil {
		return nil, err
	}

	if err = k.client.Delete(k.buildKey(job)); err != nil {
		return nil, err
	}

	return job, nil
}

// used to cache job in local memory
type MemStore struct {
	memBuf *memDbBuffer
}

func (m *MemStore) GetJob(name, region string) (*pb.Job, error) {
	jobKey := Key(fmt.Sprintf("%s-%s", until.GenerateSlug(region), until.GenerateSlug(name)))
	res, err := m.memBuf.Get(jobKey)
	if err != nil {
		return nil, err
	}
	var job pb.Job
	if err = proto.Unmarshal(res, &job); err != nil {
		return nil, err
	}
	return &job, nil
}

func (m *MemStore) SetJob(job *pb.Job) error {
	if err := until.VerifyJob(job); err != nil {
		return err
	}

	jobpb, err := proto.Marshal(job)
	if err != nil {
		return err
	}

	jobKey := Key(fmt.Sprintf("%s-%s", until.GenerateSlug(job.Region), until.GenerateSlug(job.Name)))

	if err = m.memBuf.Set(jobKey, jobpb); err != nil {
		return err
	}

	return nil
}

func (m *MemStore) DeleteJob(name, region string) error {
	jobKey := Key(fmt.Sprintf("%s-%s", until.GenerateSlug(region), until.GenerateSlug(name)))

	if err := m.memBuf.Delete(jobKey); err != nil {
		return err
	}

	return nil
}

func NewMemStore() *MemStore {
	return &MemStore{
		memBuf: NewMemDbBuffer(),
	}
}

type SQLStore struct {
	db *gorm.DB
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

func (s *SQLStore) GetExecs(jobName, region, node string, group int64) (exs []*pb.Execution, err error) {
	q := map[string]interface{}{
		"job_name": jobName,
		"region":   region,
	}
	if group > 0 {
		q["group"] = group
	}
	if node != "" {
		q["run_node_name"] = node
	}
	if err = s.db.Where(q).Find(&exs).Error; err != nil {
		return nil, err
	}
	return
}

func (s *SQLStore) GetExec(jobName, region, node string, group int64) (ex *pb.Execution, err error) {
	var exs []*pb.Execution
	exs, err = s.GetExecs(jobName, region, node, group)
	if err != nil {
		return
	}
	if len(exs) > 0 {
		return exs[0], nil
	}
	return
}

func (s *SQLStore) SetExec(ex *pb.Execution) error {
	ej, err := s.GetExec(ex.Name, ex.Region, ex.RunNodeName, ex.Group)
	if err != nil {
		return err
	}
	if ej == nil {
		if err = s.db.Create(ex).Error; err != nil {
			return err
		}
	} else {
		if err = s.db.Save(ex).Error; err != nil {
			return err
		}
	}
	return nil
}
