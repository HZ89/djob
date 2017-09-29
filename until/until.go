package until

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"unicode"

	"github.com/Knetic/govaluate"
	"github.com/gogo/protobuf/jsonpb"
	"github.com/gogo/protobuf/proto"
	"github.com/mattn/go-shellwords"

	"version.uuzu.com/zhuhuipeng/djob/errors"
	pb "version.uuzu.com/zhuhuipeng/djob/message"
	"version.uuzu.com/zhuhuipeng/djob/scheduler"
)

func VerifyJob(job *pb.Job) error {
	if job.Name == job.ParentJob {
		return errors.ErrSameJob
	}

	if job.Command == "" {
		return errors.ErrNoCmd
	}

	if job.Region == "" {
		return errors.ErrNoReg
	}

	if job.Expression == "" {
		return errors.ErrNoExp
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
			return fmt.Errorf("%s: %s", errors.ErrScheduleParse.Error(), err)
		}
	}

	return nil
}

func GenerateSlug(str string) (slug string) {
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

func GetType(obj interface{}) string {
	if t := reflect.TypeOf(obj); t.Kind() == reflect.Ptr {
		return t.Elem().Name()
	} else {
		return t.Name()
	}
}

func GetFieldValue(obj interface{}, name string) interface{} {
	val := reflect.ValueOf(obj)
	for i := 0; i < val.NumField(); i++ {
		if val.Type().Field(i).Name == name {
			return val.Field(i).Interface()
		}
	}
	return nil
}

func GetFieldType(obj interface{}, name string) string {
	t := reflect.TypeOf(obj)
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		if f.Name == name {
			return f.Type.Name()
		}
	}
	return ""
}

func Intersect(a, b []string) []string {
	set := make([]string, 0)
	//	av := reflect.ValueOf(a)
	for i := 0; i < len(a); i++ {
		el := a[i]
		if Contains(b, el) {
			set = append(set, el)
		}
	}
	return set
}

func Contains(a interface{}, e interface{}) bool {
	v := reflect.ValueOf(a)

	for i := 0; i < v.Len(); i++ {
		if v.Index(i).Interface() == e {
			return true
		}
	}
	return false
}

func PbToJSON(obj interface{}) ([]byte, error) {
	msg, ok := obj.(proto.Message)
	if !ok {
		return nil, errors.ErrType
	}
	var buf bytes.Buffer
	marshaler := &jsonpb.Marshaler{EmitDefaults: true}
	if err := marshaler.Marshal(&buf, msg); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func JsonToPb(b []byte, obj interface{}) error {
	jpb, ok := obj.(proto.Message)
	if !ok {
		return errors.ErrType
	}
	jsondec := json.NewDecoder(bytes.NewReader(b))
	unmarshaler := &jsonpb.Unmarshaler{AllowUnknownFields: false}
	if err := unmarshaler.UnmarshalNext(jsondec, jpb); err != nil {
		return err
	}
	return nil
}

type FanOutQueue struct {
	outQueues []chan struct{}
}

func (f *FanOutQueue) FanOut() {
	for _, ch := range f.outQueues {
		ch <- struct{}{}
	}
}

func (f *FanOutQueue) Sub() chan struct{} {
	ch := make(chan struct{})
	f.outQueues = append(f.outQueues, ch)
	return ch
}