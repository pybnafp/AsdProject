package controllers

import (
	"asd/app/dto"
	"asd/app/services"
	"asd/utils/common"

	"github.com/gookit/validate"
)

var ReportApi = new(ReportApiController)

type ReportApiController struct {
	BaseController
}

func (c *ReportApiController) List() {
	var req dto.ReportPageReq
	if err := c.ParseJSON(&req); err != nil {
		c.ErrorJson(400, "请求参数错误"+err.Error())
		return
	}

	if req.Offset < 0 {
		req.Offset = 0
	}

	if req.Limit == 0 {
		req.Limit = 30
	}

	if req.Limit > 1000 {
		req.Limit = 1000
	}

	// 参数校验
	v := validate.Struct(req)
	if !v.Validate() {
		c.ErrorJson(400, v.Errors.One())
		return
	}

	// userId := c.GetUserId() // 从中间件获取用户ID
	userId := 1 // TODO: !目前demo的时候展示用户1的所有报告
	list, count, err := services.Report.GetList(req, userId)
	if err != nil {
		c.ErrorJson(400, err.Error())
		return
	}

	c.JSON(common.JsonResult{
		Code:  0,
		Data:  list,
		Count: count,
	})
}
