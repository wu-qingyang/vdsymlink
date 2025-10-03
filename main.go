package main

import (
    "os"
    "fmt"
    "strconv"
    "vdsymlink-web/handlers"

    "github.com/gin-gonic/gin"
)

func main() {
    port := getPort()

    fmt.Printf("🚀 服务器启动在端口: %d\n", port)
    fmt.Printf("📎 访问地址: http://localhost:%d\n", port)

    router := gin.Default()

    // 设置常用内网代理，解决GIN警告
    router.SetTrustedProxies([]string{"127.0.0.1", "192.168.0.0/16", "10.0.0.0/8", "172.16.0.0/12"})

    // 加载静态文件和模板
    router.Static("/static", "./static")
    router.LoadHTMLGlob("templates/*")

    // 初始化处理器
    symlinkHandler := handlers.NewSymlinkHandler()

    // 路由设置
    router.GET("/", symlinkHandler.GetIndex)
    router.POST("/api/process", symlinkHandler.ProcessFiles)
    router.GET("/api/directories", symlinkHandler.ListDirectories)

    // 启动服务器
    router.Run(":" + strconv.Itoa(port))
}

// getPort 从环境变量获取端口号
func getPort() int {
    // 按优先级尝试不同的环境变量
    envVars := []string{
        "PORT",        // 云平台标准 (Heroku, Cloud Foundry, etc.)
        "APP_PORT",    // 通用应用端口
        "WEB_PORT",    // Web 服务端口
        "VD_PORT",     // 本项目特定端口
        "SERVER_PORT", // 服务器端口
    }

    for _, envVar := range envVars {
        if portStr := os.Getenv(envVar); portStr != "" {
            if port, err := strconv.Atoi(portStr); err == nil {
                if isValidPort(port) {
                    return port
                }
            }
        }
    }

    // 默认端口
    return 8080
}

// isValidPort 验证端口号是否有效
func isValidPort(port int) bool {
    return port > 0 && port < 65536
}