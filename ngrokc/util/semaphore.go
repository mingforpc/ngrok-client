package util

type Semaphore struct {
	semaphore chan int64
}

func (sem *Semaphore) Init(size int64) {
	sem.semaphore = make(chan int64, size)
}

func (sem *Semaphore) Acquire() {
	sem.semaphore <- 1
}

func (sem *Semaphore) Release() {
	<-sem.semaphore
}

func (sem *Semaphore) Close() {
	close(sem.semaphore)
}
