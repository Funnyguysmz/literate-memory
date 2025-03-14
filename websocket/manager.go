package websocket

import (
	"log"
	"sync"

	"github.com/gorilla/websocket"
)

// ClientType 客户端类型
type ClientType string

const (
	// StudentClient 学生客户端
	StudentClient ClientType = "student"
	// AdminClient 管理员客户端
	AdminClient ClientType = "admin"
)

// Client WebSocket客户端
type Client struct {
	Conn     *websocket.Conn
	UserID   string
	ExamID   string
	Type     ClientType
	SessionID string // 对于学生，这是他们的屏幕流会话ID
}

// Manager WebSocket连接管理器
type Manager struct {
	clients    map[*Client]bool
	register   chan *Client
	unregister chan *Client
	broadcast  chan []byte
	// 按考试ID和用户ID索引客户端
	examClients map[string]map[string]*Client // examID -> userID -> Client
	// 管理员客户端
	adminClients map[string][]*Client // examID -> []Client
	mutex        sync.RWMutex
}

// NewManager 创建新的WebSocket管理器
func NewManager() *Manager {
	return &Manager{
		clients:      make(map[*Client]bool),
		register:     make(chan *Client),
		unregister:   make(chan *Client),
		broadcast:    make(chan []byte),
		examClients:  make(map[string]map[string]*Client),
		adminClients: make(map[string][]*Client),
	}
}

// Start 启动WebSocket管理器
func (m *Manager) Start() {
	for {
		select {
		case client := <-m.register:
			m.registerClient(client)
		case client := <-m.unregister:
			m.unregisterClient(client)
		case message := <-m.broadcast:
			m.broadcastToAll(message)
		}
	}
}

// RegisterClient 注册客户端
func (m *Manager) RegisterClient(client *Client) {
	m.register <- client
}

// UnregisterClient 注销客户端
func (m *Manager) UnregisterClient(client *Client) {
	m.unregister <- client
}

// registerClient 内部注册客户端方法
func (m *Manager) registerClient(client *Client) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.clients[client] = true

	// 按考试和用户ID索引
	if client.Type == StudentClient {
		if _, ok := m.examClients[client.ExamID]; !ok {
			m.examClients[client.ExamID] = make(map[string]*Client)
		}
		m.examClients[client.ExamID][client.UserID] = client
	} else if client.Type == AdminClient {
		if _, ok := m.adminClients[client.ExamID]; !ok {
			m.adminClients[client.ExamID] = make([]*Client, 0)
		}
		m.adminClients[client.ExamID] = append(m.adminClients[client.ExamID], client)
	}

	log.Printf("客户端已连接: %s, 类型: %s, 考试: %s", client.UserID, client.Type, client.ExamID)
}

// unregisterClient 内部注销客户端方法
func (m *Manager) unregisterClient(client *Client) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if _, ok := m.clients[client]; ok {
		delete(m.clients, client)
		client.Conn.Close()

		// 从索引中移除
		if client.Type == StudentClient {
			if examClients, ok := m.examClients[client.ExamID]; ok {
				delete(examClients, client.UserID)
			}
		} else if client.Type == AdminClient {
			if adminClients, ok := m.adminClients[client.ExamID]; ok {
				for i, c := range adminClients {
					if c == client {
						m.adminClients[client.ExamID] = append(adminClients[:i], adminClients[i+1:]...)
						break
					}
				}
			}
		}

		log.Printf("客户端已断开连接: %s, 类型: %s, 考试: %s", client.UserID, client.Type, client.ExamID)
	}
}

// SendToAdmin 向指定考试的所有管理员发送消息
func (m *Manager) SendToAdmin(examID string, message []byte) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	if adminClients, ok := m.adminClients[examID]; ok {
		for _, client := range adminClients {
			err := client.Conn.WriteMessage(websocket.BinaryMessage, message)
			if err != nil {
				log.Printf("向管理员发送消息失败: %v", err)
				// 这里不关闭连接，让心跳检测来处理断开的连接
			}
		}
	}
}

// SendToStudent 向指定考试的特定学生发送消息
func (m *Manager) SendToStudent(examID, userID string, message []byte) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	if examClients, ok := m.examClients[examID]; ok {
		if client, ok := examClients[userID]; ok {
			err := client.Conn.WriteMessage(websocket.BinaryMessage, message)
			if err != nil {
				log.Printf("向学生发送消息失败: %v", err)
			}
		}
	}
}

// broadcastToAll 向所有客户端广播消息
func (m *Manager) broadcastToAll(message []byte) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	for client := range m.clients {
		err := client.Conn.WriteMessage(websocket.TextMessage, message)
		if err != nil {
			log.Printf("广播消息失败: %v", err)
		}
	}
}

// GetStudentClient 获取特定学生的客户端
func (m *Manager) GetStudentClient(examID, userID string) *Client {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	if examClients, ok := m.examClients[examID]; ok {
		if client, ok := examClients[userID]; ok {
			return client
		}
	}
	return nil
}

// GetAdminClients 获取特定考试的所有管理员客户端
func (m *Manager) GetAdminClients(examID string) []*Client {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	if clients, ok := m.adminClients[examID]; ok {
		return clients
	}
	return nil
} 