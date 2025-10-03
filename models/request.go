package models

type ProcessRequest struct {
    SourceDir    string `json:"sourceDir" binding:"required"`
    TargetDir    string `json:"targetDir"`
    Mode         string `json:"mode" binding:"required"` // "link", "move", "rename"
    RedirectPath string `json:"redirectPath"` // 重定向路径，用于Docker环境
}

type ProcessResponse struct {
    Success bool   `json:"success"`
    Message string `json:"message"`
    Data    any    `json:"data,omitempty"`
}