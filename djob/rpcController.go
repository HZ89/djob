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
	"reflect"

	"github.com/HZ89/djob/errors"
	"github.com/HZ89/djob/log"
	pb "github.com/HZ89/djob/message"
	"github.com/Sirupsen/logrus"
)

// implementation of Operator interface

// SendBackExecution handle job exec return
func (a *Agent) SendBackExecution(ex *pb.Execution) (err error) {

	log.FmdLoger.WithFields(logrus.Fields{
		"JobName":     ex.Name,
		"Region":      ex.Region,
		"Group":       ex.Group,
		"RunNodeName": ex.RunNodeName,
	}).Debug("RPC: Save Execution to backend")

	_, _, err = a.operationMiddleLayer(ex, pb.Ops_ADD, nil)
	if err != nil {
		log.FmdLoger.WithError(err).WithFields(logrus.Fields{
			"JobName":     ex.Name,
			"Region":      ex.Region,
			"Group":       ex.Group,
			"RunNodeName": ex.RunNodeName,
		}).Error("RPC: Save Execution to backend failed")
		return
	}
	log.FmdLoger.Debug("RPC: Execution has been saved, start modfiy jobstatus")

	return nil
}

// GetJobAndToken get job data and issue a token
func (a *Agent) GetJobAndToken(token *pb.RequestJobAndToken) (*pb.ResponseJobAndToken, error) {
	out, _, err := a.operationMiddleLayer(&pb.Job{Name: token.JobName, Region: token.JobRegion}, pb.Ops_READ, nil)
	if err != nil {
		return nil, err
	}
	if len(out) == 0 {
		log.FmdLoger.WithFields(logrus.Fields{
			"Name":   token.JobName,
			"Region": token.JobRegion,
		}).Warn("RPC: GetJob the return nothing")
		return nil, errors.ErrNotExist
	}

	job := out[0].(*pb.Job)

	// confer a token
	confer := func(obj interface{}) error {
		status, ok := obj.(*pb.JobStatus)
		if !ok {
			return errors.ErrType
		}
		c, exist := status.RunningStatus[token.ExecutionGroup]
		if !exist {
			status.RunningStatus[token.ExecutionGroup] = &pb.RunningStatus{LeftToken: job.Concurrency}
		}
		if c.LeftToken > 1 {
			c.LeftToken -= 1
			c.RunningNode = append(c.RunningNode, token.NodeName)
			status.RunningStatus[token.ExecutionGroup] = c
			return errors.ErrNil
		}

		return errors.ErrNoMoreToken
	}

	for err != errors.ErrNil {
		err = a.store.AtomicSet(&pb.JobStatus{Name: token.JobName, Region: token.JobRegion}, confer)
		if err != errors.ErrNoMoreToken {
			log.FmdLoger.WithError(err).Fatal("Agent: confer token error")
		}
	}

	if t, ok := out[0].(*pb.Job); ok {
		return t, nil
	}
	log.FmdLoger.WithFields(logrus.Fields{
		"Name":   name,
		"Region": region,
	}).Fatalf("RPC: GetJob want a %v, but get %v", reflect.TypeOf(&pb.Job{}), reflect.TypeOf(out[0]))
	return nil, errors.ErrType
}

// PerformOps func forwarding ops to remote or perform it in local
func (a *Agent) PerformOps(obj interface{}, ops pb.Ops, search *pb.Search) ([]interface{}, int32, error) {
	log.FmdLoger.Debugf("RPC: Server got a %v, ops: %s, search: %v", obj, ops, search)
	return a.operationMiddleLayer(obj, ops, search)
}
