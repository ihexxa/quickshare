package iolimiter

import (
	"context"
	"fmt"
	"sync"

	"github.com/ihexxa/quickshare/src/db"
	"github.com/ihexxa/quickshare/src/golimiter"
)

const cacheSizeLimit = 1024

type ILimiter interface {
	CanWrite(userID uint64, chunkSize int) (bool, error)
	CanRead(userID uint64, chunkSize int) (bool, error)
}

type IOLimiter struct {
	mtx             *sync.Mutex
	UploadLimiter   *golimiter.Limiter
	DownloadLimiter *golimiter.Limiter
	users           db.IUserDB
	quotaCache      map[uint64]*db.Quota
}

func NewIOLimiter(cap, cyc int, users db.IUserDB) *IOLimiter {
	return &IOLimiter{
		mtx:             &sync.Mutex{},
		UploadLimiter:   golimiter.New(cap, cyc),
		DownloadLimiter: golimiter.New(cap, cyc),
		users:           users,
		quotaCache:      map[uint64]*db.Quota{},
	}
}

func (lm *IOLimiter) CanWrite(id uint64, chunkSize int) (bool, error) {
	lm.mtx.Lock()
	defer lm.mtx.Unlock()

	quota, ok := lm.quotaCache[id]
	if !ok {
		user, err := lm.users.GetUser(context.TODO(), id) // TODO: add context
		if err != nil {
			return false, err
		}
		quota = user.Quota
		lm.quotaCache[id] = quota
	}
	if len(lm.quotaCache) > cacheSizeLimit {
		lm.clean()
	}

	return lm.UploadLimiter.Access(
		fmt.Sprint(id),
		quota.UploadSpeedLimit,
		chunkSize,
	), nil
}

func (lm *IOLimiter) CanRead(id uint64, chunkSize int) (bool, error) {
	lm.mtx.Lock()
	defer lm.mtx.Unlock()

	quota, ok := lm.quotaCache[id]
	if !ok {
		user, err := lm.users.GetUser(context.TODO(), id) // TODO: add context
		if err != nil {
			return false, err
		}
		quota = user.Quota
		lm.quotaCache[id] = quota
	}
	if len(lm.quotaCache) > cacheSizeLimit {
		lm.clean()
	}

	return lm.DownloadLimiter.Access(
		fmt.Sprint(id),
		quota.DownloadSpeedLimit,
		chunkSize,
	), nil
}

func (lm *IOLimiter) clean() {
	count := 0
	for key := range lm.quotaCache {
		delete(lm.quotaCache, key)
		if count++; count > 5 {
			break
		}
	}
}
