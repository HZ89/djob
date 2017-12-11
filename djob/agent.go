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

package djob

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Knetic/govaluate"
	"github.com/Sirupsen/logrus"
	"github.com/facebookgo/pidfile"
	"github.com/hashicorp/memberlist"
	"github.com/hashicorp/serf/serf"

	"github.com/HZ89/djob/errors"
	"github.com/HZ89/djob/log"
	pb "github.com/HZ89/djob/message"
	"github.com/HZ89/djob/rpc"
	"github.com/HZ89/djob/scheduler"
	"github.com/HZ89/djob/store"
	"github.com/HZ89/djob/util"
	"github.com/HZ89/djob/web/api"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

const (
	gracefulTime = 5 * time.Second
	// APITimeOut max api execution time
	APITimeOut = 1 * time.Second
)

// RESERVEDTAGS store protected tag that is not allowed to be modified
var RESERVEDTAGS = [...]string{"RPCADIP", "RPCADPORT", "NODE", "REGION", "SERVER", "VERSION"}

// Agent djob agent struct
type Agent struct {
	lockerChain *store.LockerChain   // a sync/map use to store obj locker
	config      *Config              // configuration
	serf        *serf.Serf           // serf obj
	eventCh     chan serf.Event      // get serf event form this chan
	rpcServer   *rpc.RpcServer       // grpc server
	store       *store.KVStore       // key value store, it can be etcd, zookeeper etc.
	memStore    *store.MemStore      // memory cache, use to cache some key value obj
	scheduler   *scheduler.Scheduler // job scheduler, scheduling based on cron syntax
	apiServer   *api.APIServer       // web api server, use gin
	version     string
	mainVersion string
	runJobCh    chan *pb.Job    // scheduler push the job ready to run into this chan
	sqlStore    *store.SQLStore // sql store, use to persistent save job and execution obj
}

