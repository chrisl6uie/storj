// Copyright (C) 2018 Storj Labs, Inc.
// See LICENSE for copying information.

package queue

import (
	"sort"
	"strconv"
	"sync"
	"testing"

	"github.com/golang/protobuf/proto"
	"github.com/stretchr/testify/assert"

	"storj.io/storj/pkg/pb"
	"storj.io/storj/storage/teststore"
)

func TestEnqueueDequeue(t *testing.T) {
	db := teststore.New()
	q := NewQueue(db)
	seg := &pb.InjuredSegment{
		Path:       "abc",
		LostPieces: []int32{},
	}
	err := q.Enqueue(seg)
	assert.NoError(t, err)

	s, err := q.Dequeue()
	assert.NoError(t, err)
	assert.True(t, proto.Equal(&s, seg))
}

func TestDequeueEmptyQueue(t *testing.T) {
	db := teststore.New()
	q := NewQueue(db)
	s, err := q.Dequeue()
	assert.Error(t, err)
	assert.Equal(t, pb.InjuredSegment{}, s)
}

func TestForceError(t *testing.T) {
	db := teststore.New()
	q := NewQueue(db)
	err := q.Enqueue(&pb.InjuredSegment{Path: "abc", LostPieces: []int32{}})
	assert.NoError(t, err)
	db.ForceError++
	item, err := q.Dequeue()
	assert.Equal(t, pb.InjuredSegment{}, item)
	assert.Error(t, err)
}

func TestSequential(t *testing.T) {
	db := teststore.New()
	q := NewQueue(db)
	const N = 100
	var addSegs []*pb.InjuredSegment
	for i := 0; i < N; i++ {
		seg := &pb.InjuredSegment{
			Path:       strconv.Itoa(i),
			LostPieces: []int32{int32(i)},
		}
		err := q.Enqueue(seg)
		assert.NoError(t, err)
		addSegs = append(addSegs, seg)
	}
	for i := 0; i < N; i++ {
		dqSeg, err := q.Dequeue()
		assert.NoError(t, err)
		assert.True(t, proto.Equal(addSegs[i], &dqSeg))
	}
}

func TestParallel(t *testing.T) {
	queue := NewQueue(teststore.New())
	const N = 100
	errs := make(chan error, N*2)
	entries := make(chan *pb.InjuredSegment, N*2)
	var wg sync.WaitGroup

	wg.Add(N)
	// Add to queue concurrently
	for i := 0; i < N; i++ {
		go func(i int) {
			defer wg.Done()
			err := queue.Enqueue(&pb.InjuredSegment{
				Path:       strconv.Itoa(i),
				LostPieces: []int32{int32(i)},
			})
			if err != nil {
				errs <- err
			}
		}(i)

	}
	wg.Wait()
	wg.Add(N)
	// Remove from queue concurrently
	for i := 0; i < N; i++ {
		go func(i int) {
			defer wg.Done()
			segment, err := queue.Dequeue()
			if err != nil {
				errs <- err
			}
			entries <- &segment
		}(i)
	}
	wg.Wait()
	close(errs)
	close(entries)

	for err := range errs {
		t.Error(err)
	}

	var items []*pb.InjuredSegment
	for segment := range entries {
		items = append(items, segment)
	}

	sort.Slice(items, func(i, k int) bool { return items[i].LostPieces[0] < items[k].LostPieces[0] })
	// check if the enqueued and dequeued elements match
	for i := 0; i < N; i++ {
		assert.Equal(t, items[i].LostPieces[0], int32(i))
	}
}
