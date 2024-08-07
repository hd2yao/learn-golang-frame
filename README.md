# learn-golang-frame
学习一下 golang 的框架使用

### ants

> github 地址：https://github.com/panjf2000/ants
> 
> 文章讲解：https://taohuawu.club/archives/high-performance-implementation-of-goroutine-pool

ants是一个高性能的 goroutine 池，实现了对大规模 goroutine 的调度管理、goroutine 复用，
允许使用者在开发并发程序的时候限制 goroutine 数量， 复用资源，达到更高效执行任务的效果。