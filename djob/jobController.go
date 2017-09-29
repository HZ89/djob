package djob

import (
	"github.com/Sirupsen/logrus"

	"version.uuzu.com/zhuhuipeng/djob/errors"
	"version.uuzu.com/zhuhuipeng/djob/log"
	pb "version.uuzu.com/zhuhuipeng/djob/message"
	"version.uuzu.com/zhuhuipeng/djob/store"
)

func (a *Agent) jobExist(job *pb.Job) bool {
	if a.store.IsLocked(job, store.OWN) {
		return true
	}

	job, err := a.store.GetJob(job.Name, job.Region)
	if err != nil && err != errors.ErrNotFound {
		log.Loger.WithFields(
			logrus.Fields{
				"Name":   job.Name,
				"Region": job.Region,
			},
		).WithError(err).Error("Agent: get job failed")
	}
	if job != nil {
		return true
	}
	return false
}

func (a *Agent) jobModify(job *pb.Job) (*pb.Job, error) {

}
