package handler

import (
	"github.com/gin-gonic/gin"

	"opentraffic-ops-backend/internal/service"
	"opentraffic-ops-backend/pkg/response"
)

// HomeHandler 首页处理器
type HomeHandler struct {
	homeService *service.HomeService
}

// NewHomeHandler 创建首页处理器
func NewHomeHandler(homeService *service.HomeService) *HomeHandler {
	return &HomeHandler{homeService: homeService}
}

// RegisterRoutes 注册首页路由
func (h *HomeHandler) RegisterRoutes(r *gin.RouterGroup) {
	home := r.Group("/rtm/home")
	{
		home.GET("/getStatisticData", h.GetStatisticData)
		home.GET("/getMonProcessAbnormalList", h.GetMonProcessAbnormalList)
	}
}

// GetStatisticData 获取首页统计数据
func (h *HomeHandler) GetStatisticData(c *gin.Context) {
	data, err := h.homeService.GetStatisticData(c.Request.Context())
	if err != nil {
		response.Error(c, err.Error())
		return
	}

	response.Success(c, data)
}

// GetMonProcessAbnormalList 获取异常进程列表（预留接口）
func (h *HomeHandler) GetMonProcessAbnormalList(c *gin.Context) {
	// 预留接口，告警模块删除后暂时返回空列表
	response.SuccessPage(c, 0, []interface{}{})
}
