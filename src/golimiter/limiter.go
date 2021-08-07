package golimiter

import (
	"fmt"
	// "math"
	"sync"
	"time"
)

const expiredCycCount = 3

type Bucket struct {
	refreshedAt time.Time
	token       int
}

func NewBucket(token int) *Bucket {
	return &Bucket{
		refreshedAt: time.Now(),
		token:       token,
	}
}

func (b *Bucket) Access(cyc, incr, decr int) bool {
	now := time.Now()

	if decr > incr {
		return false
	} else if b.token >= decr {
		b.token -= decr
		return true
	}

	if b.refreshedAt.
		Add(time.Duration(cyc) * time.Millisecond).
		After(now) {
		return false
	}
	fmt.Println(4)
	b.token = incr - decr
	b.refreshedAt = now
	return true
}

type Limiter struct {
	buckets    map[string]*Bucket
	cap        int
	cyc        int
	cleanBatch int
	mtx        *sync.RWMutex
}

func New(cap, cyc int) *Limiter {
	if cap <= 0 {
		panic("limiter: invalid cap <= 0")
	}
	if cyc <= 0 {
		panic(fmt.Sprintf("limiter: invalid cyc=%d", cyc))
	}

	return &Limiter{
		buckets:    make(map[string]*Bucket),
		cap:        cap,
		cyc:        cyc,
		cleanBatch: 10,
		mtx:        &sync.RWMutex{},
	}
}

// func NewWithcleanBatch(cap, cyc, cleanBatch int64, refill int) *Limiter {
// 	limiter := New(cap, cyc, refill)
// 	limiter.cleanBatch = cleanBatch
// 	return limiter
// }

func (l *Limiter) Access(id string, incr, decr int) bool {
	l.mtx.Lock()
	defer l.mtx.Unlock()

	b, ok := l.buckets[id]
	if !ok {
		size := len(l.buckets)
		if size > l.cap/2 {
			l.clean()
		}

		size = len(l.buckets)
		if size+1 > l.cap || incr < decr {
			return false
		}
		l.buckets[id] = NewBucket(incr - decr)
		return true
	}
	return b.Access(l.cyc, incr, decr)
}

func (l *Limiter) clean() {
	count := 0

	for key, bucket := range l.buckets {
		if bucket.refreshedAt.
			Add(time.Duration(l.cyc*expiredCycCount) * time.Millisecond).
			Before(time.Now()) {
			delete(l.buckets, key)
		}
		if count++; count >= 10 {
			break
		}
	}
}

func (l *Limiter) GetCap() int {
	l.mtx.RLock()
	defer l.mtx.RUnlock()
	return l.cap
}

func (l *Limiter) GetCyc() int {
	l.mtx.RLock()
	defer l.mtx.RUnlock()
	return l.cyc
}
