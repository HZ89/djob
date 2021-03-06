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
	"time"

	"github.com/HZ89/djob/errors"
	"github.com/HZ89/djob/log"
	pb "github.com/HZ89/djob/message"
	"github.com/HZ89/djob/store"
	"github.com/HZ89/djob/util"
	"github.com/Sirupsen/logrus"
)

// TODO: CRUD need cache

// separate local ops and remote ops
func (a *Agent) operationMiddleLayer(obj interface{}, ops pb.Ops, search *pb.Search) ([]interface{}, int32, error) {
	objRegion := util.GetFieldValue(obj, "Region")
	objName := util.GetFieldValue(obj, "Name")
	regionString, okr := objRegion.(string)
	nameString, okn := objName.(string)
	if !okr && !okn {
		return nil, 0, errors.ErrType
	}
	// have a region at lest
	if regionString == "" {
		return nil, 0, errors.ErrNoReg
	}
	// have same region maybe can be a localops
	if regionString == a.config.Region {
		job, ok := obj.(*pb.Job)
		// if this obj not a Job, or have no name, or this node have own lock, then it can be ops in this node
		if nameString == "" || a.lockerChain.HaveIt(obj, store.OWN) || !ok {
			log.FmdLoger.WithField("obj", obj).Debug("Agent: this obj localOps")
			return a.localOps(obj, ops, search)
		}
		// obj must be Job
		// if job belong to this node, just go to localops
		if job.SchedulerNodeName == a.config.Nodename {
			return a.localOps(obj, ops, search)
		}

		// if have no schedulernode find a node
		if job.SchedulerNodeName == "" {
			// find who handler this job
			owner := a.store.WhoLocked(job, store.OWN)
			if owner == "" {
				var err error
				// no one perform this, random send to some one
				owner, err = a.randomPickServer(a.config.Region)
				if err != nil {
					return nil, 0, err
				}
			}
			job.SchedulerNodeName = owner
		}

		return a.remoteOps(obj, ops, search, job.SchedulerNodeName)
	}
	// find some node in this region, proxy this action to it
	nextHandler, err := a.randomPickServer(regionString)
	if err != nil {
		return nil, 0, err
	}
	log.FmdLoger.WithField("obj", obj).Debug("Agent: this obj remoteOps")
	return a.remoteOps(obj, ops, search, nextHandler)
}

// forward ops to remote region by grpc
func (a *Agent) remoteOps(obj interface{}, ops pb.Ops, search *pb.Search, nodeName string) ([]interface{}, int32, error) {
	ip, port, err := a.getRPCConfig(nodeName)
	if err != nil {
		return nil, 0, err
	}
	log.FmdLoger.WithFields(logrus.Fields{
		"local nodeName":  a.config.Nodename,
		"remote nodeName": nodeName,
		"remote Ip":       ip,
		"remote Port":     port,
	}).Debug("Agent: proxy Ops to remote server")

	rpcClient := a.newRPCClient(ip, port)
	defer rpcClient.Shutdown()
	return rpcClient.DoOps(obj, ops, search)
}

// switch obj to each class
func (a *Agent) localOps(obj interface{}, ops pb.Ops, search *pb.Search) ([]interface{}, int32, error) {
	switch t := obj.(type) {
	case *pb.Job:
		return a.handleJobOps(t, ops, search)
	case *pb.Execution:
		return a.handleExecutionOps(t, ops, search)
	case *pb.JobStatus:
		return a.handleJobStatusOps(t, ops)
	}
	return nil, 0, errors.ErrUnknownType
}

// jobstatus CRUD
func (a *Agent) handleJobStatusOps(status *pb.JobStatus, ops pb.Ops) (out []interface{}, count int32, err error) {
	log.FmdLoger.WithFields(logrus.Fields{
		"status": status,
		"ops":    ops,
	}).Debug("Agent: ops job status")
	switch {
	case ops == pb.Ops_READ:
		_, err = a.store.Get(status)
	case ops == pb.Ops_DELETE:
		err = a.store.Delete(status)
	}
	if err != nil {
		log.FmdLoger.WithFields(logrus.Fields{
			"status": status,
			"ops":    ops,
		}).WithError(err).Debug("Agent: ops job status failed")
		return
	}
	out = append(out, status)
	return
}

