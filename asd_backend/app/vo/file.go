package vo

// FileVo 文件视图对象
type FileVo struct {
	FileID      string `json:"file_id"`      // 文件ID
	FileName    string `json:"file_name"`    // 文件名
	FilePath    string `json:"file_path"`    // 存储路径
	FileSize    int64  `json:"file_size"`    // 文件大小
	FileType    string `json:"file_type"`    // 文件类型
	ContentType string `json:"content_type"` // MIME类型
	Description string `json:"description"`  // 描述
	Metadata    string `json:"metadata"`     // 元数据
	Visibility  string `json:"visibility"`   // 可见性
	Status      string `json:"status"`       // 状态
	DownloadUrl string `json:"download_url"` // 下载链接
	CreatedAt   string `json:"created_at"`   // 创建时间
	Content     string `json:"-"`            // 文件内容
}
