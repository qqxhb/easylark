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

func TestSendText(t *testing.T) {
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

		// 处理发送消息请求
		if r.URL.Path == "/open-apis/im/v1/messages" {
			// 验证请求方法
			if r.Method != "POST" {
				t.Errorf("Expected 'POST' request, got '%s'", r.Method)
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

			// 解析请求体
			var reqBody map[string]interface{}
			json.NewDecoder(r.Body).Decode(&reqBody)

			if reqBody["receive_id"] != "chat123" {
				t.Errorf("Expected receive_id 'chat123', got '%s'", reqBody["receive_id"])
			}

			if reqBody["msg_type"] != "text" {
				t.Errorf("Expected msg_type 'text', got '%s'", reqBody["msg_type"])
			}

			content, ok := reqBody["content"].(map[string]interface{})
			if !ok {
				t.Error("Expected content to be a map")
			} else if content["text"] != "Hello, world!" {
				t.Errorf("Expected text 'Hello, world!', got '%s'", content["text"])
			}

			// 返回模拟响应
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"code": 0,
				"msg": "ok",
				"data": map[string]interface{}{
					"message_id": "om_abcdef123456",
				},
			})
		}
	}))
	defer server.Close()

	// 设置测试URL
	testBaseURL = server.URL

	// 创建自定义客户端
	client := &Client{
		AppID:     "test-app-id",
		AppSecret: "test-app-secret",
		httpClient: &http.Client{},
		tenantAccessToken: "test-token",
		tokenExpireTime:   time.Now().Add(7200 * time.Second),
	}
	
	// 初始化消息服务
	client.Message = newMessageService(client)

	// 构造请求URL和请求体
	url := testBaseURL + "/open-apis/im/v1/messages?receive_id_type=chat_id"
	reqBody := map[string]interface{}{
		"receive_id": "chat123",
		"msg_type":   "text",
		"content": map[string]interface{}{
			"text": "Hello, world!",
		},
	}
	
	// 序列化请求体
	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		t.Fatalf("marshal request body failed: %v", err)
	}
	
	// 创建请求
	req, err := http.NewRequest("POST", url, bytes.NewReader(bodyBytes))
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
			MessageID string `json:"message_id"`
		} `json:"data"`
	}
	
	if err := json.Unmarshal(respBody, &result); err != nil {
		t.Fatalf("unmarshal response body failed: %v", err)
	}
	
	if result.Code != 0 {
		t.Errorf("Expected code 0, got %d: %s", result.Code, result.Msg)
	}
	
	if result.Data.MessageID != "om_abcdef123456" {
		t.Errorf("Expected message_id 'om_abcdef123456', got '%s'", result.Data.MessageID)
	}
}

func TestSendCard(t *testing.T) {
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

		// 处理发送消息请求
		if r.URL.Path == "/open-apis/im/v1/messages" {
			// 验证请求方法
			if r.Method != "POST" {
				t.Errorf("Expected 'POST' request, got '%s'", r.Method)
			}

			// 解析请求体
			var reqBody map[string]interface{}
			json.NewDecoder(r.Body).Decode(&reqBody)

			if reqBody["receive_id"] != "chat123" {
				t.Errorf("Expected receive_id 'chat123', got '%s'", reqBody["receive_id"])
			}

			if reqBody["msg_type"] != "interact" {
				t.Errorf("Expected msg_type 'interact', got '%s'", reqBody["msg_type"])
			}

			// 返回模拟响应
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"code": 0,
				"msg": "ok",
				"data": map[string]interface{}{
					"message_id": "om_abcdef123456",
				},
			})
		}
	}))
	defer server.Close()

	// 设置测试URL
	testBaseURL = server.URL

	// 创建自定义客户端
	client := &Client{
		AppID:     "test-app-id",
		AppSecret: "test-app-secret",
		httpClient: &http.Client{},
		tenantAccessToken: "test-token",
		tokenExpireTime:   time.Now().Add(7200 * time.Second),
	}
	
	// 初始化消息服务
	client.Message = newMessageService(client)

	// 创建卡片消息
	card := NewMessageCard().SetTitle("测试卡片").AddText("这是一个测试卡片")
	
	// 构造请求URL和请求体
	url := testBaseURL + "/open-apis/im/v1/messages?receive_id_type=chat_id"
	reqBody := map[string]interface{}{
		"receive_id": "chat123",
		"msg_type":   card.Type(),
		"content":    card.Content(),
	}
	
	// 序列化请求体
	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		t.Fatalf("marshal request body failed: %v", err)
	}
	
	// 创建请求
	req, err := http.NewRequest("POST", url, bytes.NewReader(bodyBytes))
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
			MessageID string `json:"message_id"`
		} `json:"data"`
	}
	
	if err := json.Unmarshal(respBody, &result); err != nil {
		t.Fatalf("unmarshal response body failed: %v", err)
	}
	
	if result.Code != 0 {
		t.Errorf("Expected code 0, got %d: %s", result.Code, result.Msg)
	}
	
	if result.Data.MessageID != "om_abcdef123456" {
		t.Errorf("Expected message_id 'om_abcdef123456', got '%s'", result.Data.MessageID)
	}
}

