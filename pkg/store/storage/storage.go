// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: Apache-2.0

package storage

import (
	"context"
	"github.com/google/uuid"
	"github.com/onosproject/onos-lib-go/pkg/errors"
	"github.com/onosproject/onos-lib-go/pkg/logging"
	"github.com/onosproject/onos-mlb/pkg/store/event"
	"github.com/onosproject/onos-mlb/pkg/store/watcher"
	"sync"
)

var log = logging.GetLogger("store", "storage")

var _ Store = &store{} // to check interface and struct in compile time (static check)

// NewStore generates the new store
func NewStore() Store {
	watchers := watcher.NewWatchers()
	return &store{
		storage:  make(map[IDs]*Entry),
		watchers: watchers,
	}
}

// Store has all functions in this store
type Store interface {
	// Put puts key and its value
	Put(ctx context.Context, key IDs, value interface{}) (*Entry, error)

	// Get gets the element with key
	Get(ctx context.Context, key IDs) (*Entry, error)

	// ListElements gets all elements in this store
	ListElements(ctx context.Context, ch chan<- *Entry) error

	// ListKeys gets all keys in this store
	ListKeys(ctx context.Context, ch chan<- IDs) error

	// Update updates an element
	Update(ctx context.Context, entry *Entry) error

	// Delete deletes an element
	Delete(ctx context.Context, key IDs) error

	// Watch watches the event of this store
	Watch(ctx context.Context, ch chan<- event.Event) error

	// Print prints the map in this store for debugging
	Print()
}

type store struct {
	storage  map[IDs]*Entry
	mu       sync.RWMutex
	watchers *watcher.Watchers
}

func (s *store) Put(ctx context.Context, key IDs, value interface{}) (*Entry, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	entry := &Entry{
		Key:   key,
		Value: value,
	}
	s.storage[key] = entry
	s.watchers.Send(event.Event{
		Key:   key,
		Value: entry,
		Type:  Created,
	})
	return entry, nil
}

func (s *store) Get(ctx context.Context, key IDs) (*Entry, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if v, ok := s.storage[key]; ok {
		return v, nil
	}
	return nil, errors.New(errors.NotFound, "the storage entry does not exist")
}

func (s *store) ListElements(ctx context.Context, ch chan<- *Entry) error {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if len(s.storage) == 0 {
		return errors.New(errors.NotFound, "no storage entries stored")
	}

	for _, entry := range s.storage {
		ch <- entry
	}
	close(ch)
	return nil
}

func (s *store) ListKeys(ctx context.Context, ch chan<- IDs) error {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if len(s.storage) == 0 {
		return errors.New(errors.NotFound, "no storage entries stored")
	}

	for key := range s.storage {
		ch <- key
	}
	close(ch)
	return nil
}

func (s *store) Update(ctx context.Context, entry *Entry) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.storage[entry.Key]; ok {
		s.storage[entry.Key] = entry
		s.watchers.Send(event.Event{
			Key:   entry.Key,
			Value: entry,
			Type:  Updated,
		})
	}

	return errors.New(errors.NotFound, "no storage entry does not exist; put the entry first")
}

func (s *store) Delete(ctx context.Context, key IDs) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.storage, key)
	return nil
}

func (s *store) Watch(ctx context.Context, ch chan<- event.Event) error {
	id := uuid.New()
	err := s.watchers.AddWatcher(id, ch)
	if err != nil {
		log.Error(err)
		close(ch)
		return err
	}
	go func() {
		<-ctx.Done()
		err = s.watchers.RemoveWatcher(id)
		if err != nil {
			log.Error(err)
		}
		close(ch)
	}()
	return nil
}

func (s *store) Print() {
	s.mu.Lock()
	defer s.mu.Unlock()
	for k, v := range s.storage {
		log.Infof("key - %v / value - %v", k, v)
	}
}
