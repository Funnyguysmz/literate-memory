package controller

import (
	"net/http"
	"time"

	"GAI_test/config"
	"GAI_test/model"

	"github.com/gin-gonic/gin"
)

// CreateQuestion 创建题目
// @Summary 创建题目
// @Description 管理员创建新题目（单选题或判断题）
// @Accept json
// @Produce json
// @Param request body CreateQuestionRequest true "题目信息"
// @Success 200 {object} CreateQuestionResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /question/create [post]
func CreateQuestion(c *gin.Context) {
	// 验证管理员身份
	token := c.GetHeader("Authorization")
	if !validateAdminToken(token) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "无权限"})
		return
	}

	// 解析token获取管理员ID
	claims, err := validateToken(token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "token验证失败"})
		return
	}
	adminID := claims["userid"].(string)

	var req CreateQuestionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数不完整"})
		return
	}

	// 验证考试是否存在
	var exam model.Exam
	if err := config.DB.Where("exam_id = ?", req.ExamID).First(&exam).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "考试不存在"})
		return
	}

	// 验证题目类型
	var questionType model.QuestionType
	switch req.Type {
	case "single_choice":
		questionType = model.QuestionTypeSingleChoice
		// 验证单选题参数
		if req.Content == "" || req.OptionA == "" || req.OptionB == "" || req.OptionC == "" || req.OptionD == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "单选题必须提供题干和四个选项"})
			return
		}
		// 验证答案
		if req.Answer != "A" && req.Answer != "B" && req.Answer != "C" && req.Answer != "D" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "单选题答案必须是A、B、C、D之一"})
			return
		}
	case "judgment":
		questionType = model.QuestionTypeJudgment
		// 验证判断题参数
		if req.Content == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "判断题必须提供题干"})
			return
		}
		// 验证答案
		if req.Answer != "T" && req.Answer != "F" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "判断题答案必须是T或F"})
			return
		}
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "不支持的题目类型"})
		return
	}

	// 创建题目
	question := model.Question{
		ExamID:    req.ExamID,
		Type:      questionType,
		Content:   req.Content,
		OptionA:   req.OptionA,
		OptionB:   req.OptionB,
		OptionC:   req.OptionC,
		OptionD:   req.OptionD,
		Answer:    req.Answer,
		Score:     req.Score,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		CreatedBy: adminID,
	}

	if err := config.DB.Create(&question).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建题目失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "题目创建成功",
		"id":      question.ID,
		"type":    req.Type,
		"examid":  req.ExamID,
	})
}

// GetExamQuestions 获取考试题目列表
// @Summary 获取考试题目列表
// @Description 获取指定考试的所有题目
// @Accept json
// @Produce json
// @Param examid query string true "考试ID"
// @Success 200 {array} QuestionResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /question/list [get]
func GetExamQuestions(c *gin.Context) {
	// 验证管理员身份
	token := c.GetHeader("Authorization")
	if !validateAdminToken(token) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "无权限"})
		return
	}

	examID := c.Query("examid")
	if examID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "缺少考试ID"})
		return
	}

	// 验证考试是否存在
	var exam model.Exam
	if err := config.DB.Where("exam_id = ?", examID).First(&exam).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "考试不存在"})
		return
	}

	// 查询题目列表
	var questions []model.Question
	if err := config.DB.Where("exam_id = ?", examID).Find(&questions).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询题目失败"})
		return
	}

	// 构建响应
	var response []gin.H
	for _, q := range questions {
		// 转换题目类型为字符串
		typeStr := "single_choice"
		if q.Type == model.QuestionTypeJudgment {
			typeStr = "judgment"
		}

		response = append(response, gin.H{
			"id":       q.ID,
			"examid":   q.ExamID,
			"type":     typeStr,
			"content":  q.Content,
			"optionA":  q.OptionA,
			"optionB":  q.OptionB,
			"optionC":  q.OptionC,
			"optionD":  q.OptionD,
			"answer":   q.Answer,
			"score":    q.Score,
			"createAt": q.CreatedAt,
		})
	}

	c.JSON(http.StatusOK, response)
}

