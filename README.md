cacheGo
========
轻量级并发安全的golang缓存库，具备过期处理能力


## 初衷

缓存在 Web 系统中的使用有很多好处，不仅可以减少网络流量，降低客户访问延迟，还可以减轻服务器负载。

目前已经存在很多高性能的缓存系统，比如 Memcache，Redis 等，尤其是 Redis，现已经广泛用于各种 Web 服务中。既然有了这些功能完善的缓存系统，那为什么我们还需要自己实现一个缓存系统呢？这么做主要有两个原因，第一，通过动手实现我们可以了解缓存系统的工作原理。第二，Redis 之类的缓存系统都是独立存在的，如果只是开发一个简单的应用还需要单独使用 Redis 服务器，难免会过于复杂。这时候如果有一个功能完善的软件包实现了这些功能，只需要引入这个软件包就能实现缓存功能，而不需要单独使用 Redis 服务器，就最好不过了。


## 安装

确保已经安装好Go环境。参考 [install instructions](http://golang.org/doc/install.html).

安装cacheGo很简单，命令如下：

    go get github.com/BigoSprite/cacheGo

从源码编译：

    cd $GOPATH/src/github.com/BigoSprite/cacheGo
    go get -u -v
    go build && go test -v

## 例子
```go
package main

import (
	"fmt"
	"github.com/BigoSprite/cacheGo"
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
```

运行这个例子也很简单，切换到目录examples/mycachedapp/ 后执行命令:

    go run mycachedapp.go

## 参考

原作者仓库为：github.com/muesli/cache2go，本项目仅以学习为目的。

API docs can be found [here](http://godoc.org/github.com/muesli/cache2go).

[![Build Status](https://travis-ci.org/muesli/cache2go.svg?branch=master)](https://travis-ci.org/muesli/cache2go)
[![Coverage Status](https://coveralls.io/repos/github/muesli/cache2go/badge.svg?branch=master)](https://coveralls.io/github/muesli/cache2go?branch=master)
[![Go ReportCard](http://goreportcard.com/badge/muesli/cache2go)](http://goreportcard.com/report/muesli/cache2go)

