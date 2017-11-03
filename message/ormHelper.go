package message

import (
	"github.com/jinzhu/gorm"
	"github.com/segmentio/ksuid"
	"version.uuzu.com/zhuhuipeng/djob/errors"
)

func (m *Job) BeforeCreate(scope *gorm.Scope) error {
	scope.SetColumn("Id", ksuid.New().String())
	if m.ParentJob != nil {
		if m.ParentJob.Id == "" {
			return errors.ErrNoParentJobId
		}
		scope.SetColumn("ParentJob", m.ParentJob.Id)
	}
	return nil
}
