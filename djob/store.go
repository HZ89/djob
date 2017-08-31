package djob

import (
	pb "version.uuzu.com/zhuhuipeng/djob/message"

	"fmt"
	"github.com/Sirupsen/logrus"
	"github.com/docker/libkv"
	"github.com/docker/libkv/store"
	"github.com/docker/libkv/store/consul"
	"github.com/docker/libkv/store/etcd"
	"github.com/docker/libkv/store/zookeeper"
	"github.com/gogo/protobuf/proto"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"time"
)

type KVStore struct {
	Client   store.Store
	keyspace string
	backend  string
}

func init() {
	etcd.Register()
	consul.Register()
	zookeeper.Register()
}

func NewKVStore(backend string, servers []string, keyspace string) (*KVStore, error) {
	s, err := libkv.NewStore(store.Backend(backend), servers, nil)
	if err != nil {
		return nil, err
	}
	Log.WithFields(logrus.Fields{
		"backend":  backend,
		"servers":  servers,
		"keyspace": keyspace,
	}).Debug("store: Backend Connected")

	_, err = s.List(keyspace)
	if err != store.ErrKeyNotFound && err != nil {
		return nil, err
	}
	return &KVStore{
		Client:   s,
		backend:  backend,
		keyspace: keyspace,
	}, nil
}

func (k *KVStore) GetJobStatus(name, region string) (status *pb.JobStatus, err error) {
	res, err := k.Client.Get(fmt.Sprintf("%s/%s/Status/%s", k.keyspace, generateSlug(region), generateSlug(name)))
	if err != nil && err != store.ErrKeyNotFound {
		return nil, err
	}
	if res != nil {
		var es pb.JobStatus
		if err = proto.Unmarshal([]byte(res.Value), &es); err != nil {
			return nil, err
		}
		return &es, nil
	}
	return &pb.JobStatus{
		JobName: name,
		Region:  region,
	}, nil
}

func (k *KVStore) SetJobStatus(ex *pb.Execution) error {
	key := fmt.Sprintf("%s/%s/Status/%s", k.keyspace, generateSlug(ex.Region), generateSlug(ex.JobName))
	es, err := k.GetJobStatus(ex.JobName, ex.Region)
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

	Log.WithFields(logrus.Fields{
		"Name":         es.JobName,
		"Region":       es.Region,
		"SuccessCount": es.SuccessCount,
		"ErrorCount":   es.ErrorCount,
		"LastSuccess":  es.LastSuccess,
		"LastError":    es.LastError,
	}).Debug("Store: save jobstatus to kv store")

	entry, err := proto.Marshal(es)
	if err != nil {
		return err
	}
	if err := k.Client.Put(key, entry, nil); err != nil {
		return err
	}
	return nil
}

func (k *KVStore) GetJobList(region string) (jobs []*pb.Job, err error) {
	res, err := k.Client.List(fmt.Sprintf("%s/%s/Jobs/", k.keyspace, generateSlug(region)))
	if err != nil && err != store.ErrKeyNotFound {
		return nil, err
	}
	if res == nil {
		return
	}
	for _, entry := range res {
		var job pb.Job
		if err = proto.Unmarshal([]byte(entry.Value), &job); err != nil {
			return
		}
		jobs = append(jobs, &job)
	}
	return
}

func (k *KVStore) GetJob(name, region string) (*pb.Job, error) {

	res, err := k.Client.Get(fmt.Sprintf("%s/%s/Jobs/%s", k.keyspace, generateSlug(region), generateSlug(name)))
	if err != nil {
		return nil, err
	}
	var job pb.Job
	if err = proto.Unmarshal([]byte(res.Value), &job); err != nil {
		return nil, err
	}
	return &job, nil
}

func (k *KVStore) SetJob(job *pb.Job) error {
	if err := verifyJob(job); err != nil {
		return err
	}
	jobNameKey := generateSlug(job.Name)
	regionKey := generateSlug(job.Region)
	jobkey := fmt.Sprintf("%s/%s/Jobs/%s", k.keyspace, regionKey, jobNameKey)

	_, err := k.GetJob(job.Name, job.Region)
	if err != nil && err != store.ErrKeyNotFound {
		return err
	}

	jobpb, err := proto.Marshal(job)
	if err != nil {
		return err
	}
	Log.WithFields(logrus.Fields{
		"name":       jobNameKey,
		"region":     regionKey,
		"expression": job.Expression,
		"scheduler":  job.Schedule,
		"marshal":    string(jobpb),
	}).Debug("Store: save job to kv store")
	if err := k.Client.Put(jobkey, jobpb, nil); err != nil {
		return err
	}
	return nil
}

func (k *KVStore) DeleteJob(name, region string) (*pb.Job, error) {
	job, err := k.GetJob(name, region)
	if err != nil {
		return nil, err
	}

	if err := k.Client.Delete(fmt.Sprintf("%s/%s/Jobs/%s", k.keyspace, generateSlug(region), generateSlug(name))); err != nil {
		return nil, err
	}

	return job, nil
}

// used to cache job add from web api
type MemStore struct {
	memBuf *memDbBuffer
}

func (m *MemStore) GetJob(name, region string) (*pb.Job, error) {
	jobKey := Key(fmt.Sprintf("%s-%s", generateSlug(region), generateSlug(name)))
	res, err := m.memBuf.Get(jobKey)
	if err != nil {
		return nil, err
	}
	var job pb.Job
	if err := proto.Unmarshal(res, &job); err != nil {
		return nil, err
	}
	return &job, nil
}

func (m *MemStore) SetJob(job *pb.Job) error {
	if err := verifyJob(job); err != nil {
		return err
	}

	jobpb, err := proto.Marshal(job)
	if err != nil {
		return err
	}

	jobKey := Key(fmt.Sprintf("%s-%s", generateSlug(job.Region), generateSlug(job.Name)))

	if err := m.memBuf.Set(jobKey, jobpb); err != nil {
		return err
	}

	return nil
}

func (m *MemStore) DeleteJob(name, region string) error {
	jobKey := Key(fmt.Sprintf("%s-%s", generateSlug(region), generateSlug(name)))

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
	ej, err := s.GetExec(ex.JobName, ex.Region, ex.RunNodeName, ex.Group)
	if err != nil {
		return err
	}
	if ej == nil {
		if err := s.db.Create(ex).Error; err != nil {
			return err
		}
	} else {
		if err := s.db.Save(ex).Error; err != nil {
			return err
		}
	}
	return nil
}
