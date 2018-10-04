// Copyright (C) 2018 Storj Labs, Inc.
// See LICENSE for copying information.

package checker

import (
	"context"

	"storj.io/storj/pkg/datarepair/queue"
	"storj.io/storj/storage/redis"
	"storj.io/storj/storage/redis/redisserver"
)

// Config contains configurable values for checker
type Config struct {
	queueAddress string
	redisDB      int
	dbpassword   string //REMOVE
	// queueAddress string `help:"data repair queue address" default:"localhost:7777"`
}

// Run runs the checker with configured values
func (c *Config) Run(ctx context.Context) (err error) {
	addr, cleanup, err := redisserver.Start()
	if err != nil {
		return err
	}
	client, err := redis.NewClient(addr, c.dbpassword, c.redisDB)
	if err != nil {
		return err
	}
	q := queue.NewQueue(client)

	return nil
}
