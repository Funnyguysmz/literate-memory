package controller

import (
	"net/http"

	"GAI_test/config"
	"GAI_test/model"

	"github.com/gin-gonic/gin"
)

// RegisterUser 注册新用户
// @Summary 注册新用户
// @Description 注册新用户（不关联考试，仅创建基本身份）
// @Accept json
// @Produce json
// @Param request body RegisterUserRequest true "注册信息"
// @Success 200 {object} RegisterUserResponse
// @Failure 400 {object} ErrorResponse
// @Router /user/register [post]
func RegisterUser(c *gin.Context) {
	var req struct {
		UserID string `json:"userid" binding:"required"`
		Name   string `json:"name" binding:"required"`
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数不完整"})
		return
	}
	
	// 检查用户是否已存在
	var count int64
	config.DB.Model(&model.User{}).Where("user_id = ?", req.UserID).Count(&count)
	if count > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "用户ID已存在"})
		return
	}
	
	// 创建新用户
	user := model.User{
		UserID: req.UserID,
		Name:   req.Name,
	}
	
	if err := config.DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "注册失败"})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"message": "注册成功",
		"userid":  req.UserID,
		"name":    req.Name,
	})
}

// GetUser 获取用户信息
// @Summary 获取用户信息
// @Description 获取指定ID的用户信息
// @Accept json
// @Produce json
// @Param userid path string true "用户ID"
// @Success 200 {object} User
// @Failure 404 {object} ErrorResponse
// @Router /user/{userid} [get]
func GetUser(c *gin.Context) {
	userID := c.Param("userid")
	
	var user model.User
	if err := config.DB.Where("user_id = ?", userID).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"userid": user.UserID,
		"name":   user.Name,
	})
}

// 用于Swagger文档的结构体定义
// RegisterUserRequest 注册用户请求
type RegisterUserRequest struct {
	UserID string `json:"userid" example:"2021001" binding:"required"` // 用户ID（学号/工号）
	Name   string `json:"name" example:"张三" binding:"required"`      // 用户姓名
}

// RegisterUserResponse 注册用户响应
type RegisterUserResponse struct {
	Message string `json:"message" example:"注册成功"`
	UserID  string `json:"userid" example:"2021001"`
	Name    string `json:"name" example:"张三"`
}

// User 用户信息
type User struct {
	UserID string `json:"userid" example:"2021001"`
	Name   string `json:"name" example:"张三"`
}
