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

// NewPool 创建一个实例
func NewPool(size int) (*Pool, error) {
    return NewTimingPool(size, DefaultCleanIntervalTime)
}

// NewTimingPool 创建一个带有自定义定时任务的实例
func NewTimingPool(size, expiry int) (*Pool, error) {
    if size <= 0 {
        return nil, ErrInvalidPoolSize
    }
    if expiry <= 0 {
        return nil, ErrInvalidPoolExpiry
    }
    p := &Pool{
        capacity:       int32(size),
        release:        make(chan sig, 1),
        expiryDuration: time.Duration(expiry) * time.Second,
    }

    // 启动定期清理过期 worker 任务，独立 goroutine 运行，节省系统资源
    p.monitorAndClear()
    return p, nil
}