// execution CRUD
func (a *Agent) handleExecutionOps(ex *pb.Execution, ops pb.Ops, search *pb.Search) ([]interface{}, int32, error) {
	var count int32
	var err error
	switch ops {
	case pb.Ops_READ:
		var rows []*pb.Execution

		if search != nil {
			sqlExec := a.sqlStore.Model(&pb.Execution{})
			var condition *store.SearchCondition
			// if have conditions use it else use execution obj as search conditions
			if len(search.Conditions) != 0 {
				condition, err = store.NewSearchCondition(search.Conditions, search.Links)
				if err != nil {
					return nil, count, err
				}
				sqlExec = sqlExec.Where(condition)
			} else {
				sqlExec = sqlExec.Where(ex)
			}

			if search.PageNum != 0 || search.PageSize != 0 {
				sqlExec = sqlExec.PageSize(int(search.PageSize)).PageNum(int(search.PageNum))
			}

			if search.Count {
				err = sqlExec.Find(&rows).PageCount(&count).Err
			} else {
				err = sqlExec.Find(&rows).Err
			}
			if err != nil {
				return nil, count, err
			}
		} else {
			err = a.sqlStore.Model(&pb.Execution{}).Where(ex).Find(&rows).Err
			if err != nil {
				return nil, count, err
			}
		}

		out := make([]interface{}, len(rows))
		for i, row := range rows {
			out[i] = row
		}

		return out, count, err
	case pb.Ops_ADD:
		if err = a.sqlStore.Create(ex).Err; err != nil {
			return nil, 0, err
		}
		out := make([]interface{}, 1)
		out[0] = ex

		// set JobStatus counter

		{
			a.mu.Lock()
			defer a.mu.Unlock()
			var pStatus = pb.JobStatus{Name: ex.Name, Region: ex.Region}

			var tryTimes int
			lockType := &store.LockOption{
				Global:   false,
				LockType: store.RW,
			}
		TRYLOCK:
			err := a.lockerChain.AddLocker(pStatus, lockType)
			if err != nil {
				if tryTimes < 3 {
					log.FmdLoger.WithField("TryLockTiems", tryTimes).WithError(err).Debug("Agent: Modify JobStatus, try lock")
					tryTimes++
					goto TRYLOCK
				}
				return nil, 0, err
			}
			defer a.lockerChain.ReleaseLocker(pStatus, false, store.RW)

			pIndex, err := a.store.Get(&pStatus)
			if err != nil && err != errors.ErrNotExist {
				return nil, 0, err
			}
			cStatus := pb.JobStatus{
				Name:         pStatus.Name,
				Region:       pStatus.Region,
				LastError:    pStatus.LastError,
				SuccessCount: pStatus.SuccessCount,
				LastSuccess:  pStatus.LastSuccess,
				ErrorCount:   pStatus.ErrorCount,
			}
			cStatus.LastHandleAgent = ex.RunNodeName
			if ex.Succeed {
				cStatus.SuccessCount += 1
				cStatus.LastSuccess = time.Unix(0, ex.FinishTime).String()
			} else {
				cStatus.ErrorCount += 1
				cStatus.LastError = time.Unix(0, ex.FinishTime).String()
			}
			if err == errors.ErrNotExist {
				if err = a.store.Set(&cStatus, 0, nil); err != nil {
					return out, count, err
				}
			} else {
				if err = a.store.Set(&cStatus, pIndex, &pStatus); err != nil {
					return out, count, err
				}
			}
		}

		return out, count, nil
	}
	return nil, 0, errors.ErrUnknownOps
}

