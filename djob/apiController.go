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
	pb "version.uuzu.com/zhuhuipeng/djob/message"
)

func (a *Agent) ListRegions() (regions []string, err error) {
	return
}

func (a *Agent) AddJob(in *pb.Job) (out *pb.Job, err error) {
	return
}

func (a *Agent) ModifyJob(in *pb.Job) (out *pb.Job, err error) {
	return
}

func (a *Agent) DeleteJob(in *pb.Job) (out *pb.Job, err error) {
	return
}

func (a *Agent) ListJob(name, region string) (jobs []*pb.Job, err error) {
	return
}

func (a *Agent) RunJob(name, region string) (exec *pb.Execution, err error) {
	return
}

func (a *Agent) GetStatus(name, region string) (out *pb.JobStatus, err error) {
	return
}

func (a *Agent) ListExecutions(name, region string, group int64) (out []*pb.Execution, err error) {
	return
}

func (a *Agent) Search(in interface{}, search *pb.Search) (out []interface{}, count int32, err error) {
	return
}
