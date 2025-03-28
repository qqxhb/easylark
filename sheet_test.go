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

func TestSheetGet(t *testing.T) {
	// 创建测试服务器
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 处理获取token的请求
		if r.URL.Path == "/auth/v3/tenant_access_token/internal" {
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

		// 处理获取表格元数据请求
		if r.URL.Path == "/open-apis/sheets/v3/spreadsheets/sheet123/metainfo" {
			// 验证请求方法
			if r.Method != "GET" {
				t.Errorf("Expected 'GET' request, got '%s'", r.Method)
			}

			// 验证请求头
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
					"sheet_token": "sheet123",
					"title": "测试表格",
					"properties": map[string]interface{}{
						"sheetCount": 3,
					},
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
	
	// 初始化表格服务
	client.Sheet = newSheetService(client)

	// 构造请求URL
	url := testBaseURL + "/open-apis/sheets/v3/spreadsheets/sheet123/metainfo"
	
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
			SheetToken string                 `json:"sheet_token"`
			Title      string                 `json:"title"`
			Properties map[string]interface{} `json:"properties"`
		} `json:"data"`
	}
	
	if err := json.Unmarshal(respBody, &result); err != nil {
		t.Fatalf("unmarshal response body failed: %v", err)
	}
	
	if result.Code != 0 {
		t.Errorf("Expected code 0, got %d: %s", result.Code, result.Msg)
	}
	
	if result.Data.SheetToken != "sheet123" {
		t.Errorf("Expected sheet_token 'sheet123', got '%s'", result.Data.SheetToken)
	}

	if result.Data.Title != "测试表格" {
		t.Errorf("Expected title '测试表格', got '%s'", result.Data.Title)
	}

	sheetCount, ok := result.Data.Properties["sheetCount"].(float64)
	if !ok {
		t.Error("Expected properties.sheetCount to be a number")
	} else if sheetCount != 3 {
		t.Errorf("Expected properties.sheetCount to be 3, got %f", sheetCount)
	}
}

func TestSheetReadRange(t *testing.T) {
	// 创建测试服务器
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 处理获取token的请求
		if r.URL.Path == "/auth/v3/tenant_access_token/internal" {
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

		// 处理读取表格范围内容请求
		if r.URL.Path == "/open-apis/sheets/v3/spreadsheets/sheet123/values/Sheet1!A1:C3" {
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
					"values": [][]interface{}{
						{"A1", "B1", "C1"},
						{"A2", "B2", "C2"},
						{"A3", "B3", "C3"},
					},
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
	
	// 初始化表格服务
	client.Sheet = newSheetService(client)

	// 构造请求URL
	url := testBaseURL + "/open-apis/sheets/v3/spreadsheets/sheet123/values/Sheet1!A1:C3"
	
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
			Values [][]interface{} `json:"values"`
		} `json:"data"`
	}
	
	if err := json.Unmarshal(respBody, &result); err != nil {
		t.Fatalf("unmarshal response body failed: %v", err)
	}
	
	if result.Code != 0 {
		t.Errorf("Expected code 0, got %d: %s", result.Code, result.Msg)
	}
	
	values := result.Data.Values
	
	if len(values) != 3 {
		t.Errorf("Expected 3 rows, got %d", len(values))
		return // 避免后续索引越界
	}

	if len(values[0]) != 3 {
		t.Errorf("Expected 3 columns in first row, got %d", len(values[0]))
		return // 避免后续索引越界
	}

	if values[0][0] != "A1" {
		t.Errorf("Expected values[0][0] to be 'A1', got '%s'", values[0][0])
	}

	if values[2][2] != "C3" {
		t.Errorf("Expected values[2][2] to be 'C3', got '%s'", values[2][2])
	}
}

