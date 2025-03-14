package websocket

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"image/jpeg"
	"log"
	"sync"
	"time"

	"GAI_test/config"
	"GAI_test/model"
	"GAI_test/utils"

	"github.com/gorilla/websocket"
	"github.com/kbinani/screenshot"
)

// 全局WebSocket管理器
var WSManager *Manager
var once sync.Once

// InitWSManager 初始化WebSocket管理器
func InitWSManager() {
	once.Do(func() {
		WSManager = NewManager()
		go WSManager.Start()
	})
}

// ScreenMessage WebSocket消息结构
type ScreenMessage struct {
	Type      string `json:"type"`      // 消息类型: image, command, status
	UserID    string `json:"userid"`    // 用户ID
	ExamID    string `json:"examid"`    // 考试ID
	Timestamp int64  `json:"timestamp"` // 时间戳
	Data      string `json:"data"`      // 图像数据(base64)或命令数据
	SessionID string `json:"sessionid"` // 会话ID
}

// StreamManager 屏幕流管理器
type StreamManager struct {
	// 活跃的流会话
	activeSessions map[string]*StreamSession
	mutex          sync.RWMutex
}

// StreamSession 流会话
type StreamSession struct {
	SessionID    string
	UserID       string
	ExamID       string
	StreamType   string
	IntervalSecs int
	StopChan     chan struct{}
	IsActive     bool
}

// 全局流管理器
var streamManager *StreamManager

// InitStreamManager 初始化流管理器
func InitStreamManager() {
	streamManager = &StreamManager{
		activeSessions: make(map[string]*StreamSession),
	}
}

// StartScreenStream 开始屏幕流
func (sm *StreamManager) StartScreenStream(sessionID, userID, examID, streamType string, intervalSecs int) error {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	// 检查是否已有活跃会话
	if session, exists := sm.activeSessions[sessionID]; exists && session.IsActive {
		return fmt.Errorf("会话已存在且处于活跃状态")
	}

	// 创建新会话
	session := &StreamSession{
		SessionID:    sessionID,
		UserID:       userID,
		ExamID:       examID,
		StreamType:   streamType,
		IntervalSecs: intervalSecs,
		StopChan:     make(chan struct{}),
		IsActive:     true,
	}

	// 保存会话
	sm.activeSessions[sessionID] = session

	// 保存到数据库
	dbSession := model.ScreenStreamSession{
		SessionID:    sessionID,
		UserID:       userID,
		ExamID:       examID,
		StartTime:    time.Now(),
		IsActive:     true,
		StreamType:   streamType,
		IntervalSecs: intervalSecs,
	}
	if err := config.DB.Create(&dbSession).Error; err != nil {
		log.Printf("保存屏幕流会话失败: %v", err)
	}

	// 根据流类型启动不同的捕获方式
	if streamType == "interval" {
		go sm.captureScreenInterval(session)
	} else if streamType == "realtime" {
		go sm.captureScreenRealtime(session)
	}

	return nil
}

// StopScreenStream 停止屏幕流
func (sm *StreamManager) StopScreenStream(sessionID string) error {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	session, exists := sm.activeSessions[sessionID]
	if !exists || !session.IsActive {
		return fmt.Errorf("会话不存在或已停止")
	}

	// 发送停止信号
	close(session.StopChan)
	session.IsActive = false

	// 更新数据库
	var dbSession model.ScreenStreamSession
	if err := config.DB.Where("session_id = ?", sessionID).First(&dbSession).Error; err == nil {
		dbSession.EndTime = time.Now()
		dbSession.IsActive = false
		config.DB.Save(&dbSession)
	}

	return nil
}

// captureScreenInterval 定时截屏
func (sm *StreamManager) captureScreenInterval(session *StreamSession) {
	ticker := time.NewTicker(time.Duration(session.IntervalSecs) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// 捕获屏幕
			if err := sm.captureAndSend(session); err != nil {
				log.Printf("屏幕捕获失败: %v", err)
			}
		case <-session.StopChan:
			log.Printf("停止定时屏幕捕获: %s", session.SessionID)
			return
		}
	}
}

