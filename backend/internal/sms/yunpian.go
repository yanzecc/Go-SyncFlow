package sms

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// YunpianProvider 云片短信实现
type YunpianProvider struct {
	cfg ProviderConfig
}

// NewYunpianProvider 创建云片短信 Provider
func NewYunpianProvider(cfg ProviderConfig) *YunpianProvider {
	return &YunpianProvider{cfg: cfg}
}

// Name 返回提供商名称
func (p *YunpianProvider) Name() string { return "云片短信" }

// Send 发送短信
func (p *YunpianProvider) Send(phone, content string) error {
	return p.SendWithScene(phone, content, "")
}

// SendWithScene 按场景发送短信
func (p *YunpianProvider) SendWithScene(phone, content, scene string) error {
	// 参数校验
	if p.cfg.Apikey == "" {
		return fmt.Errorf("云片 APIKEY 未配置")
	}
	if phone == "" {
		return fmt.Errorf("手机号不能为空")
	}

	// 构造短信内容 - 云片需要在内容开头包含签名
	text := p.buildMessageContent(content)
	if text == "" {
		return fmt.Errorf("短信内容不能为空")
	}

	// 构造请求参数
	reqBody := url.Values{}
	reqBody.Set("apikey", p.cfg.Apikey)
	reqBody.Set("mobile", phone)
	reqBody.Set("text", text)

	// 添加可选参数（仅当有值时才添加）
	if p.cfg.Extend != "" {
		reqBody.Set("extend", p.cfg.Extend)
	}
	if p.cfg.UID != "" {
		reqBody.Set("uid", p.cfg.UID)
	}

	// 发送 HTTP 请求
	apiURL := "https://sms.yunpian.com/v2/sms/single_send.json"
	resp, err := p.sendRequest(apiURL, reqBody)
	if err != nil {
		return err
	}

	// 解析响应
	var result struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
	}

	if err := parseFormResponse(resp, &result); err != nil {
		return fmt.Errorf("解析云片响应失败: %s", string(resp))
	}

	if result.Code != 0 {
		return fmt.Errorf("云片短信发送失败: %s (code: %d)", result.Msg, result.Code)
	}

	return nil
}

// buildMessageContent 构造短信内容，自动添加签名
func (p *YunpianProvider) buildMessageContent(content string) string {
	if content == "" {
		return ""
	}

	// 如果 content 已经包含签名格式，直接返回
	if strings.HasPrefix(content, "【") && strings.Contains(content, "】") {
		return content
	}

	// 否则添加签名
	signName := p.cfg.SignName
	if signName == "" {
		signName = "云片"
	}

	return fmt.Sprintf("【%s】%s", signName, content)
}

// sendRequest 发送 HTTP 请求
func (p *YunpianProvider) sendRequest(apiURL string, reqBody url.Values) ([]byte, error) {
	req, err := http.NewRequest("POST", apiURL, strings.NewReader(reqBody.Encode()))
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %v", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json;charset=utf-8")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求云片短信 API 失败: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %v", err)
	}

	return body, nil
}

// parseFormResponse 解析表单格式响应
func parseFormResponse(body []byte, result interface{}) error {
	params, err := url.ParseQuery(string(body))
	if err != nil {
		return err
	}

	if resultMap, ok := result.(*struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
	}); ok {
		if codeStr := params.Get("code"); codeStr != "" {
			fmt.Sscanf(codeStr, "%d", &resultMap.Code)
		}
		resultMap.Msg = params.Get("msg")
	}

	return nil
}
