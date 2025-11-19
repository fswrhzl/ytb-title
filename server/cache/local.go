// 本地缓存处理器，包括存储、读取、删除、过期回收
package cache

import (
	"fmt"
	"sync"
	"time"

	"golang.org/x/sync/singleflight"
)

// 缓存数据的结构，包括值、有效期
type localItem struct {
	value      string // 缓存数据
	expiration int64  // 有效期，时间戳，单位：纳秒
}

// 本地缓存数据处理器，包括存储、读取、删除、过期回收
//
//	type LocalCache struct {
//		mu    sync.Map      // 缓存数据存储，键为字符串，值为 localItem
//		gcInt time.Duration // 垃圾回收间隔，单位：秒
//	}
type LocalCache struct {
	mu    sync.Map      // 缓存数据存储，键为字符串，值为 localItem
	gcInt time.Duration // 垃圾回收间隔，单位：秒
	group singleflight.Group
}

// 创建一个新的本地缓存处理器
func NewLocalCache(gcInt time.Duration) *LocalCache {
	c := &LocalCache{
		gcInt: gcInt,
	}
	go c.gc() // 开启后台过期回收goroutine
	return c
}

func (c *LocalCache) Get(key string) (string, int64, bool) {
	if v, ok := c.mu.Load(key); ok {
		item := v.(localItem)
		if time.Now().UnixNano() < item.expiration {
			return item.value, item.expiration, true
		}
		c.mu.Delete(key)
	}
	return "", 0, false
}

func (c *LocalCache) GetWithLoader(key string, ttl time.Duration, loader func() (string, error)) (string, error) {
	if val, _, ok := c.Get(key); ok {
		fmt.Printf("从缓存获取%s数据\n", key)
		return val, nil
	}
	val, err, _ := c.group.Do(key, func() (any, error) {
		data, err := loader()
		if err != nil {
			return "", err
		}
		return data, nil
	})
	if err != nil {
		return "", err
	}
	c.Set(key, val.(string), ttl)
	return val.(string), nil
}

func (c *LocalCache) GetWithAutoRefresh(key string, ttl time.Duration, loader func() (string, error)) (string, error) {
	if val, expiration, ok := c.Get(key); ok {
		go func() {
			// 剩余有效期少于1/10时自动刷新
			if expiration-time.Now().UnixNano() < ttl.Nanoseconds()/10 {
				data, err := loader()
				if err != nil {
					return
				}
				c.Set(key, data, ttl)
			}
		}()
		fmt.Printf("从缓存获取%s数据\n", key)
		return val, nil
	}

	return c.GetWithLoader(key, ttl, loader)
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