// prepare serf config
func (a *Agent) setupSerf() *serf.Serf {
	encryptKey, err := a.config.EncryptKey()
	if err != nil {
		log.FmdLoger.Fatal(err)
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

	log.FmdLoger.Info("Agent: Djob agent starting")
	log.FmdLoger.Debugf("Agent: Djob agent serf tag is %v", serfConfig.Tags)

	s, err := serf.Create(serfConfig)
	if err != nil {
		log.FmdLoger.Fatal(err)
		return nil
	}

	return s
}

// serfJion let serf intence jion a serf clust
func (a *Agent) serfJion(addrs []string, replay bool) (n int, err error) {
	log.FmdLoger.Infof("Agent: joining: %v replay: %v", addrs, replay)
	ignoreOld := !replay
	n, err = a.serf.Join(addrs, ignoreOld)
	if n > 0 {
		log.FmdLoger.Infof("Agent: joined: %d nodes", n)
	}
	if err != nil {
		log.FmdLoger.Warnf("Agent: error joining: %v", err)
	}
	return
}

// listen to serf events and perform the corresponding actions
// TODO: stripping serf, serf tags only supports string type, implement tags that support int, float, bool, string data type
func (a *Agent) serfEventLoop() {
	serfShutdownCh := a.serf.ShutdownCh()
	log.FmdLoger.Info("Agent: Listen for event")
	for {

		select {
		// handle serf event
		case e := <-a.eventCh:
			log.FmdLoger.WithFields(logrus.Fields{
				"event": e.String(),
			}).Debug("Agent: Received event")
			// log all member event, member jion, update, leave etc.
			//if memberevent, ok := e.(serf.MemberEvent); ok {
			//	for _, member := range memberevent.Members {
			//		log.FmdLoger.WithFields(logrus.Fields{
			//			"node":    a.config.Nodename,
			//			"members": member.Name,
			//			"event":   e.EventType(),
			//		}).Debug("Agent: Member event got")
			//	}
			//}

			// handle custom query event
			if e.EventType() == serf.EventQuery {
				query := e.(*serf.Query)

				switch qname := query.Name; qname {
				case QueryRunJob:
					log.FmdLoger.WithFields(logrus.Fields{
						"query":   query.Name,
						"payload": string(query.Payload),
						"at":      query.LTime,
					}).Debug("Agent: Running job")

					go a.receiveRunJobQuery(query)

				case QueryJobCount:
					if a.config.Server {
						log.FmdLoger.WithFields(logrus.Fields{
							"query":   query.Name,
							"payload": string(query.Payload),
							"at":      query.LTime,
						}).Debug("Agent: Server receive a job count query")

						go a.receiveJobCountQuery(query)
					}
				default:
					log.FmdLoger.Warn("Agent: get a unknow message")
					log.FmdLoger.WithFields(logrus.Fields{
						"query":   query.Name,
						"payload": string(query.Payload),
						"at":      query.LTime,
					}).Debug("Agent: get a unknow message")
				}
			}

		case <-serfShutdownCh:
			log.FmdLoger.Warn("Agent: Serf shutdown detected, quitting")
			return
		}
	}
}

// Run func start a Agent process
func (a *Agent) Run() error {
	var err error
	if err != nil {
		log.FmdLoger.Fatalln(err)
		return err
	}
	// start serf agent first
	if a.serf = a.setupSerf(); a.serf == nil {
		log.FmdLoger.Fatalln("Start serf failed!")
		return errors.ErrStartSerf
	}
	a.serfJion(a.config.SerfJoin, true)
	// TODO: add prometheus support
	// start prometheus client

	go a.serfEventLoop()

	// server start
	if a.config.Server {
		log.FmdLoger.Info("Agent: Init server")

		log.FmdLoger.WithFields(logrus.Fields{
			"dsn": a.config.DSN,
		}).Debug("Agent: Connect to database")
		sqlStroe, err := store.NewSQLStore("mysql", a.config.DSN)
		if err != nil {
			log.FmdLoger.WithError(err).Fatal("Agent: Connect to database failed")
		}
		if err = sqlStroe.Migrate(false, &pb.Execution{}, &pb.Job{}); err != nil {
			log.FmdLoger.WithError(err).Fatal("Agent: Migrate database table failed")
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
		log.FmdLoger.WithFields(logrus.Fields{
			"backend":  a.config.JobStore,
			"server":   a.config.JobStoreServers,
			"keySpace": a.config.JobStoreKeyspace,
		}).Debug("Agent: Connect to job store")
		a.store, err = store.NewKVStore(a.config.JobStore, a.config.JobStoreServers, a.config.JobStoreKeyspace)
		if err != nil {
			log.FmdLoger.WithFields(logrus.Fields{
				"backend":  a.config.JobStore,
				"servers":  a.config.JobStoreServers,
				"keyspace": a.config.JobStoreKeyspace,
			}).Debug("Agent: Connect Backend Failed")
			log.FmdLoger.WithError(err).Error("Agent: Connent Backend Failed")
			return err
		}

		// init a lockerChain
		a.lockerChain = store.NewLockerChain(a.config.Nodename, a.store)

		// run rpc server
		log.FmdLoger.WithFields(logrus.Fields{
			"BindIp":   a.config.RPCBindIP,
			"BindPort": a.config.RPCBindPort,
		}).Debug("Agent: Start RPC server")
		a.rpcServer = rpc.NewRPCServer(a.config.RPCBindIP, a.config.RPCBindPort, a, &tls)
		if err := a.rpcServer.Run(); err != nil {
			log.FmdLoger.WithError(err).Error("Agent: Start RPC Srever Failed")
			return err
		}

		// run scheduler
		a.scheduler = scheduler.New(a.runJobCh)
		a.scheduler.Start()

		// try to load job
		if a.config.LoadJobPolicy != LOADNOTHING {
			log.FmdLoger.Info("Agent: Load jobs")
			a.loadJobs(a.config.Region)
		}

		// run job
		go a.runJob()

		// start api server
		a.apiServer, err = api.NewAPIServer(a.config.APIBindIP, a.config.APIBindPort, a.config.APITokens,
			a.config.RPCTls, &keyPair, a)
		if err != nil {
			log.FmdLoger.WithError(err).Error("Agent: New API Server Failed")
			return err
		}
		a.apiServer.Run()
		log.FmdLoger.WithFields(logrus.Fields{
			"Ip":     a.config.APIBindIP,
			"Port":   a.config.APIBindPort,
			"Tokens": a.config.APITokens,
		}).Debug("Agent: API Server has been started")

		// start takeover process
		go a.takeOverJob()
	}

	return nil
}

// used to take over the unmanaged job
func (a *Agent) takeOverJob() {
	stopCh := make(chan struct{})
	for {
		watchCh, err := a.store.WatchLock(&pb.Job{Region: a.config.Region}, stopCh)
		if err == errors.ErrNotExist {
			log.FmdLoger.Debug("Agent: job lock path do not exist, sleep 5s retry")
			time.Sleep(5 * time.Second)
			continue
		}
		if err != nil {
			log.FmdLoger.WithError(err).Fatal("Agent: try to watch lock path failed")
		}
		for {
			select {
			case jobName := <-watchCh:
				log.FmdLoger.Debugf("Agent: watch got a job: %s", jobName)
				job := &pb.Job{Name: jobName, Region: a.config.Region}
				// try to lock this job
				lockType := &store.LockOption{
					Global:   true,
					LockType: store.OWN,
					TimeOut:  2 * time.Second,
				}
				err = a.lockerChain.AddLocker(job, lockType)
				if err == errors.ErrLockTimeout || err == errors.ErrRepetition {
					continue
				}
				if err != nil {
					log.FmdLoger.WithFields(logrus.Fields{
						"name":   job.Name,
						"region": job.Region,
					}).WithError(err).Fatal("Agent: lock job failed")
				}
				res, _, err := a.operationMiddleLayer(job, pb.Ops_READ, nil)
				if err != nil || len(res) == 0 {
					log.FmdLoger.WithField("jobName", job.Name).WithError(err).Warn("Agent: takeOver, read job filed")
					continue
				}
				job = res[0].(*pb.Job)

				job.SchedulerNodeName = a.config.Nodename
				_, _, err = a.operationMiddleLayer(job, pb.Ops_MODIFY, nil)
				if err != nil {
					log.FmdLoger.WithError(err).Fatal("Agent: takeOver, modify job failed")
				}
				err = a.scheduler.AddJob(job)
				if err != nil {
					log.FmdLoger.WithFields(logrus.Fields{
						"name":      job.Name,
						"region":    job.Region,
						"scheduler": job.Schedule,
					}).WithError(err).Fatal("Agent: add job to scheduler failed")
				}

				if err != nil {
					log.FmdLoger.WithFields(logrus.Fields{
						"name":   jobName,
						"region": a.config.Region,
					}).WithError(err).Fatal("Agent: get job info failed")
				}
			}
		}
	}
}

// listen to runJobCh and run a job
func (a *Agent) runJob() {
	for {
		job := <-a.runJobCh
		log.FmdLoger.WithFields(logrus.Fields{
			"Name":     job.Name,
			"Region":   job.Region,
			"cmd":      job.Command,
			"Schedule": job.Schedule,
		}).Debug("Agent: Job ready to run")
		// if job disabled just pass it
		if job.Disable {
			log.FmdLoger.WithFields(logrus.Fields{
				"Name":     job.Name,
				"Region":   job.Region,
				"cmd":      job.Command,
				"Schedule": job.Schedule,
				"disable":  job.Disable,
			}).Debug("Agent: Job is in disable status")
			continue
		}
		_, err := a.RunJob(job.Name, job.Region)
		if err != nil {
			log.FmdLoger.WithFields(logrus.Fields{
				"Name":   job.Name,
				"Region": job.Region,
			}).WithError(err).Error("Agent: Job run failed")
			continue
		}
	}
}

// load jobs from database, support three load policy:
// 1. LOADNOTHING just pass this process
// 2. LOADOWN  just load which job belong to(job.SchedulerNodeName eq this.Nodename) this node
// 3. LOADALL  just load all job and set it belong to this node
func (a *Agent) loadJobs(region string) {
	res, _, err := a.operationMiddleLayer(&pb.Job{Region: region}, pb.Ops_READ, nil)
	if err != nil {
		log.FmdLoger.WithError(err).Fatal("Agent: load job failed")
	}
	for _, i := range res {
		if t, ok := i.(*pb.Job); ok {
			load := true
			if a.store.IsLocked(t, store.OWN) {
				load = false
			}
			if a.config.LoadJobPolicy == LOADOWN && t.SchedulerNodeName != a.config.Nodename {
				load = false
			}
			if load {
				lockType := &store.LockOption{
					Global:   true,
					LockType: store.OWN,
				}
				if err := a.lockerChain.AddLocker(t, lockType); err != nil {
					log.FmdLoger.WithError(err).Fatal("Agent: load job failed")
				}
				if err := a.scheduler.AddJob(t); err != nil {
					log.FmdLoger.WithError(err).Fatal("Agent: load job failed")
				}
				log.FmdLoger.WithFields(logrus.Fields{
					"name":   t.Name,
					"region": t.Region,
				}).Debug("Agent: load a job successfully")
			}
			continue
		}
		log.FmdLoger.WithError(errors.ErrNotExpectation).Fatalf("Agent: want *message.Job but got %v", reflect.TypeOf(i))
	}
}

// Reload the agent
func (a *Agent) Reload(args []string) {
	newConf, err := newConfig(args, a.version)
	if err != nil {
		log.FmdLoger.Warn(err)
		return
	}
	a.config = newConf

	a.serf.SetTags(a.config.Tags)
}

// Stop this agent
func (a *Agent) Stop(graceful bool) int {
	if !graceful {
		return 0
	}

	gracefulCh := make(chan struct{})

	log.FmdLoger.Info("Agent: Gracefully shutting down agent...")
	go func() {
		var wg sync.WaitGroup
		if a.config.Server {
			// shutdown scheduler and unlock all jobs
			go func() {
				wg.Add(1)
				a.scheduler.Stop()
				a.lockerChain.Stop()
				wg.Done()
			}()
			// graceful shutdown rpc server
			go func() {
				wg.Add(1)
				a.rpcServer.Shutdown(gracefulTime)
				wg.Done()
			}()
			// graceful shutdown api server
			go func() {
				wg.Add(1)
				if err := a.apiServer.Stop(gracefulTime); err != nil {
					log.FmdLoger.Errorf("Agent:Graceful shutdown Api server failed: %s", err)
				}
				wg.Done()
			}()
			a.memStore.Release()
			a.sqlStore.Close()
		}
		go func() {
			wg.Add(1)
			if err := a.serf.Leave(); err != nil {
				log.FmdLoger.Errorf("Agent: Graceful shutdown down serf failed: %s", err)
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

// initialize a grpc client
// TODO: use a rpc client pool
func (a *Agent) newRPCClient(ip string, port int) *rpc.RpcClient {
	var tls *rpc.TlsOpt
	// TODO: think about the certificate automatically issued
	if a.config.RPCTls {
		tls = &rpc.TlsOpt{
			CertFile: a.config.CertFile,
			KeyFile:  a.config.KeyFile,
			CaFile:   a.config.CAFile,
		}
	}
	client, err := rpc.NewRpcClient(ip, port, tls)
	if err != nil {
		log.FmdLoger.WithError(err).Fatal("Agent: start RPC Client failed")
	}
	return client
}

// New initialization agent, args is os.agrs list
func New(args []string, version string) *Agent {
	config, err := newConfig(args, version)
	if err != nil {
		log.FmdLoger.WithError(err).Fatal("Agent:init failed")
	}
	vs := strings.Split(version, ".")
	return &Agent{
		eventCh:     make(chan serf.Event, 64),
		runJobCh:    make(chan *pb.Job),
		memStore:    store.NewMemStore(),
		config:      config,
		version:     version,
		mainVersion: fmt.Sprintf("%s.%s", vs[0], vs[1]),
	}

}

// used to sort the agent by job count
type byCount []*pb.JobCountResp

func (b byCount) Len() int           { return len(b) }
func (b byCount) Swap(i, j int)      { b[i], b[j] = b[j], b[i] }
func (b byCount) Less(i, j int) bool { return b[i].Count < b[j].Count }

// find out the node with the fewest job
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
			log.FmdLoger.WithFields(logrus.Fields{
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

// find a available node randomly
func (a *Agent) randomPickServer(region string) (string, error) {
	p, err := a.createSerfQueryParam(fmt.Sprintf("SERVER=='true'&&REGION=='%s'", region))
	if err != nil {
		return "", err
	}
	return p.FilterNodes[rand.Intn(len(p.FilterNodes))], nil
}

// constructed serf.QueryParam using expression
func (a *Agent) createSerfQueryParam(expression string) (*serf.QueryParam, error) {
	var queryParam serf.QueryParam
	fitMembers, err := a.processFilteredNodes(expression, true)
	if err != nil {
		return nil, err
	}

	if len(fitMembers) > 0 {
		var names []string
		for _, m := range fitMembers {
			names = append(names, m.Name)
		}
		queryParam = serf.QueryParam{
			FilterNodes: names,
			RequestAck:  true,
		}
		return &queryParam, nil
	}
	log.FmdLoger.WithField("expression", expression).WithError(errors.ErrCanNotFoundNode).Debug("Agent: no nodes match")
	return nil, errors.ErrCanNotFoundNode
}

// find all the node that make the expression true
func (a *Agent) processFilteredNodes(expression string, needAlievd bool) ([]serf.Member, error) {
	exp, err := govaluate.NewEvaluableExpression(expression)
	if err != nil {
		return nil, err
	}
	var fitMembers []serf.Member
	for _, member := range a.serf.Members() {
		if a.isWantedMember(exp, member) {
			fitMembers = append(fitMembers, member)
		}
	}
	return fitMembers, nil
}

// if the member is wanted, return true
func (a *Agent) isWantedMember(expression *govaluate.EvaluableExpression, member serf.Member) bool {
	wt := expression.Vars()
	sort.Sort(sort.StringSlice(wt))
	var mks []string
	for k := range member.Tags {
		mks = append(mks, k)
	}
	intersection := util.Intersect(mks, wt)
	sort.Sort(sort.StringSlice(intersection))
	if !reflect.DeepEqual(wt, intersection) {
		return false
	}
	parameters := make(map[string]interface{})
	for _, tk := range wt {
		parameters[tk] = member.Tags[tk]
	}
	result, _ := expression.Evaluate(parameters)
	return result.(bool)
}

// get grpc configuration information from tags
func (a *Agent) getRPCConfig(nodeName string) (string, int, error) {
	for _, member := range a.serf.Members() {
		if member.Name == nodeName {
			if member.Status != serf.StatusAlive {
				return "", 0, errors.ErrNodeDead
			}
			ip, iok := member.Tags["RPCADIP"]
			port, pok := member.Tags["RPCADPORT"]
			if !iok || !pok || ip == "" || port == "" {
				return "", 0, errors.ErrNodeNoRPC
			}
			iport, err := strconv.Atoi(port)
			if err != nil {
				return "", 0, err
			}
			return ip, iport, nil
		}
	}
	return "", 0, errors.ErrNotExist
}

// WritePid create pid file
func (a *Agent) WritePid() error {
	pidfile.SetPidfilePath(a.config.PidFile)
	return pidfile.Write()
}

// RemovePid remove pid file
func (a *Agent) RemovePid() error {
	return os.Remove(pidfile.GetPidfilePath())
}