func TestGetMessage(t *testing.T) {
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

		// 处理获取消息请求
		if r.URL.Path == "/open-apis/im/v1/messages/om_abcdef123456" {
			// 验证请求方法
			if r.Method != "GET" {
				t.Errorf("Expected 'GET' request, got '%s'", r.Method)
			}

			// 返回模拟响应
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"code": 0,
				"msg": "ok",
				"data": map[string]interface{}{
					"message_id": "om_abcdef123456",
					"chat_id": "chat123",
					"msg_type": "text",
					"content": map[string]interface{}{
						"text": "Hello, world!",
					},
					"create_time": 1609459200,
				},
			})
		}
	}))
	defer server.Close()

	// 设置测试URL
	testBaseURL = server.URL

	// 创建自定义客户端
	client := &Client{
		AppID:     "test-app-id",
		AppSecret: "test-app-secret",
		httpClient: &http.Client{},
		tenantAccessToken: "test-token",
		tokenExpireTime:   time.Now().Add(7200 * time.Second),
	}
	
	// 初始化消息服务
	client.Message = newMessageService(client)

	// 构造请求URL
	url := testBaseURL + "/open-apis/im/v1/messages/om_abcdef123456"
	
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
		Code int                    `json:"code"`
		Msg  string                 `json:"msg"`
		Data map[string]interface{} `json:"data"`
	}
	
	if err := json.Unmarshal(respBody, &result); err != nil {
		t.Fatalf("unmarshal response body failed: %v", err)
	}
	
	if result.Code != 0 {
		t.Errorf("Expected code 0, got %d: %s", result.Code, result.Msg)
	}
	
	if result.Data["message_id"] != "om_abcdef123456" {
		t.Errorf("Expected message_id 'om_abcdef123456', got '%s'", result.Data["message_id"])
	}

	if result.Data["chat_id"] != "chat123" {
		t.Errorf("Expected chat_id 'chat123', got '%s'", result.Data["chat_id"])
	}
}

