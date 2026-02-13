package handlers

import (
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"

	"go-syncflow/internal/dingtalk"
	"go-syncflow/internal/middleware"
	"go-syncflow/internal/models"
	"go-syncflow/internal/storage"
)

func GetUIConfig(c *gin.Context) {
	value, err := storage.GetConfig("ui")
	if err != nil {
		respondError(c, http.StatusInternalServerError, "获取配置失败")
		return
	}

	var cfg models.UIConfig
	json.Unmarshal([]byte(value), &cfg)
	respondOK(c, cfg)
}

func UpdateUIConfig(c *gin.Context) {
	var cfg models.UIConfig
	if err := c.ShouldBindJSON(&cfg); err != nil {
		respondError(c, http.StatusBadRequest, "参数错误")
		return
	}

	data, _ := json.Marshal(cfg)
	if err := storage.SetConfig("ui", string(data)); err != nil {
		respondError(c, http.StatusInternalServerError, "保存失败")
		return
	}

	middleware.RecordOperationLog(c, "系统设置", "更新界面配置", "", "")
	respondOK(c, cfg)
}

// ========== 钉钉配置 ==========

// GetDingTalkConfigFull 获取钉钉完整配置（包含密钥，仅管理员）
func GetDingTalkConfigFull(c *gin.Context) {
	value, _ := storage.GetConfig("dingtalk")

	var cfg models.DingTalkConfig
	if value != "" {
		json.Unmarshal([]byte(value), &cfg)
	}
	respondOK(c, cfg)
}

// UpdateDingTalkConfig 更新钉钉配置
func UpdateDingTalkConfig(c *gin.Context) {
	var cfg models.DingTalkConfig
	if err := c.ShouldBindJSON(&cfg); err != nil {
		respondError(c, http.StatusBadRequest, "参数错误")
		return
	}

	// 如果没传AppSecret，保留原来的
	if cfg.AppSecret == "" {
		oldValue, _ := storage.GetConfig("dingtalk")
		var oldCfg models.DingTalkConfig
		json.Unmarshal([]byte(oldValue), &oldCfg)
		cfg.AppSecret = oldCfg.AppSecret
	}

	data, _ := json.Marshal(cfg)
	if err := storage.SetConfig("dingtalk", string(data)); err != nil {
		respondError(c, http.StatusInternalServerError, "保存失败")
		return
	}

	// 重置token
	dingtalk.GetClient().ResetToken()

	middleware.RecordOperationLog(c, "系统设置", "更新钉钉配置", "", "")
	// 返回时不包含敏感信息
	cfg.AppSecret = ""
	respondOK(c, cfg)
}

// TestDingTalkConnection 测试钉钉连接
func TestDingTalkConnection(c *gin.Context) {
	if err := dingtalk.GetClient().TestConnection(); err != nil {
		respondError(c, http.StatusBadGateway, err.Error())
		return
	}
	respondOK(c, gin.H{"message": "连接成功"})
}

// GetDingTalkStatus 获取钉钉免登状态（公开接口）
func GetDingTalkStatus(c *gin.Context) {
	client := dingtalk.GetClient()
	enabled := client.IsEnabled()

	var corpId, agentId string
	if enabled {
		if cfg, err := client.GetConfig(); err == nil {
			corpId = cfg.CorpID
			agentId = cfg.AgentID
		}
	}

	respondOK(c, gin.H{
		"enabled": enabled,
		"corpId":  corpId,
		"agentId": agentId,
	})
}

// ========== HTTPS配置 ==========

const certDir = "./data/certs"

func GetHTTPSConfig(c *gin.Context) {
	value, _ := storage.GetConfig("https")

	var cfg models.HTTPSConfig
	if value != "" {
		json.Unmarshal([]byte(value), &cfg)
	} else {
		cfg.Port = "8443"
		cfg.Enabled = false
	}

	certExists := false
	keyExists := false
	if cfg.CertFile != "" {
		if _, err := os.Stat(cfg.CertFile); err == nil {
			certExists = true
		}
	}
	if cfg.KeyFile != "" {
		if _, err := os.Stat(cfg.KeyFile); err == nil {
			keyExists = true
		}
	}

	respondOK(c, gin.H{
		"enabled":     cfg.Enabled,
		"port":        cfg.Port,
		"domain":      cfg.Domain,
		"certExpiry":  cfg.CertExpiry,
		"certSubject": cfg.CertSubject,
		"certExists":  certExists,
		"keyExists":   keyExists,
	})
}