// GetQuestion 获取题目详情
// @Summary 获取题目详情
// @Description 获取指定题目的详细信息
// @Accept json
// @Produce json
// @Param id path int true "题目ID"
// @Success 200 {object} QuestionResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /question/{id} [get]
func GetQuestion(c *gin.Context) {
	// 验证管理员身份
	token := c.GetHeader("Authorization")
	if !validateAdminToken(token) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "无权限"})
		return
	}

	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "缺少题目ID"})
		return
	}

	// 查询题目
	var question model.Question
	if err := config.DB.Where("id = ?", id).First(&question).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "题目不存在"})
		return
	}

	// 转换题目类型为字符串
	typeStr := "single_choice"
	if question.Type == model.QuestionTypeJudgment {
		typeStr = "judgment"
	}

	// 构建响应
	c.JSON(http.StatusOK, gin.H{
		"id":       question.ID,
		"examid":   question.ExamID,
		"type":     typeStr,
		"content":  question.Content,
		"optionA":  question.OptionA,
		"optionB":  question.OptionB,
		"optionC":  question.OptionC,
		"optionD":  question.OptionD,
		"answer":   question.Answer,
		"score":    question.Score,
		"createAt": question.CreatedAt,
	})
}

// UpdateQuestion 更新题目
// @Summary 更新题目
// @Description 更新指定题目的信息
// @Accept json
// @Produce json
// @Param id path int true "题目ID"
// @Param request body UpdateQuestionRequest true "更新信息"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /question/{id} [put]
func UpdateQuestion(c *gin.Context) {
	// 验证管理员身份
	token := c.GetHeader("Authorization")
	if !validateAdminToken(token) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "无权限"})
		return
	}

	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "缺少题目ID"})
		return
	}

	var req UpdateQuestionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数不完整"})
		return
	}

	// 查询题目
	var question model.Question
	if err := config.DB.Where("id = ?", id).First(&question).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "题目不存在"})
		return
	}

	// 验证题目类型
	if req.Type != "" {
		switch req.Type {
		case "single_choice":
			question.Type = model.QuestionTypeSingleChoice
		case "judgment":
			question.Type = model.QuestionTypeJudgment
		default:
			c.JSON(http.StatusBadRequest, gin.H{"error": "不支持的题目类型"})
			return
		}
	}

	// 更新题目信息
	if req.Content != "" {
		question.Content = req.Content
	}
	if req.OptionA != "" {
		question.OptionA = req.OptionA
	}
	if req.OptionB != "" {
		question.OptionB = req.OptionB
	}
	if req.OptionC != "" {
		question.OptionC = req.OptionC
	}
	if req.OptionD != "" {
		question.OptionD = req.OptionD
	}
	if req.Answer != "" {
		// 验证答案
		if question.Type == model.QuestionTypeSingleChoice {
			if req.Answer != "A" && req.Answer != "B" && req.Answer != "C" && req.Answer != "D" {
				c.JSON(http.StatusBadRequest, gin.H{"error": "单选题答案必须是A、B、C、D之一"})
				return
			}
		} else if question.Type == model.QuestionTypeJudgment {
			if req.Answer != "T" && req.Answer != "F" {
				c.JSON(http.StatusBadRequest, gin.H{"error": "判断题答案必须是T或F"})
				return
			}
		}
		question.Answer = req.Answer
	}
	if req.Score > 0 {
		question.Score = req.Score
	}

	question.UpdatedAt = time.Now()

	// 保存更新
	if err := config.DB.Save(&question).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新题目失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "题目更新成功",
	})
}

// DeleteQuestion 删除题目
// @Summary 删除题目
// @Description 删除指定的题目
// @Accept json
// @Produce json
// @Param id path int true "题目ID"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /question/{id} [delete]
func DeleteQuestion(c *gin.Context) {
	// 验证管理员身份
	token := c.GetHeader("Authorization")
	if !validateAdminToken(token) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "无权限"})
		return
	}

	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "缺少题目ID"})
		return
	}

	// 查询题目是否存在
	var question model.Question
	if err := config.DB.Where("id = ?", id).First(&question).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "题目不存在"})
		return
	}

	// 删除题目
	if err := config.DB.Delete(&question).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除题目失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "题目删除成功",
	})
}

