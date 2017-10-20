package cacheGo

import (
	"sync"
)

var (
	cache = make(map[string]*CacheTable)
	mutex sync.RWMutex
)

func Cache(table string) *CacheTable {
	// Cache returns the existing cache table with given name later.
	mutex.RLock()
	t, ok := cache[table]
	mutex.RUnlock()

	// or create a new one if the table does not exist yet
	if !ok {
		mutex.Lock()
		t, ok := cache[table]
		// Double check whether the table exists or not.
		if !ok {
			t = &CacheTable{
				name:  table,
				items: make(map[interface{}]*CacheItem),
			}
			cache[table] = t
		}
		mutex.Unlock()
	}

	return t
}