func UpdateHTTPSConfig(c *gin.Context) {
	var req struct {
		Enabled bool   `json:"enabled"`
		Port    string `json:"port"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, "参数错误")
		return
	}

	value, _ := storage.GetConfig("https")
	var cfg models.HTTPSConfig
	if value != "" {
		json.Unmarshal([]byte(value), &cfg)
	}

	cfg.Enabled = req.Enabled
	if req.Port != "" {
		cfg.Port = req.Port
	} else if cfg.Port == "" {
		cfg.Port = "8443"
	}

	if cfg.Enabled && (cfg.CertFile == "" || cfg.KeyFile == "") {
		respondError(c, http.StatusBadRequest, "请先上传SSL证书")
		return
	}

	data, _ := json.Marshal(cfg)
	if err := storage.SetConfig("https", string(data)); err != nil {
		respondError(c, http.StatusInternalServerError, "保存失败")
		return
	}

	middleware.RecordOperationLog(c, "系统设置", "更新HTTPS配置", "", "")

	// 自动重启 HTTPS 服务
	if httpsRestarter != nil {
		go func() {
			if err := httpsRestarter.Restart(); err != nil {
				log.Printf("[HTTPS] 重启失败: %v", err)
			}
		}()
	}

	respondOK(c, gin.H{"message": "HTTPS配置已保存，服务正在重启", "config": cfg})
}

// HTTPSRestarter HTTPS 服务热重启接口
type HTTPSRestarter interface {
	Restart() error
}

var httpsRestarter HTTPSRestarter

// SetHTTPSRestarter 注入 HTTPS 重启器（由 main.go 调用）
func SetHTTPSRestarter(r HTTPSRestarter) {
	httpsRestarter = r
}

func UploadSSLCert(c *gin.Context) {
	certFile, err := c.FormFile("cert")
	if err != nil {
		respondError(c, http.StatusBadRequest, "请上传证书文件")
		return
	}

	keyFile, err := c.FormFile("key")
	if err != nil {
		respondError(c, http.StatusBadRequest, "请上传私钥文件")
		return
	}

	if err := os.MkdirAll(certDir, 0755); err != nil {
		respondError(c, http.StatusInternalServerError, "创建证书目录失败")
		return
	}

	certPath := filepath.Join(certDir, "server.crt")
	if err := c.SaveUploadedFile(certFile, certPath); err != nil {
		respondError(c, http.StatusInternalServerError, "保存证书文件失败")
		return
	}

	keyPath := filepath.Join(certDir, "server.key")
	if err := c.SaveUploadedFile(keyFile, keyPath); err != nil {
		respondError(c, http.StatusInternalServerError, "保存私钥文件失败")
		return
	}

	ci, err := parseCertificate(certPath)
	if err != nil {
		os.Remove(certPath)
		os.Remove(keyPath)
		respondError(c, http.StatusBadRequest, "证书无效: "+err.Error())
		return
	}

	value, _ := storage.GetConfig("https")
	var cfg models.HTTPSConfig
	if value != "" {
		json.Unmarshal([]byte(value), &cfg)
	}

	cfg.CertFile = certPath
	cfg.KeyFile = keyPath
	cfg.Domain = ci.Domain
	cfg.CertExpiry = ci.Expiry
	cfg.CertSubject = ci.Subject
	if cfg.Port == "" {
		cfg.Port = "8443"
	}

	data, _ := json.Marshal(cfg)
	if err := storage.SetConfig("https", string(data)); err != nil {
		respondError(c, http.StatusInternalServerError, "保存配置失败")
		return
	}

	middleware.RecordOperationLog(c, "系统设置", "上传SSL证书", "", "")
	respondOK(c, gin.H{
		"domain":     ci.Domain,
		"expiry":     ci.Expiry,
		"subject":    ci.Subject,
		"certExists": true,
		"keyExists":  true,
	})
}

func DeleteSSLCert(c *gin.Context) {
	value, _ := storage.GetConfig("https")
	var cfg models.HTTPSConfig
	if value != "" {
		json.Unmarshal([]byte(value), &cfg)
	}

	if cfg.CertFile != "" {
		os.Remove(cfg.CertFile)
	}
	if cfg.KeyFile != "" {
		os.Remove(cfg.KeyFile)
	}

	cfg.Enabled = false
	cfg.CertFile = ""
	cfg.KeyFile = ""
	cfg.Domain = ""
	cfg.CertExpiry = ""
	cfg.CertSubject = ""

	data, _ := json.Marshal(cfg)
	storage.SetConfig("https", string(data))

	middleware.RecordOperationLog(c, "系统设置", "删除SSL证书", "", "")
	respondOK(c, nil)
}

type certInfo struct {
	Domain  string
	Expiry  string
	Subject string
}

func parseCertificate(certPath string) (*certInfo, error) {
	certPEM, err := os.ReadFile(certPath)
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(certPEM)
	if block == nil {
		return nil, io.EOF
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, err
	}

	domain := ""
	if len(cert.DNSNames) > 0 {
		domain = cert.DNSNames[0]
	} else if cert.Subject.CommonName != "" {
		domain = cert.Subject.CommonName
	}

	return &certInfo{
		Domain:  domain,
		Expiry:  cert.NotAfter.Format(time.RFC3339),
		Subject: cert.Subject.String(),
	}, nil
}

// GetHTTPSConfigForServer 获取HTTPS配置（供服务器启动使用）
func GetHTTPSConfigForServer() *models.HTTPSConfig {
	value, _ := storage.GetConfig("https")
	if value == "" {
		return nil
	}

	var cfg models.HTTPSConfig
	json.Unmarshal([]byte(value), &cfg)

	if !cfg.Enabled || cfg.CertFile == "" || cfg.KeyFile == "" {
		return nil
	}

	if _, err := os.Stat(cfg.CertFile); err != nil {
		return nil
	}
	if _, err := os.Stat(cfg.KeyFile); err != nil {
		return nil
	}

	return &cfg
}

// DownloadDoc 下载/预览系统文档（PDF）
func DownloadDoc(c *gin.Context) {
	name := c.Param("name")
	mode := c.DefaultQuery("mode", "download")

	docMap := map[string]string{
		"technical": "技术架构文档.pdf",
		"manual":    "系统使用手册.pdf",
		"api":       "API接口文档.pdf",
	}

	filename, ok := docMap[name]
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "文档不存在"})
		return
	}

	docPaths := []string{
		filepath.Join("/opt/Go-SyncFlow/docs", filename),
		filepath.Join("..", "docs", filename),
		filepath.Join("docs", filename),
	}

	for _, p := range docPaths {
		if _, err := os.Stat(p); err == nil {
			if mode == "preview" {
				c.Header("Content-Disposition", "inline; filename=\""+filename+"\"")
			} else {
				c.Header("Content-Disposition", "attachment; filename=\""+filename+"\"")
			}
			c.Header("Content-Type", "application/pdf")
			c.File(p)
			return
		}
	}

	c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "文档文件未找到"})
}

// ListDocs 获取可用文档列表
func ListDocs(c *gin.Context) {
	docs := []gin.H{
		{
			"id":          "technical",
			"name":        "技术架构文档",
			"description": "系统功能说明、代码结构、数据库设计和服务依赖需求",
			"filename":    "技术架构文档.pdf",
			"icon":        "Setting",
		},
		{
			"id":          "manual",
			"name":        "系统使用手册",
			"description": "所有功能模块的操作指南和使用示例说明",
			"filename":    "系统使用手册.pdf",
			"icon":        "Document",
		},
		{
			"id":          "api",
			"name":        "API接口文档",
			"description": "完整的REST API调用文档，含示例和错误码",
			"filename":    "API接口文档.pdf",
			"icon":        "Connection",
					},
	}

	respondOK(c, docs)
}
