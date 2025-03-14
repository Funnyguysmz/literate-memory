package router

import (
	"GAI_test/controller"
	"GAI_test/websocket"

	"github.com/gin-gonic/gin"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func SetupRouter() *gin.Engine {
	r := gin.Default()

	// 初始化WebSocket管理器
	websocket.InitWSManager()
	// 初始化流管理器
	websocket.InitStreamManager()

	// swagger文档路由
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))

	// 用户管理接口
	r.POST("/user/register", controller.RegisterUser) // 注册新用户
	r.GET("/user/:userid", controller.GetUser)       // 获取用户信息

	// 考试管理接口
	r.POST("/exam/create", controller.CreateExam)    // 创建考试
	r.PUT("/exam/streamtype", controller.UpdateExamStreamType) // 更新考试流类型
	r.GET("/exam/streamtypes", controller.GetSupportedStreamTypes) // 获取支持的流类型
	r.POST("/exam/bind", controller.BindStudent)     // 学生绑定考试
	r.POST("/validateExamPassword", controller.ValidateExamEntryWithToken) // 登录接口
	
	// 管理接口: 批量生成密码
	r.POST("/generatePasswords", controller.GeneratePasswords)
	// 管理员获取密码接口
	r.GET("/admin/password", controller.GetAdminPassword)
	// 客户端接口: 获取未分发密码
	r.GET("/password", controller.GetPassword)
	// 举手请求接口
	r.POST("/raiseHand", controller.RaiseHand)
	// 测试接口：删除所有管理员
	r.DELETE("/admin/deleteAll", controller.DeleteAllAdmins)
	// 获取顶级管理员信息
	r.GET("/admin/backup", controller.GetAdminBackup)
	// 更新管理员信息
	r.POST("/admin/update", controller.UpdateAdmin)
	// 使token失效接口
	r.POST("/token/invalidate", controller.InvalidateToken)

	// 题目管理接口
	r.POST("/question/create", controller.CreateQuestion)    // 创建题目
	r.GET("/question/list", controller.GetExamQuestions)     // 获取考试题目列表
	r.GET("/question/:id", controller.GetQuestion)           // 获取题目详情
	r.PUT("/question/:id", controller.UpdateQuestion)        // 更新题目
	r.DELETE("/question/:id", controller.DeleteQuestion)     // 删除题目
	r.GET("/question/types", controller.GetQuestionTypes)    // 获取支持的题目类型

	// 屏幕流相关接口
	r.POST("/screen/init", controller.InitScreenStream)   // 初始化屏幕流
	r.GET("/screen/active", controller.GetActiveStreams)  // 获取活跃的屏幕流
	r.POST("/screen/stop", controller.StopScreenStream)   // 停止屏幕流

	// WebSocket接口
	r.GET("/ws/student", websocket.HandleStudentWS)  // 学生WebSocket连接
	r.GET("/ws/admin", websocket.HandleAdminWS)      // 管理员WebSocket连接

	return r
}
