package controller

import (
	"fmt"
	"net/http"
	"time"

	"GAI_test/config"
	"GAI_test/model"
	"GAI_test/utils"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
)

// 定义token黑名单
var tokenBlacklist = make(map[string]bool)

// GeneratePasswords 批量生成密码时添加角色及考试信息
// 修改：从请求中获取examid、sttime和endtime（格式采用RFC3339）
func GeneratePasswords(c *gin.Context) {
	var req struct {
		Count   int    `json:"count"`
		Role    string `json:"role"`
		ExamID  string `json:"examid"`
		STTime  string `json:"sttime"`
		EndTime string `json:"endtime"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.Count <= 0 || req.ExamID == "" || req.STTime == "" || req.EndTime == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求参数"})
		return
	}

	if req.Role == "" {
		req.Role = "student"
	}
	st, err := time.Parse(time.RFC3339, req.STTime)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "sttime格式错误"})
		return
	}
	et, err := time.Parse(time.RFC3339, req.EndTime)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "endtime格式错误"})
		return
	}

	generated := 0
	for i := 0; i < req.Count; i++ {
		pw := utils.GenerateRandomString(8)
		newPwd := model.ExamPassword{
			Password:    pw,
			Distributed: false,
			ExamID:      req.ExamID,
			StudentID:   "",
			Role:        req.Role,
			ValidFrom:   st,
			ValidUntil:  et,
		}
		if err := config.DB.Create(&newPwd).Error; err == nil {
			generated++
		}
	}
	c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("生成密码数量：%d", generated)})
}

// GetPassword 领取未分发密码
// @Summary 获取考试密码
// @Description 客户端接口：获取一个未分发的考试密码
// @Accept json
// @Produce json
// @Param userid query string true "用户ID"
// @Success 200 {object} GetPasswordResponse
// @Failure 404 {object} ErrorResponse
// @Router /password [get]
func GetPassword(c *gin.Context) {
	var pwd model.ExamPassword
	if err := config.DB.Where("distributed = ?", false).First(&pwd).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "无可用密码"})
		return
	}
	// 模拟领取时绑定考生信息，实际业务中考生ID应由客户端传入
	pwd.Distributed = true
	pwd.StudentID = c.Query("userid") // 改为userid
	config.DB.Save(&pwd)
	c.JSON(http.StatusOK, gin.H{"password": pwd.Password})
}

// GetAdminPasswordResponse 管理员密码响应
type GetAdminPasswordResponse struct {
    ExamID   string `json:"examid" example:"EXAM001"`
    AdminID  string `json:"adminid" example:"admin001"`
    Password string `json:"password" example:"a1b2c3d4"`
}

// GetAdminPassword 获取管理员密码
// @Summary 获取管理员密码
// @Description 获取指定考试的管理员密码
// @Accept json
// @Produce json
// @Param examid query string true "考试ID"
// @Success 200 {object} GetAdminPasswordResponse
// @Failure 404 {object} ErrorResponse
// @Router /admin/password [get]
func GetAdminPassword(c *gin.Context) {
    examID := c.Query("examid")
    if examID == "" {
        c.JSON(http.StatusBadRequest, gin.H{"error": "考试ID不能为空"})
        return
    }

    var pwd model.ExamPassword
    if err := config.DB.Where("distributed = ? AND exam_id = ? AND role = ?", 
        false, examID, "admin").First(&pwd).Error; err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "无可用的管理员密码"})
        return
    }

    pwd.Distributed = true
    config.DB.Save(&pwd)

    // 返回完整的管理员信息
    c.JSON(http.StatusOK, gin.H{
        "examid":   pwd.ExamID,
        "adminid":  pwd.StudentID, // StudentID字段用作adminID
        "password": pwd.Password,
    })
}

// RaiseHand 举手请求，增加token验证身份
func RaiseHand(c *gin.Context) {
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
	// 要求token中的userid与请求body中的userid一致
	var req struct {
		ExamID    string `json:"examid"`
		UserID    string `json:"userid"`  // 改为userid
		Reason    string `json:"reason"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.ExamID == "" || req.UserID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数不完整"})
		return
	}
	if claims["userid"] != req.UserID {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "token与用户不匹配"})
		return
	}

	// ...existing记录请求逻辑...
	c.JSON(http.StatusOK, gin.H{"message": "举手请求已提交"})
}

