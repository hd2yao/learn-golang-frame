package ants

import "time"

type Worker struct {
    // pool who owns this worker
    pool *Pool

    // task is a job should be done
    task chan f

    // recycleTime will be updated when putting a worker back into queue
    recycleTime time.Time
}

// run 启动一个 goroutine 来重复执行函数调用的过程
func (w *Worker) run() {
    go func() {
        // 循环监听任务队列，一旦有任务立马取出运行
        for fc := range w.task {
            if fc == nil {
                // 退出 goroutine，运行 worker 数减一
                w.pool.decRunning()
                return
            }
            fc()
            // worker 回收复用
            w.pool.putWorker(w)
        }
    }()
}
