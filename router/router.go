package router

import (
	"GAI_test/controller"

	"github.com/gin-gonic/gin"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func SetupRouter() *gin.Engine {
	r := gin.Default()

	// swagger文档路由
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))

	// 考试管理接口
	r.POST("/exam/create", controller.CreateExam)       // 创建考试
	r.POST("/exam/bind", controller.BindStudent)        // 学生绑定考试
	// 更新为使用登录接口：返回token用于后续验证
	r.POST("/validateExamPassword", controller.ValidateExamEntryWithToken)
	
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

	return r
}