// DeleteAllAdmins 删除所有管理员（仅用于测试）
// @Summary 删除所有管理员
// @Description 测试接口：删除所有管理员数据
// @Accept json
// @Produce json
// @Success 200 {object} SuccessResponse
// @Router /admin/deleteAll [delete]
func DeleteAllAdmins(c *gin.Context) {
    // 删除所有管理员密码
    config.DB.Where("role = ?", "admin").Delete(&model.ExamPassword{})
    // 删除所有管理员备份
    config.DB.Delete(&model.AdminBackup{})
    
    c.JSON(http.StatusOK, gin.H{"message": "所有管理员数据已删除"})
}

// GetAdminBackupResponse 管理员备份信息响应
type GetAdminBackupResponse struct {
    AdminID   string    `json:"adminid" example:"admin001"`
    Password  string    `json:"password" example:"admin123456"`
    ExamID    string    `json:"examid" example:"EXAM001"`
    CreatedAt time.Time `json:"created_at"`
}

// GetAdminBackup 获取顶级管理员信息
// @Summary 获取顶级管理员信息
// @Description 获取系统初始化时创建的顶级管理员信息
// @Produce json
// @Success 200 {object} GetAdminBackupResponse
// @Failure 404 {object} ErrorResponse
// @Router /admin/backup [get]
func GetAdminBackup(c *gin.Context) {
    var backup model.AdminBackup
    // 获取最新的管理员备份记录
    if err := config.DB.Order("created_at desc").First(&backup).Error; err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "未找到管理员备份信息"})
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "adminid":    backup.AdminID,
        "password":   backup.Password,
        "examid":     backup.ExamID,
        "created_at": backup.CreatedAt,
    })
}


// UpdateAdmin 更新管理员信息，验证是否为本人操作（需token）
// @Summary 更新管理员信息
// @Description 更新默认管理员的ID、密码和考试ID
// @Accept json
// @Produce json
// @Param request body UpdateAdminRequest true "更新信息"
// @Success 200 {object} SuccessResponse
// @Failure 400,404 {object} ErrorResponse
// @Router /admin/update [post]
// TODO：参数需要token验证身份，并且身份需要是本人，仅本人可以修改自己的身份信息
func UpdateAdmin(c *gin.Context) {
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
	var req UpdateAdminRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求参数"})
		return
	}
	// 检查token中userid是否与旧管理员保持一致
	if claims["userid"] != req.OldAdminID {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "不能修改他人信息"})
		return
	}

    // 验证旧的管理员信息
    var backup model.AdminBackup
    if err := config.DB.Where("admin_id = ? AND password = ? AND exam_id = ? AND is_default = ?",
        req.OldAdminID, req.OldPassword, req.OldExamID, true).First(&backup).Error; err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "管理员信息验证失败"})
        return
    }

    // 开始事务
    tx := config.DB.Begin()

    // 更新备份表中的信息
    backup.AdminID = req.NewAdminID
    backup.Password = req.NewPassword
    backup.ExamID = req.NewExamID
    backup.IsDefault = false

    // 更新密码表中的信息（如果存在）
    var pwd model.ExamPassword
    if err := tx.Where("student_id = ? AND exam_id = ? AND role = ?",
        req.OldAdminID, req.OldExamID, "admin").First(&pwd).Error; err == nil {
        pwd.StudentID = req.NewAdminID
        pwd.Password = req.NewPassword
        pwd.ExamID = req.NewExamID
        if err := tx.Save(&pwd).Error; err != nil {
            tx.Rollback()
            c.JSON(http.StatusInternalServerError, gin.H{"error": "更新失败"})
            return
        }
    }

    // 保存备份表的更改
    if err := tx.Save(&backup).Error; err != nil {
        tx.Rollback()
        c.JSON(http.StatusInternalServerError, gin.H{"error": "更新失败"})
        return
    }

    // 提交事务
    tx.Commit()

    c.JSON(http.StatusOK, gin.H{"message": "管理员信息更新成功"})
}

