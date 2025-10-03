package services

import (
    "fmt"
    "os"
    "path/filepath"
    "regexp"
    "runtime"
    "strconv"
    "strings"
)

// 预编译正则表达式
var (
    episodePatterns = []string{
        `[第]([0-9]{1,3})[话話集]`,           // 第y话、第y集
        `[Ee][Pp]?([0-9]{1,3})([^0-9]|$)`, // Ey、EPy
        `\[([0-9]{1,3})\]([^0-9]|$)`,      // [y]
        `([0-9]{1,3})[话話集]([^0-9]|$)`,   // y话、y集
    }
    seasonPatterns = []string{
        `^[Ss]([0-9]{1,2})$`,                    // S1, S01
        `[Ss]eason[._ -]*([0-9]{1,2})`,          // Season 1, Season.1
        `[Ss]([0-9]{1,2})[^0-9]`,                // S1_, S01E
        `第([0-9]{1,2})[季期]`,                   // 第1季, 第1期
    }
    precompiledEpisodeRegexes []*regexp.Regexp
    precompiledSeasonRegexes  []*regexp.Regexp
)

func init() {
    // 预编译所有正则表达式
    precompiledEpisodeRegexes = compilePatterns(episodePatterns)
    precompiledSeasonRegexes = compilePatterns(seasonPatterns)
}

type SymlinkService struct{}

func NewSymlinkService() *SymlinkService {
    return &SymlinkService{}
}

func (s *SymlinkService) ProcessFiles(sourceDir, targetDir, mode, redirectPath string) (string, error) {
    var result strings.Builder

    switch mode {
    case "rename":
        return s.renameMode(sourceDir, &result)
    case "link", "move":
        return s.linkMoveMode(sourceDir, targetDir, mode == "move", redirectPath, &result)
    default:
        return "", fmt.Errorf("不支持的模式: %s", mode)
    }
}

func (s *SymlinkService) initializeProcessing(sourceDir, targetDir string, result *strings.Builder) ([]string, string, string, string, error) {
    absSourceDir, err := filepath.Abs(sourceDir)
    if err != nil {
        return nil, "", "", "", fmt.Errorf("无法获取绝对路径: %v", err)
    }

    videoFiles, err := s.getVideoFiles(absSourceDir)
    if err != nil {
        return nil, "", "", "", err
    }

    if len(videoFiles) == 0 {
        return nil, "", "", "", fmt.Errorf("源目录 '%s' 中没有找到视频文件 (.mkv 或 .mp4)", absSourceDir)
    }

    // 对于rename模式，targetDir使用sourceDir
    effectiveTargetDir := targetDir
    if targetDir == "" {
        effectiveTargetDir = sourceDir
    }

    seriesName, seasonNumber, finalTargetDir := s.getSeriesInfo(absSourceDir, effectiveTargetDir, videoFiles)

    fmt.Fprintf(result, "使用季数: S%s\n", seasonNumber)
    fmt.Fprintf(result, "使用剧集名: %s\n", seriesName)

    return videoFiles, seriesName, seasonNumber, finalTargetDir, nil
}

func (s *SymlinkService) renameMode(sourceDir string, result *strings.Builder) (string, error) {
    videoFiles, seriesName, seasonNumber, finalTargetDir, err := s.initializeProcessing(sourceDir, sourceDir, result)
    if err != nil {
        return "", err
    }

    processedFiles := s.processFiles(videoFiles, finalTargetDir, seriesName, seasonNumber, true, true, "", result)

    if processedFiles > 0 {
        fmt.Fprintf(result, "完成! 共重命名了 %d 个文件\n", processedFiles)
    } else {
        result.WriteString("所有文件都已正确命名，无需处理\n")
    }

    return result.String(), nil
}

