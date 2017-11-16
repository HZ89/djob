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

	"github.com/HZ89/djob/util"
	"github.com/hashicorp/serf/serf"

	"github.com/HZ89/djob/errors"
	"github.com/HZ89/djob/log"
	pb "github.com/HZ89/djob/message"
)

func (a *Agent) ListRegions() (regions []string, err error) {
	err = a.memStore.Get("regions_cache", &regions)
	if err != nil && err != errors.ErrNotExist {
		return
	}
	if len(regions) != 0 {
		log.FmdLoger.Debug("Agent: regions hit cache")
		return
	}

	regionList := make(map[string]bool)
	for _, m := range a.serf.Members() {
		if m.Status == serf.StatusAlive {
			regionList[m.Tags["region"]] = true
		}

	}
	for k := range regionList {
		regions = append(regions, k)
	}
	err = a.memStore.Set("regions_cache", regions, 120*time.Second)
	if err != nil {
		return
	}
	return
}

func (a *Agent) AddJob(in *pb.Job) (out *pb.Job, err error) {
	var res []interface{}
	res, _, err = a.operationMiddleLayer(in, pb.Ops_ADD, nil)
	if err != nil {
		return
	}
	log.FmdLoger.WithField("obj", res).Debug("Agent: API AddJob got this")
	if t, ok := res[0].(*pb.Job); ok {
		out = t
		return
	}
	log.FmdLoger.Fatalf("Agent: API AddJob get type %v is not exception", res)
	return nil, errors.ErrNotExpectation
}

func (a *Agent) ModifyJob(in *pb.Job) (out *pb.Job, err error) {
	var res []interface{}
	res, _, err = a.operationMiddleLayer(in, pb.Ops_MODIFY, nil)
	if err != nil {
		return
	}
	if t, ok := res[0].(*pb.Job); ok {
		out = t
		return
	}
	log.FmdLoger.Fatalf("Agent: API AddJob get type %v is not exception", res)
	return nil, errors.ErrNotExpectation
}

func (a *Agent) DeleteJob(in *pb.Job) (out *pb.Job, err error) {
	var res []interface{}
	res, _, err = a.operationMiddleLayer(in, pb.Ops_DELETE, nil)
	if err != nil {
		return
	}
	if t, ok := res[0].(*pb.Job); ok {
		out = t
		return
	}
	log.FmdLoger.Fatalf("Agent: API AddJob get type %v is not exception", res)
	return nil, errors.ErrNotExpectation
}

// TODO: use concurrency instead of recursion
func (a *Agent) ListJob(in *pb.Job, search *pb.Search) (jobs []*pb.Job, count int32, err error) {
	if in.Region == "" {

		regions, err := a.ListRegions()
		if err != nil {
			return nil, 0, err
		}
		tj := new(pb.Job)
		if err := util.CopyField(tj, in); err != nil {
			return nil, 0, err
		}
		for _, r := range regions {
			tj.Region = r
			// reading data across regions does not support paging
			res, _, err := a.ListJob(tj, nil)
			if err != nil {
				log.FmdLoger.WithField("region", r).Error("Agent: list job in this region failed")
				return nil, 0, err
			}
			jobs = append(jobs, res...)
		}
		return jobs, 0, nil
	}

	var res []interface{}
	res, count, err = a.operationMiddleLayer(in, pb.Ops_READ, search)
	if err != nil {
		return
	}
	for _, i := range res {
		if t, ok := i.(*pb.Job); ok {
			jobs = append(jobs, t)
			continue
		}
		log.FmdLoger.Fatalf("Agent: API AddJob get type %v is not exception", res)
	}
	return
}

func (a *Agent) GetStatus(name, region string) (out *pb.JobStatus, err error) {
	in := &pb.JobStatus{Name: name, Region: region}
	var res []interface{}
	res, _, err = a.operationMiddleLayer(in, pb.Ops_READ, nil)
	if err != nil {
		return nil, err
	}
	if t, ok := res[0].(*pb.JobStatus); ok {
		out = t
		return
	}
	log.FmdLoger.Fatalf("Agent: API AddJob get type %v is not exception", res)
	return nil, errors.ErrNotExpectation
}

func (a *Agent) ListExecutions(in *pb.Execution, search *pb.Search) (out []*pb.Execution, count int32, err error) {
	if in.Region == "" {
		regions, err := a.ListRegions()
		if err != nil {
			return nil, 0, err
		}
		te := new(pb.Execution)
		if err := util.CopyField(te, in); err != nil {
			return nil, 0, err
		}
		for _, r := range regions {
			te.Region = r
			// reading data across regions does not support paging
			res, _, err := a.ListExecutions(te, nil)
			if err != nil {
				log.FmdLoger.WithField("region", r).Error("Agent: list job in this region failed")
				return nil, 0, err
			}
			out = append(out, res...)
		}
		return out, 0, nil
	}
	var res []interface{}
	res, count, err = a.operationMiddleLayer(in, pb.Ops_READ, search)
	if err != nil {
		return
	}
	for _, i := range res {
		if t, ok := i.(*pb.Execution); ok {
			out = append(out, t)
			continue
		}
		log.FmdLoger.Fatalf("Agent: API AddJob get type %v is not exception", res)
	}
	return
}

func (a *Agent) Search(in interface{}, search *pb.Search) (out []interface{}, count int32, err error) {
	switch t := in.(type) {
	case *pb.Job:
		if t.Region == "" {
			regions, err := a.ListRegions()
			if err != nil {
				return nil, 0, err
			}
			// reading data across regions does not support paging
			search.PageSize = 0
			search.PageNum = 0
			search.Count = false
			for _, r := range regions {
				t.Region = r
				var res []interface{}
				res, _, err := a.Search(t, search)
				if err != nil {
					return nil, 0, err
				}
				out = append(out, res...)
			}
			return out, 0, nil
		}
	case *pb.Execution:
		if t.Region == "" {
			regions, err := a.ListRegions()
			if err != nil {
				return nil, 0, err
			}
			// reading data across regions does not support paging
			search.PageSize = 0
			search.PageNum = 0
			search.Count = false
			for _, r := range regions {
				t.Region = r
				var res []interface{}
				res, _, err := a.Search(t, search)
				if err != nil {
					return nil, 0, err
				}
				out = append(out, res...)
			}
			return out, 0, nil
		}
	}
	out, count, err = a.operationMiddleLayer(in, pb.Ops_READ, search)
	return
}
