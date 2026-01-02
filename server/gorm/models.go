package gorm

// 标签模型
type Tag struct {
	Id   int64  `json:"id"`
	Name string `json:"name"`
}

// 频道模型
type Channel struct {
	Id           int64  `json:"id"`
	Name         string `json:"name"`
	DefaultTitle string `json:"default_title"`
}

// 频道-标签关联模型
type ChannelTag struct {
	Id        int64 `json:"id"`
	ChannelId int64 `json:"channel_id"`
	TagId     int64 `json:"tag_id"`
}

// 创建标签请求
type TagCreateRequest struct {
	Name     string  `json:"name" form:"name" binding:"required"`
	Channels []int64 `json:"channels" form:"channels" binding:"required"`
}

// 标签列表响应体
type TagResponse struct {
	Id       int64   `json:"id"`
	Name     string  `json:"name"`
	Channels []int64 `json:"channels"`
}

// 创建频道请求
type ChannelCreateRequest struct {
	Name         string  `json:"name" form:"name" binding:"required"`
	Tags         []int64 `json:"tags" form:"tags"`
	DefaultTitle string  `json:"default_title"`
}

// 更新频道请求
type ChannelUpdateRequest struct {
	Id           int64   `json:"id" form:"id" binding:"required"`
	Name         string  `json:"name" form:"name" binding:"required"`
	Tags         []int64 `json:"tags" form:"tags"`
	DefaultTitle string  `json:"default_title" form:"default_title"`
}

// 获取频道响应
type ChannelResponse struct {
	Id           int64   `json:"id"`
	Name         string  `json:"name"`
	Tags         []int64 `json:"tags"`
	DefaultTitle string  `json:"default_title"`
}