func TestCreateGroup(t *testing.T) {
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

		// 处理创建群组请求
		if r.URL.Path == "/open-apis/im/v1/chats" {
			// 验证请求方法
			if r.Method != "POST" {
				t.Errorf("Expected 'POST' request, got '%s'", r.Method)
			}

			// 解析请求体
			var reqBody map[string]interface{}
			json.NewDecoder(r.Body).Decode(&reqBody)

			if reqBody["name"] != "测试群组" {
				t.Errorf("Expected name '测试群组', got '%s'", reqBody["name"])
			}

			// 返回模拟响应
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"code": 0,
				"msg": "ok",
				"data": map[string]interface{}{
					"chat_id": "oc_abcdef123456",
				},
			})
		}
	}))
	defer server.Close()

	// 设置测试URL
	testBaseURL = server.URL

	// 创建自定义客户端
	client := &Client{
		AppID:     "test-app-id",
		AppSecret: "test-app-secret",
		httpClient: &http.Client{},
		tenantAccessToken: "test-token",
		tokenExpireTime:   time.Now().Add(7200 * time.Second),
	}
	
	// 初始化消息服务
	client.Message = newMessageService(client)

	// 创建群组请求
	req := &CreateGroupRequest{
		Name:        "测试群组",
		Description: "这是一个测试群组",
		UserIDs:     []string{"user1", "user2"},
	}
	
	// 构造请求URL和请求体
	url := testBaseURL + "/open-apis/im/v1/chats"
	
	// 序列化请求体
	bodyBytes, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("marshal request body failed: %v", err)
	}
	
	// 创建请求
	request, err := http.NewRequest("POST", url, bytes.NewReader(bodyBytes))
	if err != nil {
		t.Fatalf("create request failed: %v", err)
	}
	
	// 设置请求头
	request.Header.Set("Content-Type", "application/json; charset=utf-8")
	request.Header.Set("Authorization", "Bearer test-token")
	
	// 发送请求
	resp, err := client.httpClient.Do(request)
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
			ChatID string `json:"chat_id"`
		} `json:"data"`
	}
	
	if err := json.Unmarshal(respBody, &result); err != nil {
		t.Fatalf("unmarshal response body failed: %v", err)
	}
	
	if result.Code != 0 {
		t.Errorf("Expected code 0, got %d: %s", result.Code, result.Msg)
	}
	
	if result.Data.ChatID != "oc_abcdef123456" {
		t.Errorf("Expected chat_id 'oc_abcdef123456', got '%s'", result.Data.ChatID)
	}
}

func TestAddGroupMember(t *testing.T) {
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

		// 处理添加群成员请求
		if r.URL.Path == "/open-apis/im/v1/chats/oc_abcdef123456/members" {
			// 验证请求方法
			if r.Method != "POST" {
				t.Errorf("Expected 'POST' request, got '%s'", r.Method)
			}

			// 解析请求体
			var reqBody map[string]interface{}
			json.NewDecoder(r.Body).Decode(&reqBody)

			idList, ok := reqBody["id_list"].([]interface{})
			if !ok {
				t.Error("Expected id_list to be an array")
			} else if len(idList) != 2 {
				t.Errorf("Expected id_list to have 2 items, got %d", len(idList))
			}

			// 返回模拟响应
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"code": 0,
				"msg": "ok",
			})
		}
	}))
	defer server.Close()

	// 设置测试URL
	testBaseURL = server.URL

	// 创建自定义客户端
	client := &Client{
		AppID:     "test-app-id",
		AppSecret: "test-app-secret",
		httpClient: &http.Client{},
		tenantAccessToken: "test-token",
		tokenExpireTime:   time.Now().Add(7200 * time.Second),
	}
	
	// 初始化消息服务
	client.Message = newMessageService(client)

	// 构造请求URL和请求体
	url := testBaseURL + "/open-apis/im/v1/chats/oc_abcdef123456/members"
	reqBody := map[string]interface{}{
		"id_list": []string{"user3", "user4"},
	}
	
	// 序列化请求体
	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		t.Fatalf("marshal request body failed: %v", err)
	}
	
	// 创建请求
	request, err := http.NewRequest("POST", url, bytes.NewReader(bodyBytes))
	if err != nil {
		t.Fatalf("create request failed: %v", err)
	}
	
	// 设置请求头
	request.Header.Set("Content-Type", "application/json; charset=utf-8")
	request.Header.Set("Authorization", "Bearer test-token")
	
	// 发送请求
	resp, err := client.httpClient.Do(request)
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
	}
	
	if err := json.Unmarshal(respBody, &result); err != nil {
		t.Fatalf("unmarshal response body failed: %v", err)
	}
	
	if result.Code != 0 {
		t.Errorf("Expected code 0, got %d: %s", result.Code, result.Msg)
	}
}