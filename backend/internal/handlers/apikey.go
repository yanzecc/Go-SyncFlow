package handlers

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"go-syncflow/internal/middleware"
	"go-syncflow/internal/models"
	"go-syncflow/internal/storage"
)

// ========== 辅助函数 ==========

// generateAppID 生成 AppID（格式：ak_前缀 + 16位随机hex）
func generateAppID() string {
	b := make([]byte, 8)
	rand.Read(b)
	return "ak_" + hex.EncodeToString(b)
}

// generateAppKey 生成 AppKey（32位随机hex）
func generateAppKey() string {
	b := make([]byte, 32)
	rand.Read(b)
	return hex.EncodeToString(b)
}

// hashAppKey 对 AppKey 做 SHA256 哈希存储
func hashAppKey(key string) string {
	h := sha256.Sum256([]byte(key))
	return hex.EncodeToString(h[:])
}

// getAppKeyHint 提取 AppKey 后4位作为提示
func getAppKeyHint(key string) string {
	if len(key) < 4 {
		return "****"
	}
	return "****" + key[len(key)-4:]
}

// ========== CRUD Handlers ==========

// ListAPIKeys 获取所有 API Key 列表
func ListAPIKeys(c *gin.Context) {
	var keys []models.APIKey
	query := storage.DB.Order("created_at DESC")

	// 搜索
	if keyword := c.Query("keyword"); keyword != "" {
		like := "%" + keyword + "%"
		query = query.Where("name LIKE ? OR app_id LIKE ? OR description LIKE ?", like, like, like)
	}

	// 状态过滤
	if status := c.Query("status"); status != "" {
		if status == "active" {
			query = query.Where("is_active = ?", true)
		} else if status == "inactive" {
			query = query.Where("is_active = ?", false)
		}
	}

	if err := query.Find(&keys).Error; err != nil {
		respondError(c, http.StatusInternalServerError, "查询失败")
		return
	}

	// 构造返回结果（隐藏敏感字段）
	type apiKeyResp struct {
		models.APIKey
		PermissionList []string `json:"permissionList"`
		IPWhiteList    []string `json:"ipWhiteList"`
		IPBlackList    []string `json:"ipBlackList"`
		IsExpired      bool     `json:"isExpired"`
	}

	result := make([]apiKeyResp, 0)
	now := time.Now()
	for _, k := range keys {
		item := apiKeyResp{APIKey: k}

		// 解析 JSON 字段
		if k.Permissions != "" && k.Permissions != "[]" {
			json.Unmarshal([]byte(k.Permissions), &item.PermissionList)
		}
		if k.IPWhitelist != "" && k.IPWhitelist != "[]" {
			json.Unmarshal([]byte(k.IPWhitelist), &item.IPWhiteList)
		}
		if k.IPBlacklist != "" && k.IPBlacklist != "[]" {
			json.Unmarshal([]byte(k.IPBlacklist), &item.IPBlackList)
		}

		// 检查是否过期
		if k.ExpiresAt != nil && k.ExpiresAt.Before(now) {
			item.IsExpired = true
		}

		result = append(result, item)
	}

	respondOK(c, result)
}