func (s *SymlinkService) linkMoveMode(sourceDir, targetDir string, moveFiles bool, redirectPath string, result *strings.Builder) (string, error) {
    if err := s.validatePaths(sourceDir, targetDir); err != nil {
        return "", err
    }

    if err := s.ensureDirectoryExists(targetDir); err != nil {
        return "", fmt.Errorf("无法创建目标目录: %v", err)
    }

    videoFiles, seriesName, seasonNumber, finalTargetDir, err := s.initializeProcessing(sourceDir, targetDir, result)
    if err != nil {
        return "", err
    }

    // 只在有重定向路径时显示源文件路径
    if redirectPath != "" {
        fmt.Fprintf(result, "源文件路径: %s\n", sourceDir)
        fmt.Fprintf(result, "使用重定向路径: %s\n", redirectPath)
    }

    processedFiles := s.processFiles(videoFiles, finalTargetDir, seriesName, seasonNumber, moveFiles, false, redirectPath, result)

    action := "创建链接"
    if moveFiles {
        action = "移动文件"
    }
    if processedFiles > 0 {
        fmt.Fprintf(result, "完成! 共%s %d 个文件\n", action, processedFiles)
    } else {
        fmt.Fprintf(result, "没有文件需要%s\n", action)
    }

    return result.String(), nil
}

// 编译正则表达式模式
func compilePatterns(patterns []string) []*regexp.Regexp {
    var regexes []*regexp.Regexp
    for _, pattern := range patterns {
        regexes = append(regexes, regexp.MustCompile(pattern))
    }
    return regexes
}

// 格式化数字为两位数
func formatNumber(numStr string) string {
    if num, err := strconv.Atoi(numStr); err == nil {
        return fmt.Sprintf("%02d", num)
    }
    return numStr
}

// 确保目录存在
func (s *SymlinkService) ensureDirectoryExists(dirPath string) error {
    if err := os.MkdirAll(dirPath, 0755); err != nil {
        return fmt.Errorf("无法创建目录 %s: %v", dirPath, err)
    }
    return nil
}

// 使用预编译的正则表达式匹配模式
func matchPatterns(text string, regexes []*regexp.Regexp) (string, bool) {
    for _, re := range regexes {
        matches := re.FindStringSubmatch(text)
        if len(matches) > 1 {
            return matches[1], true
        }
    }
    return "", false
}

// 提取集数并格式化
func extractEpisodeNumber(filename string) (string, bool) {
    if episode, found := matchPatterns(filename, precompiledEpisodeRegexes); found {
        return formatNumber(episode), true
    }

    // 尝试匹配纯数字（排除季数上下文）
    re := regexp.MustCompile(`([^Ss]|^)([0-9]{1,3})([^0-9]|$)`)
    matches := re.FindStringSubmatch(filename)
    if len(matches) > 2 {
        beforePart := matches[1]
        episode := matches[2]
        // 检查是否在季数后面
        seasonRe := regexp.MustCompile(`[Ss][0-9]*$`)
        if !seasonRe.MatchString(beforePart) {
            return formatNumber(episode), true
        }
    }

    return "", false
}

// 从文件名中提取季数
func extractSeasonFromFilename(filename string) (string, bool) {
    // 使用预编译的正则匹配
    if season, found := matchPatterns(filename, precompiledSeasonRegexes); found {
        return formatNumber(season), true
    }
    return "", false
}

// 生成新文件名
func generateNewFilename(originalName, seriesName, seasonNumber, fileExtension string, isMovie bool) string {
    if isMovie {
        return seriesName + fileExtension
    }

    if episodeNumber, found := extractEpisodeNumber(originalName); found {
        return fmt.Sprintf("%s.S%sE%s%s", seriesName, seasonNumber, episodeNumber, fileExtension)
    }

    return originalName
}

// 获取视频文件列表
func (s *SymlinkService) getVideoFiles(sourceDir string) ([]string, error) {
    var videoFiles []string

    entries, err := os.ReadDir(sourceDir)
    if err != nil {
        return nil, fmt.Errorf("无法读取源目录: %v", err)
    }

    for _, entry := range entries {
        if !entry.IsDir() {
            filename := entry.Name()
            ext := strings.ToLower(filepath.Ext(filename))
            if ext == ".mkv" || ext == ".mp4" {
                fullPath := filepath.Join(sourceDir, filename)
                videoFiles = append(videoFiles, fullPath)
            }
        }
    }

    return videoFiles, nil
}

// 检测季数
func (s *SymlinkService) detectSeason(targetDir string, videoFiles []string) string {
    basename := filepath.Base(targetDir)

    // 检查目标路径是否匹配季数模式
    if season, found := matchPatterns(basename, precompiledSeasonRegexes); found {
        return formatNumber(season)
    }

    // 检查目标路径的父目录是否匹配季数模式
    parentDir := filepath.Dir(targetDir)
    parentBasename := filepath.Base(parentDir)
    if season, found := matchPatterns(parentBasename, precompiledSeasonRegexes); found {
        return formatNumber(season)
    }

    // 从源文件名中提取季数
    for _, file := range videoFiles {
        filename := filepath.Base(file)
        if season, found := extractSeasonFromFilename(filename); found {
            return season
        }
    }

    // 默认第一季
    return "01"
}

