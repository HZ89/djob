package djob

import (
	"github.com/Sirupsen/logrus"
	"github.com/hashicorp/memberlist"
	"github.com/hashicorp/serf/serf"
	"local/djob/rpc"
	"time"
)

type Agent struct {
	ShutdownCh <-chan struct{}

	schedulerCh chan string
	config    *Config
	serf      *serf.Serf
	eventCh   chan serf.Event
	ready     bool
	rpcServer *rpc.RpcServer
}

func (a *Agent) setupSerf() *serf.Serf {
	encryptKey, err := a.config.EncryptKey()
	if err != nil {
		Log.Fatal(err)
		return nil
	}

	serfConfig := serf.DefaultConfig()

	//noinspection GoBinaryAndUnaryExpressionTypesCompatibility
	serfConfig.MemberlistConfig = memberlist.DefaultWANConfig()

	serfConfig.MemberlistConfig.BindAddr = a.config.SerfBindIp
	serfConfig.MemberlistConfig.BindPort = a.config.SerfBindPort
	serfConfig.MemberlistConfig.AdvertiseAddr = a.config.SerfAdvertiseIp
	serfConfig.MemberlistConfig.AdvertisePort = a.config.SerfAdvertisePort
	serfConfig.MemberlistConfig.SecretKey = encryptKey
	serfConfig.NodeName = a.config.Nodename
	serfConfig.Tags = a.config.Tags
	serfConfig.SnapshotPath = a.config.SerfSnapshotPath
	serfConfig.CoalescePeriod = 3 * time.Second
	serfConfig.QuiescentPeriod = time.Second
	serfConfig.UserCoalescePeriod = 3 * time.Second
	serfConfig.UserQuiescentPeriod = time.Second
	serfConfig.EnableNameConflictResolution = true

	a.eventCh = make(chan serf.Event, 64)
	serfConfig.EventCh = a.eventCh

	Log.Info("agent: Djob agent starting")

	serf, err := serf.Create(serfConfig)
	if err != nil {
		Log.Fatal(err)
		return nil
	}
	return serf
}

func (a *Agent) eventLoop() {
	serfShutdownCh := a.serf.ShutdownCh()
	Log.Info("agent: Listen for Serf event")
	for {
		select {
		case e := <-a.eventCh:
			Log.WithFields(logrus.Fields{
				"event": e.String(),
			}).Debug("agent: Received event")

			if failed, ok := e.(serf.MemberEvent); ok {
				for _, member := range failed.Members {
					Log.WithFields(logrus.Fields{
						"node":   a.config.Nodename,
						"member": member.Name,
						"event":  e.EventType(),
					}).Debug("agent: Member event")
				}
			}

			// handle custom query event
			if e.EventType() == serf.EventQuery {
				query := e.(*serf.Query)

				switch qname := query.Name; qname {
				case QueryNewJob:
					if a.config.Server {
						Log.WithFields(logrus.Fields{
							"query":   query.Name,
							"payload": string(query.Payload),
							"at":      query.LTime,
						}).Debug("agent: Server receive a add new job event")

						err := a.receiveNewJobQuery(query)

						if err != nil {
							Log.WithFields(logrus.Fields{
								"query":   query.Name,
								"payload": string(query.Payload),
								"error":   err.Error(),
							}).Error("agent: Server add new job failed")
						}
					} else {
						continue
					}
				case QueryRunJob:
					Log.WithFields(logrus.Fields{
						"query":   query.Name,
						"payload": string(query.Payload),
						"at":      query.LTime,
					}).Debug("agent: Running job")
					continue
				case QueryRPCConfig:
					if a.config.Server {
						Log.WithFields(logrus.Fields{
							"query":   query.Name,
							"payload": string(query.Payload),
							"at":      query.LTime,
						}).Debug("agent: Server receive a rpc config query")
					} else {
						continue
					}
				default:
					Log.Warn("agent: get a unknow message")
					Log.WithFields(logrus.Fields{
						"query":   query.Name,
						"payload": string(query.Payload),
						"at":      query.LTime,
					}).Debug("agent: get a unknow message")
				}
			}

		case <-serfShutdownCh:
			Log.Warn("agent: Serf shutdown detected, quitting")
			return
		}
	}
}
