package dto

// 报告分页请求
type ReportPageReq struct {
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
	Type   int `json:"type"`
	Year   int `json:"year"`
}

// 添加报告请求
type ReportAddReq struct {
	UserID     int    `form:"user_id" validate:"required"`
	ReportID   string `form:"report_id" validate:"required"`
	Name       string `form:"name" validate:"required"`
	Type       int    `form:"type" validate:"required"`
	ReportDate string `form:"report_date" validate:"required"`
}
