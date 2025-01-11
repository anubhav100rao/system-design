package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

type Philosopher struct {
	id         int
	leftFork   *sync.Mutex
	rightFork  *sync.Mutex
	dineSignal chan struct{}
}

type DiningTable struct {
	forks        []*sync.Mutex
	philosophers []*Philosopher
	mu           sync.Mutex
}

func NewDiningTable(n int) *DiningTable {
	forks := make([]*sync.Mutex, n)
	for i := 0; i < n; i++ {
		forks[i] = &sync.Mutex{}
	}

	return &DiningTable{
		forks:        forks,
		philosophers: make([]*Philosopher, 0),
	}
}

func (dt *DiningTable) AddPhilosopher(id int) {
	dt.mu.Lock()
	defer dt.mu.Unlock()

	n := len(dt.forks)
	philosopher := &Philosopher{
		id:         id,
		leftFork:   dt.forks[id%n],
		rightFork:  dt.forks[(id+1)%n],
		dineSignal: make(chan struct{}),
	}
	dt.philosophers = append(dt.philosophers, philosopher)

	go philosopher.dine()
	fmt.Printf("Philosopher %d has joined the table.\n", id)
}

func (dt *DiningTable) RemovePhilosopher(id int) {
	dt.mu.Lock()
	defer dt.mu.Unlock()

	for i, philosopher := range dt.philosophers {
		if philosopher.id == id {
			close(philosopher.dineSignal)
			dt.philosophers = append(dt.philosophers[:i], dt.philosophers[i+1:]...)
			fmt.Printf("Philosopher %d has left the table.\n", id)
			break
		}
	}
}

func (p *Philosopher) dine() {
	for {
		select {
		case <-p.dineSignal:
			fmt.Printf("Philosopher %d stops dining.\n", p.id)
			return
		default:
			p.think()
			p.eat()
		}
	}
}

func (p *Philosopher) think() {
	fmt.Printf("Philosopher %d is thinking.\n", p.id)
	time.Sleep(time.Duration(rand.Intn(1000)) * time.Millisecond)
}

func (p *Philosopher) eat() {
	// Lock forks in a specific order to avoid deadlocks
	if p.id%2 == 0 {
		p.leftFork.Lock()
		p.rightFork.Lock()
	} else {
		p.rightFork.Lock()
		p.leftFork.Lock()
	}

	fmt.Printf("Philosopher %d is eating.\n", p.id)
	time.Sleep(time.Duration(rand.Intn(1000)) * time.Millisecond)

	p.rightFork.Unlock()
	p.leftFork.Unlock()
}

func RunDiningProblemSimulation() {
	rand.Seed(time.Now().UnixNano())
	table := NewDiningTable(5)

	// Add philosophers
	for i := 0; i < 5; i++ {
		table.AddPhilosopher(i)
	}

	// Dynamically add/remove philosophers
	go func() {
		for {
			time.Sleep(5 * time.Second)
			table.AddPhilosopher(rand.Intn(100))
		}
	}()

	go func() {
		for {
			time.Sleep(7 * time.Second)
			table.RemovePhilosopher(rand.Intn(5))
		}
	}()

	// Run indefinitely
	select {}
}
