package djob

import (
	pb "local/djob/message"

	"github.com/Sirupsen/logrus"
	"github.com/docker/libkv"
	"github.com/docker/libkv/store"
	"github.com/docker/libkv/store/consul"
	"github.com/docker/libkv/store/etcd"
	"github.com/docker/libkv/store/zookeeper"
	"github.com/gogo/protobuf/proto"
)

type Store struct {
	Client   store.Store
	keyspace string
	backend  string
}

func init() {
	etcd.Register()
	consul.Register()
	zookeeper.Register()
}

func NewStore(backend string, servers []string, keyspace string) (*Store, error) {
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
	return &Store{
		Client:   s,
		backend:  backend,
		keyspace: keyspace,
	}, nil
}

func (s *Store) GetJob(jobName string) (*pb.Job, error) {
	res, err := s.Client.Get(s.keyspace + "/jobs/" + jobName)
	if err != nil {
		return nil, err
	}
	job := &pb.Job{}
	if err = proto.Unmarshal([]byte(res), job); err != nil {
		return nil, err
	}
	return job, nil
}

func (s *Store) SetJob(job *Job) error {
	return nil
}

