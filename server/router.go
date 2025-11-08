package server

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"unicode/utf8"

	"fswrhzl/ytb_title/server/db"
	"fswrhzl/ytb_title/server/middleware"

	"github.com/gorilla/sessions"

	"github.com/gin-gonic/gin"
)

type (
	TitleRequest struct {
		Theme   string `form:"theme" binding:"required" json:"theme"`
		Channel int    `form:"channel" binding:"required" json:"channel"`
	}
)

// 在此处声明包级变量并初始化会导致空指针错误。因为Go的包级变量会在main函数之前初始化，而在main函数中初始化数据库连接后才拥有数据库实例，才会赋值给channelRepository和tagRepository的db字段。
// var channelRepository db.ChannelRepository = db.NewChannelRepository()
// var tagRepository db.TagRepository = db.NewTagRepository()
var (
	channelRepository db.ChannelRepository
	tagRepository     db.TagRepository
	flushChannel      = true // 是否刷新cookie中的channel数据
	flushTag          = true // 是否刷新cookie中的tag数据
	sessionStore      = sessions.NewCookieStore([]byte("ytb_title_20251106"))
)

func SetupRouter() *gin.Engine {
	channelRepository = db.NewChannelRepository()
	tagRepository = db.NewTagRepository()
	r := gin.Default()
	err := r.SetTrustedProxies(nil)
	if err != nil {
		panic(err)
	}
	// 使用 IP 限制中间件
	r.Use(middleware.IPRestrictionMiddleware())
	// 提供Vue项目的静态资源文件。web路径/assets/会被映射到./web/dist/assets目录下寻找文件
	// r.Static("/assets", "./web/dist/assets")
	api := r.Group("/api")
	{
		// 生成标题
		api.POST("/generate-title", generateTitle)
		// 获取所有频道
		api.GET("/channels", getChannels)
		// 编辑频道
		api.PUT("/channels/:id", updateChannel)
		// 新增频道
		api.POST("/channels", createChannel)
		// 删除频道
		api.DELETE("/channels/:id", deleteChannel)
		// 获取所有标签
		api.GET("/tags", getTags)
		// 新增标签
		api.POST("/tags", createTag)
		// 删除标签
		api.DELETE("/tags/:id", deleteTag)
	}
	// 404 路由
	// r.NoRoute(func(c *gin.Context) {
	// 	c.JSON(http.StatusNotFound, gin.H{
	// 		"message": "404 Not Found",
	// 	})
	// })

	// 提供入口文件。注册一个模板文件，gin会将该方法加载的文件注册成一个模板：index.html，后续可以在路由处理函数中使用c.HTML方法渲染该模板。
	r.LoadHTMLFiles("./web/dist/index.html")
	r.NoRoute(func(c *gin.Context) {
		path := c.Request.URL.Path
		if len(path) > 8 && path[:8] == "/assets/" {
			c.Status(http.StatusNotFound)
			return
		}
		c.HTML(http.StatusOK, "index.html", nil)
	})
	return r
}

// 获取所有频道
func getChannels(c *gin.Context) {
	// 从session中获取频道数据
	if !flushChannel {
		session, _ := sessionStore.Get(c.Request, "channel-session")
		if tmp, ok := session.Values["channels"]; ok {
			var channels []db.ChannelResponse
			_ = json.Unmarshal(tmp.([]byte), &channels)
			c.JSON(http.StatusOK, gin.H{
				"status":   "success",
				"message":  "获取频道成功",
				"channels": channels,
			})
			return
		}
	}

	channels, err := channelRepository.GetAllChannels()
	if err != nil {
		fmt.Printf("获取频道失败：%v\n", err)
		c.JSON(http.StatusOK, gin.H{
			"status":  err.Error(),
			"message": "获取频道失败",
		})
		return
	}
	// 将频道数据存储到session中
	session, _ := sessionStore.Get(c.Request, "channel-session")
	session.Values["channels"], _ = json.Marshal(channels)
	flushChannel = false
	err = session.Save(c.Request, c.Writer)
	if err != nil {
		fmt.Printf("channels保存频道到session失败：%v\n", err)
		flushChannel = true
	}

	c.JSON(http.StatusOK, gin.H{
		"status":   "success",
		"message":  "获取频道成功",
		"channels": channels,
	})
}

// 新增频道
func createChannel(c *gin.Context) {
	var channel db.ChannelCreateRequest
	if err := c.ShouldBind(&channel); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"status":  "error",
			"message": "错误的请求参数",
		})
		return
	}
	fmt.Printf("channel: %v+\n", channel)
	if channel.Name == "" {
		c.JSON(http.StatusOK, gin.H{
			"status":  "error",
			"message": "名称不能为空",
		})
		return
	}
	if err := channelRepository.CreateChannel(&channel); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"status":  "error",
			"message": err.Error(),
		})
		return
	}
	// 刷新channel数据
	flushChannel = true
	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "频道创建成功",
	})
}

