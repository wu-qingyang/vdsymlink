package handlers

import (
    "net/http"
    "net/url"
    "os"
    "path/filepath"
    "strings"
    "vdsymlink-web/models"
    "vdsymlink-web/services"

    "github.com/gin-gonic/gin"
)

type SymlinkHandler struct {
    service *services.SymlinkService
}

func NewSymlinkHandler() *SymlinkHandler {
    return &SymlinkHandler{
        service: services.NewSymlinkService(),
    }
}

// ProcessFiles 处理文件操作 - 支持表单和JSON
func (h *SymlinkHandler) ProcessFiles(c *gin.Context) {
    var req models.ProcessRequest

    // 根据Content-Type决定如何绑定数据
    contentType := c.GetHeader("Content-Type")
    if contentType == "application/x-www-form-urlencoded" {
        // 表单提交
        req.SourceDir = c.PostForm("sourceDir")
        req.TargetDir = c.PostForm("targetDir")
        req.Mode = c.PostForm("mode")
        req.RedirectPath = c.PostForm("redirectPath")
    } else {
        // JSON提交
        if err := c.ShouldBindJSON(&req); err != nil {
            c.JSON(http.StatusBadRequest, models.ProcessResponse{
                Success: false,
                Message: "请求参数错误: " + err.Error(),
            })
            return
        }
    }

    // 验证必填字段
    if req.SourceDir == "" {
        if contentType == "application/x-www-form-urlencoded" {
            c.HTML(http.StatusOK, "index.html", gin.H{
                "title": "VdSYMLinkTool",
                "error": "视频目录路径不能为空",
                "sourceDir": req.SourceDir,
                "targetDir": req.TargetDir,
                "mode": req.Mode,
                "redirectPath": req.RedirectPath,
            })
        } else {
            c.JSON(http.StatusBadRequest, models.ProcessResponse{
                Success: false,
                Message: "视频目录路径不能为空",
            })
        }
        return
    }

    if req.Mode == "" {
        req.Mode = "rename"
    }

    if req.Mode == "rename" {
        req.TargetDir = ""
        req.RedirectPath = ""
    } else if req.TargetDir == "" {
        if contentType == "application/x-www-form-urlencoded" {
            c.HTML(http.StatusOK, "index.html", gin.H{
                "title": "VdSYMLinkTool",
                "error": "链接/移动模式需要填写目标目录路径",
                "sourceDir": req.SourceDir,
                "targetDir": req.TargetDir,
                "mode": req.Mode,
                "redirectPath": req.RedirectPath,
            })
        } else {
            c.JSON(http.StatusBadRequest, models.ProcessResponse{
                Success: false,
                Message: "链接/移动模式需要填写目标目录路径",
            })
        }
        return
    }

    result, err := h.service.ProcessFiles(req.SourceDir, req.TargetDir, req.Mode, req.RedirectPath)

    // 关键修改：表单提交时使用重定向
    if contentType == "application/x-www-form-urlencoded" {
        if err != nil {
            // 错误情况直接返回页面（用户需要看到错误信息并修正）
            c.HTML(http.StatusOK, "index.html", gin.H{
                "title": "VdSYMLinkTool",
                "error": "处理失败: " + err.Error(),
                "sourceDir": req.SourceDir,
                "targetDir": req.TargetDir,
                "mode": req.Mode,
                "redirectPath": req.RedirectPath,
            })
        } else {
            // 成功时重定向，避免重复提交
            c.Redirect(http.StatusSeeOther, "/?success=true&result="+url.QueryEscape(result)+
                "&sourceDir="+url.QueryEscape(req.SourceDir)+
                "&targetDir="+url.QueryEscape(req.TargetDir)+
                "&mode="+url.QueryEscape(req.Mode)+
                "&redirectPath="+url.QueryEscape(req.RedirectPath))
        }
    } else {
        // JSON响应保持不变
        if err != nil {
            c.JSON(http.StatusInternalServerError, models.ProcessResponse{
                Success: false,
                Message: "处理失败: " + err.Error(),
            })
        } else {
            c.JSON(http.StatusOK, models.ProcessResponse{
                Success: true,
                Message: "处理完成",
                Data:    result,
            })
        }
    }
}

// GetIndex 显示首页（支持查询参数）
func (h *SymlinkHandler) GetIndex(c *gin.Context) {
    result := c.Query("result")
    success := c.Query("success") == "true"
    sourceDir := c.Query("sourceDir")
    targetDir := c.Query("targetDir")
    mode := c.Query("mode")
    redirectPath := c.Query("redirectPath")

    c.HTML(http.StatusOK, "index.html", gin.H{
        "title":        "VdSYMLinkTool",
        "result":       result,
        "success":      success,
        "sourceDir":    sourceDir,
        "targetDir":    targetDir,
        "mode":         mode,
        "redirectPath": redirectPath,
    })
}

func getParentPath(path string) string {
    if path == "" || path == "/" {
        return "" // 根目录没有父目录
    }

    parent := filepath.Dir(path)
    if parent == path {
        return "" // 避免无限循环
    }
    return parent
}

// ListDirectories 列出目录
func (h *SymlinkHandler) ListDirectories(c *gin.Context) {
    path := c.Query("path")
    if path == "" {
        path = "/"
    }

    var directories []map[string]string

    // 无论路径是否存在，先添加上级目录项
    parentPath := getParentPath(path)
    if parentPath != path && parentPath != "" {
        directories = append(directories, map[string]string{
            "name": "..",
            "path": parentPath,
            "type": "parent",
        })
    }

    // 尝试读取目录
    entries, err := os.ReadDir(path)
    if err != nil {
        // 即使出错，也返回上级目录信息
        c.JSON(http.StatusOK, gin.H{
            "success":     false,
            "message":     "无法读取目录: " + err.Error(),
            "currentPath": path,
            "directories": directories, // 仍然返回上级目录
        })
        return
    }

    // 使用map来去重目录名
    seen := make(map[string]bool)

    // 添加当前目录的子目录
    for _, entry := range entries {
        if entry.IsDir() {
            // 跳过隐藏目录（以 . 开头）
            if strings.HasPrefix(entry.Name(), ".") {
                continue
            }

            // 去重检查
            if !seen[entry.Name()] {
                seen[entry.Name()] = true
                fullPath := filepath.Join(path, entry.Name())
                directories = append(directories, map[string]string{
                    "name": entry.Name(),
                    "path": fullPath,
                    "type": "directory",
                })
            }
        }
    }

    c.JSON(http.StatusOK, gin.H{
        "success": true,
        "currentPath": path,
        "directories": directories,
    })
}