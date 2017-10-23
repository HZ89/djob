package djob

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"reflect"
	"sort"
	"sync"
	"time"

	"github.com/Knetic/govaluate"
	"github.com/Sirupsen/logrus"
	libstore "github.com/docker/libkv/store"
	"github.com/gogo/protobuf/proto"
	"github.com/hashicorp/memberlist"
	"github.com/hashicorp/serf/serf"

	"version.uuzu.com/zhuhuipeng/djob/errors"
	"version.uuzu.com/zhuhuipeng/djob/log"
	pb "version.uuzu.com/zhuhuipeng/djob/message"
	"version.uuzu.com/zhuhuipeng/djob/rpc"
	"version.uuzu.com/zhuhuipeng/djob/scheduler"
	"version.uuzu.com/zhuhuipeng/djob/store"
	"version.uuzu.com/zhuhuipeng/djob/util"
	"version.uuzu.com/zhuhuipeng/djob/web/api"
)

const (
	gracefulTime = 5 * time.Second
)

type Agent struct {
	lockerChain *store.LockerChain
	config      *Config
	serf        *serf.Serf
	eventCh     chan serf.Event
	ready       bool
	rpcServer   *rpc.RpcServer
	store       *store.KVStore
	memStore    *store.MemStore
	scheduler   *scheduler.Scheduler
	apiServer   *api.APIServer
	version     string
	runJobCh    chan *pb.Job
	sqlStore    *store.SQLStore
}

