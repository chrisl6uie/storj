// Copyright (C) 2018 Storj Labs, Inc.
// See LICENSE for copying information.

package queue

import (
	"context"
	"encoding/binary"
	"math/rand"
	"sync"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/zeebo/errs"

	"storj.io/storj/pkg/pb"
	"storj.io/storj/storage"
)

// RepairQueue is the interface for the data repair queue
type RepairQueue interface {
	Enqueue(qi *pb.InjuredSegment) error
	Dequeue() (pb.InjuredSegment, error)
}

// Config contains configurable values for checker
type Config struct {
	// address string `help:"data repair queue address" default:"localhost:7777"`
}

// Run runs the checker with configured values
func (c *Config) Run(ctx context.Context) (err error) {

	// TODO: Start queue server

	return err
}

// Queue implements the RepairQueue interface
type Queue struct {
	mu sync.Mutex
	db storage.KeyValueStore
}

var (
	queueError = errs.Class("data repair queue error")
)

// NewQueue returns a pointer to a new Queue instance with an initialized connection to Redis
func NewQueue(client storage.KeyValueStore) *Queue {
	return &Queue{
		mu: sync.Mutex{},
		db: client,
	}
}

// Enqueue adds a repair segment to the queue
func (q *Queue) Enqueue(qi *pb.InjuredSegment) error {
	q.mu.Lock()
	defer q.mu.Unlock()
	const dateSize = 8
	dateTime := make([]byte, dateSize)
	// TODO: this can cause conflicts when time is unstable or running on multiple computers
	// Append random 4 byte token to account for time conflicts
	binary.BigEndian.PutUint64(dateTime, uint64(time.Now().UnixNano()))
	const tokenSize = 4
	token := make([]byte, tokenSize)
	_, err := rand.Read(token)
	if err != nil {
		return queueError.New("error creating random token %s", err)
	}
	dateTime = append(dateTime, token...)
	val, err := proto.Marshal(qi)
	if err != nil {
		return queueError.New("error marshalling injured seg %s", err)
	}
	err = q.db.Put(dateTime, val)
	if err != nil {
		return queueError.New("error adding injured seg to queue %s", err)
	}
	return nil
}

// Dequeue returns the next repair segement and removes it from the queue
func (q *Queue) Dequeue() (pb.InjuredSegment, error) {
	q.mu.Lock()
	defer q.mu.Unlock()

	items, _, err := storage.ListV2(q.db, storage.ListOptions{IncludeValue: true, Limit: 1, Recursive: true})
	if err != nil {
		return pb.InjuredSegment{}, queueError.New("error getting first key %s", err)
	}
	if len(items) == 0 {
		return pb.InjuredSegment{}, queueError.New("empty database")
	}
	key := items[0].Key
	val := items[0].Value

	seg := &pb.InjuredSegment{}
	err = proto.Unmarshal(val, seg)
	if err != nil {
		return pb.InjuredSegment{}, queueError.New("error unmarshalling segment %s", err)
	}
	err = q.db.Delete(key)
	if err != nil {
		return *seg, queueError.New("error removing injured seg %s", err)
	}
	return *seg, nil
}