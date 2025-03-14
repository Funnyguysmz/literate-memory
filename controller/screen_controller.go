package controller

import (
	"net/http"
	"time"

	"GAI_test/config"
	"GAI_test/model"
	"GAI_test/utils"
	"GAI_test/websocket"

	"github.com/gin-gonic/gin"
)

// InitScreenStream 初始化屏幕流
// @Summary 初始化屏幕流
// @Description 学生端初始化屏幕流会话
// @Accept json
// @Produce json
// @Param request body InitScreenStreamRequest true "初始化请求"
// @Success 200 {object} InitScreenStreamResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /screen/init [post]
func InitScreenStream(c *gin.Context) {
	// 验证token
	tokenString := c.GetHeader("Authorization")
	if tokenString == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "缺少token"})
		return
	}
	claims, err := validateToken(tokenString)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "token验证失败"})
		return
	}

	var req struct {
		StreamType   string `json:"streamtype"` // interval, realtime
		IntervalSecs int    `json:"interval"`   // 间隔秒数
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数不完整"})
		return
	}

	// 设置默认值
	if req.StreamType == "" {
		req.StreamType = "interval"
	}
	if req.IntervalSecs <= 0 {
		req.IntervalSecs = 5 // 默认5秒
	}

	// 从token中获取用户信息
	userID := claims["userid"].(string)
	examID := claims["examid"].(string)

	// 生成会话ID
	sessionID := utils.GenerateRandomString(16)

	// 创建屏幕流会话
	session := model.ScreenStreamSession{
		SessionID:    sessionID,
		UserID:       userID,
		ExamID:       examID,
		StartTime:    time.Now(),
		IsActive:     true,
		StreamType:   req.StreamType,
		IntervalSecs: req.IntervalSecs,
	}

	if err := config.DB.Create(&session).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建会话失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   "屏幕流会话初始化成功",
		"sessionid": sessionID,
		"userid":    userID,
		"examid":    examID,
		"streamtype": req.StreamType,
		"interval":  req.IntervalSecs,
	})
}

// GetActiveStreams 获取活跃的屏幕流
// @Summary 获取活跃的屏幕流
// @Description 管理员获取指定考试的所有活跃屏幕流
// @Accept json
// @Produce json
// @Param examid query string true "考试ID"
// @Success 200 {array} ActiveStreamResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /screen/active [get]
func GetActiveStreams(c *gin.Context) {
	// 验证token
	tokenString := c.GetHeader("Authorization")
	if tokenString == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "缺少token"})
		return
	}
	claims, err := validateToken(tokenString)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "token验证失败"})
		return
	}

	// 验证是否为管理员
	role, ok := claims["role"].(string)
	if !ok || role != "admin" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "需要管理员权限"})
		return
	}

	examID := c.Query("examid")
	if examID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "缺少考试ID"})
		return
	}

	// 查询活跃的屏幕流会话
	var sessions []model.ScreenStreamSession
	if err := config.DB.Where("exam_id = ? AND is_active = ?", examID, true).Find(&sessions).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}

	// 构建响应
	var response []gin.H
	for _, session := range sessions {
		// 查询用户信息
		var user model.User
		config.DB.Where("user_id = ?", session.UserID).First(&user)

		response = append(response, gin.H{
			"sessionid":  session.SessionID,
			"userid":     session.UserID,
			"username":   user.Name,
			"examid":     session.ExamID,
			"starttime":  session.StartTime,
			"streamtype": session.StreamType,
			"interval":   session.IntervalSecs,
		})
	}

	c.JSON(http.StatusOK, response)
}

// StopScreenStream 停止屏幕流
// @Summary 停止屏幕流
// @Description 停止指定的屏幕流会话
// @Accept json
// @Produce json
// @Param request body StopScreenStreamRequest true "停止请求"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /screen/stop [post]
func StopScreenStream(c *gin.Context) {
	// 验证token
	tokenString := c.GetHeader("Authorization")
	if tokenString == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "缺少token"})
		return
	}
	claims, err := validateToken(tokenString)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "token验证失败"})
		return
	}

	var req struct {
		SessionID string `json:"sessionid"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.SessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "缺少会话ID"})
		return
	}

	// 查询会话
	var session model.ScreenStreamSession
	if err := config.DB.Where("session_id = ?", req.SessionID).First(&session).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "会话不存在"})
		return
	}

	// 验证权限（只有会话所有者或管理员可以停止）
	userID := claims["userid"].(string)
	role, _ := claims["role"].(string)
	if session.UserID != userID && role != "admin" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "无权停止此会话"})
		return
	}

	// 更新会话状态
	session.IsActive = false
	session.EndTime = time.Now()
	if err := config.DB.Save(&session).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "停止会话失败"})
		return
	}

	// 如果有活跃的流，通过StreamManager停止
	streamManager := websocket.GetStreamManager()
	if streamManager != nil {
		streamManager.StopScreenStream(req.SessionID)
	}

	c.JSON(http.StatusOK, gin.H{"message": "屏幕流已停止"})
}

// 用于Swagger文档的结构体定义
// InitScreenStreamRequest 初始化屏幕流请求
type InitScreenStreamRequest struct {
	StreamType   string `json:"streamtype" example:"interval" enums:"interval,realtime"` // 流类型：interval-定时截屏, realtime-实时直播
	IntervalSecs int    `json:"interval" example:"5"`                                   // 截屏间隔秒数（仅用于interval类型）
}

// InitScreenStreamResponse 初始化屏幕流响应
type InitScreenStreamResponse struct {
	Message    string `json:"message" example:"屏幕流会话初始化成功"`
	SessionID  string `json:"sessionid" example:"abcdef1234567890"`
	UserID     string `json:"userid" example:"2021001"`
	ExamID     string `json:"examid" example:"EXAM001"`
	StreamType string `json:"streamtype" example:"interval"`
	Interval   int    `json:"interval" example:"5"`
}

// ActiveStreamResponse 活跃屏幕流响应
type ActiveStreamResponse struct {
	SessionID  string    `json:"sessionid" example:"abcdef1234567890"`
	UserID     string    `json:"userid" example:"2021001"`
	Username   string    `json:"username" example:"张三"`
	ExamID     string    `json:"examid" example:"EXAM001"`
	StartTime  time.Time `json:"starttime" example:"2023-01-01T12:00:00Z"`
	StreamType string    `json:"streamtype" example:"interval"`
	Interval   int       `json:"interval" example:"5"`
}

// StopScreenStreamRequest 停止屏幕流请求
type StopScreenStreamRequest struct {
	SessionID string `json:"sessionid" example:"abcdef1234567890" binding:"required"`
} 