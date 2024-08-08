package ants

import (
    "sync"
    "sync/atomic"
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

// Submit 提交任务到 pool
func (p *Pool) Submit(task f) error {
    // 判断当前 pool 是否已被关闭
    if len(p.release) > 0 {
        return ErrPoolClosed
    }

    // 获取 pool 一个可用的 worker，绑定 task 执行
    w := p.getWorker()
    w.task <- task
    return nil
}

// getWorker 返回一个可用的 worker 来执行 task
func (p *Pool) getWorker() *Worker {
    var w *Worker
    // 标志变量，判断当前正在运行的 worker 数量是否已到达 pool 的容量上限
    waiting := false

    // 加锁，检测队列中是否有可用的 worker，并进行相应操作
    p.lock.Lock()
    idleWorkers := p.workers
    n := len(idleWorkers) - 1
    // 当前队列中无可用 worker
    if n < 0 {
        // 判断运行 worker 数目已达到该 Pool 的容量上限，置等待标志
        waiting = p.Running() >= p.Cap()
    } else {
        // 当前队列有可用 worker，从队列尾部取出一个使用
        w = idleWorkers[n]
        idleWorkers[n] = nil
        p.workers = idleWorkers[:n]
    }
    // 检测完成，解锁
    p.lock.Unlock()

    // Pool 容量已满，新请求等待
    if waiting {
        // 利用锁阻塞等待直到有空闲 worker
        for {
            p.lock.Lock()
            idleWorkers = p.workers
            l := len(idleWorkers) - 1
            if l < 0 {
                p.lock.Unlock()
                continue
            }
            w = idleWorkers[l]
            idleWorkers[l] = nil
            p.workers = idleWorkers[:l]
            p.lock.Unlock()
            break
        }
    } else if w == nil {
        // 当前无空闲 worker 但是 Pool 还没有满，则可以直接断开一个 worker 执行任务
        w = &Worker{
            pool: p,
            task: make(chan f, 1),
        }
        w.run()
        // 运行 worker 数加一
        p.incRunning()
    }
    return w
}

// putWorker 将 worker 放回空闲 Pool 中，goroutine 复用
func (p *Pool) putWorker(worker *Worker) {
    // 写入回收时间，亦即该 worker 的最后一次结束运行的时间
    worker.recycleTime = time.Now()
    p.lock.Lock()
    p.workers = append(p.workers, worker)
    p.lock.Unlock()
}

// ReSize 动态扩容或缩小池容量
func (p *Pool) ReSize(size int) {
    if size == p.Cap() {
        return
    }
    // 使用原子操作更新池容量
    atomic.StoreInt32(&p.capacity, int32(size))
    diff := p.Running() - size
    // 差值大于 0 表示需要关闭一些 worker
    if diff > 0 {
        for i := 0; i < diff; i++ {
            // 向 task 通道发送 nil，表示该 worker 应该停止工作
            p.getWorker().task <- nil
        }
    }
}

// periodicallyPurge 定期清理过期 Worker
func (p *Pool) periodicallyPurge() {
    heartbeat := time.NewTicker(p.expiryDuration)
    for range heartbeat.C {
        currentTime := time.Now()
        p.lock.Lock()
        idleWorkers := p.workers
        if len(idleWorkers) == 0 && p.Running() == 0 && len(p.release) > 0 {
            p.lock.Unlock()
            return
        }
        n := 0
        for i, worker := range idleWorkers {
            if currentTime.Sub(worker.recycleTime) <= p.expiryDuration {
                break
            }
            n = i
            worker.task <- nil
            idleWorkers[i] = nil
        }
        n++
        if n >= len(idleWorkers) {
            p.workers = idleWorkers[:0]
        } else {
            p.workers = idleWorkers[n:]
        }
        p.lock.Unlock()
    }
}
