package utils

import (
	"fmt"
	"github.com/IBM/ubiquity/logutil"
	"sync"
	"time"
)

//go:generate counterfeiter -o ../fakes/fake_locker.go . Locker
type Locker interface {
	WriteLock(name string)
	WriteUnlock(name string)
	ReadLock(name string)
	ReadUnlock(name string)
}

func NewLocker() Locker {
	return &locker{locks: make(map[string]*sync.RWMutex), accessLock: &sync.Mutex{}, statsLock: &sync.Mutex{}, cleanupLock: &sync.Mutex{}, stats: make(map[string]time.Time), logger: logutil.GetLogger()}
}

const (
	STALE_LOCK_TIMEOUT = 600 //in seconds
)

type locker struct {
	accessLock  *sync.Mutex
	locks       map[string]*sync.RWMutex
	statsLock   *sync.Mutex
	stats       map[string]time.Time
	cleanupLock *sync.Mutex
	logger      logutil.Logger
}

func (l *locker) WriteLock(name string) {
	defer l.logger.Trace(logutil.DEBUG, logutil.Args{{"lockName", name}})()

	defer l.updateStats(name)
	l.accessLock.Lock()
	if lock, exists := l.locks[name]; exists {
		l.accessLock.Unlock()
		lock.Lock()
		return
	}

	lock := &sync.RWMutex{}
	lock.Lock()
	l.locks[name] = lock
	l.accessLock.Unlock()
}
func (l *locker) WriteUnlock(name string) {
	defer l.logger.Trace(logutil.DEBUG, logutil.Args{{"lockName", name}})()
	defer l.updateStats(name)
	l.accessLock.Lock()
	defer l.accessLock.Unlock()
	if lock, exists := l.locks[name]; exists {
		lock.Unlock()
		return
	}

}
func (l *locker) ReadLock(name string) {
	defer l.logger.Trace(logutil.DEBUG, logutil.Args{{"lockName", name}})()
	defer l.updateStats(name)
	l.accessLock.Lock()
	if lock, exists := l.locks[name]; exists {
		l.accessLock.Unlock()
		lock.RLock()
		return
	}

	lock := &sync.RWMutex{}
	lock.RLock()
	l.locks[name] = lock
	l.accessLock.Unlock()
}
func (l *locker) ReadUnlock(name string) {
	defer l.logger.Trace(logutil.DEBUG, logutil.Args{{"lockName", name}})()
	defer l.updateStats(name)
	l.accessLock.Lock()
	defer l.accessLock.Unlock()
	if lock, exists := l.locks[name]; exists {
		lock.RUnlock()
		return
	}
}
func (l *locker) updateStats(name string) {
	defer l.logger.Trace(logutil.DEBUG, logutil.Args{{"lockName", name}})()

	l.statsLock.Lock()
	defer l.cleanup()
	defer l.statsLock.Unlock()
	if stat, exists := l.stats[name]; exists {
		stat = time.Now()
		l.stats[name] = stat
		return
	}
	stat := time.Now()
	l.stats[name] = stat
}
func (l *locker) cleanup() {
	defer l.logger.Trace(logutil.DEBUG)()

	l.cleanupLock.Lock()
	defer l.cleanupLock.Unlock()
	currentTime := time.Now()
	var statsToDelete []string
	for name, stat := range l.stats {
		if currentTime.Sub(stat).Seconds() > STALE_LOCK_TIMEOUT {
			msg := fmt.Sprint("Removing stalelock '%s' as it has exceeded configured timeout ('%d seconds')\n", name, STALE_LOCK_TIMEOUT)
			l.logger.Debug(msg)
			delete(l.locks, name)
			statsToDelete = append(statsToDelete, name)
		}
	}
	for _, name := range statsToDelete {
		delete(l.stats, name)
	}
}
