package cacheGo

import (
	"log"
	"sort"
	"sync"
	"time"
)

// CacheTable is a table within the cache
type CacheTable struct {
	sync.RWMutex
	name            string                     // 表名
	items           map[interface{}]*CacheItem // 表中所有缓存数据，底层使用map
	cleanupTimer    *time.Timer                // 负责触发清理的定时器
	cleanupInterval time.Duration              // 当前定时器持续时间
	logger          *log.Logger                // 日志
	// 回调函数
	// 当获取一个不存在的key时自动触发
	loadData func(key interface{}, args ...interface{}) *CacheItem
	// 当新增一个item到缓存时自动触发
	addedItem func(item *CacheItem)
	// 当从缓存中删除一个item时自动出发
	aboutToDeleteItem func(item *CacheItem)
}

// CacheItemPair maps key to access counter
type CacheItemPair struct {
	Key         interface{} // key
	AccessCount int64       // 访问次数计数器
}

// CacheItemPairList is a slice of CacheIemPairs that implements sort.
// Interface to sort by AccessCount.
type CacheItemPairList []CacheItemPair

// 当前缓存表中缓存的数量
func (table *CacheTable) Count() int {
	table.RLock()
	defer table.RUnlock()

	return len(table.items)
}

// 遍历所有items，可在trans中对其操作
func (table *CacheTable) Foreach(trans func(key interface{}, item *CacheItem)) {
	table.RLock()
	defer table.RUnlock()

	for k, v := range table.items {
		trans(k, v)
	}
}

// 循环过期检查
// Expiration check loop, triggered by a self-adjusting timer.
func (table *CacheTable) expirationCheck() {
	table.Lock()
	if table.cleanupTimer != nil {
		table.cleanupTimer.Stop()
	}
	if table.cleanupInterval > 0 {
		table.log("Expiration check triggered after", table.cleanupInterval, "for table", table.name)
	} else {
		table.log("Expiration check installed for table", table.name)
	}

	// To be more accurate with timers, we would need to update 'now' on every
	// loop iteration. Not sure it's really efficient though.
	now := time.Now()
	smallestDuration := 0 * time.Second
	for key, item := range table.items {
		// Cache values so we don't keep blocking the mutex.
		item.RLock()
		lifeSpan := item.lifeSpan
		accessedOn := item.accessedOn
		item.RUnlock()

		if lifeSpan == 0 {
			continue
		}
		if now.Sub(accessedOn) >= lifeSpan {
			// Item has excessed its lifespan.
			table.deleteInternal(key)
		} else {
			// Find the item chronologically closest to its end-of-lifespan.
			if smallestDuration == 0 || lifeSpan-now.Sub(accessedOn) < smallestDuration {
				smallestDuration = lifeSpan - now.Sub(accessedOn)
			}
		}
	}

	// Setup the interval for the next cleanup run.
	table.cleanupInterval = smallestDuration
	if smallestDuration > 0 {
		table.cleanupTimer = time.AfterFunc(smallestDuration, func() {
			go table.expirationCheck()
		})
	}
	table.Unlock()
}

func (table *CacheTable) addInternal(item *CacheItem) {
	// Careful: do not run this method unless the table-mutex is locked!
	// It will unlock it for the caller before running the callbacks and checks
	table.log("Adding item with key", item.key, "and lifespan of", item.lifeSpan, "to table", table.name)
	table.items[item.key] = item

	// Cache values so we don't keep blocking the mutex.
	expDur := table.cleanupInterval
	addedItem := table.addedItem
	table.Unlock()

	// Trigger callback after adding an item to cache.
	if addedItem != nil {
		addedItem(item)
	}

	// If we haven't set up any expiration check timer or found a more imminent item.
	if item.lifeSpan > 0 && (expDur == 0 || item.lifeSpan < expDur) {
		table.expirationCheck()
	}
}

// 增加一个key-value键值对到缓存
// 参数key是CacheItem结构的cache-key.
// 参数lifeSpan决定item在多长时间周期内没被访问时从缓存中删除.
// 参数data是CacheItem结构的cache-value.
func (table *CacheTable) Add(key interface{}, lifeSpan time.Duration, data interface{}) *CacheItem {
	item := NewCacheItem(key, lifeSpan, data)

	table.Lock()
	table.addInternal(item)

	return item
}

