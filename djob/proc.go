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
	"math/rand"
	"os/exec"
	"time"

	"github.com/HZ89/djob/log"
	pb "github.com/HZ89/djob/message"
	"github.com/armon/circbuf"
	"github.com/mattn/go-shellwords"
)

const (
	maxBufSize = 1048576
	maxRunTime = 1 * time.Hour
	maxLayTime = 300
)

// execute the command
func (a *Agent) execJob(job *pb.Job, ex *pb.Execution) error {
	buf, _ := circbuf.NewBuffer(maxBufSize)
	cmd := buildCmd(job)
	cmd.Stderr = buf
	cmd.Stdout = buf

	// Random sleep for a while to prevent excessive concurrency
	time.Sleep(time.Duration(rand.Intn(maxLayTime)) * time.Millisecond)

	var success bool
	ex.StartTime = time.Now().UnixNano()
	err := cmd.Start()
	if err != nil {
		success = false
	}

	if buf.TotalWritten() > buf.Size() {
		log.FmdLoger.Warnf("Proc: Job '%s' generated %d bytes of output, truncated to %d", job.Name, buf.TotalWritten(), buf.Size())
	}

	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()

	TimeLimit := maxRunTime

	userDefinedTimeLimit := time.Duration(job.MaxRunTime) * time.Second
	if userDefinedTimeLimit != 0 && userDefinedTimeLimit < TimeLimit {
		TimeLimit = userDefinedTimeLimit
	}

	select {
	case <-time.After(TimeLimit):
		log.FmdLoger.Warnf("Proc: Job '%s' reach max run time(one hour), will be kill it", job.Name)
		if err = cmd.Process.Kill(); err != nil {
			log.FmdLoger.WithError(err).Errorf("Proc: Job '%s' kill failed", job.Name)
		}
		log.FmdLoger.Warnf("Proc: Job '%s' reach max run time(one hour), has been killed", job.Name)
	case err = <-done:
		if err != nil {
			log.FmdLoger.WithError(err).Errorf("Proc: Job '%s' cmd exec error output", job.Name)
			success = false
		} else {
			success = true
		}
	}

	ex.FinishTime = time.Now().UnixNano()
	ex.Succeed = success
	ex.Output = buf.Bytes()
	ex.RunNodeName = a.config.Nodename

	// randomly select a server to receive the execution
	serverNodeName, err := a.randomPickServer(ex.Region)
	if err != nil {
		return err
	}

	ip, port, err := a.getRPCConfig(serverNodeName)
	if err != nil {
		return err
	}

	rpcClient := a.newRPCClient(ip, port)
	defer rpcClient.Shutdown()

	log.FmdLoger.WithField("execution", ex).Debug("Proc: job done send back execution")
	// send back execution
	// TODO: retry this 3/5 times
	if err = rpcClient.ExecDone(ex); err != nil {
		log.FmdLoger.WithError(err).Debug("Proc: rpc call ExecDone failed")
		return err
	}

	return nil
}

func buildCmd(job *pb.Job) (cmd *exec.Cmd) {
	var shell, flag string
	if job.Shell {
		shell = "/bin/sh"
		flag = "-c"
		cmd = exec.Command(shell, flag, job.Command)
	} else {
		args, _ := shellwords.Parse(job.Command)
		cmd = exec.Command(args[0], args[1:]...)
	}

	return
}
