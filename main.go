package main

import (
	"embed"
	"io/fs"
	"log"
	"net/http"
	"os"

	"fswrhzl/ytb_title/server"
	"fswrhzl/ytb_title/server/db"

	"github.com/joho/godotenv"
)

// 嵌入web资源文件
//go:embed web/dist/assets/*
var webFiles embed.FS

// 嵌入.env配置文件
//go:embed .env
var envFile embed.FS

func main() {
	// 加载环境变量
	if _, err := os.Stat(".env"); err == nil { // 如果当前目录下有.env文件，就加载它
		_ = godotenv.Load()
	} else { // 如果当前目录下（运行程序所在的目录）没有.env文件，就加载嵌入的.env文件（嵌入的配置文件保持打包编译时的状态）
		data, err := envFile.ReadFile(".env")
		if err != nil {
			log.Fatalf("读取嵌入的 .env 文件失败: %v", err)
		}
		_, _ = godotenv.Unmarshal(string(data))
	}

	// 初始化数据库
	if err := db.InitDatabase("server/db/data/ytb_title.db"); err != nil {
		panic(err)
	}
	defer db.Close()
	r := server.SetupRouter()
	
	// 将嵌入的文件系统根定位到 web/dist/assets，使静态路由 /assets
	// 直接映射到构建产物的资源目录，并避免暴露其他非资源文件。
	subFS, err := fs.Sub(webFiles, "web/dist/assets")
	if err != nil {
		log.Fatalf("读取嵌入的 web/dist/assets 文件失败: %v", err)
	}
	// 注册静态文件服务
	r.StaticFS("/assets", http.FS(subFS))
	if err := r.Run(":60000"); err != nil {
		panic(err)
	}
}
