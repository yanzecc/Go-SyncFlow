package sms

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// ChuanglanProvider 创蓝短信实现
type ChuanglanProvider struct {
	cfg ProviderConfig
}

// NewChuanglanProvider 创建创蓝短信 Provider
func NewChuanglanProvider(cfg ProviderConfig) *ChuanglanProvider {
	return &ChuanglanProvider{cfg: cfg}
}

// Name 返回提供商名称
func (p *ChuanglanProvider) Name() string { return "创蓝短信" }

// Send 发送短信
func (p *ChuanglanProvider) Send(phone, content string) error {
	return p.SendWithScene(phone, content, "")
}

// SendWithScene 按场景发送短信
func (p *ChuanglanProvider) SendWithScene(phone, content, scene string) error {
	// 参数校验
	if p.cfg.Account == "" {
		return fmt.Errorf("创蓝 API 账号未配置")
	}
	if p.cfg.Password == "" {
		return fmt.Errorf("创蓝 API 密码未配置")
	}
	if p.cfg.SignName == "" {
		return fmt.Errorf("短信签名未配置")
	}

	// 根据场景获取模板 ID
	templateId := p.cfg.GetTemplateCodeForScene(scene)
	if templateId == "" {
		return fmt.Errorf("短信模板 ID 未配置")
	}

	// 构造请求体
	reqBody := map[string]interface{}{
		"account":     p.cfg.Account,
		"password":    p.cfg.Password,
		"phone":       phone,
		"templateId":  templateId,
		"report":      true, // 需要状态回执
		"signName":    p.cfg.SignName,
	}

	// 处理模板参数 - 将 content 拆分为参数数组
	// 创蓝的 params 是数组格式，content 中的参数用逗号分隔
	if content != "" {
		params := strings.Split(content, ",")
		// 清理每个参数中的空格
		for i, param := range params {
			params[i] = strings.TrimSpace(param)
		}
		reqBody["params"] = params
	}

	// 发送 HTTP 请求
	apiURL := "https://smssh.253.com/msg/sms/v2/tpl/send"
	resp, err := p.sendRequest(apiURL, reqBody)
	if err != nil {
		return err
	}

	// 解析响应
	var result struct {
		Code    int    `json:"code"`
		Msg     string `json:"msg"`
		MsgId   string `json:"msgId"`
		Success bool   `json:"success"`
		Data    []struct {
			Phone  string `json:"phone"`
			MsgId  string `json:"msgId"`
			Status int    `json:"status"`
		} `json:"data"`
	}

	if err := json.Unmarshal(resp, &result); err != nil {
		return fmt.Errorf("解析创蓝响应失败: %s", string(resp))
	}

	// 根据不同的响应格式判断结果
	if result.Code != 0 && result.Code != 200 {
		return fmt.Errorf("创蓝短信发送失败: %s (code: %d)", result.Msg, result.Code)
	}

	// 检查 data 数组中的状态
	if len(result.Data) > 0 {
		for _, item := range result.Data {
			if item.Status != 0 {
				return fmt.Errorf("创蓝短信发送失败: 手机号 %s 状态码 %d", item.Phone, item.Status)
			}
		}
	}

	return nil
}

// sendRequest 发送 HTTP 请求
func (p *ChuanglanProvider) sendRequest(apiURL string, reqBody map[string]interface{}) ([]byte, error) {
	payload, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("构造请求体失败: %v", err)
	}

	req, err := http.NewRequest("POST", apiURL, bytes.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %v", err)
	}

	req.Header.Set("Content-Type", "application/json;charset=UTF-8")
	req.Header.Set("Accept", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求创蓝短信 API 失败: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %v", err)
	}

	return body, nil
}

// chuanglanSign 计算创蓝 API 签名
// 签名算法: HMAC-SHA256(password, timestamp + nonce)
func chuanglanSign(password, timestamp, nonce string) string {
	stringToSign := timestamp + nonce
	mac := hmac.New(sha256.New, []byte(password))
	mac.Write([]byte(stringToSign))
	return hex.EncodeToString(mac.Sum(nil))
}
