package djob

import (
	"github.com/hashicorp/serf/serf"
	"github.com/Sirupsen/logrus"
	"github.com/Knetic/govaluate"
	"errors"
)
var ErrCanNotFoundNode = errors.New("could not found any node can use")

func (a *Agent)createSerfQueryParam(expression string) (*serf.QueryParam, error){
	var queryParam serf.QueryParam
	exp, err := govaluate.NewEvaluableExpression(expression)
	if err != nil {
		return nil, err
	}
	parameters := make(map[string]interface{})

	suspected := make(map[string]map[string]string)
	for _, v := range exp.Vars(){
		for mk, mv := range a.membercache[v] {
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
		if result {
			foundServerName = append(foundServerName, sk)
		}
	}
	if len(foundServerName) > 0 {
		queryParam = serf.QueryParam{
			FilterNodes:foundServerName,
			RequestAck:true,
		}
		return &queryParam, nil
	} else {
		return nil, ErrCanNotFoundNode
	}
}

func (a *Agent)handleMemberCache(eventType serf.EventType, members []serf.Member) {

	a.mutex.Lock()
	defer a.mutex.Unlock()

	for _, member := range members{
		for tk, tv := range member.Tags {
			if _, exits := a.membercache[tk]; exits{
				if _, exits := a.membercache[tk][member.Name]; exits {
					if eventType == serf.EventMemberJoin || eventType == serf.EventMemberUpdate {
						a.membercache[tk][member.Name] = tv

						if eventType == serf.EventMemberJoin {
							Log.Warn("agent: get a new member event, but already have it's cache")
							Log.WithFields(logrus.Fields{
								"memberName": member.Name,
								"tags": member.Tags,
							}).Debug("agent: get a new member event, but already have it's cache")
						}

					}
					if eventType == serf.EventMemberReap || eventType == serf.EventMemberFailed || eventType == serf.EventMemberLeave {
						delete(a.membercache[tk], member.Name)
						Log.Infof("agent: delte member %s tags %s from cache", member.Name, tk)
					}
					// remove no value key
					if len(a.membercache[tk]) == 0 {
						delete(a.membercache, tk)
					}
				} else {
					if eventType == serf.EventMemberJoin || eventType == serf.EventMemberUpdate {
						a.membercache[tk][member.Name] = tv
					} else {
						Log.Warn("agent: get a member delete event, but have not cache")
					}
				}
			} else {
				if eventType == serf.EventMemberUpdate || eventType == serf.EventMemberJoin {
					a.membercache[tk][member.Name] = tv
				} else {
					Log.Warn("agent: get a member delete event, but have not cache")
				}
			}
		}
	}
}
