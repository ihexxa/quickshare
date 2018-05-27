package limiter

import (
	"sync"
	"time"
)

func now() int32 {
	return int32(time.Now().Unix())
}

func afterCyc(cyc int32) int32 {
	return int32(time.Now().Unix()) + cyc
}

func afterTtl(ttl int32) int32 {
	return int32(time.Now().Unix()) + ttl
}

type Bucket struct {
	Refresh int32
	Tokens  int16
}

func NewBucket(cyc int32, cap int16) *Bucket {
	return &Bucket{
		Refresh: afterCyc(cyc),
		Tokens:  cap,
	}
}

type Item struct {
	Expired int32
	Buckets map[int16]*Bucket
}

func NewItem(ttl int32) *Item {
	return &Item{
		Expired: afterTtl(ttl),
		Buckets: make(map[int16]*Bucket),
	}
}

type RateLimiter struct {
	items      map[string]*Item
	bucketCap  int16
	customCaps map[int16]int16
	cap        int64
	cyc        int32 // how much time, item autoclean will be executed, bucket will be refreshed
	ttl        int32 // how much time, item will be expired(but not cleaned)
	mux        sync.RWMutex
	snapshot   map[string]map[int16]*Bucket
}

func NewRateLimiter(cap int64, ttl int32, cyc int32, bucketCap int16, customCaps map[int16]int16) Limiter {
	if cap < 1 || ttl < 1 || cyc < 1 || bucketCap < 1 {
		panic("cap | bucketCap | ttl | cycle cant be less than 1")
	}

	limiter := &RateLimiter{
		items:      make(map[string]*Item, cap),
		bucketCap:  bucketCap,
		customCaps: customCaps,
		cap:        cap,
		ttl:        ttl,
		cyc:        cyc,
	}

	go limiter.autoClean()

	return limiter
}

func (limiter *RateLimiter) getBucketCap(opId int16) int16 {
	bucketCap, existed := limiter.customCaps[opId]
	if !existed {
		return limiter.bucketCap
	}
	return bucketCap
}

func (limiter *RateLimiter) Access(itemId string, opId int16) bool {
	limiter.mux.Lock()
	defer limiter.mux.Unlock()

	item, itemExisted := limiter.items[itemId]
	if !itemExisted {
		if int64(len(limiter.items)) >= limiter.cap {
			return false
		}

		limiter.items[itemId] = NewItem(limiter.ttl)
		limiter.items[itemId].Buckets[opId] = NewBucket(limiter.cyc, limiter.getBucketCap(opId)-1)
		return true
	}

	bucket, bucketExisted := item.Buckets[opId]
	if !bucketExisted {
		item.Buckets[opId] = NewBucket(limiter.cyc, limiter.getBucketCap(opId)-1)
		return true
	}

	if bucket.Refresh > now() {
		if bucket.Tokens > 0 {
			bucket.Tokens--
			return true
		}
		return false
	}

	bucket.Refresh = afterCyc(limiter.cyc)
	bucket.Tokens = limiter.getBucketCap(opId) - 1
	return true
}

func (limiter *RateLimiter) GetCap() int64 {
	return limiter.cap
}

func (limiter *RateLimiter) GetSize() int64 {
	limiter.mux.RLock()
	defer limiter.mux.RUnlock()
	return int64(len(limiter.items))
}

func (limiter *RateLimiter) ExpandCap(cap int64) bool {
	limiter.mux.RLock()
	defer limiter.mux.RUnlock()

	if cap <= int64(len(limiter.items)) {
		return false
	}

	limiter.cap = cap
	return true
}

func (limiter *RateLimiter) GetTTL() int32 {
	return limiter.ttl
}

func (limiter *RateLimiter) UpdateTTL(ttl int32) bool {
	if ttl < 1 {
		return false
	}

	limiter.ttl = ttl
	return true
}

func (limiter *RateLimiter) GetCyc() int32 {
	return limiter.cyc
}

func (limiter *RateLimiter) UpdateCyc(cyc int32) bool {
	if limiter.cyc < 1 {
		return false
	}

	limiter.cyc = cyc
	return true
}

func (limiter *RateLimiter) Snapshot() map[string]map[int16]*Bucket {
	return limiter.snapshot
}

func (limiter *RateLimiter) autoClean() {
	for {
		if limiter.cyc == 0 {
			break
		}
		time.Sleep(time.Duration(int64(limiter.cyc) * 1000000000))
		limiter.clean()
	}
}

// clean may add affect other operations, do frequently?
func (limiter *RateLimiter) clean() {
	limiter.snapshot = make(map[string]map[int16]*Bucket)
	now := now()

	limiter.mux.RLock()
	defer limiter.mux.RUnlock()
	for key, item := range limiter.items {
		if item.Expired <= now {
			delete(limiter.items, key)
		} else {
			limiter.snapshot[key] = item.Buckets
		}
	}
}

// Only for test
func (limiter *RateLimiter) exist(id string) bool {
	limiter.mux.RLock()
	defer limiter.mux.RUnlock()

	_, existed := limiter.items[id]
	return existed
}

// Only for test
func (limiter *RateLimiter) truncate() {
	limiter.mux.RLock()
	defer limiter.mux.RUnlock()

	for key, _ := range limiter.items {
		delete(limiter.items, key)
	}
}

// Only for test
func (limiter *RateLimiter) get(id string) (*Item, bool) {
	limiter.mux.RLock()
	defer limiter.mux.RUnlock()

	item, existed := limiter.items[id]
	return item, existed
}
