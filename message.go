package easylark

import (
	"fmt"
)

// MessageService 消息相关API服务
type MessageService struct {
	client *Client
}

// newMessageService 创建消息服务
func newMessageService(client *Client) *MessageService {
	return &MessageService{client: client}
}

// MessageType 消息类型
type MessageType string

const (
	MessageTypeText     MessageType = "text"     // 文本消息
	MessageTypePost     MessageType = "post"     // 富文本消息
	MessageTypeImage    MessageType = "image"    // 图片消息
	MessageTypeInteract MessageType = "interact" // 消息卡片
)

// MessageContent 消息内容接口
type MessageContent interface {
	Type() MessageType
	Content() map[string]interface{}
}

// TextContent 文本消息内容
type TextContent struct {
	Text string
}

// Type 实现MessageContent接口
func (t *TextContent) Type() MessageType {
	return MessageTypeText
}

// Content 实现MessageContent接口
func (t *TextContent) Content() map[string]interface{} {
	return map[string]interface{}{
		"text": t.Text,
	}
}

// MessageCard 消息卡片
type MessageCard struct {
	title   string
	elements []interface{}
}

// NewMessageCard 创建一个新的消息卡片
func NewMessageCard() *MessageCard {
	return &MessageCard{
		elements: make([]interface{}, 0),
	}
}

// SetTitle 设置卡片标题
func (c *MessageCard) SetTitle(title string) *MessageCard {
	c.title = title
	return c
}

// AddText 添加文本内容
func (c *MessageCard) AddText(text string) *MessageCard {
	c.elements = append(c.elements, map[string]interface{}{
		"tag":  "plain_text",
		"text": text,
	})
	return c
}

// Type 实现MessageContent接口
func (c *MessageCard) Type() MessageType {
	return MessageTypeInteract
}

// Content 实现MessageContent接口
func (c *MessageCard) Content() map[string]interface{} {
	return map[string]interface{}{
		"elements": []interface{}{
			map[string]interface{}{
				"tag": "div",
				"text": map[string]interface{}{
					"tag":     "plain_text",
					"content": c.title,
				},
				"fields": c.elements,
			},
		},
	}
}

