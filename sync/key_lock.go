package sync

import "sync"

type KeyLock interface {
	Lock(key string)
	Unlock(key string)
	RLock(key string)
	RUnlock(key string)
}

type keyLock struct {
	locks sync.Map
}

func NewKeyLock() KeyLock {
	return &keyLock{
		locks: sync.Map{},
	}
}

func (l *keyLock) getLock(key string) *sync.RWMutex {
	if lock, ok := l.locks.Load(key); ok {
		return lock.(*sync.RWMutex)
	}

	lock := &sync.RWMutex{}
	l.locks.Store(key, lock)
	return lock
}

func (l *keyLock) Lock(key string) {
	l.getLock(key).Lock()
}

func (l *keyLock) Unlock(key string) {
	l.getLock(key).Unlock()
}

func (l *keyLock) RLock(key string) {
	l.getLock(key).RLock()
}

func (l *keyLock) RUnlock(key string) {
	l.getLock(key).RUnlock()
}
