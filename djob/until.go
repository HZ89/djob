package djob

import (
	"fmt"
	"reflect"
	"sort"
	"strings"
	"unicode"

	"github.com/Knetic/govaluate"
	"github.com/gogo/protobuf/proto"
	"github.com/hashicorp/serf/serf"
	"github.com/mattn/go-shellwords"
	pb "version.uuzu.com/zhuhuipeng/djob/message"
	"version.uuzu.com/zhuhuipeng/djob/scheduler"
)

func (a *Agent) createSerfQueryParam(expression string) (*serf.QueryParam, error) {
	var queryParam serf.QueryParam
	nodeNames, err := a.processFilteredNodes(expression)
	if err != nil {
		return nil, err
	}

	if len(nodeNames) > 0 {
		queryParam = serf.QueryParam{
			FilterNodes: nodeNames,
			RequestAck:  true,
		}
		return &queryParam, nil
	}
	return nil, ErrCanNotFoundNode
}

func (a *Agent) processFilteredNodes(expression string) ([]string, error) {
	exp, err := govaluate.NewEvaluableExpression(expression)
	if err != nil {
		return nil, err
	}
	wt := exp.Vars()
	sort.Sort(sort.StringSlice(wt))
	var nodeNames []string
	for _, member := range a.serf.Members() {
		if member.Status == serf.StatusAlive {
			var mtks []string
			for k := range member.Tags {
				mtks = append(mtks, k)
			}
			intersection := intersect([]string(mtks), wt)
			sort.Sort(sort.StringSlice(intersection))
			if reflect.DeepEqual(wt, intersection) {
				parameters := make(map[string]interface{})
				for _, tk := range wt {
					parameters[tk] = member.Tags[tk]
				}
				result, err := exp.Evaluate(parameters)
				if err != nil {
					return nil, err
				}
				if result.(bool) {
					nodeNames = append(nodeNames, member.Name)
				}
			}
		}
	}
	return nodeNames, nil
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

	if _, err := govaluate.NewEvaluableExpression(job.Expression); err != nil {
		return err
	}

	if !job.Shell {
		_, err := shellwords.Parse(job.Command)
		if err != nil {
			return err
		}
	}

	if job.ParentJob == "" {
		if _, err := scheduler.Prepare(job.Schedule); err != nil {
			return fmt.Errorf("%s: %s", ErrScheduleParse.Error(), err)
		}
	}

	return nil
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

func genrateResultPb(status int, message string) ([]byte, error) {
	r := pb.Result{
		Status:  int32(status),
		Message: message,
	}
	pb, err := proto.Marshal(&r)
	if err != nil {
		return nil, err
	}
	return pb, nil
}

func getType(obj interface{}) string {
	if t := reflect.TypeOf(obj); t.Kind() == reflect.Ptr {
		return t.Elem().Name()
	} else {
		return t.Name()
	}
}

func intersect(a, b []string) []string {
	set := make([]string, 0)
	//	av := reflect.ValueOf(a)
	for i := 0; i < len(a); i++ {
		el := a[i]
		if contains(b, el) {
			set = append(set, el)
		}
	}
	return set
}

func contains(a interface{}, e interface{}) bool {
	v := reflect.ValueOf(a)

	for i := 0; i < v.Len(); i++ {
		if v.Index(i).Interface() == e {
			return true
		}
	}
	return false
}
