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
	item.RLock()
	defer item.Unlock()

	return item.key
}

func (item *CacheItem) Data() interface{} {
	item.RLock()
	defer item.Unlock()

	return item.data
}

func (item *CacheItem) LifeSpan() time.Duration {
	return item.lifeSpan
}

func (item *CacheItem) CreateOn() time.Time {
	item.RLock()
	defer item.Unlock()

	return item.createdOn
}

func (item *CacheItem) AccessOn() time.Time {
	item.RLock()
	defer item.Unlock()

	return item.accessedOn
}

func (item *CacheItem) AccessCount() int64 {
	item.RLock()
	defer item.Unlock()

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
