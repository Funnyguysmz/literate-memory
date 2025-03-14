package websocket

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

// 配置WebSocket升级器
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	// 允许所有跨域请求
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// 消息类型
const (
	MessageTypeCommand = "command" // 命令消息
	MessageTypeImage   = "image"   // 图像消息
	MessageTypeStatus  = "status"  // 状态消息
)

// 命令类型
const (
	CommandStart = "start" // 开始屏幕共享
	CommandStop  = "stop"  // 停止屏幕共享
)

// HandleStudentWS 处理学生WebSocket连接
func HandleStudentWS(c *gin.Context) {
	// 获取参数
	userID := c.Query("userid")
	examID := c.Query("examid")
	token := c.Query("token")

	// 验证参数
	if userID == "" || examID == "" || token == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "缺少必要参数"})
		return
	}

	// 升级HTTP连接为WebSocket
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("WebSocket升级失败: %v", err)
		return
	}

	// 创建客户端
	client := &Client{
		Conn:   conn,
		UserID: userID,
		ExamID: examID,
		Type:   StudentClient,
	}

	// 注册客户端
	WSManager.RegisterClient(client)

	// 处理连接
	go handleConnection(client)
}

// HandleAdminWS 处理管理员WebSocket连接
func HandleAdminWS(c *gin.Context) {
	// 获取参数
	userID := c.Query("userid")
	examID := c.Query("examid")
	token := c.Query("token")

	// 验证参数
	if userID == "" || examID == "" || token == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "缺少必要参数"})
		return
	}

	// 升级HTTP连接为WebSocket
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("WebSocket升级失败: %v", err)
		return
	}

	// 创建客户端
	client := &Client{
		Conn:   conn,
		UserID: userID,
		ExamID: examID,
		Type:   AdminClient,
	}

	// 注册客户端
	WSManager.RegisterClient(client)

	// 处理连接
	go handleConnection(client)
}

// handleConnection 处理WebSocket连接
func handleConnection(client *Client) {
	// 设置连接关闭处理
	defer func() {
		WSManager.UnregisterClient(client)
	}()

	// 设置读取超时
	client.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	client.Conn.SetPongHandler(func(string) error {
		client.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	// 启动心跳检测
	go ping(client)

	// 读取消息循环
	for {
		_, message, err := client.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket错误: %v", err)
			}
			break
		}

		// 处理消息
		handleMessage(client, message)
	}
}

// handleMessage 处理WebSocket消息
func handleMessage(client *Client, message []byte) {
	// 解析消息类型
	var msg struct {
		Type string `json:"type"`
	}
	if err := json.Unmarshal(message, &msg); err != nil {
		log.Printf("解析消息失败: %v", err)
		return
	}

	// 根据消息类型处理
	switch msg.Type {
	case MessageTypeCommand:
		// 处理命令消息
		HandleScreenCommand(client, message)
	case MessageTypeStatus:
		// 处理状态消息
		// 暂不实现
	default:
		log.Printf("未知消息类型: %s", msg.Type)
	}
}

// ping 发送心跳包
func ping(client *Client) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := client.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				WSManager.UnregisterClient(client)
				return
			}
		}
	}
} 