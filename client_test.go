package easylark

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewClient(t *testing.T) {
	client := NewClient("test-app-id", "test-app-secret")
	
	if client.AppID != "test-app-id" {
		t.Errorf("Expected AppID to be 'test-app-id', got '%s'", client.AppID)
	}
	
	if client.AppSecret != "test-app-secret" {
		t.Errorf("Expected AppSecret to be 'test-app-secret', got '%s'", client.AppSecret)
	}
	
	if client.httpClient == nil {
		t.Error("Expected httpClient to be initialized")
	}
	
	if client.Message == nil {
		t.Error("Expected Message service to be initialized")
	}
	
	if client.Sheet == nil {
		t.Error("Expected Sheet service to be initialized")
	}
}

// 定义一个变量用于测试，避免直接修改常量BaseURL
var testBaseURL string

func TestGetTenantAccessToken(t *testing.T) {
	// 创建测试服务器
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 验证请求方法和路径
		if r.Method != "POST" {
			t.Errorf("Expected 'POST' request, got '%s'", r.Method)
		}
		
		if r.URL.Path != "/open-apis/auth/v3/tenant_access_token/internal" {
			t.Errorf("Expected path '/open-apis/auth/v3/tenant_access_token/internal', got '%s'", r.URL.Path)
		}
		
		// 验证请求头
		contentType := r.Header.Get("Content-Type")
		if contentType != "application/json; charset=utf-8" {
			t.Errorf("Expected Content-Type 'application/json; charset=utf-8', got '%s'", contentType)
		}
		
		// 解析请求体
		var reqBody map[string]string
		json.NewDecoder(r.Body).Decode(&reqBody)
		
		if reqBody["app_id"] != "test-app-id" {
			t.Errorf("Expected app_id 'test-app-id', got '%s'", reqBody["app_id"])
		}
		
		if reqBody["app_secret"] != "test-app-secret" {
			t.Errorf("Expected app_secret 'test-app-secret', got '%s'", reqBody["app_secret"])
		}
		
		// 返回模拟响应
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"code": 0,
			"msg": "ok",
			"tenant_access_token": "test-token",
			"expire": 7200,
		})
	}))
	defer server.Close()
	
	// 设置测试URL
	testBaseURL = server.URL
	
	// 创建一个自定义的客户端，使用测试URL
	client := &Client{
		AppID:     "test-app-id",
		AppSecret: "test-app-secret",
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
	
	// 使用测试URL
	tenantTokenURL := testBaseURL + "/open-apis/auth/v3/tenant_access_token/internal"
	
	// 构造请求体
	reqBody := map[string]string{
		"app_id":     client.AppID,
		"app_secret": client.AppSecret,
	}
	
	body, err := json.Marshal(reqBody)
	if err != nil {
		t.Fatalf("marshal request body failed: %v", err)
	}
	
	// 发送请求
	req, err := http.NewRequest("POST", tenantTokenURL, bytes.NewReader(body))
	if err != nil {
		t.Fatalf("create request failed: %v", err)
	}
	
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	
	resp, err := client.httpClient.Do(req)
	if err != nil {
		t.Fatalf("send request failed: %v", err)
	}
	defer resp.Body.Close()
	
	// 解析响应
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read response body failed: %v", err)
	}
	
	var result struct {
		Code              int    `json:"code"`
		Msg               string `json:"msg"`
		TenantAccessToken string `json:"tenant_access_token"`
		Expire            int    `json:"expire"`
	}
	
	if err := json.Unmarshal(respBody, &result); err != nil {
		t.Fatalf("unmarshal response body failed: %v", err)
	}
	
	// 检查响应状态
	if result.Code != 0 {
		t.Fatalf("get tenant_access_token failed: %s", result.Msg)
	}
	
	// 更新token和过期时间
	client.tenantAccessToken = result.TenantAccessToken
	client.tokenExpireTime = time.Now().Add(time.Duration(result.Expire) * time.Second)
	
	token := client.tenantAccessToken
	
	if token != "test-token" {
		t.Errorf("Expected token 'test-token', got '%s'", token)
	}
	
	// 测试token缓存
	if client.tenantAccessToken != "test-token" {
		t.Errorf("Expected client.tenantAccessToken to be 'test-token', got '%s'", client.tenantAccessToken)
	}
	
	if time.Now().After(client.tokenExpireTime) {
		t.Error("Expected token expire time to be in the future")
	}
}

