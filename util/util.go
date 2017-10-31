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

package util

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"
	"strings"
	"unicode"

	"github.com/Knetic/govaluate"
	"github.com/gogo/protobuf/jsonpb"
	"github.com/gogo/protobuf/proto"
	"github.com/mattn/go-shellwords"

	"encoding/gob"

	"version.uuzu.com/zhuhuipeng/djob/errors"
	pb "version.uuzu.com/zhuhuipeng/djob/message"
	"version.uuzu.com/zhuhuipeng/djob/scheduler"
)

/**
 * Parses url with the given regular expression and returns the
 * group values defined in the expression.
 *
 */
func RegexpGetParams(regEx, url string) (paramsMap map[string]string) {

	var compRegEx = regexp.MustCompile(regEx)
	match := compRegEx.FindStringSubmatch(url)

	paramsMap = make(map[string]string)
	for i, name := range compRegEx.SubexpNames() {
		if i > 0 && i <= len(match) {
			paramsMap[name] = match[i]
		}
	}
	return
}

func RegexpMatch(regEx, url string) bool {

	var compRegEx = regexp.MustCompile(regEx)
	return compRegEx.MatchString(url)
}

func VerifyJob(job *pb.Job) error {
	if job.Name == job.ParentJob.Name {
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

	if job.ParentJob == nil {
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

func GetBytes(data interface{}) ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(data)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func GetInterface(bts []byte, data interface{}) error {
	buf := bytes.NewBuffer(bts)
	dec := gob.NewDecoder(buf)
	err := dec.Decode(data)
	if err != nil {
		return err
	}
	return nil
}