// CreateAPIKey 创建新 API Key
func CreateAPIKey(c *gin.Context) {
	var req struct {
		Name        string   `json:"name" binding:"required"`
		Description string   `json:"description"`
		AppID       string   `json:"appId"`       // 可选自定义
		AppKey      string   `json:"appKey"`      // 可选自定义
		Permissions []string `json:"permissions"` // 权限范围
		IPWhitelist []string `json:"ipWhitelist"`
		IPBlacklist []string `json:"ipBlacklist"`
		RateLimit   int      `json:"rateLimit"`
		ExpiresAt   string   `json:"expiresAt"` // ISO8601 格式
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, "参数错误: "+err.Error())
		return
	}

	// AppID: 自定义或自动生成
	appID := req.AppID
	if appID == "" {
		appID = generateAppID()
	} else {
		// 检查自定义 AppID 格式
		if len(appID) < 4 || len(appID) > 64 {
			respondError(c, http.StatusBadRequest, "AppID 长度须在4-64位之间")
			return
		}
	}

	// 检查 AppID 唯一性
	var count int64
	storage.DB.Model(&models.APIKey{}).Where("app_id = ?", appID).Count(&count)
	if count > 0 {
		respondError(c, http.StatusConflict, "AppID 已存在")
		return
	}

	// AppKey: 自定义或自动生成
	rawAppKey := req.AppKey
	if rawAppKey == "" {
		rawAppKey = generateAppKey()
	} else if len(rawAppKey) < 16 {
		respondError(c, http.StatusBadRequest, "AppKey 长度不得少于16位")
		return
	}

	// 频率限制默认值
	rateLimit := req.RateLimit
	if rateLimit <= 0 {
		rateLimit = 60
	}

	// 序列化 JSON 字段
	permsJSON, _ := json.Marshal(req.Permissions)
	wlJSON, _ := json.Marshal(req.IPWhitelist)
	blJSON, _ := json.Marshal(req.IPBlacklist)

	apiKey := models.APIKey{
		AppID:       appID,
		AppKey:      hashAppKey(rawAppKey),
		AppKeyHint:  getAppKeyHint(rawAppKey),
		Name:        req.Name,
		Description: req.Description,
		Permissions: string(permsJSON),
		IPWhitelist: string(wlJSON),
		IPBlacklist: string(blJSON),
		RateLimit:   rateLimit,
		IsActive:    true,
		CreatedBy:   middleware.GetUserID(c),
	}

	// 过期时间
	if req.ExpiresAt != "" {
		t, err := time.Parse(time.RFC3339, req.ExpiresAt)
		if err != nil {
			t, err = time.Parse("2006-01-02", req.ExpiresAt)
			if err != nil {
				respondError(c, http.StatusBadRequest, "过期时间格式错误")
				return
			}
		}
		apiKey.ExpiresAt = &t
	}

	if err := storage.DB.Create(&apiKey).Error; err != nil {
		log.Printf("[APIKey] 创建失败: %v", err)
		respondError(c, http.StatusInternalServerError, "创建失败")
		return
	}

	middleware.RecordOperationLog(c, "API密钥", "创建API密钥", fmt.Sprintf("AppID: %s, 名称: %s", appID, req.Name), "")

	// 返回创建结果，**首次显示明文 AppKey**
	respondOK(c, gin.H{
		"id":     apiKey.ID,
		"appId":  apiKey.AppID,
		"appKey": rawAppKey, // 仅创建时返回明文
		"name":   apiKey.Name,
		"message": "API密钥创建成功，请妥善保存 AppKey，后续将无法再次查看完整密钥。",
	})
}