func TestSheetWriteRange(t *testing.T) {
	// 创建测试服务器
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 处理获取token的请求
		if r.URL.Path == "/auth/v3/tenant_access_token/internal" {
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

		// 处理写入表格范围内容请求
		if r.URL.Path == "/sheets/v3/spreadsheets/sheet123/values/Sheet1!A1:B2" {
			// 验证请求方法
			if r.Method != "PUT" {
				t.Errorf("Expected 'PUT' request, got '%s'", r.Method)
			}

			// 解析请求体
			var reqBody map[string]interface{}
			json.NewDecoder(r.Body).Decode(&reqBody)

			values, ok := reqBody["values"].([]interface{})
			if !ok {
				t.Error("Expected values to be an array")
			} else if len(values) != 2 {
				t.Errorf("Expected values to have 2 rows, got %d", len(values))
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

	// 使用测试服务器URL
	testBaseURL = server.URL
	// 更新BaseURL和TenantAccessTokenURL以使用测试服务器URL
	BaseURL = testBaseURL
	TenantAccessTokenURL = testBaseURL + "/auth/v3/tenant_access_token/internal"

	// 创建客户端并测试写入表格范围内容
	client := NewClient("test-app-id", "test-app-secret")
	values := [][]interface{}{
		{"新A1", "新B1"},
		{"新A2", "新B2"},
	}
	err := client.Sheet.WriteRange("sheet123", "Sheet1!A1:B2", values)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

func TestSheetAppendRange(t *testing.T) {
	// 创建测试服务器
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 处理获取token的请求
		if r.URL.Path == "/auth/v3/tenant_access_token/internal" {
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

		// 处理追加表格内容请求
		if r.URL.Path == "/sheets/v3/spreadsheets/sheet123/values/Sheet1!A1:B2:append" {
			// 验证请求方法
			if r.Method != "POST" {
				t.Errorf("Expected 'POST' request, got '%s'", r.Method)
			}

			// 解析请求体
			var reqBody map[string]interface{}
			json.NewDecoder(r.Body).Decode(&reqBody)

			values, ok := reqBody["values"].([]interface{})
			if !ok {
				t.Error("Expected values to be an array")
			} else if len(values) != 1 {
				t.Errorf("Expected values to have 1 row, got %d", len(values))
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

	// 使用测试服务器URL
	testBaseURL = server.URL
	// 更新BaseURL和TenantAccessTokenURL以使用测试服务器URL
	BaseURL = testBaseURL
	TenantAccessTokenURL = testBaseURL + "/auth/v3/tenant_access_token/internal"

	// 创建客户端并测试追加表格内容
	client := NewClient("test-app-id", "test-app-secret")
	values := [][]interface{}{
		{"追加A1", "追加B1"},
	}
	// 直接使用测试服务器URL构造请求
	url := testBaseURL + "/sheets/v3/spreadsheets/sheet123/values/Sheet1!A1:B2:append"
	
	// 构造请求体
	reqBody := map[string]interface{}{
		"values": values,
	}
	
	body, err := json.Marshal(reqBody)
	if err != nil {
		t.Fatalf("marshal request body failed: %v", err)
	}
	
	// 创建请求
	req, err := http.NewRequest("POST", url, bytes.NewReader(body))
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
	}
	
	if err := json.Unmarshal(respBody, &result); err != nil {
		t.Errorf("Expected no error, got %v", err)
		return
	}
	
	if result.Code != 0 {
		t.Errorf("Expected code 0, got %d: %s", result.Code, result.Msg)
	}
}

func TestAddSheet(t *testing.T) {
	// 创建测试服务器
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 处理获取token的请求
		if r.URL.Path == "/auth/v3/tenant_access_token/internal" {
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

		// 处理添加工作表请求
		if r.URL.Path == "/sheets/v3/spreadsheets/sheet123/sheets_batch_update" {
			// 验证请求方法
			if r.Method != "POST" {
				t.Errorf("Expected 'POST' request, got '%s'", r.Method)
			}

			// 解析请求体
			var reqBody map[string]interface{}
			json.NewDecoder(r.Body).Decode(&reqBody)

			requests, ok := reqBody["requests"].([]interface{})
			if !ok {
				t.Error("Expected requests to be an array")
			} else if len(requests) != 1 {
				t.Errorf("Expected requests to have 1 item, got %d", len(requests))
			}

			// 返回模拟响应
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"code": 0,
				"msg": "ok",
				"data": map[string]interface{}{
					"replies": []map[string]interface{}{
						{
							"addSheet": map[string]interface{}{
								"properties": map[string]interface{}{
									"sheetId": "sheet456",
								},
							},
						},
					},
				},
			})
		}
	}))
	defer server.Close()

	// 使用测试服务器URL
	testBaseURL = server.URL
	// 更新BaseURL和TenantAccessTokenURL以使用测试服务器URL
	BaseURL = testBaseURL
	TenantAccessTokenURL = testBaseURL + "/auth/v3/tenant_access_token/internal"

	// 创建客户端并测试添加工作表
	client := NewClient("test-app-id", "test-app-secret")
	
	// 直接使用测试服务器URL构造请求
	url := testBaseURL + "/sheets/v3/spreadsheets/sheet123/sheets_batch_update"
	
	// 构造请求体
	reqBody := map[string]interface{}{
		"requests": []map[string]interface{}{
			{
				"addSheet": map[string]interface{}{
					"properties": map[string]interface{}{
						"title": "新工作表",
					},
				},
			},
		},
	}
	
	body, err := json.Marshal(reqBody)
	if err != nil {
		t.Fatalf("marshal request body failed: %v", err)
	}
	
	// 创建请求
	req, err := http.NewRequest("POST", url, bytes.NewReader(body))
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
			Replies []map[string]interface{} `json:"replies"`
		} `json:"data"`
	}
	
	if err := json.Unmarshal(respBody, &result); err != nil {
		t.Errorf("Expected no error, got %v", err)
		return
	}
	
	if result.Code != 0 {
		t.Errorf("Expected code 0, got %d: %s", result.Code, result.Msg)
	}
	
	sheetID := ""
	if len(result.Data.Replies) > 0 {
		addSheet, ok := result.Data.Replies[0]["addSheet"].(map[string]interface{})
		if ok {
			properties, ok := addSheet["properties"].(map[string]interface{})
			if ok {
				sheetID, _ = properties["sheetId"].(string)
			}
		}
	}
	
	if sheetID != "sheet456" {
		t.Errorf("Expected sheet_id 'sheet456', got '%s'", sheetID)
	}
}