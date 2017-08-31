package djob

import (
	"time"
	pb "version.uuzu.com/zhuhuipeng/djob/message"
	"version.uuzu.com/zhuhuipeng/djob/rpc"

	"errors"
	"fmt"
	"sync"

	"github.com/Sirupsen/logrus"
	"github.com/docker/libkv/store"
	"github.com/hashicorp/memberlist"
	"github.com/hashicorp/serf/serf"
	"io/ioutil"
	"sort"
	"version.uuzu.com/zhuhuipeng/djob/scheduler"
	"version.uuzu.com/zhuhuipeng/djob/web/api"
)

var (
	ErrLockTimeout = errors.New("locking timeout")
	LockTimeOut    = 2 * time.Second
)

const (
	gracefulTime      = 5 * time.Second
	sqlMaxOpenConnect = 100
	sqlMaxIdleConnect = 20
)

type Agent struct {
	jobLockers map[string]store.Locker
	config     *Config
	serf       *serf.Serf
	eventCh    chan serf.Event
	ready      bool
	rpcServer  *rpc.RpcServer
	rpcClient  *rpc.RpcClient
	store      *KVStore
	memStore   *MemStore
	mutex      *sync.Mutex
	scheduler  *scheduler.Scheduler
	apiServer  *api.APIServer
	version    string
	runJobCh   chan *pb.Job
	sqlStore   *SQLStore
}

func (a *Agent) setupSerf() *serf.Serf {
	encryptKey, err := a.config.EncryptKey()
	if err != nil {
		Log.Fatal(err)
		return nil
	}

	serfConfig := serf.DefaultConfig()

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

	serfConfig.EventCh = a.eventCh

	serfConfig.LogOutput = ioutil.Discard
	serfConfig.MemberlistConfig.LogOutput = ioutil.Discard

	Log.Info("Agent: Djob agent starting")

	s, err := serf.Create(serfConfig)
	if err != nil {
		Log.Fatal(err)
		return nil
	}
	//a.memberCache = make(map[string]map[string]string)

	return s
}

// serfJion let serf intence jion a serf clust
func (a *Agent) serfJion(addrs []string, replay bool) (n int, err error) {
	Log.Infof("Agent: joining: %v replay: %v", addrs, replay)
	ignoreOld := !replay
	n, err = a.serf.Join(addrs, ignoreOld)
	if n > 0 {
		Log.Infof("Agent: joined: %d nodes", n)
	}
	if err != nil {
		Log.Warnf("Agent: error joining: %v", err)
	}
	return
}

func (a *Agent) isLocked(name, region string, obj interface{}) bool {
	lockkey := fmt.Sprintf("%s/%s/%s_locks/%s", a.store.keyspace, generateSlug(region), getType(obj), generateSlug(name))
	l, err := a.store.Client.Get(lockkey)
	if err != nil && err != store.ErrKeyNotFound {
		Log.WithField("key", lockkey).WithError(err).Fatal("Agent get key failed")
	}
	if l != nil {
		return true
	}
	return false
}

func (a *Agent) lock(name, region string, obj interface{}) (store.Locker, error) {

	t := getType(obj)

	lockkey := fmt.Sprintf("%s/%s/%s_locks/%s", a.store.keyspace, generateSlug(region), t, generateSlug(name))

	//l, err := a.store.Client.NewLock(lockkey, &store.LockOptions{RenewLock:reNewCh})
	l, err := a.store.Client.NewLock(lockkey, &store.LockOptions{})
	if err != nil {
		Log.WithField("key", lockkey).WithError(err).Fatal("Agent: New lock failed")
	}

	errCh := make(chan error)
	freeCh := make(chan struct{})
	timeoutCh := time.After(LockTimeOut)
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

func (a *Agent) serfEventLoop() {
	serfShutdownCh := a.serf.ShutdownCh()
	Log.Info("Agent: Listen for event")
	for {

		select {
		// handle serf event
		case e := <-a.eventCh:
			Log.WithFields(logrus.Fields{
				"event": e.String(),
			}).Debug("Agent: Received event")

			if memberevent, ok := e.(serf.MemberEvent); ok {
				for _, member := range memberevent.Members {
					Log.WithFields(logrus.Fields{
						"node":    a.config.Nodename,
						"members": member.Name,
						"event":   e.EventType(),
					}).Debug("Agent: Member event got")
				}
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
						}).Debug("Agent: Server receive a add new job event")

						go a.receiveNewJobQuery(query)
					}
				case QueryRunJob:
					Log.WithFields(logrus.Fields{
						"query":   query.Name,
						"payload": string(query.Payload),
						"at":      query.LTime,
					}).Debug("Agent: Running job")

					go a.receiveRunJobQuery(query)

				case QueryRPCConfig:
					if a.config.Server {
						Log.WithFields(logrus.Fields{
							"query":   query.Name,
							"payload": string(query.Payload),
							"at":      query.LTime,
						}).Debug("Agent: Server receive a rpc config query")

						go a.receiveGetRPCConfigQuery(query)
					}
				case QueryJobCount:
					if a.config.Server {
						Log.WithFields(logrus.Fields{
							"query":   query.Name,
							"payload": string(query.Payload),
							"at":      query.LTime,
						}).Debug("Agent: Server receive a job count query")

						go a.receiveJobCountQuery(query)
					}
				case QueryJobDelete:
					if a.config.Server {
						Log.WithFields(logrus.Fields{
							"query":   query.Name,
							"payload": string(query.Payload),
							"at":      query.LTime,
						}).Debug("Agent: Server receive a job delete query")

						go a.receiveJobDeleteQuery(query)
					}
				default:
					Log.Warn("Agent: get a unknow message")
					Log.WithFields(logrus.Fields{
						"query":   query.Name,
						"payload": string(query.Payload),
						"at":      query.LTime,
					}).Debug("Agent: get a unknow message")
				}
			}

		case <-serfShutdownCh:
			Log.Warn("Agent: Serf shutdown detected, quitting")
			return
		}
	}
}

