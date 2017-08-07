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

func NewStore(backend string, servers []string, keyspace string) (*KVStore, error) {
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

func (s *KVStore) GetJob(jobName string) (*pb.Job, error) {
	jobNamekey := generateSlug(jobName)
	res, err := s.Client.Get(s.keyspace + "/jobs/" + jobNamekey)
	if err != nil {
		return nil, err
	}
	job := &pb.Job{}
	if err = proto.Unmarshal([]byte(res.Value), job); err != nil {
		return nil, err
	}
	return job, nil
}

func (s *KVStore) SetJob(job *pb.Job) error {
	if err := verifyJob(job); err != nil {
		return err
	}
	jobName := generateSlug(job.Name)
	jobkey := fmt.Sprintf("%s/jobs/%s", s.keyspace, jobName)

	v, err := proto.Marshal(job)
	if err != nil {
		return err
	}
	Log.WithFields(logrus.Fields{})
	return nil
}
