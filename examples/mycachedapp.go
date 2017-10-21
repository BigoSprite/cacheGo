package main

import (
	"fmt"
	"github.com/cacheGo"
	"time"
)

type myStruct struct {
	text     string
	moreData []byte
}

func main() {
	// 创建一个缓存表对象，使用的是单例模式
	cache := cacheGo.Cache("myCache")

	// 准备数据
	val := myStruct{
		text:     "This is a test!",
		moreData: []byte{},
	}

	// 向缓存中增加要缓存的数据val，过期时间是5s，这里的key暂设置为someKey
	cache.Add("someKey", 5*time.Second, &val)

	// 从缓存中获取key对应的数据
	res, err := cache.Value("someKey")
	if err == nil {
		fmt.Println("Fount value in cache: ", res.Data().(*myStruct).text)
	} else {
		fmt.Println("Error retrieving value from cache:", err)
	}

	// 睡眠6s，以使someKey对应的缓存数据过期
	time.Sleep(6 * time.Second)
	res, err = cache.Value("someKey")
	if err != nil {
		fmt.Println("Item is not cached (anymore).")
	}

	// 现在someKey已经过期
	// Add another item that never expires.
	cache.Add("someKey", 0, &val)

	// cacheGo supports a few handy callbacks and loading mechanisms.
	cache.SetAboutToDeleteItemCallback(func(e *cacheGo.CacheItem) {
		fmt.Println("Deleting:", e.Key(), e.Data().(*myStruct).text, e.CreatedOn())
	})

	// 从缓存中删除key对应的数据
	cache.Delete("someKey")

	// 清除所有缓存
	cache.Flush()
}
