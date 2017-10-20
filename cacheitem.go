package cacheGo

import (
	"sync"
	"time"
)

func NewCacheItem(key interface{}, lifeSpan time.Duration, data interface{}) *CacheItem {
	timeNow := time.Now()
	return &CacheItem{
		key:           key,
		data:          data,
		lifeSpan:      lifeSpan,
		createdOn:     timeNow,
		accessedOn:    timeNow,
		accessCount:   0,
		aboutToExpire: nil,
	}
}

type CacheItem struct {
	sync.RWMutex

	key  interface{} // key
	data interface{} // value

	lifeSpan      time.Duration // life
	createdOn     time.Time     // created time
	accessedOn    time.Time
	accessCount   int64
	aboutToExpire func(key interface{})
}

func (item *CacheItem) Key() interface{} {
	// immutable
	return item.key
}

func (item *CacheItem) Data() interface{} {
	// immutable
	return item.data
}

func (item *CacheItem) LifeSpan() time.Duration {
	// immutable
	return item.lifeSpan
}

func (item *CacheItem) CreatedOn() time.Time {
	// immutable
	return item.createdOn
}

func (item *CacheItem) AccessedOn() time.Time {
	// Because accessedOn is mutable, lock it before reading.
	item.RLock()
	defer item.RUnlock()

	return item.accessedOn
}

func (item *CacheItem) AccessCount() int64 {
	// Because accessCount is mutable, lock it before reading.
	item.RLock()
	defer item.RUnlock()

	return item.accessCount
}

func (item *CacheItem) SetAboutToExpireCallback(f func(interface{})) {
	item.Lock()
	defer item.Unlock()

	item.aboutToExpire = f
}

func (item *CacheItem) KeepAlive() {
	item.Lock()
	defer item.Unlock()

	item.accessedOn = time.Now()
	item.accessCount++
}
