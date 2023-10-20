package entity

import (
	"errors"
	"fmt"
	"sync"
)

type IDPool struct {
	mu        sync.Mutex
	start     int64
	end       int64
	available []int64
	used      map[int64]bool
}

func NewIDPool(start, end int64) (*IDPool, error) {
	if start <= 0 {
		return nil, errors.New("start id must be greater than 0")
	}
	if start >= end {
		return nil, errors.New("start id must be less than end id")
	}

	pool := &IDPool{
		start: start,
		end:   end,
		used:  make(map[int64]bool),
	}

	for id := start; id < end; id++ {
		pool.available = append(pool.available, id)
	}

	return pool, nil
}

func (p *IDPool) Get() (int64, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if len(p.available) == 0 {
		return 0, fmt.Errorf("ID pool is empty")
	}

	id := p.available[0]
	p.available = p.available[1:]
	p.used[id] = true
	return id, nil
}

func (p *IDPool) Put(id int64) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.used[id] {
		delete(p.used, id)
		p.available = append(p.available, id)
	}
}
