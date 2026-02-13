package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"go-syncflow/internal/middleware"
	"go-syncflow/internal/models"
	"go-syncflow/internal/storage"
)

// ========== API 调用日志 ==========

// ListAPIAccessLogs 查询 API 调用日志
func ListAPIAccessLogs(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))
	authType := c.Query("authType")
	appID := c.Query("appId")
	method := c.Query("method")
	statusCode := c.Query("statusCode")
	ip := c.Query("ip")
	path := c.Query("path")
	startDate := c.Query("startDate")
	endDate := c.Query("endDate")

	if page < 1 {
		page = 1
	}

	query := storage.DB.Model(&models.APIAccessLog{})
	if authType != "" {
		query = query.Where("auth_type = ?", authType)
	}
	if appID != "" {
		query = query.Where("app_id = ?", appID)
	}
	if method != "" {
		query = query.Where("method = ?", method)
	}
	if statusCode != "" {
		sc, _ := strconv.Atoi(statusCode)
		query = query.Where("status_code = ?", sc)
	}
	if ip != "" {
		query = query.Where("ip LIKE ?", "%"+ip+"%")
	}
	if path != "" {
		query = query.Where("path LIKE ?", "%"+path+"%")
	}
	if startDate != "" {
		query = query.Where("created_at >= ?", startDate+" 00:00:00")
	}
	if endDate != "" {
		query = query.Where("created_at <= ?", endDate+" 23:59:59")
	}

	var total int64
	query.Count(&total)

	var logs []models.APIAccessLog
	query.Order("created_at DESC").Offset((page - 1) * size).Limit(size).Find(&logs)

	respondOK(c, gin.H{"list": logs, "total": total})
}

// GetAPIAccessLogStats 获取 API 调用统计
func GetAPIAccessLogStats(c *gin.Context) {
	today := time.Now().Format("2006-01-02")

	// 今日总调用数
	var todayTotal int64
	storage.DB.Model(&models.APIAccessLog{}).Where("created_at >= ?", today+" 00:00:00").Count(&todayTotal)

	// 今日成功数
	var todaySuccess int64
	storage.DB.Model(&models.APIAccessLog{}).Where("created_at >= ? AND status_code < 400", today+" 00:00:00").Count(&todaySuccess)

	// 平均响应时间
	var avgDuration float64
	storage.DB.Model(&models.APIAccessLog{}).Where("created_at >= ?", today+" 00:00:00").Select("COALESCE(AVG(duration),0)").Scan(&avgDuration)

	// Top5 AppID
	type appStat struct {
		AppID string `json:"appId"`
		Count int64  `json:"count"`
	}
	var topApps []appStat
	storage.DB.Model(&models.APIAccessLog{}).
		Select("app_id, COUNT(*) as count").
		Where("created_at >= ? AND app_id != ''", today+" 00:00:00").
		Group("app_id").
		Order("count DESC").
		Limit(5).
		Find(&topApps)

	successRate := float64(0)
	if todayTotal > 0 {
		successRate = float64(todaySuccess) / float64(todayTotal) * 100
	}

	respondOK(c, gin.H{
		"todayTotal":   todayTotal,
		"successRate":  fmt.Sprintf("%.1f%%", successRate),
		"avgDuration":  fmt.Sprintf("%.0fms", avgDuration),
		"topApps":      topApps,
	})
}

// ========== API 调用日志记录中间件 ==========

// APIAccessLogMiddleware 记录 API 调用日志的中间件
func APIAccessLogMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// 只记录 /api/open 路径的调用
		if !strings.HasPrefix(c.Request.URL.Path, "/api/open") {
			c.Next()
			return
		}

		c.Next()

		// 异步写入日志
		go func() {
			duration := int(time.Since(start).Milliseconds())

			authType := "jwt"
			appID := ""
			var apiKeyID uint
			var userID uint
			username := ""

			if v, exists := c.Get("apiKeyId"); exists {
				authType = "apikey"
				apiKeyID, _ = v.(uint)
			}
			if v, exists := c.Get("apiKeyAppID"); exists {
				appID, _ = v.(string)
			}
			if v, exists := c.Get("userID"); exists {
				userID, _ = v.(uint)
			}
			if v, exists := c.Get("username"); exists {
				username, _ = v.(string)
			}
			if username == "" && appID != "" {
				username = appID
			}

			// 脱敏请求体
			reqBody := ""
			if body, exists := c.Get("requestBody"); exists {
				reqBody = sanitizeBody(body.(string))
			}

			// 脱敏查询参数
			query := sanitizeQuery(c.Request.URL.RawQuery)

			errMsg := ""
			if c.Writer.Status() >= 400 {
				if v, exists := c.Get("errorMessage"); exists {
					errMsg, _ = v.(string)
				}
			}

			storage.DB.Create(&models.APIAccessLog{
				AuthType:     authType,
				AppID:        appID,
				APIKeyID:     apiKeyID,
				UserID:       userID,
				Username:     username,
				Method:       c.Request.Method,
				Path:         c.Request.URL.Path,
				Query:        query,
				RequestBody:  reqBody,
				StatusCode:   c.Writer.Status(),
				ResponseSize: c.Writer.Size(),
				IP:           c.ClientIP(),
				UserAgent:    c.Request.UserAgent(),
				Duration:     duration,
				ErrorMessage: errMsg,
			})
		}()
	}
}

