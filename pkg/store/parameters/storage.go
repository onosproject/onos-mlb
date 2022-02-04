// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: Apache-2.0

package paramstorage

import (
	"context"
	"github.com/atomix/go-client/pkg/client/errors"
	"sync"
)

// NewStore generates a store object to save parameters into a map
func NewStore() Store {
	return &store{
		storage: make(map[string]int),
	}
}

// Store includes all functions for parameter storage
type Store interface {
	// Put puts parameter key and value
	Put(ctx context.Context, key string, value int) error

	// Get gets parameter value with key
	Get(ctx context.Context, key string) (int, error)

	// Update updates parameter value with key
	Update(ctx context.Context, key string, value int) error
}

type store struct {
	storage map[string]int
	mu      sync.RWMutex
}

func (s *store) Put(ctx context.Context, key string, value int) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.storage[key] = value
	return nil
}

func (s *store) Get(ctx context.Context, key string) (int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if _, ok := s.storage[key]; !ok {
		return 0, errors.NewNotFound("key not found")
	}
	return s.storage[key], nil
}

func (s *store) Update(ctx context.Context, key string, value int) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.storage[key] = value
	return nil
}