func (table *CacheTable) deleteInternal(key interface{}) (*CacheItem, error) {
	r, ok := table.items[key]
	if !ok {
		return nil, ErrKeyNotFound
	}

	// Cache value so we don't keep blocking the mutex.
	aboutToDeleteItem := table.aboutToDeleteItem
	table.Unlock()

	// Trigger callbacks before deleting an item from cache.
	if aboutToDeleteItem != nil {
		aboutToDeleteItem(r)
	}

	r.RLock()
	defer r.RUnlock()
	if r.aboutToExpire != nil {
		r.aboutToExpire(key)
	}

	table.Lock()
	table.log("Deleting item with key", key, "created on", r.createdOn, "and hit", r.accessCount, "times from table", table.name)
	delete(table.items, key)

	return r, nil
}

// 从缓存中删除key对应的item
// 参数key是CacheItem结构的cache-key.
func (table *CacheTable) Delete(key interface{}) (*CacheItem, error) {
	table.Lock()
	defer table.Unlock()

	return table.deleteInternal(key)
}

// Exists 返回是否删除缓存中key对应的item. Unlike the Value method
// Exists neither tries to fetch data via the loadData callback nor does it
// keep the item alive in the cache.
func (table *CacheTable) Exists(key interface{}) bool {
	table.RLock()
	defer table.RUnlock()

	// ok-idom
	_, ok := table.items[key]

	return ok
}

// NotFoundAdd 返回item是否存在于缓存中. Unlike the Exists
// method this also adds data if they key could not be found.
func (table *CacheTable) NotFoundAdd(key interface{}, lifeSpan time.Duration, data interface{}) bool {
	table.Lock()

	if _, ok := table.items[key]; ok {
		table.Unlock()
		return false
	}

	item := NewCacheItem(key, lifeSpan, data)
	table.addInternal(item)

	return true
}

// Value 返回缓存中key对应的item，并标记它为kept alive.
// 你可以增加一些额外的参数给回调函数DataLoader.
func (table *CacheTable) Value(key interface{}, args ...interface{}) (*CacheItem, error) {

	table.RLock()
	vItem, ok := table.items[key]
	loadData := table.loadData
	table.RUnlock()

	if ok {
		// 更新访问次数accessCount和最后一次访问时间accessedOn
		vItem.KeepAlive()
		return vItem, nil
	}

	// Item doesn't exist in cache. Try and fetch it with a data-loader.
	if loadData != nil {
		item := loadData(key, args...)
		if item != nil {
			table.Add(key, item.lifeSpan, item.data)
			return item, nil
		}
		return nil, ErrKeyNotFoundOrLoadable
	}
	return nil, ErrKeyNotFound
}

func (table *CacheTable) SetDataLoader(f func(interface{}, ...interface{}) *CacheItem) {
	table.Lock()
	defer table.Unlock()

	table.loadData = f
}

func (table *CacheTable) SetAddedItemCallback(f func(*CacheItem)) {
	table.Lock()
	defer table.Unlock()

	table.addedItem = f
}

func (table *CacheTable) SetAboutToDeleteItemCallback(f func(*CacheItem)) {
	table.Lock()
	defer table.Unlock()

	table.aboutToDeleteItem = f
}

func (table *CacheTable) SetLogger(logger *log.Logger) {
	table.Lock()
	defer table.Unlock()

	// 把日志库中的日志变量赋值给CacheTable中logger属性
	table.logger = logger
}

// 清除缓存表中所有的items
func (table *CacheTable) Flush() {
	table.Lock()
	defer table.Unlock()

	table.log("Flush table", table.name)

	table.items = make(map[interface{}]*CacheItem)
	table.cleanupInterval = 0
	if table.cleanupTimer != nil {
		table.cleanupTimer.Stop()
	}
}

func (p CacheItemPairList) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}

func (p CacheItemPairList) Len() int {
	return len(p)
}

func (p CacheItemPairList) Less(i, j int) bool {
	return p[i].AccessCount > p[j].AccessCount
}

// MostAccessed 返回缓存表中访问最多的items
func (table *CacheTable) MostAccessed(count int64) []*CacheItem {
	table.RLock()
	defer table.RUnlock()

	// 获得缓存表中的所有数据后存储到切片p中，并计数和排序
	p := make(CacheItemPairList, len(table.items))
	i := 0
	for k, v := range table.items {
		p[i] = CacheItemPair{k, v.accessCount}
		i++
	}
	sort.Sort(p)

	var r []*CacheItem
	c := int64(0)
	for _, v := range p {
		if c >= count {
			break
		}

		item, ok := table.items[v.Key]
		if ok {
			r = append(r, item)
		}
		c++
	}

	return r
}

// Internal logging method for convenience.
func (table *CacheTable) log(v ...interface{}) {
	if table.logger == nil {
		return
	}

	table.logger.Println(v)
}
