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
	"github.com/Sirupsen/logrus"

	"version.uuzu.com/zhuhuipeng/djob/errors"
	"version.uuzu.com/zhuhuipeng/djob/log"
	pb "version.uuzu.com/zhuhuipeng/djob/message"
	"version.uuzu.com/zhuhuipeng/djob/store"
	"version.uuzu.com/zhuhuipeng/djob/util"
)

// separate local ops and remote ops
func (a *Agent) operationMiddleLayer(obj interface{}, ops pb.Ops, search *pb.Search) ([]interface{}, int, error) {
	objRegion := util.GetFieldValue(obj, "Region")
	objName := util.GetFieldValue(obj, "Name")
	if objRegion == nil || objName == nil {
		return nil, 0, errors.ErrType
	}

	if objRegion.(string) == a.config.Region {
		_, ok := obj.(*pb.Job)
		if a.lockerChain.HaveIt(objName.(string)) || !ok {
			return a.localOps(obj, ops, search)
		}
		owner := a.store.WhoLocked(obj, store.OWN)
		if owner == "" {
			var err error
			owner, err = a.minimalLoadServer(a.config.Region)
			if err != nil {
				return nil, 0, err
			}
		}
		return a.remoteOps(obj, ops, search, owner)
	}
	nextHandler, err := a.randomPickServer(objRegion.(string))
	if err != nil {
		return nil, 0, err
	}
	return a.remoteOps(obj, ops, search, nextHandler)
}

// forward ops to remote region by grpc
func (a *Agent) remoteOps(obj interface{}, ops pb.Ops, search *pb.Search, nodeName string) ([]interface{}, int, error) {
	ip, port, err := a.sendGetRPCConfigQuery(nodeName)
	if err != nil {
		return nil, 0, err
	}
	log.Loger.WithFields(logrus.Fields{
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
func (a *Agent) localOps(obj interface{}, ops pb.Ops, search *pb.Search) ([]interface{}, int, error) {
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

func (a *Agent) handleJobStatusOps(status *pb.JobStatus, ops pb.Ops) (out []interface{}, count int, err error) {
	switch {
	case ops == pb.Ops_MODIFY || ops == pb.Ops_ADD:
		if out[0], err = a.store.SetJobStatus(status); err != nil {
			return
		}
		return
	case ops == pb.Ops_READ:
		out[0], err = a.store.GetJobStatus(status)
		if err != nil {
			return
		}
		return
	case ops == pb.Ops_DELETE:
		out[0], err = a.store.DeleteJobStatus(status)
		if err != nil {
			return
		}
	}
	return nil, 0, errors.ErrUnknownOps
}

func (a *Agent) handleExecutionOps(ex *pb.Execution, ops pb.Ops, search *pb.Search) ([]interface{}, int, error) {
	var count int
	var err error
	switch ops {
	case pb.Ops_READ:
		var rows []*pb.Execution

		if search != nil {
			var condition *store.SearchCondition
			condition, err = store.NewSearchCondition(search.Conditions, search.Links)
			if err != nil {
				return nil, count, err
			}
			if search.Count {
				err = a.sqlStore.Model(&pb.Execution{}).Where(condition).PageSize(int(search.PageSize)).PageNum(int(search.PageNum)).Find(rows).PageCount(count).Err
			} else {
				err = a.sqlStore.Model(&pb.Execution{}).Where(condition).PageSize(int(search.PageSize)).PageNum(int(search.PageNum)).Find(rows).Err
			}
			if err != nil {
				return nil, count, err
			}
		} else {
			err = a.sqlStore.Model(&pb.Execution{}).Where(ex).Find(rows).Err
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
		return out, count, nil
	}
	return nil, 0, errors.ErrUnknownOps
}

func (a *Agent) handleJobOps(job *pb.Job, ops pb.Ops, search *pb.Search) ([]interface{}, int, error) {
	var err error
	var count int
	switch {
	case ops == pb.Ops_ADD || ops == pb.Ops_MODIFY:
		if err = a.sqlStore.Create(job).Err; err != nil {
			return nil, count, err
		}
		if err = a.scheduler.AddJob(job); err != nil {
			return nil, count, err
		}
		out := make([]interface{}, 1)
		out[0] = job
		return out, count, err
	case ops == pb.Ops_DELETE:
		a.scheduler.DeleteJob(job)
		if err = a.sqlStore.Delete(job).Err; err != nil {
			return nil, count, err
		}
		out := make([]interface{}, 1)
		out[0] = job
		return out, count, err
	case ops == pb.Ops_READ:
		var rows []*pb.Job

		if search != nil {
			var condition *store.SearchCondition
			condition, err = store.NewSearchCondition(search.Conditions, search.Links)
			if err != nil {
				return nil, count, err
			}
			if search.Count {
				err = a.sqlStore.Model(&pb.Job{}).Where(condition).PageSize(int(search.PageSize)).PageNum(int(search.PageNum)).Find(rows).PageCount(count).Err
			} else {
				err = a.sqlStore.Model(&pb.Job{}).Where(condition).PageSize(int(search.PageSize)).PageNum(int(search.PageNum)).Find(rows).Err
			}
			if err != nil {
				return nil, count, err
			}
		} else {
			err = a.sqlStore.Model(&pb.Job{}).Where(job).Find(rows).Err
			if err != nil {
				return nil, count, err
			}
		}

		out := make([]interface{}, len(rows))

		for i, row := range rows {
			out[i] = row
		}

		return out, count, err
	}
	return nil, 0, errors.ErrUnknownOps
}