// GetQuestionTypes 获取支持的题目类型
// @Summary 获取支持的题目类型
// @Description 获取系统当前支持的题目类型
// @Produce json
// @Success 200 {object} QuestionTypesResponse
// @Router /question/types [get]
func GetQuestionTypes(c *gin.Context) {
	// 定义当前支持的题目类型
	questionTypes := []gin.H{
		{
			"type": "single_choice",
			"name": "单选题",
			"description": "从四个选项中选择一个正确答案",
		},
		{
			"type": "judgment",
			"name": "判断题",
			"description": "判断题目的正误，答案为T（正确）或F（错误）",
		},
	}

	c.JSON(http.StatusOK, gin.H{
		"questiontypes": questionTypes,
	})
}

// 用于Swagger文档的结构体定义
// CreateQuestionRequest 创建题目请求
type CreateQuestionRequest struct {
	ExamID   string `json:"examid" binding:"required" example:"EXAM001"` // 考试ID
	Type     string `json:"type" binding:"required" example:"single_choice" enums:"single_choice,judgment"` // 题目类型
	Content  string `json:"content" binding:"required" example:"以下哪个是正确的？"` // 题干
	OptionA  string `json:"optionA" example:"选项A"` // 选项A（单选题必填）
	OptionB  string `json:"optionB" example:"选项B"` // 选项B（单选题必填）
	OptionC  string `json:"optionC" example:"选项C"` // 选项C（单选题必填）
	OptionD  string `json:"optionD" example:"选项D"` // 选项D（单选题必填）
	Answer   string `json:"answer" binding:"required" example:"A"` // 正确答案
	Score    int    `json:"score" binding:"required" example:"5"` // 分值
}

// CreateQuestionResponse 创建题目响应
type CreateQuestionResponse struct {
	Message string `json:"message" example:"题目创建成功"`
	ID      uint   `json:"id" example:"1"`
	Type    string `json:"type" example:"single_choice"`
	ExamID  string `json:"examid" example:"EXAM001"`
}

// QuestionResponse 题目响应
type QuestionResponse struct {
	ID       uint      `json:"id" example:"1"`
	ExamID   string    `json:"examid" example:"EXAM001"`
	Type     string    `json:"type" example:"single_choice"`
	Content  string    `json:"content" example:"以下哪个是正确的？"`
	OptionA  string    `json:"optionA" example:"选项A"`
	OptionB  string    `json:"optionB" example:"选项B"`
	OptionC  string    `json:"optionC" example:"选项C"`
	OptionD  string    `json:"optionD" example:"选项D"`
	Answer   string    `json:"answer" example:"A"`
	Score    int       `json:"score" example:"5"`
	CreateAt time.Time `json:"createAt" example:"2023-01-01T12:00:00Z"`
}

// UpdateQuestionRequest 更新题目请求
type UpdateQuestionRequest struct {
	Type    string `json:"type" example:"single_choice" enums:"single_choice,judgment"` // 题目类型
	Content string `json:"content" example:"以下哪个是正确的？"` // 题干
	OptionA string `json:"optionA" example:"选项A"` // 选项A
	OptionB string `json:"optionB" example:"选项B"` // 选项B
	OptionC string `json:"optionC" example:"选项C"` // 选项C
	OptionD string `json:"optionD" example:"选项D"` // 选项D
	Answer  string `json:"answer" example:"A"` // 正确答案
	Score   int    `json:"score" example:"5"` // 分值
}

// QuestionTypesResponse 题目类型响应
type QuestionTypesResponse struct {
	QuestionTypes []QuestionTypeInfo `json:"questiontypes"`
}

// QuestionTypeInfo 题目类型信息
type QuestionTypeInfo struct {
	Type        string `json:"type" example:"single_choice"`
	Name        string `json:"name" example:"单选题"`
	Description string `json:"description" example:"从四个选项中选择一个正确答案"`
} 