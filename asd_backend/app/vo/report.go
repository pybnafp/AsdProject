package vo

// ReportVo 文件视图对象
type ReportVo struct {
	ReportID         string `json:"report_id"` // 报告ID
	UserID           int    `json:"user_id"`
	Name             string `json:"name"`   // 患者姓名
	Type             int    `json:"type"`   // 类型
	Status           int    `json:"status"` // 状态
	OriginalFileName string `json:"original_file_name"`
	OriginalFileUrl  string `json:"original_file_url"`
	OriginalFileSize int64  `json:"original_file_size"`
	ReportFileName   string `json:"report_file_name"`
	ReportFileUrl    string `json:"report_file_url"`
	ReportFileSize   int64  `json:"report_file_size"`
	ReportDate       string `json:"report_date"` // 报告日期
	CreatedAt        string `json:"created_at"`  // 创建时间
	Content          string `json:"-"`
}
