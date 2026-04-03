package fiber_handle

import (
	"fmt"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	recover2 "github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/xsxdot/gokit/logger"
	"go.uber.org/zap"
)

// AppConfig 为 Fiber 应用级可选行为。当前字段用于反向代理 / CDN 场景下客户端 IP（c.IP 等）的可信来源。
//
// 零值表示不启用 TrustedProxyCheck，与 Fiber 默认一致。
// 若服务部署在 Cloudflare、Nginx 等之后且依赖真实客户端 IP，应显式设置 EnableTrustedProxyCheck 及 TrustedProxies、ProxyHeader；
// TrustedProxies 请配置为实际反代或 CDN 出口网段，避免滥用 0.0.0.0/0 导致任意来源伪造 IP。
type AppConfig struct {
	EnableTrustedProxyCheck bool   `json:"enableTrustedProxyCheck"`
	TrustedProxies          string `json:"trustedProxies"`
	ProxyHeader             string `json:"proxyHeader"`
}

// NewApp 使用 cfg 创建 Fiber 应用（默认 BodyLimit、统一 ErrorHandler、CORS、recover、健康检查、静态资源等）。
func NewApp(cfg AppConfig) *fiber.App {
	fc := fiber.Config{
		BodyLimit:    10 * 1024 * 1024,
		ErrorHandler: ErrHandler,
	}
	if cfg.EnableTrustedProxyCheck {
		fc.EnableTrustedProxyCheck = true
		fc.TrustedProxies = strings.Split(cfg.TrustedProxies, ",")
		fc.ProxyHeader = cfg.ProxyHeader
	}
	app := fiber.New(fc)
	app.Use(Cors())
	app.Use(recover2.New(recover2.Config{
		Next:             nil,
		EnableStackTrace: true,
		StackTraceHandler: func(c *fiber.Ctx, e interface{}) {
			fmt.Println(e)
		},
	}))
	//app.Use(pprof.New())
	app.Use(HealthCheck(HealthCheckConfig{Path: "/health"}))

	RegisterStaticFiles(app, "./web", "/")

	return app
}

// GetApp 等价于 NewApp(AppConfig{})：不启用信任代理，与 Fiber 默认 IP 语义一致。
func GetApp() *fiber.App {
	return NewApp(AppConfig{})
}

// RegisterStaticFiles 配置静态文件服务，用于提供Vue打包的前端页面
func RegisterStaticFiles(app *fiber.App, staticPath string, prefixPath string) {
	if staticPath == "" {
		return
	}

	// 注册静态文件目录
	app.Static(prefixPath, staticPath, fiber.Static{
		Compress:      true,             // 启用压缩
		ByteRange:     true,             // 启用字节范围请求
		Browse:        false,            // 禁止目录浏览
		Index:         "index.html",     // 默认索引文件
		CacheDuration: 10 * time.Minute, // 设置缓存时间（例如10天）
	})

	// 处理SPA路由 - 将所有不匹配的请求重定向到index.html
	// 注意：这应该在API路由配置之后添加
	app.Get("*", func(c *fiber.Ctx) error {
		// 检查请求的路径是否以API路径开头，如果是API请求则跳过
		path := c.Path()
		if strings.HasPrefix(path, "/api") {
			return c.Next()
		}
		if strings.HasPrefix(path, "/admin") {
			return c.Next()
		}
		if strings.HasPrefix(path, "/client") {
			return c.Next()
		}
		if strings.HasPrefix(path, "/third") || strings.HasPrefix(path, "/internal") || strings.HasPrefix(path, "/gateway") {
			return c.Next()
		}
		if strings.HasPrefix(path, "/health") {
			return c.Next()
		}

		// 静态资源文件检查（不重定向常见的静态资源请求）
		if len(path) > 0 {
			ext := getFileExtension(path)
			if ext == ".js" || ext == ".css" || ext == ".png" || ext == ".jpg" ||
				ext == ".jpeg" || ext == ".gif" || ext == ".svg" || ext == ".ico" ||
				ext == ".woff" || ext == ".woff2" || ext == ".ttf" || ext == ".eot" {
				return c.Next()
			}
		}

		// 重定向到index.html以支持SPA路由
		return c.SendFile(staticPath + "/index.html")
	})

	logger.GetLogger().Info("已注册静态文件服务", zap.String("路径", staticPath), zap.String("前缀", prefixPath))
}

// 辅助函数：获取文件扩展名
func getFileExtension(path string) string {
	for i := len(path) - 1; i >= 0; i-- {
		if path[i] == '.' {
			return path[i:]
		}
		if path[i] == '/' {
			break
		}
	}
	return ""
}
