package cacheGo

import (
	"sync"
	"time"
)

// cacheitem.go文件主要是实现缓存相关的属性

// 初始化CacheItem
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

// CacheItem 结构
// 缓存主要的属性集
type CacheItem struct {
	sync.RWMutex

	key  interface{} // key
	data interface{} // value

	lifeSpan      time.Duration         // 缓存存在的生命周期
	createdOn     time.Time             // 缓存创建的时间
	accessedOn    time.Time             // 缓存最后一次被访问的时间
	accessCount   int64                 // 被访问的次数
	aboutToExpire func(key interface{}) // 回调函数，删除缓存中的数据前被触发
}

// 以下为CacheItem结构的方法
// 注意：因为key, data, lifeSpan, createdOn属性为不变的属性，所以对它们的访问不需要加锁（同步数据）；
// 而accessedOn, accessCount, aboutToExpire属性是可变的，所以应对它们加锁以同步，避免数据的不一致。

// 获取key
func (item *CacheItem) Key() interface{} {
	// immutable
	return item.key
}

// 获取value
func (item *CacheItem) Data() interface{} {
	// immutable
	return item.data
}

// 获取生命周期
func (item *CacheItem) LifeSpan() time.Duration {
	// immutable
	return item.lifeSpan
}

// 获取创建时间
func (item *CacheItem) CreatedOn() time.Time {
	// immutable
	return item.createdOn
}

// 获取最后一次访问时间
func (item *CacheItem) AccessedOn() time.Time {
	// Because accessedOn is mutable, lock it before reading.
	item.RLock()
	defer item.RUnlock()

	return item.accessedOn
}

// 获取访问次数
func (item *CacheItem) AccessCount() int64 {
	// Because accessCount is mutable, lock it before reading.
	item.RLock()
	defer item.RUnlock()

	return item.accessCount
}

// 设置aboutToExpire回调函数，可以在删除之前做自行一些处理，如备份等
func (item *CacheItem) SetAboutToExpireCallback(f func(interface{})) {
	item.Lock()
	defer item.Unlock()

	item.aboutToExpire = f
}

// 更新访问时间和访问次数
func (item *CacheItem) KeepAlive() {
	item.Lock()
	defer item.Unlock()

	item.accessedOn = time.Now()
	item.accessCount++
}
