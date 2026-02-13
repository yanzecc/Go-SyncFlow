package main

import (
	"context"
	"crypto/tls"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"

	"go-syncflow/internal/handlers"
	"go-syncflow/internal/ldapserver"
	"go-syncflow/internal/middleware"
	"go-syncflow/internal/services"
	"go-syncflow/internal/storage"
)

func main() {
	// 设置默认时区为东八区
	loc, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		log.Printf("加载时区失败，使用固定偏移: %v", err)
		loc = time.FixedZone("CST", 8*3600)
	}
	time.Local = loc

	dbPath := getEnv("DB_PATH", "./data/app.db")

	if err := storage.InitDB(dbPath); err != nil {
		log.Fatalf("初始化数据库失败: %v", err)
	}

	// 初始化 IP 白名单
	middleware.InitIPWhitelist(storage.DB)

	// 初始化 RSA 密钥对（密码传输加密）
	services.InitRSAKeyPair()

	// 初始化默认通知通道和模板
	services.InitDefaultChannels()
	services.InitDefaultTemplates()

	router := gin.Default()
	handlers.RegisterRoutes(router)

	// 启动钉钉定时同步（旧方式，保持兼容）
	handlers.StartSyncScheduler()

	// 启动上下游定时同步调度器
	handlers.StartUpstreamSchedulers()
	handlers.StartDownstreamSchedulers()

	// 启动日志清理调度器
	handlers.StartLogCleanupScheduler()

	// 启动 LDAP 服务器
	ldapSrv := ldapserver.NewLDAPServer()
	handlers.SetLDAPServer(ldapSrv)
	ldapCfg := ldapserver.GetLDAPConfig()
	if ldapCfg.Enabled {
		go func() {
			if err := ldapSrv.Start(ldapCfg); err != nil {
				log.Printf("LDAP 服务启动失败: %v", err)
			}
		}()
	}

	// HTTP服务
	httpAddr := getEnv("LISTEN_ADDR", ":8080")

	// HTTPS 热重启管理器
	httpsManager := &HTTPSManager{router: router}
	handlers.SetHTTPSRestarter(httpsManager)

	// 启动时尝试启动 HTTPS
	httpsManager.Start()

	log.Printf("HTTP服务启动: http://localhost%s", httpAddr)
	if err := router.Run(httpAddr); err != nil {
		log.Fatalf("HTTP服务启动失败: %v", err)
	}
}

// HTTPSManager 管理 HTTPS 服务的热重启
type HTTPSManager struct {
	router *gin.Engine
	server *http.Server
}

func (m *HTTPSManager) Start() {
	cfg := handlers.GetHTTPSConfigForServer()
	if cfg == nil {
		log.Println("[HTTPS] 未配置或未启用，跳过")
		return
	}
	addr := ":" + cfg.Port
	tlsCert, err := tls.LoadX509KeyPair(cfg.CertFile, cfg.KeyFile)
	if err != nil {
		log.Printf("[HTTPS] 加载证书失败: %v", err)
		return
	}
	m.server = &http.Server{
		Addr:    addr,
		Handler: m.router,
		TLSConfig: &tls.Config{
			Certificates: []tls.Certificate{tlsCert},
		},
	}
	go func() {
		log.Printf("[HTTPS] 服务启动: https://0.0.0.0%s", addr)
		if err := m.server.ListenAndServeTLS("", ""); err != nil && err != http.ErrServerClosed {
			log.Printf("[HTTPS] 服务异常: %v", err)
		}
	}()
}

func (m *HTTPSManager) Restart() error {
	// 先关闭旧服务
	if m.server != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		m.server.Shutdown(ctx)
		m.server = nil
		log.Println("[HTTPS] 旧服务已关闭")
	}
	// 启动新服务
	m.Start()
	return nil
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