// sanitizeBody 脱敏请求体
func sanitizeBody(body string) string {
	if len(body) > 2048 {
		body = body[:2048] + "...(truncated)"
	}

	// 替换敏感字段
	sensitiveFields := regexp.MustCompile(`(?i)"(password|secret|appKey|appSecret|token|bindPassword|dbPassword)":\s*"[^"]*"`)
	body = sensitiveFields.ReplaceAllString(body, `"$1":"***"`)

	return body
}

// sanitizeQuery 脱敏查询参数
func sanitizeQuery(query string) string {
	if query == "" {
		return ""
	}
	sensitiveParams := regexp.MustCompile(`(?i)(token|secret|password|key)=[^&]*`)
	return sensitiveParams.ReplaceAllString(query, "$1=***")
}

// ========== 日志保留设置 ==========

// LogRetentionConfig 日志保留配置
type LogRetentionConfig struct {
	LoginLogDays      int    `json:"loginLogDays"`
	OperationLogDays  int    `json:"operationLogDays"`
	SyncLogDays       int    `json:"syncLogDays"`
	APIAccessLogDays  int    `json:"apiAccessLogDays"`
	SecurityEventDays int    `json:"securityEventDays"`
	AlertLogDays      int    `json:"alertLogDays"`
	LoginAttemptDays  int    `json:"loginAttemptDays"`
	AutoCleanEnabled  bool   `json:"autoCleanEnabled"`
	CleanTime         string `json:"cleanTime"`
}

func defaultLogRetentionConfig() LogRetentionConfig {
	return LogRetentionConfig{
		LoginLogDays:      90,
		OperationLogDays:  90,
		SyncLogDays:       60,
		APIAccessLogDays:  30,
		SecurityEventDays: 180,
		AlertLogDays:      90,
		LoginAttemptDays:  30,
		AutoCleanEnabled:  true,
		CleanTime:         "03:00",
	}
}

// GetLogRetention 获取日志保留配置
func GetLogRetention(c *gin.Context) {
	cfg := getLogRetentionConfig()
	respondOK(c, cfg)
}

// UpdateLogRetention 更新日志保留配置
func UpdateLogRetention(c *gin.Context) {
	var cfg LogRetentionConfig
	if err := c.ShouldBindJSON(&cfg); err != nil {
		respondError(c, http.StatusBadRequest, "参数错误")
		return
	}

	data, _ := json.Marshal(cfg)
	storage.SetConfig("log_retention", string(data))

	// 重新配置清理调度器
	RestartLogCleanupScheduler()

	middleware.RecordOperationLog(c, "日志管理", "更新日志保留配置", "", "")
	respondOK(c, nil)
}

// CleanLogsNow 立即执行日志清理
func CleanLogsNow(c *gin.Context) {
	result := cleanExpiredLogs()
	middleware.RecordOperationLog(c, "日志管理", "手动清理日志", "", result)
	respondOK(c, gin.H{"message": result})
}

// GetLogRetentionStats 获取各类日志统计
func GetLogRetentionStats(c *gin.Context) {
	type logStat struct {
		Type     string `json:"type"`
		Label    string `json:"label"`
		Count    int64  `json:"count"`
		SizeMB   string `json:"sizeMb"`
	}

	tables := []struct {
		model interface{}
		label string
		typ   string
	}{
		{&models.LoginLog{}, "登录日志", "loginLog"},
		{&models.OperationLog{}, "操作日志", "operationLog"},
		{&models.SyncLog{}, "同步日志", "syncLog"},
		{&models.APIAccessLog{}, "API调用日志", "apiAccessLog"},
		{&models.SecurityEvent{}, "安全事件", "securityEvent"},
		{&models.AlertLog{}, "告警日志", "alertLog"},
		{&models.LoginAttempt{}, "登录尝试", "loginAttempt"},
	}

	stats := make([]logStat, 0, len(tables))
	for _, t := range tables {
		var count int64
		storage.DB.Model(t.model).Count(&count)

		// 估算磁盘占用（SQLite 无法精确获取表大小，按平均行大小估算）
		avgRowSize := 300.0 // bytes
		sizeMB := float64(count) * avgRowSize / 1024 / 1024

		stats = append(stats, logStat{
			Type:   t.typ,
			Label:  t.label,
			Count:  count,
			SizeMB: fmt.Sprintf("%.1f", sizeMB),
		})
	}

	respondOK(c, stats)
}

func getLogRetentionConfig() LogRetentionConfig {
	cfg := defaultLogRetentionConfig()
	value, err := storage.GetConfig("log_retention")
	if err == nil && value != "" {
		json.Unmarshal([]byte(value), &cfg)
	}
	return cfg
}

// ========== 自动清理引擎 ==========

