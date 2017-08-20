package db

import "sync"

type ShardedLock struct {
	nShards uint32
	mu      []sync.RWMutex
}

func NewShardedLock(nShards uint32) *ShardedLock {
	return &ShardedLock{
		nShards: nShards,
		mu:      make([]sync.RWMutex, nShards),
	}
}

func (l *ShardedLock) RLock(id uint32) {
	l.mu[id%l.nShards].RLock()
}

func (l *ShardedLock) RUnlock(id uint32) {
	l.mu[id%l.nShards].RUnlock()
}

func (l *ShardedLock) Lock(id uint32) {
	l.mu[id%l.nShards].Lock()
}

func (l *ShardedLock) Unlock(id uint32) {
	l.mu[id%l.nShards].Unlock()
}
