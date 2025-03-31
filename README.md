# EasyLark

EasyLark 是一个简单易用的飞书(Lark)开放平台 API 的 Go SDK，旨在帮助开发者更轻松地集成飞书的功能到自己的应用中。

## 功能特点

- 简洁易用的 API 设计
- 完整的类型定义
- 内置错误处理
- 支持飞书身份验证
- 支持消息发送功能
- 支持表格操作功能

## 安装

```bash
go get github.com/yourusername/easylark
```

## 快速开始

### 初始化客户端

```go
import "github.com/yourusername/easylark"

func main() {
    client := easylark.NewClient("your-app-id", "your-app-secret")
    
    // 使用客户端调用API
    // ...
}
```

### 发送消息

```go
// 发送文本消息到群组
err := client.Message.SendText("chat_id", "Hello from EasyLark!")
if err != nil {
    // 处理错误
}

// 发送富文本消息
card := easylark.NewMessageCard().SetTitle("标题").AddText("正文内容")
err = client.Message.SendCard("chat_id", card)
if err != nil {
    // 处理错误
}
```

### 操作表格

```go
// 获取表格元数据
sheet, err := client.Sheet.Get("sheet_token")
if err != nil {
    // 处理错误
}

// 读取表格内容
values, err := client.Sheet.ReadRange("sheet_token", "Sheet1!A1:C3")
if err != nil {
    // 处理错误
}

// 写入表格内容
err = client.Sheet.WriteRange("sheet_token", "Sheet1!A1:B2", [][]interface{}{
    {"姓名", "年龄"},
    {"张三", 25},
})
if err != nil {
    // 处理错误
}
```

## 文档

详细的API文档请参考[这里](https://godoc.org/github.com/qqxhb/easylark)。

## 贡献

欢迎提交问题和PR！

## 许可证

MIT