// Job CRUD
func (a *Agent) handleJobOps(job *pb.Job, ops pb.Ops, search *pb.Search) ([]interface{}, int32, error) {
	var err error
	var count int32

	switch {
	case ops == pb.Ops_ADD:
		if err = util.VerifyJob(job); err != nil {
			log.FmdLoger.WithField("job", job).Debug("Agent: job is invalid")
			return nil, count, err
		}

		if job.ParentJob != nil {
			if job.ParentJob.Region != job.Region {
				return nil, count, errors.ErrParentNotInSameRegion
			}
			var parent pb.Job
			if err = a.sqlStore.Model(&pb.Job{}).Where(job.ParentJob).Find(&parent).Err; err != nil {
				if err == errors.ErrNotExist {
					return nil, count, errors.ErrParentNotExist
				}
				return nil, count, err
			}
			log.FmdLoger.WithFields(logrus.Fields{
				"sub-job":    job,
				"parent-job": parent,
			}).Debug("Agent: parent found")
			job.ParentJob = &parent
			job.ParentJobName = job.ParentJob.Name
		}

		job.SchedulerNodeName = a.config.Nodename

		log.FmdLoger.WithField("job", job).Debug("Agent: ready to save job to SQL")
		if err = a.sqlStore.Create(job).Err; err != nil {
			return nil, count, err
		}

		// set own locker on the job
		lockType := &store.LockOption{
			Global:   true,
			LockType: store.OWN,
		}
		if err = a.lockerChain.AddLocker(job, lockType); err != nil {
			a.sqlStore.Delete(job)
			return nil, count, err
		}

		// initialize a jobStatus to kv store
		status := pb.JobStatus{Name: job.Name, Region: job.Region}
		if err = a.store.Set(&status, 0, nil); err != nil {
			a.sqlStore.Delete(job)
			a.lockerChain.ReleaseLocker(job, true, store.OWN)
			return nil, count, err
		}

		log.FmdLoger.WithField("job", job).Debug("Agent: add job to scheduler")
		if err = a.scheduler.AddJob(job); err != nil {
			a.sqlStore.Delete(job)
			a.lockerChain.ReleaseLocker(job, true, store.OWN)
			a.store.Delete(&status)
			return nil, count, err
		}

		out := make([]interface{}, 1)
		out[0] = job
		return out, count, err
	case ops == pb.Ops_MODIFY:
		if err = util.VerifyJob(job); err != nil {
			return nil, count, err
		}

		if job.ParentJob != nil {
			if job.ParentJob.Region != job.Region {
				return nil, count, errors.ErrParentNotInSameRegion
			}
			var parent pb.Job
			if err = a.sqlStore.Model(&pb.Job{}).Where(job.ParentJob).Find(&parent).Err; err != nil {
				if err == errors.ErrNotExist {
					return nil, count, errors.ErrParentNotExist
				}
				return nil, count, err
			}
			job.ParentJobName = parent.Name
		}

		var oldJob pb.Job
		if err = a.sqlStore.Model(&pb.Job{}).Where(&pb.Job{Name: job.Name, Region: job.Region}).Find(&oldJob).Err; err != nil {
			return nil, count, err
		}
		if err = a.sqlStore.Model(&pb.Job{}).Modify(job).Err; err != nil {
			return nil, count, err
		}
		if err = a.scheduler.AddJob(job); err != nil {
			log.FmdLoger.WithError(err).Fatal("Agent: save job into scheduler failed")
		}
		ttl, _ := util.GetTimeInterval(job.Schedule)
		if err = a.memStore.Set(job.Region+"/"+job.Name, job, ttl*2); err != nil {
			log.FmdLoger.WithError(err).Fatal("Agent: set memory cache failed")
		}
		out := make([]interface{}, 1)
		out[0] = job
		return out, count, err
	case ops == pb.Ops_DELETE:
		var subJobs []*pb.Job
		if err = a.sqlStore.Model(&pb.Job{}).Where(&pb.Job{ParentJobName: job.Name}).Find(&subJobs).Err; err != nil && err != errors.ErrNotExist {
			return nil, count, err
		}
		if len(subJobs) != 0 {
			return nil, count, errors.ErrHaveSubJob
		}

		a.scheduler.DeleteJob(job)
		if err = a.sqlStore.Delete(job).Err; err != nil {
			return nil, count, err
		}
		if err = a.store.Delete(&pb.JobStatus{Name: job.Name, Region: job.Region}); err != nil {
			return nil, count, err
		}
		a.lockerChain.ReleaseLocker(job, true, store.OWN)
		if err = a.memStore.Delete(job.Region + "/" + job.Name); err != nil {
			log.FmdLoger.WithError(err).Fatal("Agent delete memory cache failed")
		}
		out := make([]interface{}, 1)
		out[0] = job
		return out, count, err
	case ops == pb.Ops_READ:
		var rows []*pb.Job

		if search != nil {
			sqlExec := a.sqlStore.Model(&pb.Job{})
			var condition *store.SearchCondition
			// if have conditions use it else use job obj as search conditions
			if len(search.Conditions) != 0 {
				condition, err = store.NewSearchCondition(search.Conditions, search.Links)
				if err != nil {
					return nil, count, err
				}
				sqlExec = sqlExec.Where(condition)
			} else {
				sqlExec = sqlExec.Where(job)
			}

			if search.PageNum != 0 || search.PageSize != 0 {
				sqlExec = sqlExec.PageSize(int(search.PageSize)).PageNum(int(search.PageNum))
			}

			if search.Count {
				err = sqlExec.Find(&rows).PageCount(&count).Err
			} else {
				err = sqlExec.Find(&rows).Err
			}
			if err != nil {
				return nil, count, err
			}
		} else {
			var cjob pb.Job
			// if job have region and name look cache first
			if job.Name != "" && job.Region != "" {
				err = a.memStore.Get(job.Region+"/"+job.Name, &cjob)
				if err != nil && err != errors.ErrNotExist {
					return nil, count, err
				}
				// hit cache
				if err == nil {
					log.FmdLoger.WithField("job", cjob).Debug("Agent: get job from cache")
					rows = append(rows, &cjob)
				}
			}
			// if miss cache, search in database
			if len(rows) == 0 {
				err = a.sqlStore.Model(&pb.Job{}).Where(job).Find(&rows).Err
				if err != nil {
					return nil, count, err
				}
			}
		}

		out := make([]interface{}, len(rows))

		for i, row := range rows {
			if row.ParentJobName != "" {
				pjob := pb.Job{Name: row.ParentJobName, Region: row.Region}
				if err = a.sqlStore.Model(&pb.Job{}).Where(&pjob).Find(&pjob).Err; err != nil {
					return nil, count, err
				}
				row.ParentJob = &pjob
				row.ParentJobName = ""
			}
			// set data cache into memory cache with double schedule time interval as cache ttl
			ttl, _ := util.GetTimeInterval(row.Schedule)
			if err = a.memStore.Set(row.Region+"/"+row.Name, row, ttl*2); err != nil {
				log.FmdLoger.WithError(err).Fatal("Agent: set memory cache failed")
			}
			out[i] = row
		}

		return out, count, err
	}
	return nil, 0, errors.ErrUnknownOps
}

