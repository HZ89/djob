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

package store

import (
	"fmt"
	"math"
	"reflect"
	"strconv"

	"github.com/fatih/structs"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"

	"github.com/HZ89/djob/errors"
	"github.com/HZ89/djob/log"
	"github.com/HZ89/djob/util"
)

const (
	sqlMaxOpenConnect = 100
	sqlMaxIdleConnect = 20
	defaultPageSize   = 10
)

// SQLStore database obj
type SQLStore struct {
	db           *gorm.DB      // gorm db obj
	sqlCondition *sqlCondition // sql where contition
	pageSize     int           // page size
	pageNum      int           // page num
	count        int64         // max page num
	Err          error         // error
}

// NewSQLStore init a SQLStore struct
func NewSQLStore(backend, dsn string) (*SQLStore, error) {

	db, err := gorm.Open(backend, dsn)
	if err != nil {
		return nil, err
	}

	db.DB().SetMaxOpenConns(sqlMaxOpenConnect)
	db.DB().SetMaxIdleConns(sqlMaxIdleConnect)

	if log.FmdLoger.Logger.Level.String() == "debug" {
		db.LogMode(true)
	}

	return &SQLStore{
		db: db,
	}, nil
}

func (s *SQLStore) clone() *SQLStore {
	store := SQLStore{db: s.db, sqlCondition: s.sqlCondition, pageSize: s.pageSize, pageNum: s.pageNum, Err: s.Err, count: s.count}
	return &store
}

// Close database connect
func (s *SQLStore) Close() error {
	return s.db.Close()
}

// Migrate auto migrate table struct
func (s *SQLStore) Migrate(drop bool, valus ...interface{}) error {
	if !drop {
		if err := s.db.AutoMigrate(valus...).Error; err != nil {
			return err
		}
		return nil
	}
	return nil
}

// Model set the model to be manipulated
func (s *SQLStore) Model(obj interface{}) *SQLStore {
	s.db = s.db.Model(obj)
	return s
}

// Where build where condition, can be a obj or SearchContion
func (s *SQLStore) Where(obj interface{}) *SQLStore {
	n := s.clone()
	switch t := obj.(type) {
	case *SearchCondition:
		condition, err := newSQLCondition(t, n.db.Value)
		if err != nil {
			n.Err = err
			return n
		}
		n.db = n.db.Where(condition.condition, condition.values...)
	default:
		n.db = n.db.Where(obj)
	}
	return n
}

// PageSize set page size
func (s *SQLStore) PageSize(i int) *SQLStore {
	n := s.clone()
	n.pageSize = i
	return n
}

// PageNum which page need scan into result set
func (s *SQLStore) PageNum(i int) *SQLStore {
	n := s.clone()
	if n.pageSize == 0 {
		n.pageSize = defaultPageSize
	}
	if i == 0 {
		i = 1
	}
	offSet := (i - 1) * n.pageSize
	limit := n.pageSize
	n.db = n.db.Limit(limit).Offset(offSet)
	return n
}

// PageCount func cal max page num
func (s *SQLStore) PageCount(out interface{}) *SQLStore {
	n := s.clone()
	n.Err = n.db.Offset(0).Count(&s.count).Error

	if n.Err != nil {
		return n
	}

	rf := math.Ceil(float64(s.count) / float64(s.pageSize))
	typ := util.IndirectType(reflect.TypeOf(out))
	val := util.Indirect(reflect.ValueOf(out))
	kind := typ.Kind()
	if int(kind) > 1 && int(kind) < 6 {
		val.SetInt(int64(rf))
		return n
	}
	n.Err = errors.ErrType
	return n
}

// Find func perform find in database
func (s *SQLStore) Find(out interface{}) *SQLStore {
	n := s.clone()
	if n.Err != nil {
		return n
	}
	e := n.db.Find(out).Error
	if e == gorm.ErrRecordNotFound {
		n.Err = errors.ErrNotExist
	} else {
		n.Err = e
	}
	return n
}

// Create func perform create in database
func (s *SQLStore) Create(obj interface{}) *SQLStore {
	n := s.clone()
	old := reflect.New(util.IndirectType(reflect.TypeOf(obj))).Interface()
	// should use primary key build where condition but just Job can be modify, so just copy Name and Region
	util.CopyField(old, obj, "Name", "Region")
	err := n.db.Where(old).First(old).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		n.Err = err
		return n
	}
	if err == gorm.ErrRecordNotFound {
		n.Err = n.db.Create(obj).Error
		return n
	}
	n.Err = errors.ErrRepetition
	return n
}

