package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

const (
	trays           = 8
	fixturesPerTray = 12
	totalFixtures   = trays * fixturesPerTray
)

type Product struct {
	ID       int
	WorkTime time.Duration
}

// Work is the job performed when the product is assigned to a fixture.
func (p *Product) Work(fixtureID int) {
	fmt.Printf("[Fixture %02d] Working on Product %03d for %v\n", fixtureID, p.ID, p.WorkTime)
	time.Sleep(p.WorkTime)
	fmt.Printf("[Fixture %02d] Finished Product %03d\n", fixtureID, p.ID)
}

type Fixture struct {
	ID    int
	JobCh chan *Product
}

func newFixture(id int, wg *sync.WaitGroup, availableCh chan *Fixture) *Fixture {
	f := &Fixture{
		ID:    id,
		JobCh: make(chan *Product),
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		for product := range f.JobCh {
			product.Work(f.ID)
			availableCh <- f // Mark self as available again
		}
	}()
	return f
}

func main() {
	rand.Seed(time.Now().UnixNano())
	var wg sync.WaitGroup

	fixtures := make([]*Fixture, totalFixtures)
	availableFixtures := make(chan *Fixture, totalFixtures)

	// Start fixtures and fill availability pool
	for i := 0; i < totalFixtures; i++ {
		f := newFixture(i, &wg, availableFixtures)
		fixtures[i] = f
		availableFixtures <- f
	}

	// Product generator
	go func() {
		productID := 0
		for {
			productID++
			workTime := time.Duration(rand.Intn(4)+3) * time.Second
			product := &Product{ID: productID, WorkTime: workTime}

			fixture := <-availableFixtures
			fixture.JobCh <- product

			time.Sleep(1 * time.Second) // new product every second
		}
	}()

	// Run for 60 seconds and then exit
	time.Sleep(60 * time.Second)
	fmt.Println("Shutting down...")

	// Clean up: close all fixture channels
	for _, f := range fixtures {
		close(f.JobCh)
	}

	wg.Wait()
}
