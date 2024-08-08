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
