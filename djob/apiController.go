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
	"math"
	"sort"
	"strings"
	"time"

	"github.com/HZ89/djob/errors"
	"github.com/HZ89/djob/log"
	pb "github.com/HZ89/djob/message"
	"github.com/HZ89/djob/util"
	"github.com/hashicorp/serf/serf"
)

// ListNode get node info, support paging
// TODO: Need more elegant realization
func (a *Agent) ListNode(node *pb.Node, search *pb.Search) (nodes []*pb.Node, count int32, err error) {
	var expression string
	var res []*pb.Node
	var members []serf.Member
	if node == nil {
		members = a.serf.Members()
	} else {
		if node.Name != "" {
			expression += fmt.Sprintf("NAME == '%s' && ", node.Name)
		}
		if node.Region != "" {
			expression += fmt.Sprintf("REGION == '%s' && ", node.Region)
		}
		if node.Role == "server" {
			expression += fmt.Sprintf("SERVER == 'true' && ")
		}
		if node.Role == "agent" {
			expression += fmt.Sprintf("SERVER == 'false' && ")
		}
		for k, v := range node.Tags {
			expression += fmt.Sprintf("%s == '%s' && ", k, v)
		}
		expression = strings.TrimRight(expression, "&& ")
		log.FmdLoger.Debugf("Agent: listNode use expression %s", expression)
		members, err = a.processFilteredNodes(expression, node.Alived)
		if err != nil {
			return nil, 0, err
		}
	}

	for _, m := range members {
		n := &pb.Node{Name: m.Name, Region: m.Tags["REGION"], Version: m.Tags["VERSION"]}
		n.Tags = make(map[string]string)
		for k, v := range m.Tags {
			var found bool
			// pass reserved tags
			for _, i := range RESERVEDTAGS {
				if k == i {
					found = true
				}
			}
			if !found {
				n.Tags[k] = v
			}
		}

		if m.Tags["SERVER"] == "true" {
			n.Role = "server"
		} else {
			n.Role = "agent"
		}

		if m.Status == serf.StatusAlive {
			n.Alived = true
		}
		res = append(res, n)
	}

	nodes = res
	// paging
	var pageNum int32 = 1
	var pageSize int32 = 10

	if search != nil {
		sort.Sort(sortNodes(res))
		if search.PageSize != 0 {
			pageSize = search.PageSize
		}
		if search.PageNum != 0 {
			pageNum = search.PageNum
		}
		end := pageNum * pageSize
		start := end - pageSize

		if start > int32(len(res)) {
			return nil, 0, errors.ErrNotExist
		}

		if end > int32(len(res)) {
			nodes = res[start:]
		} else {
			nodes = res[start:end]
		}

		if search.Count {
			count = int32(math.Ceil(float64(len(res)) / float64(pageSize)))
		}
	}
	return
}

type sortNodes []*pb.Node

func (s sortNodes) Len() int {
	return len(s)
}

func (s sortNodes) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s sortNodes) Less(i, j int) bool {
	if s[i].Region != s[j].Region {
		return s[i].Region < s[j].Region
	}
	return s[i].Name < s[j].Name
}

// ListRegions list all available regions
// search in memory cache first, if no cache, set it with 120s ttl
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
			regionList[m.Tags["REGION"]] = true
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

// AddJob add job
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

// ModifyJob func modify job
func (a *Agent) ModifyJob(in *pb.Job) (out *pb.Job, err error) {

	old, _, err := a.operationMiddleLayer(&pb.Job{Name: in.Name, Region: in.Region}, pb.Ops_READ, nil)
	if err != nil {
		return nil, err
	}
	if len(old) != 1 {
		return nil, errors.ErrNotExist
	}

	in.SchedulerNodeName = old[0].(*pb.Job).SchedulerNodeName

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

// delete job
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

// ListJob find a job
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
				log.FmdLoger.WithField("REGION", r).Error("Agent: list job in this region failed")
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

// GetStatus get job status from kv store
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

// ListExecutions get job executions
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
				log.FmdLoger.WithField("REGION", r).Error("Agent: list job in this region failed")
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

// Search job or execution
// TODO: use concurrency instead of recursion
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
