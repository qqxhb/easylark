package easylark

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"time"
)

// 定义变量用于测试，避免直接修改常量
var (
	// API基础URL
	BaseURL string = "https://open.feishu.cn/open-apis"
	// 获取tenant_access_token的URL
	TenantAccessTokenURL string = "https://open.feishu.cn/open-apis/auth/v3/tenant_access_token/internal"
)

// Client 飞书API客户端
type Client struct {
	AppID     string
	AppSecret string
	httpClient *http.Client
	
	// 认证相关
	tenantAccessToken string
	tokenExpireTime   time.Time
	
	// API服务
	Message *MessageService
	Sheet   *SheetService
}

// NewClient 创建一个新的飞书API客户端
func NewClient(appID, appSecret string) *Client {
	c := &Client{
		AppID:     appID,
		AppSecret: appSecret,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
	
	// 初始化各服务
	c.Message = newMessageService(c)
	c.Sheet = newSheetService(c)
	
	return c
}

// GetTenantAccessToken 获取tenant_access_token
func (c *Client) GetTenantAccessToken() (string, error) {
	// 如果token未过期，直接返回
	if c.tenantAccessToken != "" && time.Now().Before(c.tokenExpireTime) {
		return c.tenantAccessToken, nil
	}
	
	// 构造请求体
	reqBody := map[string]string{
		"app_id":     c.AppID,
		"app_secret": c.AppSecret,
	}
	
	body, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("marshal request body failed: %w", err)
	}
	
	// 发送请求
	req, err := http.NewRequest("POST", TenantAccessTokenURL, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("create request failed: %w", err)
	}
	
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("send request failed: %w", err)
	}
	defer resp.Body.Close()
	
	// 解析响应
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read response body failed: %w", err)
	}
	
	var result struct {
		Code              int    `json:"code"`
		Msg               string `json:"msg"`
		TenantAccessToken string `json:"tenant_access_token"`
		Expire            int    `json:"expire"`
	}
	
	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", fmt.Errorf("unmarshal response body failed: %w", err)
	}
	
	// 检查响应状态
	if result.Code != 0 {
		return "", fmt.Errorf("get tenant_access_token failed: %s", result.Msg)
	}
	
	// 更新token和过期时间
	c.tenantAccessToken = result.TenantAccessToken
	c.tokenExpireTime = time.Now().Add(time.Duration(result.Expire) * time.Second)
	
	return c.tenantAccessToken, nil
}

// DoRequest 发送HTTP请求
func (c *Client) DoRequest(method, path string, body interface{}, result interface{}) error {
	// 获取认证token
	token, err := c.GetTenantAccessToken()
	if err != nil {
		return err
	}
	
	// 构造请求URL
	url := BaseURL + path
	
	// 构造请求体
	var reqBody io.Reader
	if body != nil {
		bodyBytes, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("marshal request body failed: %w", err)
		}
		reqBody = bytes.NewReader(bodyBytes)
	}
	
	// 创建请求
	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		return fmt.Errorf("create request failed: %w", err)
	}
	
	// 设置请求头
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("Authorization", "Bearer "+token)
	
	// 发送请求
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("send request failed: %w", err)
	}
	defer resp.Body.Close()
	
	// 读取响应体
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response body failed: %w", err)
	}
	
	// 解析响应
	if result != nil {
		if err := json.Unmarshal(respBody, result); err != nil {
			return fmt.Errorf("unmarshal response body failed: %w", err)
		}
	}
	
	return nil
}

// APIResponse 通用API响应结构
type APIResponse struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

// Error API错误
type Error struct {
	Code    int
	Message string
}

// Error 实现error接口
func (e *Error) Error() string {
	return fmt.Sprintf("easylark API error: code=%d, message=%s", e.Code, e.Message)
}

// UploadFile 上传文件
func (c *Client) UploadFile(path string, fileBytes []byte, fileName string) (string, error) {
	// 获取认证token
	token, err := c.GetTenantAccessToken()
	if err != nil {
		return "", err
	}
	
	// 构造请求URL
	url := BaseURL + path
	
	// 创建一个buffer用于构造multipart/form-data请求
	var requestBody bytes.Buffer
	multipartWriter := multipart.NewWriter(&requestBody)
	
	// 添加文件部分
	filePart, err := multipartWriter.CreateFormFile("file", fileName)
	if err != nil {
		return "", fmt.Errorf("create form file failed: %w", err)
	}
	
	// 写入文件内容
	if _, err := filePart.Write(fileBytes); err != nil {
		return "", fmt.Errorf("write file content failed: %w", err)
	}
	
	// 关闭multipart writer
	if err := multipartWriter.Close(); err != nil {
		return "", fmt.Errorf("close multipart writer failed: %w", err)
	}
	
	// 创建请求
	req, err := http.NewRequest("POST", url, &requestBody)
	if err != nil {
		return "", fmt.Errorf("create request failed: %w", err)
	}
	
	// 设置请求头
	req.Header.Set("Content-Type", multipartWriter.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+token)
	
	// 发送请求
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("send request failed: %w", err)
	}
	defer resp.Body.Close()
	
	// 读取响应体
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read response body failed: %w", err)
	}
	
	// 解析响应
	var result struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
		Data struct {
			FileKey string `json:"file_key"`
		} `json:"data"`
	}
	
	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", fmt.Errorf("unmarshal response body failed: %w", err)
	}
	
	// 检查响应状态
	if result.Code != 0 {
		return "", fmt.Errorf("upload file failed: %s", result.Msg)
	}
	
	return result.Data.FileKey, nil
}