// 编辑频道
func updateChannel(c *gin.Context) {
	var channelUpdateRequest db.ChannelUpdateRequest
	if err := c.ShouldBindJSON(&channelUpdateRequest); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"error":   "参数错误",
			"message": "数据格式错误",
		})
		return
	}
	if channelUpdateRequest.Name == "" {
		c.JSON(http.StatusOK, gin.H{
			"status":  "error",
			"message": "名称不能为空",
		})
		return
	}
	if err := channelRepository.UpdateChannel(&channelUpdateRequest); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"status":  "error",
			"message": err.Error(),
		})
		return
	}
	// 刷新channel数据
	flushChannel = true
	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "频道修改成功",
	})
}

// 删除频道
func deleteChannel(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"error": "错误的请求参数"})
		return
	}
	if err = channelRepository.DeleteChannel(id); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"status":  "error",
			"message": err.Error(),
		})
	}
	// 刷新channel数据
	flushChannel = true
	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "频道删除成功",
	})
}

// 获取所有标签
func getTags(c *gin.Context) {
	// 从session中获取标签数据
	if !flushTag {
		session, _ := sessionStore.Get(c.Request, "tag-session")
		if tmp, ok := session.Values["tags"]; ok {
			var tags []db.TagResponse
			json.Unmarshal(tmp.([]byte), &tags)
			c.JSON(http.StatusOK, gin.H{
				"status":  "success",
				"message": "获取标签成功",
				"tags":    tags,
			})
			return
		}
	}

	tags, err := tagRepository.ListTags()
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"status":  "error",
			"message": err.Error(),
		})
		return
	}
	// 将标签数据存储到session中
	session, _ := sessionStore.Get(c.Request, "tag-session")
	session.Values["tags"], _ = json.Marshal(tags)
	flushTag = false
	err = session.Save(c.Request, c.Writer)
	if err != nil {
		flushTag = true
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "获取标签成功",
		"tags":    tags,
	})
}

// 新增标签
func createTag(c *gin.Context) {
	var tagCreateRequest db.TagCreateRequest
	if err := c.ShouldBindJSON(&tagCreateRequest); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"status":  "error",
			"message": "错误的请求参数",
		})
		return
	}
	fmt.Printf("tagInfo: %v\n", tagCreateRequest)
	if tagCreateRequest.Name == "" || len(tagCreateRequest.Channels) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"status":  "error",
			"message": "标签名、频道不能为空",
		})
		return
	}
	// 创建标签
	if err := tagRepository.CreateTag(&tagCreateRequest); err != nil {
		// 关于http.StatusOK状态的使用，能够给出明确提示，且不泄露内部信息的错误，都应该返回http.StatusOK状态码
		c.JSON(http.StatusOK, gin.H{
			"status":  "error",
			"message": err.Error(),
		})
		return
	}
	// 刷新tag数据
	flushTag = true
	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "标签创建成功",
	})
}

// 删除标签
func deleteTag(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"error": "ID 格式错误"})
		return
	}
	if err := tagRepository.DeleteTag(id); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"status":  "error",
			"message": err.Error(),
		})
		return
	}
	// 刷新tag数据
	flushTag = true
	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "标签删除成功",
	})
}

// 生成标题
func generateTitle(c *gin.Context) {
	var titleRequest TitleRequest
	if err := c.ShouldBind(&titleRequest); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"status":  "error",
			"message": "错误的请求参数",
		})
		return
	}
	if utf8.RuneCountInString(titleRequest.Theme) > 100 {
		c.JSON(http.StatusOK, gin.H{
			"status":  "error",
			"message": "标题长度不能超过100个字符",
		})
		return
	}
	channelSession, _ := sessionStore.Get(c.Request, "channel-session")
	// 从session中获取指定的频道下的标签信息
	var tagIds []int64
	if tmp, ok := channelSession.Values["channels"]; ok {
		var channels []db.ChannelResponse
		json.Unmarshal(tmp.([]byte), &channels)
		for _, channel := range channels {
			if channel.Id == int64(titleRequest.Channel) {
				tagIds = channel.Tags
				break
			}
		}
	}
	fmt.Printf("tagIds: %v\n", tagIds)
	var finalTitle string = titleRequest.Theme
	if len(tagIds) > 0 {
		// 从session中获取标签名称
		needTags := make([]string, 0)
		tagSession, _ := sessionStore.Get(c.Request, "tag-session")
		if tmp, ok := tagSession.Values["tags"]; ok {
			var tags []db.TagResponse
			json.Unmarshal(tmp.([]byte), &tags)
			for _, tagId := range tagIds {
				for _, tag := range tags {
					if tag.Id == int64(tagId) {
						needTags = append(needTags, tag.Name)
					}
				}
			}
		}
		for utf8.RuneCountInString(finalTitle) < 100 {
			// 从needTags中随机选择一个标签
			tmpIndex := rand.Intn(len(needTags))
			tmp := " #" + needTags[tmpIndex]
			// 从needTags中删除已选择的标签
			needTags = append(needTags[:tmpIndex], needTags[tmpIndex+1:]...)
			if utf8.RuneCountInString(finalTitle)+utf8.RuneCountInString(tmp) > 100 {
				break
			}
			finalTitle += tmp
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "生成标题成功",
		"title":   finalTitle,
	})
}
