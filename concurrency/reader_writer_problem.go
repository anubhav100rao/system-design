package main

import (
	"fmt"
	"sync"
	"time"
)

type FairReadWriteLock struct {
	mu             sync.Mutex
	readers        int
	writersWaiting int
	isWriting      bool
	readCond       *sync.Cond
	writeCond      *sync.Cond
}

func NewFairReadWriteLock() *FairReadWriteLock {
	lock := &FairReadWriteLock{}
	lock.readCond = sync.NewCond(&lock.mu)
	lock.writeCond = sync.NewCond(&lock.mu)
	return lock
}

func (lock *FairReadWriteLock) StartRead() {
	lock.mu.Lock()
	defer lock.mu.Unlock()

	// Wait if a writer is active or writers are queued
	for lock.isWriting || lock.writersWaiting > 0 {
		lock.readCond.Wait()
	}
	lock.readers++
}

func (lock *FairReadWriteLock) EndRead() {
	lock.mu.Lock()
	defer lock.mu.Unlock()

	lock.readers--
	// Signal writers only when all readers are done
	if lock.readers == 0 {
		lock.writeCond.Signal()
	}
}

func (lock *FairReadWriteLock) StartWrite() {
	lock.mu.Lock()
	defer lock.mu.Unlock()

	lock.writersWaiting++
	// Wait until no readers or writers are active
	for lock.readers > 0 || lock.isWriting {
		lock.writeCond.Wait()
	}
	lock.writersWaiting--
	lock.isWriting = true
}

func (lock *FairReadWriteLock) EndWrite() {
	lock.mu.Lock()
	defer lock.mu.Unlock()

	lock.isWriting = false
	// Prioritize waking waiting writers first
	if lock.writersWaiting > 0 {
		lock.writeCond.Signal()
	} else {
		lock.readCond.Broadcast() // Wake all pending readers
	}
}

// Example usage
func main() {
	rwLock := NewFairReadWriteLock()
	data := 0

	// Reader Goroutine
	reader := func(id int) {
		for {
			rwLock.StartRead()
			fmt.Printf("Reader %d: read data = %d\n", id, data)
			time.Sleep(100 * time.Millisecond)
			rwLock.EndRead()
		}
	}

	// Writer Goroutine
	writer := func(id int) {
		for {
			rwLock.StartWrite()
			data++
			fmt.Printf("Writer %d: wrote data = %d\n", id, data)
			time.Sleep(200 * time.Millisecond)
			rwLock.EndWrite()
		}
	}

	// Start 3 readers and 2 writers
	for i := 0; i < 3; i++ {
		go reader(i)
	}
	for i := 0; i < 2; i++ {
		go writer(i)
	}

	// Keep main alive
	select {}
}