// Run func start a Agent process
func (a *Agent) Run() error {
	var err error
	if err != nil {
		Log.Fatalln(err)
		return err
	}
	if a.serf = a.setupSerf(); a.serf == nil {
		Log.Fatalln("Start serf failed!")
		return errors.New("Start serf failed")
	}
	a.serfJion(a.config.SerfJoin, true)
	// TODO: add prometheus support
	// start prometheus client

	if a.config.Server {
		Log.Info("Agent: Init server")

		Log.WithFields(logrus.Fields{
			"dsn": a.config.DSN,
		}).Debug("Agent: Connect to database")
		sqlStroe, err := NewSQLStore("mysql", a.config.DSN)
		if err != nil {
			Log.WithError(err).Fatal("Agent: Connect to database failed")
		}
		if err := sqlStroe.Migrate(false, &pb.Execution{}); err != nil {
			Log.WithError(err).Fatal("Agent: Migrate database table failed")
		}
		a.sqlStore = sqlStroe

		var tls rpc.TlsOpt
		var keyPair api.KayPair
		if a.config.RPCTls {
			tls = rpc.TlsOpt{
				CertFile: a.config.CertFile,
				KeyFile:  a.config.KeyFile,
				CaFile:   a.config.CAFile,
			}
			keyPair = api.KayPair{
				Key:  a.config.KeyFile,
				Cert: a.config.CertFile,
			}
		}

		// connect to kv store
		Log.WithFields(logrus.Fields{
			"backend":  a.config.JobStore,
			"server":   a.config.JobStoreServers,
			"keySpace": a.config.JobStoreKeyspace,
		}).Debug("Agent: Connect to job store")
		a.store, err = NewKVStore(a.config.JobStore, a.config.JobStoreServers, a.config.JobStoreKeyspace)
		if err != nil {
			Log.WithFields(logrus.Fields{
				"backend":  a.config.JobStore,
				"servers":  a.config.JobStoreServers,
				"keyspace": a.config.JobStoreKeyspace,
			}).Debug("Agent: Connect Backend Failed")
			Log.WithError(err).Error("Agent: Connent Backend Failed")
			return err
		}
		// run rpc server
		Log.WithFields(logrus.Fields{
			"BindIp":   a.config.RPCBindIP,
			"BindPort": a.config.RPCBindPort,
		}).Debug("Agent: Start RPC server")
		a.rpcServer = rpc.NewRPCServer(a.config.RPCBindIP, a.config.RPCBindPort, a, &tls)
		if err := a.rpcServer.Run(); err != nil {
			Log.WithError(err).Error("Agent: Start RPC Srever Failed")
			return err
		}

		// run scheduler
		a.scheduler = scheduler.New(a.runJobCh)
		a.scheduler.Start()

		// try to load job
		Log.Info("Agent: Load jobs from KV store")
		a.loadJobs(a.config.Region)

		// wait run job
		go func() {
			for {
				job := <-a.runJobCh
				Log.WithFields(logrus.Fields{
					"JobName":  job.Name,
					"Region":   job.Region,
					"cmd":      job.Command,
					"Schedule": job.Schedule,
				}).Debug("Agent: Job ready to run")
				ex := pb.Execution{
					SchedulerNodeName: job.SchedulerNodeName,
					JobName:           job.Name,
					Region:            job.Region,
					Group:             time.Now().UnixNano(),
					Retries:           0,
				}
				go a.sendRunJobQuery(&ex)
				Log.WithFields(logrus.Fields{
					"JobName": ex.JobName,
					"Region":  ex.Region,
					"Group":   ex.Group,
				}).Debug("Agent: Job Run serf query has been send")
			}
		}()

		// start api server
		a.apiServer, err = api.NewAPIServer(a.config.APIBindIP, a.config.APIBindPort, Log,
			a.config.APITokens, a.config.RPCTls, &keyPair, a)
		if err != nil {
			Log.WithError(err).Error("Agent: New API Server Failed")
			return err
		}
		a.apiServer.Run()
	}

	go a.serfEventLoop()
	return nil
}

