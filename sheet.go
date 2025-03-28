package easylark

import (
	"fmt"
)

// SheetService 表格相关API服务
type SheetService struct {
	client *Client
}

// newSheetService 创建表格服务
func newSheetService(client *Client) *SheetService {
	return &SheetService{client: client}
}

// Sheet 表格信息
type Sheet struct {
	SheetToken string                 `json:"sheet_token"`
	Title      string                 `json:"title"`
	Properties map[string]interface{} `json:"properties"`
}

// Get 获取表格元数据
func (s *SheetService) Get(sheetToken string) (*Sheet, error) {
	path := fmt.Sprintf("/sheets/v3/spreadsheets/%s/metainfo", sheetToken)
	
	var result struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
		Data *Sheet `json:"data"`
	}
	
	err := s.client.DoRequest("GET", path, nil, &result)
	if err != nil {
		return nil, err
	}
	
	if result.Code != 0 {
		return nil, &Error{Code: result.Code, Message: result.Msg}
	}
	
	return result.Data, nil
}

// SheetValues 表格值
type SheetValues struct {
	Values [][]interface{} `json:"values"`
}

// ReadRange 读取表格范围内容
func (s *SheetService) ReadRange(sheetToken, rangeStr string) ([][]interface{}, error) {
	path := fmt.Sprintf("/sheets/v3/spreadsheets/%s/values/%s", sheetToken, rangeStr)
	
	var result struct {
		Code int         `json:"code"`
		Msg  string      `json:"msg"`
		Data SheetValues `json:"data"`
	}
	
	err := s.client.DoRequest("GET", path, nil, &result)
	if err != nil {
		return nil, err
	}
	
	if result.Code != 0 {
		return nil, &Error{Code: result.Code, Message: result.Msg}
	}
	
	return result.Data.Values, nil
}

// WriteRange 写入表格范围内容
func (s *SheetService) WriteRange(sheetToken, rangeStr string, values [][]interface{}) error {
	path := fmt.Sprintf("/sheets/v3/spreadsheets/%s/values/%s", sheetToken, rangeStr)
	
	reqBody := map[string]interface{}{
		"values": values,
	}
	
	var result APIResponse
	err := s.client.DoRequest("PUT", path, reqBody, &result)
	if err != nil {
		return err
	}
	
	if result.Code != 0 {
		return &Error{Code: result.Code, Message: result.Msg}
	}
	
	return nil
}

// AppendRange 追加表格内容
func (s *SheetService) AppendRange(sheetToken, rangeStr string, values [][]interface{}) error {
	path := fmt.Sprintf("/sheets/v3/spreadsheets/%s/values/%s:append", sheetToken, rangeStr)
	
	reqBody := map[string]interface{}{
		"values": values,
	}
	
	var result APIResponse
	err := s.client.DoRequest("POST", path, reqBody, &result)
	if err != nil {
		return err
	}
	
	if result.Code != 0 {
		return &Error{Code: result.Code, Message: result.Msg}
	}
	
	return nil
}

// ClearRange 清除表格范围内容
func (s *SheetService) ClearRange(sheetToken, rangeStr string) error {
	path := fmt.Sprintf("/sheets/v3/spreadsheets/%s/values/%s:clear", sheetToken, rangeStr)
	
	var result APIResponse
	err := s.client.DoRequest("POST", path, nil, &result)
	if err != nil {
		return err
	}
	
	if result.Code != 0 {
		return &Error{Code: result.Code, Message: result.Msg}
	}
	
	return nil
}

// AddSheet 添加工作表
func (s *SheetService) AddSheet(sheetToken, title string) (string, error) {
	path := fmt.Sprintf("/sheets/v3/spreadsheets/%s/sheets_batch_update", sheetToken)
	
	reqBody := map[string]interface{}{
		"requests": []map[string]interface{}{
			{
				"addSheet": map[string]interface{}{
					"properties": map[string]interface{}{
						"title": title,
					},
				},
			},
		},
	}
	
	var result struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
		Data struct {
			Replies []struct {
				AddSheet struct {
					Properties struct {
						SheetID string `json:"sheetId"`
					} `json:"properties"`
				} `json:"addSheet"`
			} `json:"replies"`
		} `json:"data"`
	}
	
	err := s.client.DoRequest("POST", path, reqBody, &result)
	if err != nil {
		return "", err
	}
	
	if result.Code != 0 {
		return "", &Error{Code: result.Code, Message: result.Msg}
	}
	
	if len(result.Data.Replies) == 0 || result.Data.Replies[0].AddSheet.Properties.SheetID == "" {
		return "", fmt.Errorf("failed to get sheet ID")
	}
	
	return result.Data.Replies[0].AddSheet.Properties.SheetID, nil
}

