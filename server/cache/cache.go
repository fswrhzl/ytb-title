// 缓存统一接口，提供缓存操作方法
package cache

import "time"

type Cache interface {
	Get(key string) (string, bool)
	Set(key string, value string, ttl time.Duration)
	Delete(key string)
}