func (a *Agent) loadJobs(region string) {
	jobs, err := a.store.GetJobList(region)
	if err != nil && err != store.ErrKeyNotFound {
		Log.WithError(err).Fatal("Agent: Get job list failed")
	}
	if len(jobs) == 0 {
		return
	}
	for _, job := range jobs {
		if !a.isLocked(job.Name, job.Region, &pb.Job{}) {
			locker, err := a.lock(job.Name, job.Region, &pb.Job{})
			if err != nil {
				Log.WithError(err).Fatal("Agent: Lock job failed")
			}
			a.addLocker(job.Name, locker)
			a.scheduler.AddJob(job)
		}
	}
}

func (a *Agent) Reload(args []string) {
	newConf, err := newConfig(args, a.version)
	if err != nil {
		Log.Warn(err)
		return
	}
	a.config = newConf

	a.serf.SetTags(a.config.Tags)
}

func (a *Agent) Stop(graceful bool) int {

	if !graceful {
		return 0
	}

	gracefulCh := make(chan struct{})

	Log.Info("Agent: Gracefully shutting down agent...")
	go func() {
		var wg sync.WaitGroup
		if a.config.Server {
			// shutdown scheduler and unlock all jobs
			go func() {
				wg.Add(1)
				a.scheduler.Stop()
				for _, locker := range a.jobLockers {
					locker.Unlock()
				}
				wg.Done()
			}()
			// graceful shutdown rpc server
			go func() {
				wg.Add(1)
				a.rpcServer.Shutdown(gracefulTime)
				wg.Done()
			}()
			// gracefull shutdown api server
			go func() {
				wg.Add(1)
				if err := a.apiServer.Stop(gracefulTime); err != nil {
					Log.Errorf("Agent:Graceful shutdown Api server failed: %s", err)
				}
				wg.Done()
			}()
		}
		go func() {
			wg.Add(1)
			if err := a.serf.Leave(); err != nil {
				Log.Errorf("Agent: Graceful shutdown down serf failed: %s", err)
			}
			a.serf.Shutdown()
			wg.Done()
		}()
		wg.Wait()
		a.sqlStore.Close()
		gracefulCh <- struct{}{}
	}()

	select {
	case <-time.After(gracefulTime):
		return 1
	case <-gracefulCh:
		return 0
	}
}

// TODO: use a rpc client pool
func (a *Agent) newRPCClient(ip string, port int) *rpc.RpcClient {
	var tls *rpc.TlsOpt
	if a.config.RPCTls {
		tls = &rpc.TlsOpt{
			CertFile: a.config.CertFile,
			KeyFile:  a.config.KeyFile,
			CaFile:   a.config.CAFile,
		}
	}
	client, err := rpc.NewRpcClient(ip, port, tls)
	if err != nil {
		Log.WithError(err).Fatal("Agent: start RPC Client failed")
	}
	return client
}

func New(args []string, version string) *Agent {
	config, err := newConfig(args, version)
	if err != nil {
		Log.WithError(err).Fatal("Agent:init failed")
	}
	return &Agent{
		jobLockers: make(map[string]store.Locker),
		eventCh:    make(chan serf.Event, 64),
		runJobCh:   make(chan *pb.Job),
		memStore:   NewMemStore(),
		config:     config,
		version:    version,
		mutex:      &sync.Mutex{},
	}

}

type byCount []*pb.JobCountResp

func (b byCount) Len() int           { return len(b) }
func (b byCount) Swap(i, j int)      { b[i], b[j] = b[j], b[i] }
func (b byCount) Less(i, j int) bool { return b[i].Count < b[j].Count }

func (a *Agent) minimalLoadServer(region string) (string, error) {

	resCh := make(chan []*pb.JobCountResp)
	errCh := make(chan error)

	go func() {
		jobCounts, err := a.sendJobCountQuery(region)
		if err != nil {
			errCh <- err
		}
		resCh <- jobCounts
	}()

	select {
	case err := <-errCh:
		return "", err
	case res := <-resCh:
		if res == nil || len(res) == 0 {
			Log.WithFields(logrus.Fields{
				"node":     a.config.Nodename,
				"function": "minimalLoadServer",
			}).Fatal("Agent: []*pb.JobCountResp is nil")
		}
		sort.Sort(byCount(res))
		return res[0].Name, nil
	case <-time.After(APITimeOut):
		return "", ErrTimeOut
	}
}

func (a *Agent) haveIt(name string) bool {
	a.mutex.Lock()
	_, exist := a.jobLockers[name]
	a.mutex.Unlock()
	return exist
}

func (a *Agent) addLocker(name string, locker store.Locker) {
	a.mutex.Lock()
	a.jobLockers[name] = locker
	a.mutex.Unlock()
}

func (a *Agent) deleteLocker(name string) error {
	a.mutex.Lock()
	defer a.mutex.Unlock()
	locker, exist := a.jobLockers[name]
	if exist {
		if err := locker.Unlock(); err != nil {
			return err
		}
		delete(a.jobLockers, name)
		return nil
	}
	return ErrNotExist
}