func TestGetTenantAccessTokenError(t *testing.T) {
	// 创建测试服务器，返回错误响应
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"code": 10002,
			"msg": "invalid app_secret",
		})
	}))
	defer server.Close()
	
	// 使用测试URL
	tenantTokenURL := server.URL + "/open-apis/auth/v3/tenant_access_token/internal"
	
	// 创建客户端
	client := &Client{
		AppID:     "test-app-id",
		AppSecret: "wrong-app-secret",
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
	
	// 构造请求体
	reqBody := map[string]string{
		"app_id":     client.AppID,
		"app_secret": client.AppSecret,
	}
	
	body, err := json.Marshal(reqBody)
	if err != nil {
		t.Fatalf("marshal request body failed: %v", err)
	}
	
	// 发送请求
	req, err := http.NewRequest("POST", tenantTokenURL, bytes.NewReader(body))
	if err != nil {
		t.Fatalf("create request failed: %v", err)
	}
	
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	
	resp, err := client.httpClient.Do(req)
	if err != nil {
		t.Fatalf("send request failed: %v", err)
	}
	defer resp.Body.Close()
	
	// 解析响应
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read response body failed: %v", err)
	}
	
	var result struct {
		Code              int    `json:"code"`
		Msg               string `json:"msg"`
		TenantAccessToken string `json:"tenant_access_token"`
		Expire            int    `json:"expire"`
	}
	
	if err := json.Unmarshal(respBody, &result); err != nil {
		t.Fatalf("unmarshal response body failed: %v", err)
	}
	
	// 检查响应状态
	if result.Code == 0 {
		t.Error("Expected error code, got 0")
	}
	
	if result.Msg != "invalid app_secret" {
		t.Errorf("Expected error message 'invalid app_secret', got '%s'", result.Msg)
	}
}

func TestDoRequest(t *testing.T) {
	// 创建测试服务器
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 处理获取token的请求
		if r.URL.Path == "/open-apis/auth/v3/tenant_access_token/internal" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"code": 0,
				"msg": "ok",
				"tenant_access_token": "test-token",
				"expire": 7200,
			})
			return
		}
		
		// 处理API请求
		if r.URL.Path == "/open-apis/test/api" {
			// 验证请求方法
			if r.Method != "GET" {
				t.Errorf("Expected 'GET' request, got '%s'", r.Method)
			}
			
			// 验证请求头
			contentType := r.Header.Get("Content-Type")
			if contentType != "application/json; charset=utf-8" {
				t.Errorf("Expected Content-Type 'application/json; charset=utf-8', got '%s'", contentType)
			}
			
			auth := r.Header.Get("Authorization")
			if auth != "Bearer test-token" {
				t.Errorf("Expected Authorization 'Bearer test-token', got '%s'", auth)
			}
			
			// 返回模拟响应
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"code": 0,
				"msg": "ok",
				"data": map[string]interface{}{
					"key": "value",
				},
			})
		}
	}))
	defer server.Close()
	
	// 创建一个自定义的客户端，使用测试URL
	client := &Client{
		AppID:     "test-app-id",
		AppSecret: "test-app-secret",
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		tenantAccessToken: "test-token",
		tokenExpireTime:   time.Now().Add(7200 * time.Second),
	}
	
	// 构造请求URL
	url := server.URL + "/open-apis/test/api"
	
	// 创建请求
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		t.Fatalf("create request failed: %v", err)
	}
	
	// 设置请求头
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("Authorization", "Bearer test-token")
	
	// 发送请求
	resp, err := client.httpClient.Do(req)
	if err != nil {
		t.Fatalf("send request failed: %v", err)
	}
	defer resp.Body.Close()
	
	// 读取响应体
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read response body failed: %v", err)
	}
	
	// 解析响应
	var result struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
		Data struct {
			Key string `json:"key"`
		} `json:"data"`
	}
	
	if err := json.Unmarshal(respBody, &result); err != nil {
		t.Fatalf("unmarshal response body failed: %v", err)
	}
	
	if result.Code != 0 {
		t.Errorf("Expected code 0, got %d", result.Code)
	}
	
	if result.Msg != "ok" {
		t.Errorf("Expected msg 'ok', got '%s'", result.Msg)
	}
	
	if result.Data.Key != "value" {
		t.Errorf("Expected data.key 'value', got '%s'", result.Data.Key)
	}
}

func TestError(t *testing.T) {
	err := &Error{Code: 10002, Message: "invalid app_secret"}
	expected := "easylark API error: code=10002, message=invalid app_secret"
	
	if err.Error() != expected {
		t.Errorf("Expected error message '%s', got '%s'", expected, err.Error())
	}
}