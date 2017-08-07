package djob

import (
	"local/djob/rpc"
	"time"

	"errors"
	"fmt"
	"sync"

	"github.com/Sirupsen/logrus"
	"github.com/docker/libkv/store"
	"github.com/hashicorp/memberlist"
	"github.com/hashicorp/serf/serf"
	"local/djob/scheduler"
)

var (
	ErrLockTimeout = errors.New("locking timeout")
	LockTimeout    = 2 * time.Second
)

type Agent struct {
	ShutdownCh  <-chan struct{}
	jobLockers  map[string]store.Locker
	config      *Config
	serf        *serf.Serf
	eventCh     chan serf.Event
	ready       bool
	rpcServer   *rpc.RpcServer
	store       *KVStore
	membercache map[string]map[string]string
	mutex       sync.Mutex
	scheduler   *scheduler.Scheduler
}

func (a *Agent) setupSerf() *serf.Serf {
	encryptKey, err := a.config.EncryptKey()
	if err != nil {
		Log.Fatal(err)
		return nil
	}

	serfConfig := serf.DefaultConfig()

	//noinspection GoBinaryAndUnaryExpressionTypesCompatibility
	serfConfig.MemberlistConfig = memberlist.DefaultWANConfig()

	serfConfig.MemberlistConfig.BindAddr = a.config.SerfBindIP
	serfConfig.MemberlistConfig.BindPort = a.config.SerfBindPort
	serfConfig.MemberlistConfig.AdvertiseAddr = a.config.SerfAdvertiseIP
	serfConfig.MemberlistConfig.AdvertisePort = a.config.SerfAdvertisePort
	serfConfig.MemberlistConfig.SecretKey = encryptKey
	serfConfig.NodeName = a.config.Nodename
	serfConfig.Tags = a.config.Tags
	serfConfig.SnapshotPath = a.config.SerfSnapshotPath
	serfConfig.CoalescePeriod = 3 * time.Second
	serfConfig.QuiescentPeriod = time.Second
	serfConfig.UserCoalescePeriod = 3 * time.Second
	serfConfig.UserQuiescentPeriod = time.Second
	serfConfig.EnableNameConflictResolution = true

	a.eventCh = make(chan serf.Event, 64)
	serfConfig.EventCh = a.eventCh

	Log.Info("agent: Djob agent starting")

	serf, err := serf.Create(serfConfig)
	if err != nil {
		Log.Fatal(err)
		return nil
	}
	//a.membercache = make(map[string]map[string]string)

	return serf
}

// serfJion let serf intence jion a serf clust
func (a *Agent) serfJion(addrs []string, replay bool) (n int, err error) {
	Log.Infof("agent: joining: %v replay: %v", addrs, replay)
	ignoreOld := !replay
	n, err = a.serf.Join(addrs, ignoreOld)
	if n > 0 {
		Log.Infof("agent: joined: %d nodes", n)
	}
	if err != nil {
		Log.Warnf("agent: error joining: %v", err)
	}
	return
}

func (a *Agent) lockJob(jobName string) (store.Locker, error) {

	//reNewCh := make(chan struct{})

	lockkey := fmt.Sprintf("%s/job_locks/%s", a.store.keyspace, jobName)

	//l, err := a.store.Client.NewLock(lockkey, &store.LockOptions{RenewLock:reNewCh})
	l, err := a.store.Client.NewLock(lockkey, &store.LockOptions{})
	if err != nil {
		Log.WithField("jobName", jobName).WithError(err).Fatal("agent: New lock failed")
	}

	errCh := make(chan error)
	freeCh := make(chan struct{})
	timeoutCh := time.After(LockTimeout)
	stoplockingCh := make(chan struct{})

	go func() {
		_, err = l.Lock(stoplockingCh)
		if err != nil {
			errCh <- err
			return
		}
		freeCh <- struct{}{}
	}()

	select {
	case <-freeCh:
		return l, nil
	case err := <-errCh:
		return nil, err
	case <-timeoutCh:
		stoplockingCh <- struct{}{}
		return nil, ErrLockTimeout
	}
}

func (a *Agent) mainLoop() {
	serfShutdownCh := a.serf.ShutdownCh()
	Log.Info("agent: Listen for event")
	for {

		select {
		// handle serf event
		case e := <-a.eventCh:
			Log.WithFields(logrus.Fields{
				"event": e.String(),
			}).Debug("agent: Received event")

			if memberevent, ok := e.(serf.MemberEvent); ok {
				var memberNames []string
				for _, member := range memberevent.Members {
					memberNames = append(memberNames, member.Name)
				}
				Log.WithFields(logrus.Fields{
					"node":    a.config.Nodename,
					"members": memberNames,
					"event":   e.EventType(),
				}).Debug("agent: Member event got")

				/* go a.handleMemberCache(memberevent.Type, memberevent.Members) */
			}

			// handle custom query event
			if e.EventType() == serf.EventQuery {
				query := e.(*serf.Query)

				switch qname := query.Name; qname {
				case QueryNewJob:
					if a.config.Server {
						Log.WithFields(logrus.Fields{
							"query":   query.Name,
							"payload": string(query.Payload),
							"at":      query.LTime,
						}).Debug("agent: Server receive a add new job event")

						go a.receiveNewJobQuery(query)
					}
				case QueryRunJob:
					Log.WithFields(logrus.Fields{
						"query":   query.Name,
						"payload": string(query.Payload),
						"at":      query.LTime,
					}).Debug("agent: Running job")
				case QueryRPCConfig:
					if a.config.Server {
						Log.WithFields(logrus.Fields{
							"query":   query.Name,
							"payload": string(query.Payload),
							"at":      query.LTime,
						}).Debug("agent: Server receive a rpc config query")
					}
				default:
					Log.Warn("agent: get a unknow message")
					Log.WithFields(logrus.Fields{
						"query":   query.Name,
						"payload": string(query.Payload),
						"at":      query.LTime,
					}).Debug("agent: get a unknow message")
				}
			}

		case <-serfShutdownCh:
			Log.Warn("agent: Serf shutdown detected, quitting")
			return
		}
	}
}

// Run func start a Agent process
func (a *Agent) Run(args []string) int {
	var err error
	a.config, err = NewConfig(args)
	if err != nil {
		Log.Fatalln(err)
	}
	if a.serf = a.setupSerf(); a.serf == nil {
		Log.Fatalln("Start serf failed!")
	}
	a.serfJion(a.config.SerfJoin, true)

	// TODO: add prometheus support
	// start prometheus client
	var tls rpc.TlsOpt
	if a.config.RPCTls {
		tls = rpc.TlsOpt{
			CertFile: a.config.RPCCertFile,
			KeyFile:  a.config.RPCKeyFile,
			CaFile:   a.config.RPCCAfile,
		}
	}

	if a.config.Server {
		a.store, err = NewStore(a.config.JobStore, a.config.JobStoreServers, a.config.JobStoreKeyspace)
		if err != nil {
			Log.WithFields(logrus.Fields{
				"backend":  a.config.JobStore,
				"servers":  a.config.JobStoreServers,
				"keyspace": a.config.JobStoreKeyspace,
			}).Debug("Connect Backend Failed")
			Log.Fatalf("Connent Backend Failed: %s", err)
		}
		// run rpc server
		a.rpcServer = rpc.NewRPCserver(a.config.RPCBindIP, a.config.RPCBindPort, a, &tls)
		err = a.rpcServer.Run()
		if err != nil {
			Log.Fatalf("Start rpc server Failed: %s", err)
		}
		a.scheduler = scheduler.New()

	}
}
