package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"regexp"

	"ghproxy/config"
	"ghproxy/logger"
	"ghproxy/proxy"

	"github.com/gin-gonic/gin"
)

var cfg *config.Config
var logw = logger.Logw
var router *gin.Engine
var configfile = "/data/ghproxy/config/config.yaml"

var (
	exps = []*regexp.Regexp{
		regexp.MustCompile(`^(?:https?://)?github\.com/([^/]+)/([^/]+)/(?:releases|archive)/.*`),
		regexp.MustCompile(`^(?:https?://)?github\.com/([^/]+)/([^/]+)/(?:blob|raw)/.*`),
		regexp.MustCompile(`^(?:https?://)?github\.com/([^/]+)/([^/]+)/(?:info|git-).*`),
		regexp.MustCompile(`^(?:https?://)?raw\.github(?:usercontent|)\.com/([^/]+)/([^/]+)/.+?/.+`),
		regexp.MustCompile(`^(?:https?://)?gist\.github\.com/([^/]+)/.+?/.+`),
	}
)

func loadConfig() {
	var err error
	// 初始化配置
	cfg, err = config.LoadConfig(configfile)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	fmt.Printf("Loaded config: %v\n", cfg)
}

func setupLogger() {
	// 初始化日志模块
	var err error
	err = logger.Init(cfg.LogFilePath) // 传递日志文件路径
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	logw("Logger initialized")
	logw("Init Completed")
}

func init() {
	loadConfig()
	setupLogger()

	// 设置 Gin 模式
	gin.SetMode(gin.ReleaseMode)

	// 初始化路由
	router = gin.Default()

	// 定义路由
	router.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "https://ghproxy0rtt.1888866.xyz/")
	})

	router.GET("/api", api)

	// 健康检查
	router.GET("/api/healthcheck", func(c *gin.Context) {
		c.String(http.StatusOK, "OK")
	})

	// 未匹配路由处理
	router.NoRoute(func(c *gin.Context) {
		proxy.NoRouteHandler(cfg)(c)
	})
}

func main() {
	// 启动服务器
	err := router.Run(fmt.Sprintf("%s:%d", cfg.Host, cfg.Port))
	if err != nil {
		log.Fatalf("Error starting server: %v\n", err)
	}

	fmt.Println("Program finished")
}

func api(c *gin.Context) {
	// 设置响应头
	c.Writer.Header().Set("Content-Type", "application/json")
	json.NewEncoder(c.Writer).Encode(map[string]interface{}{
		"MaxResponseBodySize": cfg.SizeLimit,
	})
}