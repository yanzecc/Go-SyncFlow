package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"go-syncflow/internal/models"
	"go-syncflow/internal/storage"
)

func ListLoginLogs(c *gin.Context) {
	pageIndex, _ := strconv.Atoi(c.DefaultQuery("pageIndex", "0"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	username := c.Query("username")
	status := c.Query("status")

	var logs []models.LoginLog
	var total int64

	ip := c.Query("ip")
	startDate := c.Query("startDate")
	endDate := c.Query("endDate")

	query := storage.DB.Model(&models.LoginLog{})
	if username != "" {
		query = query.Where("username LIKE ?", "%"+username+"%")
	}
	if status != "" {
		s, _ := strconv.Atoi(status)
		query = query.Where("status = ?", s)
	}
	if ip != "" {
		query = query.Where("ip LIKE ?", "%"+ip+"%")
	}
	if startDate != "" {
		query = query.Where("created_at >= ?", startDate+" 00:00:00")
	}
	if endDate != "" {
		query = query.Where("created_at <= ?", endDate+" 23:59:59")
	}

	query.Count(&total)
	query.Offset(pageIndex * pageSize).Limit(pageSize).Order("id desc").Find(&logs)

	respondList(c, logs, total)
}

func ListOperationLogs(c *gin.Context) {
	pageIndex, _ := strconv.Atoi(c.DefaultQuery("pageIndex", "0"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	username := c.Query("username")
	module := c.Query("module")
	startDate := c.Query("startDate")
	endDate := c.Query("endDate")

	var logs []models.OperationLog
	var total int64

	query := storage.DB.Model(&models.OperationLog{})
	if username != "" {
		query = query.Where("username LIKE ?", "%"+username+"%")
	}
	if module != "" {
		query = query.Where("module = ?", module)
	}
	if startDate != "" {
		query = query.Where("created_at >= ?", startDate+" 00:00:00")
	}
	if endDate != "" {
		query = query.Where("created_at <= ?", endDate+" 23:59:59")
	}

	query.Count(&total)
	query.Offset(pageIndex * pageSize).Limit(pageSize).Order("id desc").Find(&logs)

	respondList(c, logs, total)
}

// ListAllSyncLogs 查询所有同步器的同步日志
func ListAllSyncLogs(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))
	status := c.Query("status")
	event := c.Query("event")
	username := c.Query("username")
	startDate := c.Query("startDate")
	endDate := c.Query("endDate")

	var total int64
	var logs []models.SyncLog

	direction := c.Query("direction")

	query := storage.DB.Model(&models.SyncLog{})
	if direction != "" {
		query = query.Where("direction = ?", direction)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if event != "" {
		query = query.Where("trigger_event = ?", event)
	}
	if username != "" {
		query = query.Where("username LIKE ?", "%"+username+"%")
	}
	if startDate != "" {
		query = query.Where("created_at >= ?", startDate+" 00:00:00")
	}
	if endDate != "" {
		query = query.Where("created_at <= ?", endDate+" 23:59:59")
	}

	query.Count(&total)
	query.Order("created_at DESC").Offset((page - 1) * size).Limit(size).Find(&logs)

	respondOK(c, gin.H{"list": logs, "total": total})
}

// ========== 日志导出 (CSV/Excel) ==========

func escapeCSV(s string) string {
	s = strings.ReplaceAll(s, "\"", "\"\"")
	if strings.ContainsAny(s, ",\"\n\r") {
		return "\"" + s + "\""
	}
	return s
}

func fmtTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format("2006-01-02 15:04:05")
}

// ExportLoginLogs 导出登录日志
func ExportLoginLogs(c *gin.Context) {
	var logs []models.LoginLog
	query := storage.DB.Model(&models.LoginLog{})
	username := c.Query("username")
	status := c.Query("status")
	if username != "" {
		query = query.Where("username LIKE ?", "%"+username+"%")
	}
	if status != "" {
		s, _ := strconv.Atoi(status)
		query = query.Where("status = ?", s)
	}
	query.Order("id desc").Limit(10000).Find(&logs)

	// BOM + CSV
	var b strings.Builder
	b.WriteString("\xEF\xBB\xBF") // UTF-8 BOM for Excel
	b.WriteString("ID,用户名,IP地址,状态,备注,时间\n")
	for _, l := range logs {
		statusText := "登录失败"
		if l.Status == 1 {
			statusText = "登录成功"
		}
		b.WriteString(fmt.Sprintf("%d,%s,%s,%s,%s,%s\n",
			l.ID, escapeCSV(l.Username), escapeCSV(l.IP),
			statusText, escapeCSV(l.Message), fmtTime(l.CreatedAt)))
	}

	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", "attachment; filename=login_logs.csv")
	c.String(http.StatusOK, b.String())
}

// ExportOperationLogs 导出操作日志
func ExportOperationLogs(c *gin.Context) {
	var logs []models.OperationLog
	query := storage.DB.Model(&models.OperationLog{})
	username := c.Query("username")
	module := c.Query("module")
	if username != "" {
		query = query.Where("username LIKE ?", "%"+username+"%")
	}
	if module != "" {
		query = query.Where("module = ?", module)
	}
	query.Order("id desc").Limit(10000).Find(&logs)

	var b strings.Builder
	b.WriteString("\xEF\xBB\xBF")
	b.WriteString("ID,用户名,模块,操作,目标,详情,IP地址,时间\n")
	for _, l := range logs {
		b.WriteString(fmt.Sprintf("%d,%s,%s,%s,%s,%s,%s,%s\n",
			l.ID, escapeCSV(l.Username), escapeCSV(l.Module),
			escapeCSV(l.Action), escapeCSV(l.Target), escapeCSV(l.Content),
			escapeCSV(l.IP), fmtTime(l.CreatedAt)))
	}

	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", "attachment; filename=operation_logs.csv")
	c.String(http.StatusOK, b.String())
}

// ExportSyncLogs 导出同步日志
func ExportSyncLogs(c *gin.Context) {
	var logs []models.SyncLog
	query := storage.DB.Model(&models.SyncLog{})
	status := c.Query("status")
	event := c.Query("event")
	username := c.Query("username")
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if event != "" {
		query = query.Where("trigger_event = ?", event)
	}
	if username != "" {
		query = query.Where("username LIKE ?", "%"+username+"%")
	}
	query.Order("created_at DESC").Limit(10000).Find(&logs)

	var b strings.Builder
	b.WriteString("\xEF\xBB\xBF")
	b.WriteString("ID,时间,触发方式,事件,用户,状态,概要,详情,耗时(ms)\n")
	for _, l := range logs {
		b.WriteString(fmt.Sprintf("%d,%s,%s,%s,%s,%s,%s,%s,%d\n",
			l.ID, fmtTime(l.CreatedAt), escapeCSV(l.TriggerType),
			escapeCSV(l.TriggerEvent), escapeCSV(l.Username),
			escapeCSV(l.Status), escapeCSV(l.Message),
			escapeCSV(l.Detail), l.Duration))
	}

	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", "attachment; filename=sync_logs.csv")
	c.String(http.StatusOK, b.String())
}

// ========== 系统日志（登录+操作合并） ==========

type SystemLogEntry struct {
	ID        uint      `json:"id"`
	CreatedAt time.Time `json:"createdAt"`
	Username  string    `json:"username"`
	LogType   string    `json:"logType"`   // login / operation
	Summary   string    `json:"summary"`
	IP        string    `json:"ip"`
	Status    string    `json:"status"`
}

func ListSystemLogs(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))
	logType := c.Query("type")       // login / operation / 空=全部
	keyword := c.Query("keyword")
	startDate := c.Query("startDate")
	endDate := c.Query("endDate")

	if page < 1 {
		page = 1
	}
	offset := (page - 1) * size

	// 构建 WHERE 条件
	var whereParts []string
	var args []interface{}
	if keyword != "" {
		whereParts = append(whereParts, "username LIKE ?")
		args = append(args, "%"+keyword+"%")
	}
	if startDate != "" {
		whereParts = append(whereParts, "created_at >= ?")
		args = append(args, startDate+" 00:00:00")
	}
	if endDate != "" {
		whereParts = append(whereParts, "created_at <= ?")
		args = append(args, endDate+" 23:59:59")
	}
	whereClause := ""
	if len(whereParts) > 0 {
		whereClause = " WHERE " + strings.Join(whereParts, " AND ")
	}

	// UNION 查询：登录日志 + 操作日志
	loginSelect := fmt.Sprintf(
		`SELECT id, created_at, username, 'login' as log_type,
		 CASE WHEN status=1 THEN '登录成功' ELSE '登录失败: ' || COALESCE(message,'') END as summary,
		 ip, CASE WHEN status=1 THEN 'success' ELSE 'failed' END as status
		 FROM login_logs%s`, whereClause)

	opSelect := fmt.Sprintf(
		`SELECT id, created_at, username, 'operation' as log_type,
		 module || ' - ' || action || ': ' || COALESCE(target,'') as summary,
		 ip, 'success' as status
		 FROM operation_logs%s`, whereClause)

	var unionQuery string
	switch logType {
	case "login":
		unionQuery = loginSelect
	case "operation":
		unionQuery = opSelect
	default:
		unionQuery = loginSelect + " UNION ALL " + opSelect
	}

	// 计算总数
	countSQL := fmt.Sprintf("SELECT COUNT(*) FROM (%s)", unionQuery)
	// 合并 args（UNION ALL 时参数要翻倍）
	countArgs := make([]interface{}, 0)
	countArgs = append(countArgs, args...)
	if logType == "" {
		countArgs = append(countArgs, args...)
	}

	var total int64
	storage.DB.Raw(countSQL, countArgs...).Scan(&total)

	// 查询数据
	dataSQL := fmt.Sprintf("SELECT * FROM (%s) AS t ORDER BY created_at DESC LIMIT ? OFFSET ?", unionQuery)
	dataArgs := make([]interface{}, 0)
	dataArgs = append(dataArgs, countArgs...)
	dataArgs = append(dataArgs, size, offset)

	var logs []SystemLogEntry
	storage.DB.Raw(dataSQL, dataArgs...).Scan(&logs)

	respondOK(c, gin.H{"list": logs, "total": total})
}