// DeleteSheet 删除工作表
func (s *SheetService) DeleteSheet(sheetToken, sheetID string) error {
	path := fmt.Sprintf("/sheets/v3/spreadsheets/%s/sheets_batch_update", sheetToken)
	
	reqBody := map[string]interface{}{
		"requests": []map[string]interface{}{
			{
				"deleteSheet": map[string]interface{}{
					"sheetId": sheetID,
				},
			},
		},
	}
	
	var result APIResponse
	err := s.client.DoRequest("POST", path, reqBody, &result)
	if err != nil {
		return err
	}
	
	if result.Code != 0 {
		return &Error{Code: result.Code, Message: result.Msg}
	}
	
	return nil
}

// SheetInfo 工作表信息
type SheetInfo struct {
	SheetID    string                 `json:"sheetId"`
	Title      string                 `json:"title"`
	Index      int                    `json:"index"`
	Properties map[string]interface{} `json:"properties"`
}

// GetSheets 获取工作表列表
func (s *SheetService) GetSheets(sheetToken string) ([]SheetInfo, error) {
	path := fmt.Sprintf("/sheets/v3/spreadsheets/%s/sheets/query", sheetToken)
	
	var result struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
		Data struct {
			Sheets []SheetInfo `json:"sheets"`
		} `json:"data"`
	}
	
	err := s.client.DoRequest("GET", path, nil, &result)
	if err != nil {
		return nil, err
	}
	
	if result.Code != 0 {
		return nil, &Error{Code: result.Code, Message: result.Msg}
	}
	
	return result.Data.Sheets, nil
}

// CellStyle 单元格样式
type CellStyle struct {
	ForegroundColor *Color                `json:"foregroundColor,omitempty"`
	BackgroundColor *Color                `json:"backgroundColor,omitempty"`
	FontFamily     string                `json:"fontFamily,omitempty"`
	FontSize       int                   `json:"fontSize,omitempty"`
	Bold           bool                  `json:"bold,omitempty"`
	Italic         bool                  `json:"italic,omitempty"`
	Strikethrough  bool                  `json:"strikethrough,omitempty"`
	Underline      bool                  `json:"underline,omitempty"`
	HorizontalAlign string               `json:"horizontalAlign,omitempty"`
	VerticalAlign   string               `json:"verticalAlign,omitempty"`
}

// Color 颜色
type Color struct {
	Red   float64 `json:"red"`
	Green float64 `json:"green"`
	Blue  float64 `json:"blue"`
	Alpha float64 `json:"alpha,omitempty"`
}

// SetCellStyle 设置单元格样式
func (s *SheetService) SetCellStyle(sheetToken, sheetID, rangeStr string, style *CellStyle) error {
	path := fmt.Sprintf("/sheets/v3/spreadsheets/%s/sheets_batch_update", sheetToken)
	
	reqBody := map[string]interface{}{
		"requests": []map[string]interface{}{
			{
				"updateCells": map[string]interface{}{
					"range": map[string]interface{}{
						"sheetId": sheetID,
						"range": rangeStr,
					},
					"style": style,
				},
			},
		},
	}
	
	var result APIResponse
	err := s.client.DoRequest("POST", path, reqBody, &result)
	if err != nil {
		return err
	}
	
	if result.Code != 0 {
		return &Error{Code: result.Code, Message: result.Msg}
	}
	
	return nil
}

// MergeCells 合并单元格
func (s *SheetService) MergeCells(sheetToken, sheetID, rangeStr string) error {
	path := fmt.Sprintf("/sheets/v3/spreadsheets/%s/sheets_batch_update", sheetToken)
	
	reqBody := map[string]interface{}{
		"requests": []map[string]interface{}{
			{
				"mergeCells": map[string]interface{}{
					"range": map[string]interface{}{
						"sheetId": sheetID,
						"range": rangeStr,
					},
				},
			},
		},
	}
	
	var result APIResponse
	err := s.client.DoRequest("POST", path, reqBody, &result)
	if err != nil {
		return err
	}
	
	if result.Code != 0 {
		return &Error{Code: result.Code, Message: result.Msg}
	}
	
	return nil
}

// DimensionType 维度类型
type DimensionType string

const (
	DimensionTypeRow    DimensionType = "ROW"    // 行
	DimensionTypeColumn DimensionType = "COLUMN" // 列
)

// SetDimension 设置行高或列宽
func (s *SheetService) SetDimension(sheetToken, sheetID string, dimensionType DimensionType, startIndex, endIndex int, pixelSize int) error {
	path := fmt.Sprintf("/sheets/v3/spreadsheets/%s/sheets_batch_update", sheetToken)
	
	reqBody := map[string]interface{}{
		"requests": []map[string]interface{}{
			{
				"updateDimensionProperties": map[string]interface{}{
					"range": map[string]interface{}{
						"sheetId": sheetID,
						"dimension": dimensionType,
						"startIndex": startIndex,
						"endIndex": endIndex,
					},
					"properties": map[string]interface{}{
						"pixelSize": pixelSize,
					},
					"fields": "pixelSize",
				},
			},
		},
	}
	
	var result APIResponse
	err := s.client.DoRequest("POST", path, reqBody, &result)
	if err != nil {
		return err
	}
	
	if result.Code != 0 {
		return &Error{Code: result.Code, Message: result.Msg}
	}
	
	return nil
}