func (a *Agent) setupSerf() *serf.Serf {
	encryptKey, err := a.config.EncryptKey()
	if err != nil {
		log.Loger.Fatal(err)
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

	log.Loger.Info("Agent: Djob agent starting")

	s, err := serf.Create(serfConfig)
	if err != nil {
		log.Loger.Fatal(err)
		return nil
	}
	//a.memberCache = make(map[string]map[string]string)

	return s
}

// serfJion let serf intence jion a serf clust
func (a *Agent) serfJion(addrs []string, replay bool) (n int, err error) {
	log.Loger.Infof("Agent: joining: %v replay: %v", addrs, replay)
	ignoreOld := !replay
	n, err = a.serf.Join(addrs, ignoreOld)
	if n > 0 {
		log.Loger.Infof("Agent: joined: %d nodes", n)
	}
	if err != nil {
		log.Loger.Warnf("Agent: error joining: %v", err)
	}
	return
}

func (a *Agent) serfEventLoop() {
	serfShutdownCh := a.serf.ShutdownCh()
	log.Loger.Info("Agent: Listen for event")
	for {

		select {
		// handle serf event
		case e := <-a.eventCh:
			log.Loger.WithFields(logrus.Fields{
				"event": e.String(),
			}).Debug("Agent: Received event")

			if memberevent, ok := e.(serf.MemberEvent); ok {
				for _, member := range memberevent.Members {
					log.Loger.WithFields(logrus.Fields{
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
				case QueryModJob:
					if a.config.Server {
						log.Loger.WithFields(logrus.Fields{
							"query":   query.Name,
							"payload": string(query.Payload),
							"at":      query.LTime,
						}).Debug("Agent: Server receive a add new job event")

						go a.receiveNewJobQuery(query)
					}
				case QueryRunJob:
					log.Loger.WithFields(logrus.Fields{
						"query":   query.Name,
						"payload": string(query.Payload),
						"at":      query.LTime,
					}).Debug("Agent: Running job")

					go a.receiveRunJobQuery(query)

				case QueryRPCConfig:
					if a.config.Server {
						log.Loger.WithFields(logrus.Fields{
							"query":   query.Name,
							"payload": string(query.Payload),
							"at":      query.LTime,
						}).Debug("Agent: Server receive a rpc config query")

						go a.receiveGetRPCConfigQuery(query)
					}
				case QueryJobCount:
					if a.config.Server {
						log.Loger.WithFields(logrus.Fields{
							"query":   query.Name,
							"payload": string(query.Payload),
							"at":      query.LTime,
						}).Debug("Agent: Server receive a job count query")

						go a.receiveJobCountQuery(query)
					}
				case QueryJobDelete:
					if a.config.Server {
						log.Loger.WithFields(logrus.Fields{
							"query":   query.Name,
							"payload": string(query.Payload),
							"at":      query.LTime,
						}).Debug("Agent: Server receive a job delete query")

						go a.receiveJobDeleteQuery(query)
					}
				default:
					log.Loger.Warn("Agent: get a unknow message")
					log.Loger.WithFields(logrus.Fields{
						"query":   query.Name,
						"payload": string(query.Payload),
						"at":      query.LTime,
					}).Debug("Agent: get a unknow message")
				}
			}

		case <-serfShutdownCh:
			log.Loger.Warn("Agent: Serf shutdown detected, quitting")
			return
		}
	}
}

// Run func start a Agent process
func (a *Agent) Run() error {
	var err error
	if err != nil {
		log.Loger.Fatalln(err)
		return err
	}
	if a.serf = a.setupSerf(); a.serf == nil {
		log.Loger.Fatalln("Start serf failed!")
		return errors.ErrStartSerf
	}
	a.serfJion(a.config.SerfJoin, true)
	// TODO: add prometheus support
	// start prometheus client

	if a.config.Server {
		log.Loger.Info("Agent: Init server")

		log.Loger.WithFields(logrus.Fields{
			"dsn": a.config.DSN,
		}).Debug("Agent: Connect to database")
		sqlStroe, err := store.NewSQLStore("mysql", a.config.DSN)
		if err != nil {
			log.Loger.WithError(err).Fatal("Agent: Connect to database failed")
		}
		if err = sqlStroe.Migrate(false, &pb.Execution{}, &pb.Job{}); err != nil {
			log.Loger.WithError(err).Fatal("Agent: Migrate database table failed")
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
		log.Loger.WithFields(logrus.Fields{
			"backend":  a.config.JobStore,
			"server":   a.config.JobStoreServers,
			"keySpace": a.config.JobStoreKeyspace,
		}).Debug("Agent: Connect to job store")
		a.store, err = store.NewKVStore(a.config.JobStore, a.config.JobStoreServers, a.config.JobStoreKeyspace)
		if err != nil {
			log.Loger.WithFields(logrus.Fields{
				"backend":  a.config.JobStore,
				"servers":  a.config.JobStoreServers,
				"keyspace": a.config.JobStoreKeyspace,
			}).Debug("Agent: Connect Backend Failed")
			log.Loger.WithError(err).Error("Agent: Connent Backend Failed")
			return err
		}
		// run rpc server
		log.Loger.WithFields(logrus.Fields{
			"BindIp":   a.config.RPCBindIP,
			"BindPort": a.config.RPCBindPort,
		}).Debug("Agent: Start RPC server")
		a.rpcServer = rpc.NewRPCServer(a.config.RPCBindIP, a.config.RPCBindPort, a, &tls)
		if err := a.rpcServer.Run(); err != nil {
			log.Loger.WithError(err).Error("Agent: Start RPC Srever Failed")
			return err
		}

		// run scheduler
		a.scheduler = scheduler.New(a.runJobCh)
		a.scheduler.Start()

		// try to load job
		log.Loger.Info("Agent: Load jobs from KV store")
		a.loadJobs(a.config.Region)

		// wait run job
		go func() {
			for {
				job := <-a.runJobCh
				log.Loger.WithFields(logrus.Fields{
					"JobName":  job.Name,
					"Region":   job.Region,
					"cmd":      job.Command,
					"Schedule": job.Schedule,
				}).Debug("Agent: Job ready to run")
				ex := pb.Execution{
					SchedulerNodeName: job.SchedulerNodeName,
					Name:              job.Name,
					Region:            job.Region,
					Group:             time.Now().UnixNano(),
					Retries:           0,
				}
				go a.sendRunJobQuery(&ex)
				log.Loger.WithFields(logrus.Fields{
					"JobName": ex.Name,
					"Region":  ex.Region,
					"Group":   ex.Group,
				}).Debug("Agent: Job Run serf query has been send")
			}
		}()

		// start api server
		a.apiServer, err = api.NewAPIServer(a.config.APIBindIP, a.config.APIBindPort, a.config.APITokens,
			a.config.RPCTls, &keyPair, a)
		if err != nil {
			log.Loger.WithError(err).Error("Agent: New API Server Failed")
			return err
		}
		a.apiServer.Run()
	}

	go a.serfEventLoop()
	return nil
}

func (a *Agent) loadJobs(region string) {
	jobs, err := a.store.GetRegionJobs(region)
	if err != nil && err != libstore.ErrKeyNotFound {
		log.Loger.WithError(err).Fatal("Agent: Get job list failed")
	}
	if len(jobs) == 0 {
		return
	}
	for _, job := range jobs {
		if !a.store.IsLocked(job, store.OWN) {
			locker, err := a.store.Lock(job, store.OWN, a.config.Nodename)
			if err != nil {
				log.Loger.WithError(err).Fatal("Agent: Lock job failed")
			}
			err = a.lockerChain.AddLocker(job.Name, locker)
			if err != nil {
				log.Loger.WithError(err).Fatal("Agent: Save locker failed")
			}
			a.scheduler.AddJob(job)
		}
	}
}

func (a *Agent) Reload(args []string) {
	newConf, err := newConfig(args, a.version)
	if err != nil {
		log.Loger.Warn(err)
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

	log.Loger.Info("Agent: Gracefully shutting down agent...")
	go func() {
		var wg sync.WaitGroup
		if a.config.Server {
			// shutdown scheduler and unlock all jobs
			go func() {
				wg.Add(1)
				a.scheduler.Stop()
				a.lockerChain.ReleaseAll()
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
					log.Loger.Errorf("Agent:Graceful shutdown Api server failed: %s", err)
				}
				wg.Done()
			}()
		}
		go func() {
			wg.Add(1)
			if err := a.serf.Leave(); err != nil {
				log.Loger.Errorf("Agent: Graceful shutdown down serf failed: %s", err)
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
		log.Loger.WithError(err).Fatal("Agent: start RPC Client failed")
	}
	return client
}

func New(args []string, version string) *Agent {
	config, err := newConfig(args, version)
	if err != nil {
		log.Loger.WithError(err).Fatal("Agent:init failed")
	}
	return &Agent{
		lockerChain: store.NewLockerChain(),
		eventCh:     make(chan serf.Event, 64),
		runJobCh:    make(chan *pb.Job),
		memStore:    store.NewMemStore(),
		config:      config,
		version:     version,
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
			log.Loger.WithFields(logrus.Fields{
				"node":     a.config.Nodename,
				"function": "minimalLoadServer",
			}).Fatal("Agent: []*pb.JobCountResp is nil")
		}
		sort.Sort(byCount(res))
		return res[0].Name, nil
	case <-time.After(APITimeOut):
		return "", errors.ErrTimeOut
	}
}

func (a *Agent) randomPickServer(region string) (string, error) {
	p, err := a.createSerfQueryParam(fmt.Sprintf("server=='true'&&region=='%s'", region))
	if err != nil {
		return "", err
	}
	return p.FilterNodes[rand.Intn(len(p.FilterNodes))], nil
}

func (a *Agent) createSerfQueryParam(expression string) (*serf.QueryParam, error) {
	var queryParam serf.QueryParam
	nodeNames, err := a.processFilteredNodes(expression)
	if err != nil {
		return nil, err
	}

	if len(nodeNames) > 0 {
		queryParam = serf.QueryParam{
			FilterNodes: nodeNames,
			RequestAck:  true,
		}
		return &queryParam, nil
	}
	return nil, errors.ErrCanNotFoundNode
}

func (a *Agent) processFilteredNodes(expression string) ([]string, error) {
	exp, err := govaluate.NewEvaluableExpression(expression)
	if err != nil {
		return nil, err
	}
	wt := exp.Vars()
	sort.Sort(sort.StringSlice(wt))
	var nodeNames []string
	for _, member := range a.serf.Members() {
		if member.Status == serf.StatusAlive {
			var mtks []string
			for k := range member.Tags {
				mtks = append(mtks, k)
			}
			intersection := util.Intersect([]string(mtks), wt)
			sort.Sort(sort.StringSlice(intersection))
			if reflect.DeepEqual(wt, intersection) {
				parameters := make(map[string]interface{})
				for _, tk := range wt {
					parameters[tk] = member.Tags[tk]
				}
				result, err := exp.Evaluate(parameters)
				if err != nil {
					return nil, err
				}
				if result.(bool) {
					nodeNames = append(nodeNames, member.Name)
				}
			}
		}
	}
	return nodeNames, nil
}

func (a *Agent) genrateResultPb(status int, message string) ([]byte, error) {
	r := pb.QueryResult{
		Status:  int32(status),
		Node:    a.config.Nodename,
		Message: message,
	}
	rpb, err := proto.Marshal(&r)
	if err != nil {
		return nil, err
	}
	return rpb, nil
}