// CreateExam 创建新考试
// @Summary 创建新考试
// @Description 管理员创建新的考试
// @Accept json
// @Produce json
// @Param request body CreateExamRequest true "考试信息"
// @Success 200 {object} CreateExamResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /exam/create [post]
func CreateExam(c *gin.Context) {
	// 验证管理员身份
	token := c.GetHeader("Authorization")
	if !validateAdminToken(token) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "无权限"})
		return
	}

	var req struct {
		ExamID     string `json:"examid"`
		StreamType string `json:"streamtype"` // 流类型：interval-定时截屏, realtime-实时直播
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求参数"})
		return
	}

	// 设置默认流类型
	if req.StreamType == "" {
		req.StreamType = "interval" // 默认使用定时截屏
	}

	// 验证流类型是否有效
	if req.StreamType != "interval" && req.StreamType != "realtime" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的流类型"})
		return
	}

	// 如果没有提供考试ID，则自动生成一个
	if req.ExamID == "" {
		// 生成一个随机的考试ID，格式为 EXAM + 6位随机字符
		req.ExamID = "EXAM" + utils.GenerateRandomString(6)
		
		// 确保生成的ID不重复
		var count int64
		for {
			config.DB.Model(&model.Exam{}).Where("exam_id = ?", req.ExamID).Count(&count)
			if count == 0 {
				break
			}
			req.ExamID = "EXAM" + utils.GenerateRandomString(6)
		}
	} else {
		// 检查考试ID是否已存在
		var count int64
		config.DB.Model(&model.Exam{}).Where("exam_id = ?", req.ExamID).Count(&count)
		if count > 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "考试ID已存在"})
			return
		}
	}

	exam := model.Exam{
		ExamID:     req.ExamID,
		Status:     "preparing",
		StreamType: req.StreamType,
	}

	if err := config.DB.Create(&exam).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建考试失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    "考试创建成功",
		"examid":     exam.ExamID,
		"streamtype": exam.StreamType,
	})
}

// BindUser 用户绑定考试 (重命名原BindStudent)
// @Summary 用户绑定考试
// @Description 用户到达考场后绑定考试并获取密码
// @Accept json
// @Produce json
// @Param request body BindUserRequest true "绑定信息"
// @Success 200 {object} BindUserResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /exam/bind [post]
func BindStudent(c *gin.Context) {
	var req struct {
		ExamID string `json:"examid"`
		UserID string `json:"userid"`  // 改为userid
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数不完整"})
		return
	}

	// 检查用户是否已注册
	var user model.User
	if err := config.DB.Where("user_id = ?", req.UserID).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "用户未注册，请先注册"})
		return
	}

	// 验证考试是否存在
	var exam model.Exam
	if err := config.DB.Where("exam_id = ?", req.ExamID).First(&exam).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "考试不存在"})
		return
	}

	// 检查用户是否已绑定此考试
	var existingPwd model.ExamPassword
	if err := config.DB.Where("exam_id = ? AND student_id = ?", 
		req.ExamID, req.UserID).First(&existingPwd).Error; err == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "该用户已绑定考试"})
		return
	}

	// 生成密码和token
	password := utils.GenerateRandomString(8)
	token := generateToken(req.UserID, req.ExamID, "student")

	// 创建密码记录
	pwd := model.ExamPassword{
		Password:   password,
		ExamID:    req.ExamID,
		StudentID: req.UserID,  // 使用StudentID字段存储userid
		Role:      "student",
		ValidFrom: time.Now(),
		ValidUntil: time.Now().Add(24 * time.Hour),
		Token:     token,
	}

	if err := config.DB.Create(&pwd).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "密码创建失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "绑定成功",
		"examid":   req.ExamID,
		"userid":   req.UserID,
		"username": user.Name,  // 添加用户姓名
		"password": password,
	})
}

// ValidateExamEntryWithToken 登录接口，验证考试密码后返回token
// @Summary 考试系统登录
// @Description 验证考试密码是否有效并返回用于后续请求的token
// @Accept json
// @Produce json
// @Param request body ValidateExamEntryRequest true "登录验证信息"
// @Success 200 {object} ValidateExamEntryResponse
// @Failure 400 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /validateExamPassword [post]
func ValidateExamEntryWithToken(c *gin.Context) {
	var req struct {
		ExamID   string `json:"examid"`
		UserID   string `json:"userid"`  // 改为userid
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数不完整"})
		return
	}

	// 首先检查用户是否存在
	var count int64
	config.DB.Model(&model.ExamPassword{}).Where("exam_id = ? AND student_id = ?", 
		req.ExamID, req.UserID).Count(&count)
	if count == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在，请先注册"})
		return
	}

	var pwd model.ExamPassword
	if err := config.DB.Where("exam_id = ? AND student_id = ? AND password = ?", 
		req.ExamID, req.UserID, req.Password).First(&pwd).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "验证失败"})
		return
	}

	now := time.Now()
	if now.Before(pwd.ValidFrom) || now.After(pwd.ValidUntil) {
		c.JSON(http.StatusForbidden, gin.H{"error": "密码不在有效时间内"})
		return
	}

	// 返回token用于后续请求的身份验证
	c.JSON(http.StatusOK, gin.H{
		"message": "验证成功",
		"token":   pwd.Token,
		"role":    pwd.Role,
		"examid":  pwd.ExamID,
		"userid":  pwd.StudentID,  // 这里响应中使用userid
	})
}