// Modify func perform modify
func (s *SQLStore) Modify(obj interface{}) *SQLStore {
	n := s.clone()
	old := reflect.New(util.IndirectType(reflect.TypeOf(obj))).Interface()
	//oldslice := reflect.SliceOf(util.IndirectType(reflect.TypeOf(obj)))
	// should use primary key build where condition but just Job can be modify, so just copy Name and Region
	util.CopyField(old, obj, "Name", "Region")
	err := n.db.Where(old).Find(old).Error
	if err == gorm.ErrRecordNotFound {
		n.Err = errors.ErrNotExist
		return n
	}
	if err != nil {
		n.Err = err
		return n
	}
	// gorm only update non blank value in struct fields, so use a map to update.
	m := structs.Map(obj)
	n.Err = n.db.Model(obj).Updates(m).Error
	return n
}

// Delete func perform delete
func (s *SQLStore) Delete(obj interface{}) *SQLStore {
	n := s.clone()
	out := reflect.New(util.IndirectType(reflect.TypeOf(obj))).Interface()
	err := n.db.Where(obj).First(out).Error
	if err != nil {
		n.Err = err
		return n
	}
	n.Err = n.db.Delete(out).Error
	if n.Err == nil {
		obj = out
	}
	return n
}

// SearchCondition save search condition
type SearchCondition struct {
	conditions []map[string]string // user input search condition,eg. a = 1 , b <> abc
	linkSymbol []logicSymbol       // user input link symbol, eg. AND
}

// NewSearchCondition serialized user input generated SearchCondition
func NewSearchCondition(conditions, links []string) (*SearchCondition, error) {

	if len(conditions) != len(links)+1 {
		return nil, errors.ErrLinkNum
	}

	regex := `^(?P<column>\w+)\s*(?P<symbol>=|>=|>|<=|<|<>|=)\s*(?P<value>\w+)$`
	s := new(SearchCondition)

	for _, condition := range conditions {
		if !util.RegexpMatch(regex, condition) {
			return nil, errors.ErrConditionFormat
		}
		s.conditions = append(s.conditions, util.RegexpGetParams(regex, condition))
	}

	for _, link := range links {
		if symbol, ok := StringToLogicSymbol(link); ok {
			s.linkSymbol = append(s.linkSymbol, symbol)
		} else {
			return nil, errors.ErrNotSupportSymbol
		}
	}
	return s, nil
}

type logicSymbol int

var logicSymbolNames = map[int]string{
	1: "OR",
	2: "AND",
	3: "=",
	4: ">=",
	5: ">",
	6: "<=",
	7: "<",
	8: "<>",
}

var logicSymbolValues = map[string]int{
	"OR":  1,
	"AND": 2,
	"=":   3,
	">=":  4,
	">":   5,
	"<=":  6,
	"<":   7,
	"<>":  8,
}

func (l logicSymbol) String() string {
	return logicSymbolNames[int(l)]
}

// StringToLogicSymbol used to conversion string to logicSymbol
func StringToLogicSymbol(s string) (logicSymbol, bool) {
	if v, ok := logicSymbolValues[s]; ok {
		return logicSymbol(v), true
	}
	return 0, false
}

type sqlCondition struct {
	condition string        // string of where condition eg. a = ? and b <> ?
	values    []interface{} // value of where condition eg. 1,"avc"
}

// use SearchCondition generated sql where condition
func newSQLCondition(search *SearchCondition, obj interface{}) (*sqlCondition, error) {
	s := new(sqlCondition)
	for n, condition := range search.conditions {
		s.condition += fmt.Sprintf("%s %s ?", condition["column"], condition["symbol"])
		fieldKind := util.GetFieldKind(obj, condition["column"])
		switch fieldKind {
		case reflect.Int32:
			v, err := strconv.ParseInt(condition["value"], 10, 32)
			if err != nil {
				return nil, errors.New(95236, err.Error())
			}
			s.values = append(s.values, v)
		case reflect.Int64:
			v, err := strconv.ParseInt(condition["value"], 10, 64)
			if err != nil {
				return nil, errors.New(95236, err.Error())
			}
			s.values = append(s.values, v)
		case reflect.String:
			s.values = append(s.values, condition["value"])
		case reflect.Bool:
			v, err := strconv.ParseBool(condition["value"])
			if err != nil {
				return nil, errors.New(95236, err.Error())
			}
			s.values = append(s.values, v)
		default:
			return nil, errors.New(95236, "not support feild type")
		}
		if n < len(search.linkSymbol) {
			s.condition += fmt.Sprintf(" %s ", search.linkSymbol[n])
		}
	}
	return s, nil
}
