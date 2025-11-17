// 本地缓存处理器，包括存储、读取、删除、过期回收
package cache

import (
	"sync"
	"time"
)

// 缓存数据的结构，包括值、有效期
type localItem struct {
	value      string // 缓存数据
	expiration int64  // 有效期，时间戳，单位：纳秒
}

// 本地缓存数据处理器，包括存储、读取、删除、过期回收
type LocalCache struct {
	mu    sync.Map      // 缓存数据存储，键为字符串，值为 localItem
	gcInt time.Duration // 垃圾回收间隔，单位：秒
}

// 创建一个新的本地缓存处理器
func NewLocalCache(gcInt time.Duration) *LocalCache {
	c := &LocalCache{
		gcInt: gcInt,
	}
	go c.gc() // 开启后台过期回收goroutine
	return c
}

func (c *LocalCache) Get(key string) (string, bool) {
	if v, ok := c.mu.Load(key); ok {
		item := v.(localItem)
		if time.Now().UnixNano() < item.expiration {
			return item.value, true
		}
		c.mu.Delete(key)
	}
	return "", false
}

func (c *LocalCache) Set(key string, value string, ttl time.Duration) {
	c.mu.Store(key, localItem{
		value:      value,
		expiration: time.Now().Add(ttl).UnixNano(),
	})
}

func (c *LocalCache) Delete(key string) {
	c.mu.Delete(key)
}

// 开启后台死循环，定期清理过期缓存项
func (c *LocalCache) gc() {
	for {
		time.Sleep(c.gcInt)
		c.mu.Range(func(k, v any) bool {
			item := v.(localItem) // 类型断言
			if time.Now().UnixNano() > item.expiration {
				c.mu.Delete(k)
			}
			return true
		})
	}
}