// InvalidateToken 使得token失效，添加至黑名单（token有效期为当天）
func InvalidateToken(c *gin.Context) {
	var req struct {
		Token string `json:"token"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.Token == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的token参数"})
		return
	}
	tokenBlacklist[req.Token] = true
	c.JSON(http.StatusOK, gin.H{"message": "token已失效"})
}

// generateToken 生成JWT token时添加role字段，token有效期为当天
func generateToken(userID, examID, role string) string {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"userid": userID,
		"examid": examID,
		"role":   role,
		"exp":    time.Now().Add(24 * time.Hour).Unix(),
	})
	tokenString, err := token.SignedString([]byte("your-secret-key"))
	if err != nil {
		return ""
	}
	return tokenString
}

// validateToken 辅助函数：解析token、检查是否在黑名单中
func validateToken(tokenString string) (map[string]interface{}, error) {
	if tokenBlacklist[tokenString] {
		return nil, fmt.Errorf("token已失效")
	}
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte("your-secret-key"), nil
	})
	if err != nil || !token.Valid {
		return nil, fmt.Errorf("无效token")
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("解析claims失败")
	}
	return claims, nil
}

// validateAdminToken 判断token是否指向管理员身份
func validateAdminToken(tokenString string) bool {
	claims, err := validateToken(tokenString)
	if err != nil {
		return false
	}
	role, ok := claims["role"].(string)
	return ok && role == "admin"
}

// validateStudentToken 判断token是否指向学生身份
func validateStudentToken(tokenString string) bool {
	claims, err := validateToken(tokenString)
	if err != nil {
		return false
	}
	role, ok := claims["role"].(string)
	return ok && role == "student"
}

// 用于 Swagger 文档的请求响应结构体定义
type GeneratePasswordsRequest struct {
	Count   int    `json:"count" example:"10"`
	Role    string `json:"role" example:"student"`
	ExamID  string `json:"examid" example:"EXAM001"`
	STTime  string `json:"sttime" example:"2024-02-17T15:30:45Z"`
	EndTime string `json:"endtime" example:"2024-02-17T16:30:45Z"`
}

type GeneratePasswordsResponse struct {
	Message string `json:"message" example:"生成密码数量：10"`
}

type GetPasswordResponse struct {
	Password string `json:"password" example:"a1b2c3d4"`
}

type ValidateExamEntryRequest struct {
	ExamID   string `json:"examid" example:"EXAM001"`
	UserID   string `json:"userid" example:"2021001"`
	Password string `json:"password" example:"a1b2c3d4"`
}

// ValidateExamEntryResponse 登录验证响应
type ValidateExamEntryResponse struct {
	Message   string `json:"message" example:"验证成功"`
	Token     string `json:"token" example:"eyJhbGciOiJIUzI1NiIs..."`
	Role      string `json:"role" example:"student" enums:"student,admin"`
	IsDefault bool   `json:"is_default" example:"false"`
	ExamID    string `json:"examid" example:"EXAM001"`
	UserID    string `json:"userid" example:"2021001"`
}

// TokenInvalidateRequest token失效请求
type TokenInvalidateRequest struct {
	Token string `json:"token" example:"eyJhbGciOiJIUzI1NiIs..." binding:"required"`
}

// TokenInvalidateResponse token失效响应
type TokenInvalidateResponse struct {
	Message string `json:"message" example:"token已失效"`
}

// RaiseHandRequest 举手请求
type RaiseHandRequest struct {
	ExamID  string `json:"examid" example:"EXAM001" binding:"required"`
	UserID  string `json:"userid" example:"2021001" binding:"required"` // 改为userid
	Reason  string `json:"reason" example:"无法获取密码" binding:"required"`
}

// RaiseHandResponse 举手响应
type RaiseHandResponse struct {
	Message string `json:"message" example:"举手请求已提交"`
}

type SuccessResponse struct {
	Message string `json:"message" example:"操作成功"`
}

type ErrorResponse struct {
	Error string `json:"error" example:"请求参数错误"`
}

// UpdateAdminRequest 更新管理员信息请求
type UpdateAdminRequest struct {
	OldAdminID  string `json:"old_adminid" example:"admin" binding:"required"`
	OldPassword string `json:"old_password" example:"admin123" binding:"required"`
	OldExamID   string `json:"old_examid" example:"EXAM001" binding:"required"`
	NewAdminID  string `json:"new_adminid" example:"customAdmin" binding:"required"`
	NewPassword string `json:"new_password" example:"customPassword" binding:"required"`
	NewExamID   string `json:"new_examid" example:"CUSTOM001" binding:"required"`
}

// CreateExamRequest 创建考试请求
type CreateExamRequest struct {
	ExamID     string `json:"examid" example:"EXAM002"` // 考试ID（可选，不提供则自动生成）
	StreamType string `json:"streamtype" example:"interval" enums:"interval,realtime"` // 流类型：interval-定时截屏, realtime-实时直播
}

// CreateExamResponse 创建考试响应
type CreateExamResponse struct {
	Message    string `json:"message" example:"考试创建成功"`
	ExamID     string `json:"examid" example:"EXAM002"`
	StreamType string `json:"streamtype" example:"interval"`
}

// BindUserRequest 用户绑定考试请求 (原BindStudentRequest)
type BindUserRequest struct {
	ExamID string `json:"examid" example:"EXAM001"`
	UserID string `json:"userid" example:"2021001"`
}

// BindUserResponse 用户绑定考试响应 (原BindStudentResponse)
type BindUserResponse struct {
	Message  string `json:"message" example:"绑定成功"`
	ExamID   string `json:"examid" example:"EXAM001"`
	UserID   string `json:"userid" example:"2021001"`
	Password string `json:"password" example:"a1b2c3d4"`
}

// UpdateExamStreamType 更新考试流类型
// @Summary 更新考试流类型
// @Description 管理员更新指定考试的屏幕流类型
// @Accept json
// @Produce json
// @Param request body UpdateExamStreamTypeRequest true "更新请求"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /exam/streamtype [put]
func UpdateExamStreamType(c *gin.Context) {
	// 验证管理员身份
	token := c.GetHeader("Authorization")
	if !validateAdminToken(token) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "无权限"})
		return
	}

	var req struct {
		ExamID     string `json:"examid" binding:"required"`
		StreamType string `json:"streamtype" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数不完整"})
		return
	}

	// 验证流类型是否有效
	if req.StreamType != "interval" && req.StreamType != "realtime" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的流类型"})
		return
	}

	// 查询考试是否存在
	var exam model.Exam
	if err := config.DB.Where("exam_id = ?", req.ExamID).First(&exam).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "考试不存在"})
		return
	}

	// 更新流类型
	exam.StreamType = req.StreamType
	if err := config.DB.Save(&exam).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "考试流类型更新成功",
		"examid": req.ExamID,
		"streamtype": req.StreamType,
	})
}

