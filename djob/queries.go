package djob

import "github.com/hashicorp/serf/serf"

const (
	QueryNewJob    = "job:new"
	QueryRunJob    = "job:run"
	QueryRPCConfig = "rpc:config"
)

func (a *Agent) SendNewJobQuery(jobName string) {
	var params *serf.QueryParam
	filtertags := make(map[string]string)
	filtertags["server"] = "true"
	filtertags["region"]
}

func (a *Agent) receiveNewJobQuery(query *serf.Query) error {
	return nil
}
