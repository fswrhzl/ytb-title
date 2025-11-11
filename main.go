package main

import (
	"embed"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"time"

	"fswrhzl/ytb_title/server"
	// "fswrhzl/ytb_title/server/db"
	mGorm "fswrhzl/ytb_title/server/gorm"

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
	loadEnv()
	// 初始化数据库
	if err := mGorm.InitDatabase("server/db/data/ytb_title.db"); err != nil {
		panic(err)
	}
	defer mGorm.Close()
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
	// gin启动后会阻塞Run方法后的代码，放到子协程中启动。
	go func() {
		if err := r.Run(":50000"); err != nil {
			log.Fatalf("服务器启动失败: %v", err)
			panic(err)
		}
	}()
	// 服务启动后自动打开浏览器
	url := "http://127.0.0.1:50000"
	if waitForServer(url, 10) {
		log.Printf("服务器启动成功")
		_ = openBrowser(url)
	}
	// 防止服务器启动前主协程退出
	select {}
}

// 加载环境变量
func loadEnv() {
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
}

func openBrowser(url string) error {
	var cmd string
	var args []string

	switch runtime.GOOS {
	case "windows":
		cmd = "cmd"
		args = []string{"/c", "start"}
	case "darwin":
		cmd = "open"
	default: // Linux 和其他 Unix 系统
		cmd = "xdg-open"
	}

	args = append(args, url)
	return exec.Command(cmd, args...).Start()
}

// 测试服务器是否启动
func waitForServer(url string, maxAttempts int) bool {
	log.Printf("等待服务器启动...")
	for range maxAttempts {
		resp, err := http.Get(url)
		// 响应成功，服务器启动成功
		if err == nil && resp.StatusCode == http.StatusOK {
			resp.Body.Close()
			return true
		}
		time.Sleep(2 * time.Second)
	}
	log.Fatalf("服务器启动超时")
	return false
}
