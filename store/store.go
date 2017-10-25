package store

import (
	"fmt"
	"reflect"
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
	LockTimeOut       = 2 * time.Second
	waitTimeOut       = 50 * time.Millisecond
	defaultPageSize   = 10
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

// TODO: use sync.Map replace map
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

func (l *LockerChain) GetLocker(name string) (libstore.Locker, error) {
	if l.HaveIt(name) {
		return l.chain[name], nil
	}
	return nil, errors.ErrNotExist
}

func (l *LockerChain) AddLocker(name string, locker libstore.Locker) error {
	if l.HaveIt(name) {
		return errors.ErrRepetition
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
	name := util.GetFieldValue(obj, "Name")
	region := util.GetFieldValue(obj, "Region")
	if name == nil || region == nil {
		log.Loger.WithFields(logrus.Fields{
			"ObjType":         util.GetType(obj),
			"NameFieldType":   util.GetFieldType(obj, "Name"),
			"RegionFieldType": util.GetFieldType(obj, "Region"),
		}).Debug("Store: build lock key failed")
		return "", errors.ErrType
	}
	key = fmt.Sprintf("%s/%s/%s_locks/%s/%s",
		k.keyspace,
		util.GenerateSlug(region.(string)),
		util.GetType(obj),
		util.GenerateSlug(name.(string)),
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
				k.keyspace, util.GenerateSlug(obj.(pb.Job).Region))
		}
		return fmt.Sprintf("%s/%s/Jobs/%s",
			k.keyspace, util.GenerateSlug(obj.(pb.Job).Region),
			util.GenerateSlug(obj.(pb.Job).Name))
	case pb.JobStatus:
		return fmt.Sprintf("%s/%s/Status/%s",
			k.keyspace, util.GenerateSlug(obj.(pb.JobStatus).Region),
			util.GenerateSlug(obj.(pb.JobStatus).Name))
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
			err = util.JsonToPb(kp.Value, v)
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

func (k *KVStore) untilUnlock(obj interface{}, timeout time.Duration) bool {
	var round int

	for k.IsLocked(obj, RW) {
		if round >= 5 {
			return false
		}
		round += 1
		time.Sleep(timeout)
	}
	return true
}

func (k *KVStore) DeleteJobStatus(status *pb.JobStatus) (out *pb.JobStatus, err error) {
	if !k.untilUnlock(status, waitTimeOut) {
		return nil, errors.ErrLockTimeout
	}
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
// wait JobStatus lock 5 times
func (k *KVStore) GetJobStatus(in *pb.JobStatus) (out *pb.JobStatus, err error) {
	// if job do not exist just return status not exist
	_, err = k.GetJob(in.Name, in.Region)
	if err != nil {
		return
	}

	out = &pb.JobStatus{
		Name:   in.Name,
		Region: in.Region,
	}
	if !k.untilUnlock(in, waitTimeOut) {
		return nil, errors.ErrLockTimeout
	}

	res, err := k.client.Get(k.buildKey(out))
	if err != nil && err != libstore.ErrKeyNotFound {
		return nil, err
	}

	// fill data to status
	if res != nil {
		if err = util.JsonToPb(res.Value, out); err != nil {
			return nil, err
		}
		log.Loger.WithFields(logrus.Fields{
			"Name":         out.Name,
			"Region":       out.Region,
			"SuccessCount": out.SuccessCount,
			"ErrorCount":   out.ErrorCount,
			"LastError":    out.LastError,
			"OrginData":    string(res.Value),
		}).Debug("Store: get jobstatus from kv store")
		return
	}

	return
}

// set job status to kv store
// safely set jobstatus need lock it with RW locker first
func (k *KVStore) SetJobStatus(status *pb.JobStatus) (*pb.JobStatus, error) {

	entry, err := util.PbToJSON(status)
	if err != nil {
		return nil, err
	}
	log.Loger.WithFields(logrus.Fields{
		"Name":         status.Name,
		"Region":       status.Region,
		"SuccessCount": status.SuccessCount,
		"ErrorCount":   status.ErrorCount,
		"LastSuccess":  status.LastSuccess,
		"LastError":    status.LastError,
		"OriginData":   string(entry),
	}).Debug("Store: save job status to kv store")

	if err = k.client.Put(k.buildKey(status), entry, nil); err != nil {
		return nil, err
	}
	return status, nil
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
			key := fmt.Sprintf("%s/%s/Jobs/", k.keyspace, util.GenerateSlug(region))
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
			if err = util.JsonToPb(entry.Value, job); err != nil {
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
	if err := util.VerifyJob(job); err != nil {
		return err
	}

	_, err := k.GetJob(job.Name, job.Region)

	if err != nil && err != libstore.ErrKeyNotFound {
		return err
	}

	json, err := util.PbToJSON(job)
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
	jobKey := Key(fmt.Sprintf("%s-%s", util.GenerateSlug(region), util.GenerateSlug(name)))
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
	if err := util.VerifyJob(job); err != nil {
		return err
	}

	jobpb, err := proto.Marshal(job)
	if err != nil {
		return err
	}

	jobKey := Key(fmt.Sprintf("%s-%s", util.GenerateSlug(job.Region), util.GenerateSlug(job.Name)))

	if err = m.memBuf.Set(jobKey, jobpb); err != nil {
		return err
	}

	return nil
}

func (m *MemStore) DeleteJob(name, region string) error {
	jobKey := Key(fmt.Sprintf("%s-%s", util.GenerateSlug(region), util.GenerateSlug(name)))

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
		n.db = s.db.Where(condition.condition, condition.values...)
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
	n.Err = s.db.Find(out).Error
	return n
}

func (s *SQLStore) Create(obj interface{}) *SQLStore {
	n := s.clone()
	out := reflect.New(reflect.TypeOf(obj)).Interface()
	err := s.db.Where(obj).First(out).Error
	if err != nil {
		n.Err = err
		return n
	}
	if out != nil {
		n.Err = errors.ErrRepetition
		return n
	}
	n.Err = s.db.Create(obj).Error
	return n
}

func (s *SQLStore) Modify(obj interface{}) *SQLStore {
	n := s.clone()
	out := reflect.New(reflect.TypeOf(obj)).Interface()
	err := s.db.Where(obj).First(out).Error
	if err != nil {
		n.Err = err
		return n
	}
	if out == nil {
		n.Err = errors.ErrNotExist
		return n
	}
	n.Err = s.db.Model(out).Updates(obj).Error
	return n
}

func (s *SQLStore) Delete(obj interface{}) *SQLStore {
	n := s.clone()
	out := reflect.New(reflect.TypeOf(obj)).Interface()
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

const (
	LogicSymbol_OR  LogicSymbol = 1
	LogicSymbol_AND LogicSymbol = 2
	LogicSymbol_EQ  LogicSymbol = 3
	LogicSymbol_GE  LogicSymbol = 4
	LogicSymbol_GT  LogicSymbol = 5
	LogicSymbol_LE  LogicSymbol = 6
	LogicSymbol_LT  LogicSymbol = 7
	LogicSymbol_NE  LogicSymbol = 8
)

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
