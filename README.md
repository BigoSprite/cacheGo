cacheGo
========
轻量级并发安全的golang缓存库，具备过期处理能力

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
```

运行这个例子也很简单，切换到目录examples/mycachedapp/ 后执行命令:

    go run mycachedapp.go

## 参考

API docs can be found [here](http://godoc.org/github.com/muesli/cache2go).

[![Build Status](https://travis-ci.org/muesli/cache2go.svg?branch=master)](https://travis-ci.org/muesli/cache2go)
[![Coverage Status](https://coveralls.io/repos/github/muesli/cache2go/badge.svg?branch=master)](https://coveralls.io/github/muesli/cache2go?branch=master)
[![Go ReportCard](http://goreportcard.com/badge/muesli/cache2go)](http://goreportcard.com/report/muesli/cache2go)