// GetSupportedStreamTypes 获取支持的流类型
// @Summary 获取支持的流类型
// @Description 获取系统当前支持的屏幕流类型
// @Produce json
// @Success 200 {object} StreamTypesResponse
// @Router /exam/streamtypes [get]
func GetSupportedStreamTypes(c *gin.Context) {
	// 定义当前支持的流类型
	streamTypes := []gin.H{
		{
			"type": "interval",
			"name": "定时截屏",
			"description": "每隔几秒截取一次屏幕，适合低带宽环境",
		},
		{
			"type": "realtime",
			"name": "实时直播",
			"description": "高频率截屏（约10帧/秒），提供接近实时的体验",
		},
	}

	c.JSON(http.StatusOK, gin.H{
		"streamtypes": streamTypes,
	})
}

// UpdateExamStreamTypeRequest 更新考试流类型请求
type UpdateExamStreamTypeRequest struct {
	ExamID     string `json:"examid" example:"EXAM001" binding:"required"`
	StreamType string `json:"streamtype" example:"realtime" binding:"required" enums:"interval,realtime"`
}

// StreamTypesResponse 流类型响应
type StreamTypesResponse struct {
	StreamTypes []StreamTypeInfo `json:"streamtypes"`
}

// StreamTypeInfo 流类型信息
type StreamTypeInfo struct {
	Type        string `json:"type" example:"interval"`
	Name        string `json:"name" example:"定时截屏"`
	Description string `json:"description" example:"每隔几秒截取一次屏幕，适合低带宽环境"`
}
