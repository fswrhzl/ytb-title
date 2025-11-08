package main

import (
	"embed"
	"io/fs"
	"log"
	"net/http"
	"os"

	"fswrhzl/ytb_title/server"
	"fswrhzl/ytb_title/server/db"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

// 嵌入web资源文件
//
//go:embed all:web/dist/*
var webFiles embed.FS

// 嵌入.env配置文件
//
//go:embed .env
var envFile embed.FS

func main() {
	// 加载环境变量
	if _, err := os.Stat(".env"); err == nil { // 如果当前目录下有.env文件，就加载它
		if err := godotenv.Load(".env"); err != nil {
			log.Fatalf("无法加载外部 .env: %v", err)
		}
	} else { // 如果当前目录下（运行程序所在的目录）没有.env文件，就加载嵌入的.env文件（嵌入的配置文件保持打包编译时的状态）
		envData, err := envFile.ReadFile(".env")
		if err != nil {
			log.Fatalf("读取嵌入的 .env 文件失败: %v", err)
		}
		var envMap map[string]string
		if envMap, err = godotenv.Unmarshal(string(envData)); err != nil {
			log.Fatalf("无法解析内嵌 .env: %v", err)
		} else {
			// 将解析的项目环境变量存入系统环境变量
			for k, v := range envMap {
				os.Setenv(k, v)
			}
		}
	}
	// 初始化数据库
	if err := db.InitDatabase("server/db/data/ytb_title.db"); err != nil {
		panic(err)
	}
	defer db.Close()
	r := server.SetupRouter()
	// 将嵌入的文件系统根定位到 web/dist/assets，使静态路由 /assets
	// 直接映射到构建产物的资源目录，并避免暴露其他非资源文件。
	staticFS, err := fs.Sub(webFiles, "web/dist/assets")
	if err != nil {
		log.Fatalf("读取嵌入的 web/dist/assets 文件失败: %v", err)
	}
	// 提供静态资源路由
	r.StaticFS("/assets", http.FS(staticFS))
	r.NoRoute(func(c *gin.Context) {
		path := c.Request.URL.Path
		if len(path) > 8 && path[:8] == "/assets/" {
			c.Status(http.StatusNotFound)
			return
		}
		indexHTML, err := webFiles.ReadFile("web/dist/index.html")
		if err != nil {
			c.Status(http.StatusInternalServerError)
			return
		}
		c.Data(http.StatusOK, "text/html; charset=utf-8", indexHTML)
	})

	if err := r.Run(":50000"); err != nil {
		panic(err)
	}
}
