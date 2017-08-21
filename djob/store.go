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

func (k *KVStore) GetJobList(region string) ([]*pb.Job, error) {
	res, err := k.Client.List(fmt.Sprintf("%s/%s/jobs/", k.keyspace, generateSlug(region)))
	if err != nil && err != store.ErrKeyNotFound {
		return nil, err
	}
	if res == nil {
		return []*pb.Job{}, nil
	}
	jobs := make([]*pb.Job, 0)
	for _, entry := range res {
		var job pb.Job
		if err := proto.Unmarshal(entry.Value, &job); err != nil {
			return nil, err
		}
		jobs = append(jobs, &job)
	}
	return jobs, nil
}

func (k *KVStore) GetJob(name, region string) (*pb.Job, error) {

	res, err := k.Client.Get(fmt.Sprintf("%s/%s/jobs/%s", k.keyspace, generateSlug(region), generateSlug(name)))
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
	jobkey := fmt.Sprintf("%s/%s/jobs/%s", k.keyspace, regionKey, jobNameKey)

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
	}).Debug("Store: save job to etcd")
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

	if err := k.Client.Delete(fmt.Sprintf("%s/%s/jobs/%s", k.keyspace, generateSlug(region), generateSlug(name))); err != nil {
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
