package message

import (
	"github.com/jinzhu/gorm"
	"github.com/segmentio/ksuid"
)

func (m *Job) BeforeCreate(scope *gorm.Scope) error {
	scope.SetColumn("Id", ksuid.New().String())
	return nil
}
