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
	"strings"
	"time"

	"github.com/HZ89/djob/errors"
	"github.com/HZ89/djob/log"
	"github.com/HZ89/djob/util"
)

const separator = "###"

// MemStore used to cache job in local memory
type MemStore struct {
	memBuf *memDbBuffer // leveldb memory buffer
	stopCh chan struct{}
}

// base struct save in MemStore
type entry struct {
	DeadTime time.Time // the dead time of this entry
	Data     []byte    // user save data
}

func (m *MemStore) Delete(key string) error {
	return m.memBuf.Delete(Key(key))
}

func (m *MemStore) Set(key string, in interface{}, ttl time.Duration) (err error) {
	var value []byte
	e := &entry{}
	if ttl != 0 {
		e.DeadTime = time.Now().Add(ttl)
	} else {
		e.DeadTime = time.Time{}
	}
	e.Data, err = util.GetBytes(in)
	if err != nil {
		return
	}
	value, err = util.GetBytes(e)
	if err != nil {
		return
	}
	return m.memBuf.Set(Key(key), value)
}

func (m *MemStore) Get(key string, out interface{}) (err error) {
	var e *entry
	now := time.Now()
	e, err = m.getEntry(Key(key))
	if err != nil {
		return err
	}

	if !m.isOverdue(e, now) {
		return util.GetInterface(e.Data, out)
	}
	return errors.ErrNotExist
}

func (m *MemStore) getEntry(key Key) (*entry, error) {
	var (
		v   []byte
		err error
		e   entry
	)
	v, err = m.memBuf.Get(key)
	if err != nil {
		return nil, err
	}
	if len(v) == 0 {
		return nil, errors.ErrNotExist
	}
	err = util.GetInterface(v, &e)
	if err != nil {
		return nil, err
	}
	return &e, nil
}

// timeing seek and clear expired key
func (m *MemStore) handleOverdueKey() {
	go func() {
		for true {
			select {
			case <-time.After(time.Second):
				m.releaseOverdueKey()
			case <-m.stopCh:
				return
			}
		}
	}()
}

// seek and clear
func (m *MemStore) releaseOverdueKey() {
	now := time.Now()
	iter, err := m.memBuf.Seek(nil)
	if err != nil {
		log.FmdLoger.Fatalf("Store: memBuf Seek failed: %v", err)
	}
	defer iter.Close()
	// seek
	for ; iter.Valid(); iter.Next() {
		key := iter.Key()
		var e *entry
		e, err = m.getEntry(key)
		if err != nil && err != errors.ErrNotExist {
			log.FmdLoger.WithError(err).Fatalf("Store: memBuf Seek get key filed: key: %s, value: %v", string(key), e)
		}
		// clear
		if e != nil && m.isOverdue(e, now) {
			if err = m.memBuf.Delete(key); err != nil {
				log.FmdLoger.WithField("key", string(key)).WithError(err).Fatal("Store: purge overdue key failed")
			}
		}
	}
}

func (m *MemStore) isOverdue(e *entry, now time.Time) bool {
	if e.DeadTime.IsZero() {
		return false
	}
	return e.DeadTime.Before(now)
}

func (m *MemStore) buildKey(name string, ttl time.Duration) (Key, error) {
	a := strings.Split(name, separator)
	if len(a) > 1 {
		return nil, errors.ErrIllegalCharacter
	}
	if ttl == 0 {
		return Key(name), nil
	}
	overtime := time.Now().Add(ttl)
	key := fmt.Sprintf("%s%s%s", name, separator, overtime.Format(time.RFC3339))
	return Key(key), nil
}

// Release used to stop MemStore
func (m *MemStore) Release() {
	m.stopCh <- struct{}{}
}

// NewMemStore init a MemStore and start a goroutine handle expired key
func NewMemStore() *MemStore {
	ms := &MemStore{
		memBuf: NewMemDbBuffer(),
		stopCh: make(chan struct{}),
	}
	ms.handleOverdueKey()
	return ms
}
