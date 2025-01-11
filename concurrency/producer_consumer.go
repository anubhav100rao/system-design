package main

import (
	"fmt"
	"sync"
	"time"
)

type RWLock struct {
	mu           sync.Mutex
	readers      int
	writerActive bool
	cond         *sync.Cond
}

func NewRWLock() *RWLock {
	lock := &RWLock{}
	lock.cond = sync.NewCond(&lock.mu)
	return lock
}

// Reader wants to acquire a read lock
func (rw *RWLock) RLock() {
	rw.mu.Lock()
	defer rw.mu.Unlock()

	for rw.writerActive {
		rw.cond.Wait() // Wait until the writer finishes
	}

	rw.readers++ // Increment the count of active readers
}

// Reader releases the read lock
func (rw *RWLock) RUnlock() {
	rw.mu.Lock()
	defer rw.mu.Unlock()

	rw.readers-- // Decrement the count of active readers
	if rw.readers == 0 {
		rw.cond.Broadcast() // Notify waiting writers if no readers remain
	}
}

// Writer wants to acquire a write lock
func (rw *RWLock) Lock() {
	rw.mu.Lock()
	defer rw.mu.Unlock()

	for rw.writerActive || rw.readers > 0 {
		rw.cond.Wait() // Wait until no readers or writers are active
	}

	rw.writerActive = true // Mark writer as active
}

// Writer releases the write lock
func (rw *RWLock) Unlock() {
	rw.mu.Lock()
	defer rw.mu.Unlock()

	rw.writerActive = false // Mark writer as inactive
	rw.cond.Broadcast()     // Notify all waiting readers and writers
}

func main() {
	lock := NewRWLock()
	var wg sync.WaitGroup

	// Simulating multiple readers
	for i := 1; i <= 5; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < 3; j++ {
				lock.RLock()
				fmt.Printf("Reader %d is reading.\n", id)
				time.Sleep(100 * time.Millisecond)
				lock.RUnlock()
				fmt.Printf("Reader %d finished reading.\n", id)
				time.Sleep(100 * time.Millisecond)
			}
		}(i)
	}

	// Simulating multiple writers
	for i := 1; i <= 2; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < 2; j++ {
				lock.Lock()
				fmt.Printf("Writer %d is writing.\n", id)
				time.Sleep(200 * time.Millisecond)
				lock.Unlock()
				fmt.Printf("Writer %d finished writing.\n", id)
				time.Sleep(300 * time.Millisecond)
			}
		}(i)
	}

	wg.Wait()
	fmt.Println("All tasks completed.")
}
