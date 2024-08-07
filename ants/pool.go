package ants

import (
    "sync"
    "time"
)

type sig struct{}

type f func() error

type Pool struct {
    // capacity of the pool
    capacity int32

    // running is the number of the currently running goroutines
    running int32

    // expiryDuration set the expired time (second) of every worker
    expiryDuration time.Duration

    // workers is a slice that store the available workers
    workers []*Worker

    // release is used to notice the pool to closed itself
    release chan sig

    // lock for synchronous operation
    lock sync.Mutex

    // once is used to ensure that the pool shutdown is performed only once
    once sync.Once
}