// 路径验证
func (s *SymlinkService) validatePaths(sourceDir, targetDir string) error {
    absSourceDir, _ := filepath.Abs(sourceDir)
    absTargetDir, _ := filepath.Abs(targetDir)

    // 检查源路径和目标路径是否相同
    if absSourceDir == absTargetDir {
        return fmt.Errorf("源路径和目标路径不能相同: %s", sourceDir)
    }

    // 检查源目录是否存在
    if _, err := os.Stat(sourceDir); os.IsNotExist(err) {
        return fmt.Errorf("源路径不存在: %s", sourceDir)
    }

    // 检查源目录是否为空
    entries, err := os.ReadDir(sourceDir)
    if err != nil {
        return fmt.Errorf("无法读取源目录: %v", err)
    }
    if len(entries) == 0 {
        return fmt.Errorf("源目录为空: %s", sourceDir)
    }

    return nil
}

// 确定剧集目标目录
func (s *SymlinkService) determineSeriesTargetDir(sourceDir, targetDir, seasonNumber string) string {
    sourceBasename := filepath.Base(sourceDir)
    targetBasename := filepath.Base(targetDir)

    var finalDir string

    if targetBasename == sourceBasename {
        finalDir = filepath.Join(targetDir, "S"+seasonNumber)
    } else {
        finalDir = filepath.Join(targetDir, sourceBasename, "S"+seasonNumber)
    }

    // 确保目录被创建
    s.ensureDirectoryExists(finalDir)

    return finalDir
}

// 确定最终目标目录结构
func (s *SymlinkService) determineTargetDirectory(sourceDir, targetDir string, videoFiles []string, isSeasonDir bool) (string, string, string) {
    absSourceDir, _ := filepath.Abs(sourceDir)
    absTargetDir, _ := filepath.Abs(targetDir)
    targetBasename := filepath.Base(absTargetDir)

    var seasonNumber, seriesName, finalTargetDir string

    if isSeasonDir {
        // 目标路径是季数目录
        if season, found := matchPatterns(targetBasename, precompiledSeasonRegexes); found {
            seasonNumber = formatNumber(season)
        } else {
            seasonNumber = "01"
        }

        // 获取目标路径的父目录名作为剧集名
        targetParentDir := filepath.Dir(absTargetDir)
        seriesName = filepath.Base(targetParentDir)
        finalTargetDir = absTargetDir
    } else {
        // 使用源目录名作为剧集名
        seriesName = filepath.Base(absSourceDir)
        seasonNumber = s.detectSeason(absTargetDir, videoFiles)

        if len(videoFiles) == 1 {
            // 单个文件，认为是电影
            finalTargetDir = absTargetDir
        } else {
            finalTargetDir = s.determineSeriesTargetDir(absSourceDir, absTargetDir, seasonNumber)
        }
    }

    return seriesName, seasonNumber, finalTargetDir
}

// 处理文件冲突
func (s *SymlinkService) handleFileConflict(targetFile string) error {
    // 只删除符号链接
    fileInfo, err := os.Lstat(targetFile)
    if err != nil {
        if os.IsNotExist(err) {
            return nil // 文件不存在，无需处理
        }
        return err // 其他错误
    }

    // 如果是符号链接，则删除
    if fileInfo.Mode()&os.ModeSymlink != 0 {
        if err := os.Remove(targetFile); err != nil {
            return fmt.Errorf("无法删除已存在的符号链接: %v", err)
        }
    }

    return nil
}