// SendMessage 发送消息
func (s *MessageService) SendMessage(chatID string, content MessageContent) error {
	path := "/im/v1/messages?receive_id_type=chat_id"
	
	reqBody := map[string]interface{}{
		"receive_id":      chatID,
		"msg_type":        content.Type(),
		"content":         content.Content(),
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

// SendText 发送文本消息
func (s *MessageService) SendText(chatID, text string) error {
	content := &TextContent{Text: text}
	return s.SendMessage(chatID, content)
}

// SendCard 发送卡片消息
func (s *MessageService) SendCard(chatID string, card *MessageCard) error {
	return s.SendMessage(chatID, card)
}

// GetMessage 获取消息
func (s *MessageService) GetMessage(messageID string) (map[string]interface{}, error) {
	path := fmt.Sprintf("/im/v1/messages/%s", messageID)
	
	var result struct {
		Code int                    `json:"code"`
		Msg  string                 `json:"msg"`
		Data map[string]interface{} `json:"data"`
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

// PostContent 富文本消息内容
type PostContent struct {
	ZhCn *PostBody `json:"zh_cn,omitempty"`
	EnUs *PostBody `json:"en_us,omitempty"`
}

// PostBody 富文本消息体
type PostBody struct {
	Title   string         `json:"title"`
	Content [][]PostElement `json:"content"`
}

// PostElement 富文本元素
type PostElement struct {
	Tag      string                 `json:"tag"`
	Text     string                 `json:"text,omitempty"`
	Href     string                 `json:"href,omitempty"`
	UserId   string                 `json:"user_id,omitempty"`
	Attrs    map[string]interface{} `json:"attrs,omitempty"`
}

// Type 实现MessageContent接口
func (p *PostContent) Type() MessageType {
	return MessageTypePost
}

// Content 实现MessageContent接口
func (p *PostContent) Content() map[string]interface{} {
	return map[string]interface{}{
		"post": map[string]interface{}{
			"zh_cn": p.ZhCn,
			"en_us": p.EnUs,
		},
	}
}

// NewPostContent 创建富文本消息内容
func NewPostContent() *PostContent {
	return &PostContent{}
}

// WithZhCn 设置中文内容
func (p *PostContent) WithZhCn(title string, content [][]PostElement) *PostContent {
	p.ZhCn = &PostBody{
		Title:   title,
		Content: content,
	}
	return p
}

// WithEnUs 设置英文内容
func (p *PostContent) WithEnUs(title string, content [][]PostElement) *PostContent {
	p.EnUs = &PostBody{
		Title:   title,
		Content: content,
	}
	return p
}

// SendPost 发送富文本消息
func (s *MessageService) SendPost(chatID string, post *PostContent) error {
	return s.SendMessage(chatID, post)
}

// ImageContent 图片消息内容
type ImageContent struct {
	ImageKey string `json:"image_key"`
}

// Type 实现MessageContent接口
func (i *ImageContent) Type() MessageType {
	return MessageTypeImage
}

// Content 实现MessageContent接口
func (i *ImageContent) Content() map[string]interface{} {
	return map[string]interface{}{
		"image_key": i.ImageKey,
	}
}

// SendImage 发送图片消息
func (s *MessageService) SendImage(chatID string, imageKey string) error {
	content := &ImageContent{ImageKey: imageKey}
	return s.SendMessage(chatID, content)
}

// UploadImage 上传图片并获取image_key
func (s *MessageService) UploadImage(imageBytes []byte, imageName string) (string, error) {
	path := "/im/v1/images"
	return s.client.UploadFile(path, imageBytes, imageName)
}

// CreateGroupRequest 创建群组请求
type CreateGroupRequest struct {
	Name        string   `json:"name"`
	Description string   `json:"description,omitempty"`
	UserIDs     []string `json:"user_ids,omitempty"`
}

// CreateGroup 创建群组
func (s *MessageService) CreateGroup(req *CreateGroupRequest) (string, error) {
	path := "/im/v1/chats"
	
	var result struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
		Data struct {
			ChatID string `json:"chat_id"`
		} `json:"data"`
	}
	
	err := s.client.DoRequest("POST", path, req, &result)
	if err != nil {
		return "", err
	}
	
	if result.Code != 0 {
		return "", &Error{Code: result.Code, Message: result.Msg}
	}
	
	return result.Data.ChatID, nil
}

// GetGroupInfo 获取群组信息
func (s *MessageService) GetGroupInfo(chatID string) (map[string]interface{}, error) {
	path := fmt.Sprintf("/im/v1/chats/%s", chatID)
	
	var result struct {
		Code int                    `json:"code"`
		Msg  string                 `json:"msg"`
		Data map[string]interface{} `json:"data"`
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

// AddGroupMember 添加群成员
func (s *MessageService) AddGroupMember(chatID string, userIDs []string) error {
	path := fmt.Sprintf("/im/v1/chats/%s/members", chatID)
	
	reqBody := map[string]interface{}{
		"id_list": userIDs,
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

// RemoveGroupMember 移除群成员
func (s *MessageService) RemoveGroupMember(chatID string, userIDs []string) error {
	path := fmt.Sprintf("/im/v1/chats/%s/members", chatID)
	
	// 构造请求URL参数
	queryParams := "?"
	for i, userID := range userIDs {
		if i > 0 {
			queryParams += "&"
		}
		queryParams += fmt.Sprintf("id_list=%s", userID)
	}
	path += queryParams
	
	var result APIResponse
	err := s.client.DoRequest("DELETE", path, nil, &result)
	if err != nil {
		return err
	}
	
	if result.Code != 0 {
		return &Error{Code: result.Code, Message: result.Msg}
	}
	
	return nil
}

// FileContent 文件消息内容
type FileContent struct {
	FileKey string `json:"file_key"`
}

// Type 实现MessageContent接口
func (f *FileContent) Type() MessageType {
	return "file"
}

// Content 实现MessageContent接口
func (f *FileContent) Content() map[string]interface{} {
	return map[string]interface{}{
		"file_key": f.FileKey,
	}
}

// SendFile 发送文件消息
func (s *MessageService) SendFile(chatID string, fileKey string) error {
	content := &FileContent{FileKey: fileKey}
	return s.SendMessage(chatID, content)
}

// UploadFile 上传文件并获取file_key
func (s *MessageService) UploadFile(fileBytes []byte, fileName string) (string, error) {
	path := "/im/v1/files"
	return s.client.UploadFile(path, fileBytes, fileName)
}