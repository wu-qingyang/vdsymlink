class DirectoryBrowser {
    constructor(inputId, browseBtnId) {
        this.input = document.getElementById(inputId);
        this.browseBtn = document.getElementById(browseBtnId);
        this.dropdown = null;
        this.currentPath = '';
        this.autoBrowseEnabled = true;
        this.isLoading = false;
        this.lastProcessedPath = '';

        this.init();
    }

    init() {
        this.createBrowseButton();
        this.createDropdown();
        this.bindEvents();
    }

    createBrowseButton() {
        if (!this.browseBtn) {
            const button = document.createElement('button');
            button.type = 'button';
            button.id = this.input.id + '-browse';
            button.className = 'browse-btn';
            button.innerHTML = '📁 浏览';
            button.title = '浏览目录';

            this.input.parentNode.insertBefore(button, this.input.nextSibling);
            this.browseBtn = button;
        }
    }

    createDropdown() {
        this.dropdown = document.createElement('div');
        this.dropdown.className = 'directory-dropdown';
        this.dropdown.innerHTML = `
            <div class="dropdown-header">
                <span class="current-path" id="current-path"></span>
                <button type="button" class="close-dropdown">&times;</button>
            </div>
            <div class="dropdown-loading" style="display: none;">加载中...</div>
            <div class="directory-list" id="directory-list"></div>
        `;

        this.input.parentNode.appendChild(this.dropdown);
    }

    bindEvents() {
        this.browseBtn.addEventListener('click', (e) => {
            e.stopPropagation();
            this.toggleDropdown();
        });

        this.input.addEventListener('input', (e) => {
            this.handleInputChange(e.target.value);
        });

        this.input.addEventListener('focus', () => {
            if (this.dropdown.style.display !== 'block' && this.input.value) {
                this.showDropdownForCurrentPath();
            }
        });

        this.dropdown.querySelector('.close-dropdown').addEventListener('click', (e) => {
            e.stopPropagation();
            this.hideDropdown();
        });

        this.dropdown.querySelector('.current-path').addEventListener('click', (e) => {
            e.stopPropagation();
            this.selectCurrentPath();
        });

        document.addEventListener('click', (e) => {
            if (!this.input.contains(e.target) &&
                !this.dropdown.contains(e.target) &&
                !this.browseBtn.contains(e.target)) {
                this.hideDropdown();
            }
        });

        this.input.addEventListener('keydown', (e) => {
            if (e.key === 'Escape') {
                this.hideDropdown();
            }
        });

        this.dropdown.addEventListener('click', (e) => {
            e.stopPropagation();
        });
    }

    handleInputChange(value) {
        if (!this.autoBrowseEnabled || this.isLoading) return;

        if (this.dropdown.style.display !== 'block' && (value.endsWith('/') || value.endsWith('\\'))) {
            let pathToLoad = value.slice(0, -1); // 移除结尾的斜杠

            // 特殊处理根目录
            if (pathToLoad === '') {
                pathToLoad = '/';
            }

            this.showDropdownForPath(pathToLoad);
            return;
        }

        if (this.dropdown.style.display === 'block') {
            this.handlePathChange(value);
        }
    }

    handlePathChange(value) {
        if (value.endsWith('/') || value.endsWith('\\')) {
            const pathWithoutSlash = value.slice(0, -1);
            if (pathWithoutSlash !== this.lastProcessedPath && pathWithoutSlash.length > 0) {
                this.lastProcessedPath = pathWithoutSlash;
                this.loadDirectories(pathWithoutSlash);
            }
        } else {
            const parentPath = this.getParentPath(value);
            if (parentPath && parentPath !== this.lastProcessedPath) {
                this.lastProcessedPath = parentPath;
                this.loadDirectories(parentPath);
            }
        }
    }

    getParentPath(path) {
        if (!path) return '';
        const normalizedPath = path.replace(/\\/g, '/');
        const lastSlashIndex = normalizedPath.lastIndexOf('/');
        return lastSlashIndex === -1 ? '' : normalizedPath.substring(0, lastSlashIndex);
    }

    async showDropdownForCurrentPath() {
        this.showDropdown();
        const value = this.input.value;

        if (value.endsWith('/') || value.endsWith('\\')) {
            const pathWithoutSlash = value.slice(0, -1);
            if (pathWithoutSlash.length > 0) {
                await this.loadDirectories(pathWithoutSlash);
            } else {
                await this.loadDirectories();
            }
        } else {
            const parentPath = this.getParentPath(value);
            if (parentPath) {
                await this.loadDirectories(parentPath);
            } else {
                await this.loadDirectories();
            }
        }
    }

    async showDropdownForPath(path) {
        this.showDropdown();
        await this.loadDirectories(path);
    }

    toggleDropdown() {
        if (this.dropdown.style.display === 'block') {
            this.hideDropdown();
        } else {
            this.showDropdownForCurrentPath();
        }
    }

    async showDropdown() {
        this.dropdown.style.display = 'block';
    }

    hideDropdown() {
        this.dropdown.style.display = 'none';
        this.lastProcessedPath = '';
    }

    selectCurrentPath() {
        if (this.currentPath) {
            this.input.value = this.currentPath === '/' ? '/' : this.currentPath + '/';
            this.input.focus();
            this.hideDropdown();
        }
    }

    async loadDirectories(path = '') {
        if (this.isLoading) return;

        this.isLoading = true;

        const directoryList = this.dropdown.querySelector('#directory-list');
        const currentPathSpan = this.dropdown.querySelector('#current-path');
        const loadingElement = this.dropdown.querySelector('.dropdown-loading');

        loadingElement.style.display = 'block';
        directoryList.innerHTML = '';

        try {
            const params = new URLSearchParams();
            if (path) params.append('path', path);

            const response = await fetch(`/api/directories?${params}`);
            const data = await response.json();

            // 无论成功还是失败，都更新当前路径
            this.currentPath = data.currentPath || path;
            currentPathSpan.textContent = this.currentPath === '/' ? '/' : this.currentPath + '/';

            const fragment = document.createDocumentFragment();

            // 无论 success 是 true 还是 false，都渲染目录列表
            if (data.directories && data.directories.length > 0) {
                data.directories.forEach(dir => {
                    const dirElement = document.createElement('div');
                    dirElement.className = 'directory-item';

                    if (dir.type === 'parent') {
                        dirElement.className += ' parent-directory';
                        dirElement.innerHTML = `
                            <span class="dir-icon">↶</span>
                            <span class="dir-name">${this.escapeHtml(dir.name)}</span>
                            <span class="dir-desc">（上级目录）</span>
                        `;
                    } else {
                        dirElement.innerHTML = `
                            <span class="dir-icon">📁</span>
                            <span class="dir-name">${this.escapeHtml(dir.name)}</span>
                        `;
                    }

                    dirElement.addEventListener('click', (e) => {
                        e.stopPropagation();
                        this.input.value = dir.path === '/' ? '/' : dir.path + '/';
                        this.input.focus();
                        this.loadDirectories(dir.path);
                    });

                    fragment.appendChild(dirElement);
                });
            }

            // 如果有错误信息，在目录后面显示
            if (!data.success && data.message) {
                const errorElement = document.createElement('div');
                errorElement.className = 'error';
                errorElement.textContent = `错误: ${data.message}`;
                fragment.appendChild(errorElement);
            }

            // 如果既没有目录也没有错误信息，显示空状态
            if (fragment.children.length === 0) {
                directoryList.innerHTML = '<div class="empty">该目录下没有子目录</div>';
            } else {
                directoryList.appendChild(fragment);
            }

        } catch (error) {
            this.currentPath = path;
            currentPathSpan.textContent = this.currentPath === '/' ? '/' : this.currentPath + '/';
            directoryList.innerHTML = `<div class="error">加载失败: ${error.message}</div>`;
        } finally {
            loadingElement.style.display = 'none';
            this.isLoading = false;
        }
    }

    escapeHtml(unsafe) {
        return unsafe
            .replace(/&/g, "&amp;")
            .replace(/</g, "&lt;")
            .replace(/>/g, "&gt;")
            .replace(/"/g, "&quot;")
            .replace(/'/g, "&#039;");
    }
}

// 表单提交处理 - 使用 AJAX
function handleFormSubmit(event) {
    event.preventDefault(); // 阻止默认的表单提交行为

    const form = document.getElementById('mainForm');

    // 收集表单数据
    const formData = {
        sourceDir: document.getElementById('sourceDir').value,
        targetDir: document.getElementById('targetDir').value,
        mode: document.querySelector('input[name="mode"]:checked').value,
        redirectPath: document.getElementById('redirectPath').value
    };

    // 使用 JSON 格式提交
    fetch('/api/process', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify(formData)
    })
    .then(response => response.json())
    .then(data => {
        // 显示处理结果
        showProcessResult(data);
    })
    .catch(error => {
        // 显示错误信息
        showProcessResult({
            success: false,
            message: '网络错误: ' + error.message
        });
    });

    return false;
}

// 显示处理结果
function showProcessResult(data) {
    // 移除之前的结果
    const oldResult = document.querySelector('.result-container');
    if (oldResult) {
        oldResult.remove();
    }

    // 创建新的结果容器
    const resultDiv = document.createElement('div');
    resultDiv.className = `result-container ${data.success ? 'success' : 'error'}`;

    let resultContent = '';
    if (data.success) {
        resultContent = `
            <h3>处理结果:</h3>
            <pre>${data.data || data.message}</pre>
        `;
    } else {
        resultContent = `
            <h3>错误:</h3>
            <pre>${data.message}</pre>
        `;
    }

    resultDiv.innerHTML = resultContent;

    // 插入到表单后面
    const form = document.getElementById('mainForm');
    form.parentNode.insertBefore(resultDiv, form.nextSibling);

    // 滚动到结果位置
    resultDiv.scrollIntoView({ behavior: 'smooth' });
}

// 模式切换功能
function toggleMode() {
    const renameMode = document.querySelector('input[name="mode"][value="rename"]').checked;
    const linkMode = document.querySelector('input[name="mode"][value="link"]').checked;
    const targetDirGroup = document.getElementById('targetDirGroup');
    const redirectPathGroup = document.getElementById('redirectPathGroup');

    if (renameMode) {
        targetDirGroup.style.display = 'none';
        redirectPathGroup.style.display = 'none';
    } else {
        targetDirGroup.style.display = 'block';
        if (linkMode) {
            redirectPathGroup.style.display = 'block';
        } else {
            redirectPathGroup.style.display = 'none';
        }
    }
}

// 页面初始化
function initializeApp() {
    toggleMode();

    // 初始化目录浏览器
    new DirectoryBrowser('sourceDir', 'sourceDir-browse');
    new DirectoryBrowser('targetDir', 'targetDir-browse');

    // 添加表单提交事件监听
    const form = document.getElementById('mainForm');
    if (form) {
        form.addEventListener('submit', handleFormSubmit);
    }
}

// 当页面加载完成时初始化
document.addEventListener('DOMContentLoaded', initializeApp);