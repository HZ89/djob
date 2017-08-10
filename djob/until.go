package djob

import (
	"errors"

	"github.com/Knetic/govaluate"
	"github.com/Sirupsen/logrus"
	"github.com/hashicorp/serf/serf"
	pb "version.uuzu.com/zhuhuipeng/djob/message"
	"strings"
	"unicode"
)

var (
	ErrCanNotFoundNode = errors.New("could not found any node can use")
	ErrSameJob         = errors.New("This job set himself as his parent")
	ErrNoCmd           = errors.New("A job must have a Command")
	ErrNoReg           = errors.New("A job must have a region")
	ErrNoExp           = errors.New("A job must have a Expression")
)

func (a *Agent) createSerfQueryParam(expression string) (*serf.QueryParam, error) {
	var queryParam serf.QueryParam
	exp, err := govaluate.NewEvaluableExpression(expression)
	if err != nil {
		return nil, err
	}
	parameters := make(map[string]interface{})

	suspected := make(map[string]map[string]string)
	for _, v := range exp.Vars() {
		for mk, mv := range a.memberCache[v] {
			suspected[mk][v] = mv
		}
	}
	var foundServerName []string
	for sk, sv := range suspected {
		for _, v := range exp.Vars() {
			if tv, exits := sv[v]; exits {
				parameters[v] = tv
			}
		}
		result, err := exp.Evaluate(parameters)
		if err != nil {
			return nil, err
		}
		if result.(bool) {
			foundServerName = append(foundServerName, sk)
		}
	}
	if len(foundServerName) > 0 {
		queryParam = serf.QueryParam{
			FilterNodes: foundServerName,
			RequestAck:  true,
		}
		return &queryParam, nil
	}
	return nil, ErrCanNotFoundNode
}

func (a *Agent) handleMemberCache(eventType serf.EventType, members []serf.Member) {

	a.mutex.Lock()
	defer a.mutex.Unlock()

	for _, member := range members {
		for tk, tv := range member.Tags {
			if _, exits := a.memberCache[tk]; exits {
				if _, exits := a.memberCache[tk][member.Name]; exits {
					if eventType == serf.EventMemberJoin || eventType == serf.EventMemberUpdate {
						a.memberCache[tk][member.Name] = tv

						if eventType == serf.EventMemberJoin {
							Log.Warn("agent: get a new member event, but already have it's cache")
							Log.WithFields(logrus.Fields{
								"memberName": member.Name,
								"tags":       member.Tags,
							}).Debug("agent: get a new member event, but already have it's cache")
						}

					}
					if eventType == serf.EventMemberReap || eventType == serf.EventMemberFailed || eventType == serf.EventMemberLeave {
						delete(a.memberCache[tk], member.Name)
						Log.Infof("agent: delte member %s tags %s from cache", member.Name, tk)
					}
					// remove no value key
					if len(a.memberCache[tk]) == 0 {
						delete(a.memberCache, tk)
					}
				} else {
					if eventType == serf.EventMemberJoin || eventType == serf.EventMemberUpdate {
						a.memberCache[tk][member.Name] = tv
					} else {
						Log.Warn("agent: get a member delete event, but have not cache")
					}
				}
			} else {
				if eventType == serf.EventMemberUpdate || eventType == serf.EventMemberJoin {
					a.memberCache[tk][member.Name] = tv
				} else {
					Log.Warn("agent: get a member delete event, but have not cache")
				}
			}
		}
	}
}

func verifyJob(job *pb.Job) error {
	if job.Name == job.ParentJob {
		return ErrSameJob
	}

	if job.Command == "" {
		return ErrNoCmd
	}

	if job.Region == "" {
		return ErrNoReg
	}

	if job.Expression == "" {
		return ErrNoExp
	}

}

func generateSlug(str string) (slug string) {
	return strings.Map(func(r rune) rune {
		switch {
		case r == ' ', r == '-':
			return '-'
		case r == '_', unicode.IsLetter(r), unicode.IsDigit(r):
			return r
		default:
			return -1
		}
	}, strings.ToLower(strings.TrimSpace(str)))
}