var logCleanupTicker *time.Ticker
var logCleanupDone chan bool

// StartLogCleanupScheduler 启动日志清理调度器
func StartLogCleanupScheduler() {
	cfg := getLogRetentionConfig()
	if !cfg.AutoCleanEnabled {
		return
	}

	// 解析清理时间
	cleanTime := cfg.CleanTime
	if cleanTime == "" {
		cleanTime = "03:00"
	}

	// 计算到下一个清理时间的间隔
	now := time.Now()
	parts := strings.SplitN(cleanTime, ":", 2)
	hour, _ := strconv.Atoi(parts[0])
	minute := 0
	if len(parts) > 1 {
		minute, _ = strconv.Atoi(parts[1])
	}

	next := time.Date(now.Year(), now.Month(), now.Day(), hour, minute, 0, 0, now.Location())
	if next.Before(now) {
		next = next.Add(24 * time.Hour)
	}

	delay := next.Sub(now)
	log.Printf("[日志清理] 下次清理时间: %s (%.0f分钟后)", next.Format("2006-01-02 15:04"), delay.Minutes())

	// 先延迟到清理时间，然后每24小时执行一次
	go func() {
		time.Sleep(delay)
		log.Println("[日志清理] 执行定时清理...")
		cleanExpiredLogs()

		logCleanupTicker = time.NewTicker(24 * time.Hour)
		logCleanupDone = make(chan bool)

		for {
			select {
			case <-logCleanupDone:
				return
			case <-logCleanupTicker.C:
				log.Println("[日志清理] 执行定时清理...")
				cleanExpiredLogs()
			}
		}
	}()
}

// RestartLogCleanupScheduler 重启日志清理调度器
func RestartLogCleanupScheduler() {
	if logCleanupTicker != nil {
		logCleanupTicker.Stop()
	}
	if logCleanupDone != nil {
		close(logCleanupDone)
		logCleanupDone = nil
	}
	StartLogCleanupScheduler()
}

// cleanExpiredLogs 清理过期日志
func cleanExpiredLogs() string {
	cfg := getLogRetentionConfig()
	now := time.Now()

	type cleanTarget struct {
		tableName string
		label     string
		days      int
	}

	targets := []cleanTarget{
		{"login_logs", "登录日志", cfg.LoginLogDays},
		{"operation_logs", "操作日志", cfg.OperationLogDays},
		{"sync_logs", "同步日志", cfg.SyncLogDays},
		{"api_access_logs", "API调用日志", cfg.APIAccessLogDays},
		{"security_events", "安全事件", cfg.SecurityEventDays},
		{"alert_logs", "告警日志", cfg.AlertLogDays},
		{"login_attempts", "登录尝试", cfg.LoginAttemptDays},
	}

	var results []string
	totalDeleted := int64(0)

	for _, t := range targets {
		if t.days <= 0 {
			continue
		}
		cutoff := now.AddDate(0, 0, -t.days)
		result := storage.DB.Exec("DELETE FROM "+t.tableName+" WHERE created_at < ?", cutoff)
		deleted := result.RowsAffected
		if deleted > 0 {
			log.Printf("[日志清理] %s: 删除 %d 条 (保留 %d 天)", t.label, deleted, t.days)
			results = append(results, fmt.Sprintf("%s删除%d条", t.label, deleted))
			totalDeleted += deleted
		}
	}

	if len(results) == 0 {
		return "没有需要清理的日志"
	}
	return fmt.Sprintf("共清理 %d 条: %s", totalDeleted, strings.Join(results, ", "))
}

// ExportAPIAccessLogs 导出API调用日志
func ExportAPIAccessLogs(c *gin.Context) {
	var logs []models.APIAccessLog
	query := storage.DB.Model(&models.APIAccessLog{})

	authType := c.Query("authType")
	appID := c.Query("appId")
	if authType != "" {
		query = query.Where("auth_type = ?", authType)
	}
	if appID != "" {
		query = query.Where("app_id = ?", appID)
	}

	query.Order("created_at DESC").Limit(10000).Find(&logs)

	var b strings.Builder
	b.WriteString("\xEF\xBB\xBF") // UTF-8 BOM
	b.WriteString("ID,时间,认证方式,调用者,方法,路径,状态码,耗时(ms),IP,错误\n")
	for _, l := range logs {
		b.WriteString(fmt.Sprintf("%d,%s,%s,%s,%s,%s,%d,%d,%s,%s\n",
			l.ID,
			l.CreatedAt.Format("2006-01-02 15:04:05"),
			l.AuthType,
			escapeCSVField(l.Username),
			l.Method,
			escapeCSVField(l.Path),
			l.StatusCode,
			l.Duration,
			l.IP,
			escapeCSVField(l.ErrorMessage),
		))
	}

	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", "attachment; filename=api_access_logs.csv")
	c.String(http.StatusOK, b.String())
}

func escapeCSVField(s string) string {
	s = strings.ReplaceAll(s, "\"", "\"\"")
	if strings.ContainsAny(s, ",\"\n\r") {
		return "\"" + s + "\""
	}
	return s
}