// RunJob implement api interface
func (a *Agent) RunJob(name, region string) (*pb.Execution, error) {
	in := &pb.Job{Name: name, Region: region}
	// proxy job run action to the right region
	if region != a.config.Region {
		remoteServer, err := a.randomPickServer(region)
		if err != nil {
			log.FmdLoger.WithFields(logrus.Fields{
				"Region": region,
			}).WithError(err).Error("Agent: can not find server from the region")
			return nil, err
		}
		return a.remoteRunJob(in, remoteServer)
	}

	res, _, err := a.operationMiddleLayer(in, pb.Ops_READ, nil)
	if err != nil {
		log.FmdLoger.WithFields(logrus.Fields{
			"name":   name,
			"region": region,
		}).WithError(err).Error("Agent: RunJob try to get job info failed")
		return nil, err
	}
	var job *pb.Job
	var ok bool
	if job, ok = res[0].(*pb.Job); !ok {
		return nil, errors.ErrNotExpectation
	}

	// send job run action to job scheduler node
	if job.SchedulerNodeName != a.config.Nodename {
		return a.remoteRunJob(job, job.SchedulerNodeName)
	}

	return a.localRunJob(job)
}

func (a *Agent) remoteRunJob(job *pb.Job, remoteServer string) (*pb.Execution, error) {
	rsIp, rsPort, err := a.getRPCConfig(remoteServer)
	if err != nil {
		log.FmdLoger.WithFields(logrus.Fields{
			"Region":   job.Region,
			"nodeName": remoteServer,
		}).WithError(err).Error("Agent: RunJob get rpc config filed")
		return nil, err
	}
	rpcClient := a.newRPCClient(rsIp, rsPort)
	return rpcClient.ProxyJobRun(job.Name, job.Region)
}

func (a *Agent) localRunJob(job *pb.Job) (*pb.Execution, error) {
	ex := &pb.Execution{
		Name:              job.Name,
		Region:            job.Region,
		SchedulerNodeName: a.config.Nodename,
		Group:             time.Now().UnixNano(),
	}
	log.FmdLoger.WithFields(logrus.Fields{
		"name":   ex.Name,
		"region": ex.Region,
		"group":  ex.Group,
	}).Debug("Agent: Ready to perform the execution")
	go a.sendRunJobQuery(ex, job)
	return ex, nil
}