// captureScreenRealtime 实时屏幕捕获
func (sm *StreamManager) captureScreenRealtime(session *StreamSession) {
	// 实时模式下，我们使用更高的频率
	ticker := time.NewTicker(100 * time.Millisecond) // 100ms约等于10帧/秒
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// 捕获屏幕
			if err := sm.captureAndSend(session); err != nil {
				log.Printf("屏幕捕获失败: %v", err)
			}
		case <-session.StopChan:
			log.Printf("停止实时屏幕捕获: %s", session.SessionID)
			return
		}
	}
}

// captureAndSend 捕获屏幕并发送
func (sm *StreamManager) captureAndSend(session *StreamSession) error {
	// 获取主显示器的屏幕截图
	bounds := screenshot.GetDisplayBounds(0)
	img, err := screenshot.CaptureRect(bounds)
	if err != nil {
		return fmt.Errorf("截图失败: %v", err)
	}

	// 压缩图像
	var buf bytes.Buffer
	quality := 50 // 降低质量以减小数据大小
	if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: quality}); err != nil {
		return fmt.Errorf("图像编码失败: %v", err)
	}

	// 转换为base64
	base64Data := base64.StdEncoding.EncodeToString(buf.Bytes())

	// 创建消息
	message := ScreenMessage{
		Type:      "image",
		UserID:    session.UserID,
		ExamID:    session.ExamID,
		Timestamp: time.Now().Unix(),
		Data:      base64Data,
		SessionID: session.SessionID,
	}

	// 序列化消息
	jsonData, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("消息序列化失败: %v", err)
	}

	// 发送到所有管理员
	WSManager.SendToAdmin(session.ExamID, jsonData)

	// 可选：保存截图记录到数据库
	if session.StreamType == "interval" {
		// 只有定时模式才保存记录，实时模式频率太高
		capture := model.ScreenCapture{
			SessionID:   session.SessionID,
			CaptureTime: time.Now(),
			// 如果需要保存图像，这里可以保存路径
		}
		config.DB.Create(&capture)
	}

	return nil
}

// GetStreamManager 获取流管理器实例
func GetStreamManager() *StreamManager {
	return streamManager
}

// HandleScreenCommand 处理屏幕命令
func HandleScreenCommand(client *Client, message []byte) {
	var cmd struct {
		Command   string `json:"command"`   // start, stop
		StreamType string `json:"streamtype"` // interval, realtime
		Interval  int    `json:"interval"`  // 间隔秒数
	}

	if err := json.Unmarshal(message, &cmd); err != nil {
		log.Printf("解析命令失败: %v", err)
		return
	}

	sm := GetStreamManager()

	switch cmd.Command {
	case "start":
		// 生成会话ID
		sessionID := utils.GenerateRandomString(16)
		
		// 设置默认值
		if cmd.StreamType == "" {
			cmd.StreamType = "interval"
		}
		if cmd.Interval <= 0 {
			cmd.Interval = 5 // 默认5秒
		}
		
		// 开始屏幕流
		if err := sm.StartScreenStream(sessionID, client.UserID, client.ExamID, cmd.StreamType, cmd.Interval); err != nil {
			log.Printf("启动屏幕流失败: %v", err)
			return
		}
		
		// 更新客户端会话ID
		client.SessionID = sessionID
		
		// 发送确认消息
		response := ScreenMessage{
			Type:      "status",
			UserID:    client.UserID,
			ExamID:    client.ExamID,
			Timestamp: time.Now().Unix(),
			Data:      "started",
			SessionID: sessionID,
		}
		jsonResp, _ := json.Marshal(response)
		client.Conn.WriteMessage(websocket.TextMessage, jsonResp)
		
	case "stop":
		if client.SessionID == "" {
			log.Printf("无活跃会话")
			return
		}
		
		// 停止屏幕流
		if err := sm.StopScreenStream(client.SessionID); err != nil {
			log.Printf("停止屏幕流失败: %v", err)
			return
		}
		
		// 发送确认消息
		response := ScreenMessage{
			Type:      "status",
			UserID:    client.UserID,
			ExamID:    client.ExamID,
			Timestamp: time.Now().Unix(),
			Data:      "stopped",
			SessionID: client.SessionID,
		}
		jsonResp, _ := json.Marshal(response)
		client.Conn.WriteMessage(websocket.TextMessage, jsonResp)
		
		// 清除会话ID
		client.SessionID = ""
	}
} 