package djob

import (
	"github.com/armon/circbuf"
	"github.com/mattn/go-shellwords"
	"github.com/prometheus/common/log"
	"os/exec"
	"time"
	pb "version.uuzu.com/zhuhuipeng/djob/message"
)

const (
	maxBufSize = 1048576
	maxRunTime = 1 * time.Hour
)

func (a *Agent) execJob(job *pb.Job, ex *pb.Execution) error {
	buf, _ := circbuf.NewBuffer(maxBufSize)
	cmd := buildCmd(job)
	cmd.Stderr = buf
	cmd.Stdout = buf

	var success bool
	ex.StartTime = time.Now().UnixNano()
	err := cmd.Start()
	if err != nil {
		success = false
	}

	if buf.TotalWritten() > buf.Size() {
		Log.Warnf("Proc: Job '%s' generated %d bytes of output, truncated to %d", job.Name, buf.TotalWritten(), buf.Size())
	}

	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()

	select {
	case <-time.After(maxRunTime):
		Log.Warnf("Proc: Job '%s' reach max run time(one hour), will be kill it", job.Name)
		if err := cmd.Process.Kill(); err != nil {
			Log.WithError(err).Errorf("Proc: Job '%s' kill failed", job.Name)
		}
		log.Warnf("Proc: Job '%s' reach max run time(one hour), has been killed", job.Name)
	case err := <-done:
		if err != nil {
			Log.WithError(err).Errorf("Proc: Job '%s' cmd exec error output", job.Name)
			success = false
		} else {
			success = true
		}
	}

	ex.FinishTime = time.Now().UnixNano()
	ex.Succeed = success
	ex.Output = buf.Bytes()
	ex.RunNodeName = a.config.Nodename

	ip, port, err := a.sendGetRPCConfigQuery("")
	if err != nil {
		return err
	}

	rpcClient := a.newRPCClient(ip, port)
	defer rpcClient.Shutdown()

	if err := rpcClient.ExecDone(ex); err != nil {
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
