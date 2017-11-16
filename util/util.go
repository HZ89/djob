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
	"database/sql"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"
	"strings"
	"unicode"

	"github.com/HZ89/djob/errors"
	"github.com/HZ89/djob/log"
	pb "github.com/HZ89/djob/message"
	"github.com/HZ89/djob/scheduler"
	"github.com/Knetic/govaluate"
	"github.com/Sirupsen/logrus"
	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
	"github.com/mattn/go-shellwords"
)

// Copy field from a struct to another.
// if have fieldNames just copy this field, if not copy all same name field
func CopyField(toValue, fromValue interface{}, fieldNames ...string) (err error) {
	var (
		isSlice      bool
		amount       = 1
		from         = Indirect(reflect.ValueOf(fromValue))
		to           = Indirect(reflect.ValueOf(toValue))
		fieldNameMap = make(map[string]bool)
	)
	for _, v := range fieldNames {
		fieldNameMap[v] = true
	}

	if !to.CanAddr() {
		return errors.ErrCopyToUnaddressable
	}

	// Return is from value is invalid
	if !from.IsValid() {
		return
	}

	// Just set it if possible to assign and no fieldNames
	if len(fieldNames) == 0 && from.Type().AssignableTo(to.Type()) {
		to.Set(from)
		return
	}

	fromType := IndirectType(from.Type())
	toType := IndirectType(to.Type())

	if fromType.Kind() != reflect.Struct || toType.Kind() != reflect.Struct {
		return
	}

	if to.Kind() == reflect.Slice {
		isSlice = true
		if from.Kind() == reflect.Slice {
			amount = from.Len()
		}
	}

	for i := 0; i < amount; i++ {
		var dest, source reflect.Value

		if isSlice {
			// source
			if from.Kind() == reflect.Slice {
				source = Indirect(from.Index(i))
			} else {
				source = Indirect(from)
			}

			// dest
			dest = Indirect(reflect.New(toType).Elem())
		} else {
			source = Indirect(from)
			dest = Indirect(to)
		}

		// Copy from field to field or method
		for _, field := range deepFields(fromType) {
			name := field.Name
			if len(fieldNames) != 0 {
				if _, ok := fieldNameMap[name]; !ok {
					continue
				}
			}

			if fromField := source.FieldByName(name); fromField.IsValid() {
				// has field
				if toField := dest.FieldByName(name); toField.IsValid() {
					if toField.CanSet() {
						if !set(toField, fromField) {
							if err := CopyField(toField.Addr().Interface(), fromField.Interface()); err != nil {
								return err
							}
						}
					}
				} else {
					// try to set to method
					var toMethod reflect.Value
					if dest.CanAddr() {
						toMethod = dest.Addr().MethodByName(name)
					} else {
						toMethod = dest.MethodByName(name)
					}

					if toMethod.IsValid() && toMethod.Type().NumIn() == 1 && fromField.Type().AssignableTo(toMethod.Type().In(0)) {
						toMethod.Call([]reflect.Value{fromField})
					}
				}
			}
		}

		// Copy from method to field
		for _, field := range deepFields(toType) {
			name := field.Name

			var fromMethod reflect.Value
			if source.CanAddr() {
				fromMethod = source.Addr().MethodByName(name)
			} else {
				fromMethod = source.MethodByName(name)
			}

			if fromMethod.IsValid() && fromMethod.Type().NumIn() == 0 && fromMethod.Type().NumOut() == 1 {
				if toField := dest.FieldByName(name); toField.IsValid() && toField.CanSet() {
					values := fromMethod.Call([]reflect.Value{})
					if len(values) >= 1 {
						set(toField, values[0])
					}
				}
			}
		}

		if isSlice {
			if dest.Addr().Type().AssignableTo(to.Type().Elem()) {
				to.Set(reflect.Append(to, dest.Addr()))
			} else if dest.Type().AssignableTo(to.Type().Elem()) {
				to.Set(reflect.Append(to, dest))
			}
		}
	}
	return
}

func deepFields(reflectType reflect.Type) []reflect.StructField {
	var fields []reflect.StructField

	if reflectType = IndirectType(reflectType); reflectType.Kind() == reflect.Struct {
		for i := 0; i < reflectType.NumField(); i++ {
			v := reflectType.Field(i)
			if v.Anonymous {
				fields = append(fields, deepFields(v.Type)...)
			} else {
				fields = append(fields, v)
			}
		}
	}

	return fields
}

func Indirect(reflectValue reflect.Value) reflect.Value {
	for reflectValue.Kind() == reflect.Ptr {
		reflectValue = reflectValue.Elem()
	}
	return reflectValue
}

func IndirectType(reflectType reflect.Type) reflect.Type {
	for reflectType.Kind() == reflect.Ptr || reflectType.Kind() == reflect.Slice {
		reflectType = reflectType.Elem()
	}
	return reflectType
}

func set(to, from reflect.Value) bool {
	if from.IsValid() {
		if to.Kind() == reflect.Ptr {
			//set `to` to nil if from is nil
			if from.Kind() == reflect.Ptr && from.IsNil() {
				to.Set(reflect.Zero(to.Type()))
				return true
			} else if to.IsNil() {
				to.Set(reflect.New(to.Type().Elem()))
			}
			to = to.Elem()
		}

		if from.Type().ConvertibleTo(to.Type()) {
			to.Set(from.Convert(to.Type()))
		} else if scanner, ok := to.Addr().Interface().(sql.Scanner); ok {
			scanner.Scan(from.Interface())
		} else if from.Kind() == reflect.Ptr {
			return set(to, from.Elem())
		} else {
			return false
		}
	}
	return true
}

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

	if job.Name == "" || job.Region == "" {
		return errors.ErrArgs
	}
	if job.Name == job.ParentJobName {
		return errors.ErrSameJob
	}

	if _, err := scheduler.Prepare(job.Schedule); err != nil {
		if job.ParentJobName == "" {
			return fmt.Errorf("%s: %s", errors.ErrScheduleParse.Error(), err)
		}
		log.FmdLoger.WithFields(logrus.Fields{
			"Name":          job.Name,
			"Region":        job.Region,
			"Schedule":      job.Schedule,
			"ParentJobName": job.ParentJobName,
		}).WithError(err).Warn("Agent: Job schedule no effort")
		job.Schedule = ""
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
	val := Indirect(reflect.ValueOf(obj))

	for i := 0; i < val.NumField(); i++ {
		if val.Type().Field(i).Name == name {
			return val.Field(i).Interface()
		}
	}
	return nil
}

func GetFieldKind(obj interface{}, name string) reflect.Kind {
	t := IndirectType(reflect.TypeOf(obj))

	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		if f.Name == name {
			return f.Type.Kind()
		}
	}
	return reflect.Invalid
}

func Intersect(a, b []string) []string {
	s := make([]string, 0)
	//	av := reflect.ValueOf(a)
	for i := 0; i < len(a); i++ {
		el := a[i]
		if Contains(b, el) {
			s = append(s, el)
		}
	}
	return s
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