// UpdateAPIKey 更新 API Key 信息（不含重置密钥）
func UpdateAPIKey(c *gin.Context) {
	id := c.Param("id")

	var apiKey models.APIKey
	if err := storage.DB.First(&apiKey, id).Error; err != nil {
		respondError(c, http.StatusNotFound, "API密钥不存在")
		return
	}

	var req struct {
		Name        string   `json:"name"`
		Description string   `json:"description"`
		Permissions []string `json:"permissions"`
		IPWhitelist []string `json:"ipWhitelist"`
		IPBlacklist []string `json:"ipBlacklist"`
		RateLimit   int      `json:"rateLimit"`
		IsActive    *bool    `json:"isActive"`
		ExpiresAt   *string  `json:"expiresAt"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, "参数错误")
		return
	}

	updates := map[string]interface{}{}

	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.Description != "" || c.Request.ContentLength > 0 {
		updates["description"] = req.Description
	}
	if req.IsActive != nil {
		updates["is_active"] = *req.IsActive
	}
	if req.RateLimit > 0 {
		updates["rate_limit"] = req.RateLimit
	}

	// 权限
	if req.Permissions != nil {
		permsJSON, _ := json.Marshal(req.Permissions)
		updates["permissions"] = string(permsJSON)
	}

	// IP 白名单
	if req.IPWhitelist != nil {
		wlJSON, _ := json.Marshal(req.IPWhitelist)
		updates["ip_whitelist"] = string(wlJSON)
	}

	// IP 黑名单
	if req.IPBlacklist != nil {
		blJSON, _ := json.Marshal(req.IPBlacklist)
		updates["ip_blacklist"] = string(blJSON)
	}

	// 过期时间
	if req.ExpiresAt != nil {
		if *req.ExpiresAt == "" {
			updates["expires_at"] = nil // 清除过期时间
		} else {
			t, err := time.Parse(time.RFC3339, *req.ExpiresAt)
			if err != nil {
				t, err = time.Parse("2006-01-02", *req.ExpiresAt)
				if err != nil {
					respondError(c, http.StatusBadRequest, "过期时间格式错误")
					return
				}
			}
			updates["expires_at"] = &t
		}
	}

	if len(updates) == 0 {
		respondError(c, http.StatusBadRequest, "无更新内容")
		return
	}

	if err := storage.DB.Model(&apiKey).Updates(updates).Error; err != nil {
		respondError(c, http.StatusInternalServerError, "更新失败")
		return
	}

	middleware.RecordOperationLog(c, "API密钥", "更新API密钥", fmt.Sprintf("AppID: %s", apiKey.AppID), "")
	respondOK(c, gin.H{"message": "更新成功"})
}

// ResetAPIKey 重置 AppKey（生成新密钥）
func ResetAPIKey(c *gin.Context) {
	id := c.Param("id")

	var apiKey models.APIKey
	if err := storage.DB.First(&apiKey, id).Error; err != nil {
		respondError(c, http.StatusNotFound, "API密钥不存在")
		return
	}

	var req struct {
		AppKey string `json:"appKey"` // 可选自定义新密钥
	}
	c.ShouldBindJSON(&req)

	rawAppKey := req.AppKey
	if rawAppKey == "" {
		rawAppKey = generateAppKey()
	} else if len(rawAppKey) < 16 {
		respondError(c, http.StatusBadRequest, "AppKey 长度不得少于16位")
		return
	}

	if err := storage.DB.Model(&apiKey).Updates(map[string]interface{}{
		"app_key":      hashAppKey(rawAppKey),
		"app_key_hint": getAppKeyHint(rawAppKey),
	}).Error; err != nil {
		respondError(c, http.StatusInternalServerError, "重置失败")
		return
	}

	middleware.RecordOperationLog(c, "API密钥", "重置API密钥", fmt.Sprintf("AppID: %s", apiKey.AppID), "")

	respondOK(c, gin.H{
		"appId":   apiKey.AppID,
		"appKey":  rawAppKey,
		"message": "密钥已重置，请妥善保存新的 AppKey。",
	})
}

// DeleteAPIKey 删除 API Key
func DeleteAPIKey(c *gin.Context) {
	id := c.Param("id")

	var apiKey models.APIKey
	if err := storage.DB.First(&apiKey, id).Error; err != nil {
		respondError(c, http.StatusNotFound, "API密钥不存在")
		return
	}

	if err := storage.DB.Delete(&apiKey).Error; err != nil {
		respondError(c, http.StatusInternalServerError, "删除失败")
		return
	}

	middleware.RecordOperationLog(c, "API密钥", "删除API密钥", fmt.Sprintf("AppID: %s, 名称: %s", apiKey.AppID, apiKey.Name), "")
	respondOK(c, gin.H{"message": "删除成功"})
}

// GetAPIKey 获取单个 API Key 详情
func GetAPIKey(c *gin.Context) {
	id := c.Param("id")

	var apiKey models.APIKey
	if err := storage.DB.First(&apiKey, id).Error; err != nil {
		respondError(c, http.StatusNotFound, "API密钥不存在")
		return
	}

	respondOK(c, apiKey)
}

// ToggleAPIKeyStatus 启用/禁用 API Key
func ToggleAPIKeyStatus(c *gin.Context) {
	id := c.Param("id")

	var apiKey models.APIKey
	if err := storage.DB.First(&apiKey, id).Error; err != nil {
		respondError(c, http.StatusNotFound, "API密钥不存在")
		return
	}

	newStatus := !apiKey.IsActive
	if err := storage.DB.Model(&apiKey).Update("is_active", newStatus).Error; err != nil {
		respondError(c, http.StatusInternalServerError, "操作失败")
		return
	}

	action := "启用"
	if !newStatus {
		action = "禁用"
	}
	middleware.RecordOperationLog(c, "API密钥", action+"API密钥", fmt.Sprintf("AppID: %s", apiKey.AppID), "")
	respondOK(c, gin.H{"message": action + "成功", "isActive": newStatus})
}

// ========== API Key 认证中间件 ==========

// APIKeyAuthMiddleware 通过 AppID/AppKey 认证的中间件
func APIKeyAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		appID := c.GetHeader("X-App-ID")
		appKey := c.GetHeader("X-App-Key")

		// 同时支持 query 参数
		if appID == "" {
			appID = c.Query("app_id")
		}
		if appKey == "" {
			appKey = c.Query("app_key")
		}

		if appID == "" || appKey == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "缺少 X-App-ID 或 X-App-Key"})
			c.Abort()
			return
		}

		// 查找 API Key
		var apiKeyRecord models.APIKey
		if err := storage.DB.Where("app_id = ?", appID).First(&apiKeyRecord).Error; err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "AppID 无效"})
			c.Abort()
			return
		}

		// 检查状态
		if !apiKeyRecord.IsActive {
			c.JSON(http.StatusForbidden, gin.H{"success": false, "message": "API密钥已被禁用"})
			c.Abort()
			return
		}

		// 检查过期
		if apiKeyRecord.ExpiresAt != nil && apiKeyRecord.ExpiresAt.Before(time.Now()) {
			c.JSON(http.StatusForbidden, gin.H{"success": false, "message": "API密钥已过期"})
			c.Abort()
			return
		}

		// 验证 AppKey
		if hashAppKey(appKey) != apiKeyRecord.AppKey {
			c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "AppKey 无效"})
			c.Abort()
			return
		}

		clientIP := c.ClientIP()

		// 检查 IP 黑名单
		if apiKeyRecord.IPBlacklist != "" && apiKeyRecord.IPBlacklist != "[]" {
			var blacklist []string
			if json.Unmarshal([]byte(apiKeyRecord.IPBlacklist), &blacklist) == nil {
				for _, ip := range blacklist {
					if matchIP(clientIP, ip) {
						c.JSON(http.StatusForbidden, gin.H{"success": false, "message": "IP 在黑名单中"})
						c.Abort()
						return
					}
				}
			}
		}

		// 检查 IP 白名单（如果配置了白名单则仅允许白名单内的IP）
		if apiKeyRecord.IPWhitelist != "" && apiKeyRecord.IPWhitelist != "[]" {
			var whitelist []string
			if json.Unmarshal([]byte(apiKeyRecord.IPWhitelist), &whitelist) == nil && len(whitelist) > 0 {
				allowed := false
				for _, ip := range whitelist {
					if matchIP(clientIP, ip) {
						allowed = true
						break
					}
				}
				if !allowed {
					c.JSON(http.StatusForbidden, gin.H{"success": false, "message": "IP 不在白名单中"})
					c.Abort()
					return
				}
			}
		}

		// 更新使用统计（异步）
		go func() {
			now := time.Now()
			storage.DB.Model(&models.APIKey{}).Where("id = ?", apiKeyRecord.ID).Updates(map[string]interface{}{
				"last_used_at": now,
				"last_used_ip": clientIP,
				"usage_count":  apiKeyRecord.UsageCount + 1,
			})
		}()

		// 将 API Key 信息注入上下文
		c.Set("apiKeyId", apiKeyRecord.ID)
		c.Set("apiKeyAppId", apiKeyRecord.AppID)
		c.Set("apiKeyName", apiKeyRecord.Name)
		// 标记为 API Key 认证（区别于 JWT 认证）
		c.Set("authType", "apikey")

		// 设置一个管理员用户身份（API Key 以创建者身份执行操作）
		c.Set("userId", apiKeyRecord.CreatedBy)
		c.Set("username", "apikey:"+apiKeyRecord.AppID)

		c.Next()
	}
}

// matchIP 简单 IP 匹配（支持精确匹配和 CIDR 前缀匹配）
func matchIP(clientIP, pattern string) bool {
	if clientIP == pattern {
		return true
	}
	// 简单前缀匹配（如 192.168.1.* 写法）
	if strings.HasSuffix(pattern, ".*") {
		prefix := strings.TrimSuffix(pattern, "*")
		if strings.HasPrefix(clientIP, prefix) {
			return true
		}
	}
	// CIDR 匹配（如 10.0.0.0/8）
	if strings.Contains(pattern, "/") {
		// 简单实现：比较网络前缀
		parts := strings.SplitN(pattern, "/", 2)
		if len(parts) == 2 {
			prefix := parts[0]
			// 粗略匹配，对于 /8, /16, /24
			switch parts[1] {
			case "8":
				p1 := strings.SplitN(prefix, ".", 2)
				c1 := strings.SplitN(clientIP, ".", 2)
				return len(p1) > 0 && len(c1) > 0 && p1[0] == c1[0]
			case "16":
				p1 := strings.SplitN(prefix, ".", 3)
				c1 := strings.SplitN(clientIP, ".", 3)
				return len(p1) >= 2 && len(c1) >= 2 && p1[0] == c1[0] && p1[1] == c1[1]
			case "24":
				p1 := strings.SplitN(prefix, ".", 4)
				c1 := strings.SplitN(clientIP, ".", 4)
				return len(p1) >= 3 && len(c1) >= 3 && p1[0] == c1[0] && p1[1] == c1[1] && p1[2] == c1[2]
			}
		}
	}
	return false
}

// ========== API Key 安全限制中间件 ==========

// APIKeySafetyMiddleware 开放 API 安全限制
// 禁止通过 API 操作 admin 用户、分配超级管理员角色
func APIKeySafetyMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.Request.URL.Path
		method := c.Request.Method

		// 仅对涉及用户操作的路径进行检查
		// 检查是否操作指定用户（/open/users/:id 路径）
		if strings.Contains(path, "/open/users/") && method != "POST" {
			idStr := c.Param("id")
			if idStr != "" {
				var targetUser models.User
				if storage.DB.First(&targetUser, idStr).Error == nil {
					if targetUser.Username == "admin" {
						c.JSON(http.StatusForbidden, gin.H{
							"success": false,
							"message": "API 安全限制：不允许通过 API 操作管理员账户",
						})
						c.Abort()
						return
					}
				}
			}
		}

		// 检查创建用户时是否包含超级管理员角色
		if method == "POST" && strings.HasSuffix(path, "/open/users") {
			if containsSuperAdminRole(c) {
				c.JSON(http.StatusForbidden, gin.H{
					"success": false,
					"message": "API 安全限制：不允许通过 API 分配超级管理员角色",
				})
				c.Abort()
				return
			}
		}

		// 检查更新用户时是否尝试分配超级管理员角色
		if method == "PUT" && strings.Contains(path, "/open/users/") && !strings.Contains(path, "/status") && !strings.Contains(path, "/reset-password") {
			if containsSuperAdminRole(c) {
				c.JSON(http.StatusForbidden, gin.H{
					"success": false,
					"message": "API 安全限制：不允许通过 API 分配超级管理员角色",
				})
				c.Abort()
				return
			}
		}

		c.Next()
	}
}

// containsSuperAdminRole 检查请求体中是否包含超级管理员角色
func containsSuperAdminRole(c *gin.Context) bool {
	// 读取并恢复 body
	bodyBytes, err := io.ReadAll(c.Request.Body)
	if err != nil {
		return false
	}
	c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	var req struct {
		RoleIDs []uint `json:"roleIds"`
	}
	if json.Unmarshal(bodyBytes, &req) != nil || len(req.RoleIDs) == 0 {
		return false
	}

	// 查找 super_admin 角色 ID
	var superAdmin models.Role
	if storage.DB.Where("code = ?", "super_admin").First(&superAdmin).Error != nil {
		return false
	}

	for _, rid := range req.RoleIDs {
		if rid == superAdmin.ID {
			return true
		}
	}
	return false
}