// 处理单个文件
func (s *SymlinkService) processSingleFile(file, finalTargetDir, seriesName, seasonNumber string, moveFiles bool, isMovie bool, isRenameMode bool, redirectPath string, result *strings.Builder) (bool, error) {
    filename := filepath.Base(file)
    fileExtension := filepath.Ext(filename)

    newFilename := generateNewFilename(filename, seriesName, seasonNumber, fileExtension, isMovie)
    targetFile := filepath.Join(finalTargetDir, newFilename)

    // 在重命名模式下，如果新旧文件名相同，说明文件已经正确命名，跳过
    if isRenameMode && filename == newFilename {
        fmt.Fprintf(result, "文件 '%s' 已正确命名，跳过\n", filename)
        return false, nil
    }

    // 处理文件冲突
    if err := s.handleFileConflict(targetFile); err != nil {
        return false, err
    }

    // 移动或创建符号链接
    err := s.linkMoveFile(file, targetFile, moveFiles, filename, newFilename, isRenameMode, redirectPath, result)
    return err == nil, err
}

// 移动或链接文件
func (s *SymlinkService) linkMoveFile(source, target string, moveFiles bool, originalName, newName string, isRenameMode bool, redirectPath string, result *strings.Builder) error {
    var err error

    if moveFiles {
        // 移动文件：source -> target
        err = os.Rename(source, target)
    } else {
        // 创建符号链接：target -> source（新文件指向原始文件）
        linkTarget := source
        if redirectPath != "" {
            // 计算重定向路径
            linkTarget = s.calculateRedirectPath(source, redirectPath)
        }
        err = os.Symlink(linkTarget, target)
    }

    if err != nil {
        return s.formatFileOperationError(err, moveFiles, newName)
    }

    // 统一显示格式
    if isRenameMode {
        // 格式化命名只显示文件名
        fmt.Fprintf(result, "重命名文件: %s -> %s\n", originalName, newName)
    } else {
        action := "创建链接"

        if moveFiles {
            action = "移动文件"
            // 移动文件显示完整路径
            fmt.Fprintf(result, "%s: %s -> %s\n", action, source, target)
        } else {
            // 创建链接显示完整路径
            linkTarget := source
            if redirectPath != "" {
                linkTarget = s.calculateRedirectPath(source, redirectPath)
            }
            fmt.Fprintf(result, "%s: %s -> %s\n", action, target, linkTarget)
        }
    }
    return nil
}

// 计算重定向路径
func (s *SymlinkService) calculateRedirectPath(originalPath, redirectPath string) string {
    // 获取源文件的直接父目录名
    parentDir := filepath.Base(filepath.Dir(originalPath))
    filename := filepath.Base(originalPath)

    // 组合：重定向路径 + 父目录名 + 文件名
    return filepath.Join(redirectPath, parentDir, filename)
}

// 格式化文件操作错误信息
func (s *SymlinkService) formatFileOperationError(err error, moveFiles bool, filename string) error {
    if moveFiles {
        return fmt.Errorf("无法移动 '%s': %v", filename, err)
    } else {
        if runtime.GOOS == "windows" {
            return fmt.Errorf("创建符号链接失败，请以管理员身份运行或启用开发者模式: %v", err)
        } else {
            return fmt.Errorf("无法创建符号链接 '%s': %v", filename, err)
        }
    }
}

// 处理文件（移动或创建链接）
func (s *SymlinkService) processFiles(videoFiles []string, finalTargetDir, seriesName, seasonNumber string, moveFiles bool, isRenameMode bool, redirectPath string, result *strings.Builder) int {
    processedFiles := 0

    // 确保目标目录存在
    if err := s.ensureDirectoryExists(finalTargetDir); err != nil {
        fmt.Fprintf(result, "错误: %v\n", err)
        return 0
    }

    isMovie := len(videoFiles) == 1

    for _, file := range videoFiles {
        processed, err := s.processSingleFile(file, finalTargetDir, seriesName, seasonNumber, moveFiles, isMovie, isRenameMode, redirectPath, result)
        if err != nil {
            fmt.Fprintf(result, "错误: %v\n", err)
            continue
        }
        if processed {
            processedFiles++
        }
    }

    return processedFiles
}

// 智能获取剧集名和季数
func (s *SymlinkService) getSeriesInfo(sourceDir, targetDir string, videoFiles []string) (string, string, string) {
    absTargetDir, _ := filepath.Abs(targetDir)
    targetBasename := filepath.Base(absTargetDir)

    // 检查目标路径是否已经是季数目录
    isSeasonDir := regexp.MustCompile(`^[Ss][0-9]`).MatchString(targetBasename)

    return s.determineTargetDirectory(sourceDir, targetDir, videoFiles, isSeasonDir)
}