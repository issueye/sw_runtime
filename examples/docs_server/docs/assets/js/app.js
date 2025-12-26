// 简化的语法高亮函数
function highlightCode() {
    const codeBlocks = document.querySelectorAll('.code-example');
    
    codeBlocks.forEach(block => {
        let code = block.textContent || block.innerText;
        
        // 转义HTML
        code = code.replace(/&/g, '&amp;')
                  .replace(/</g, '&lt;')
                  .replace(/>/g, '&gt;');
        
        // 应用语法高亮
        code = applyHighlighting(code);
        
        // 格式化代码
        code = formatCode(code);
        
        block.innerHTML = `<pre><code>${code}</code></pre>`;
    });
}

function applyHighlighting(code) {
    // 高亮注释
    code = code.replace(/(\/\/.*?)$/gm, '<span class="comment">$1</span>');
    code = code.replace(/(\/\*[\s\S]*?\*\/)/g, '<span class="comment">$1</span>');
    
    // 高亮字符串
    code = code.replace(/(['"`])((?:\\.|(?!\1)[^\\])*?)\1/g, '<span class="string">$1$2$3</span>');
    
    // 高亮数字
    code = code.replace(/\b(\d+(?:\.\d+)?)\b/g, '<span class="number">$1</span>');
    
    // 高亮关键字
    const keywords = [
        'const', 'let', 'var', 'function', 'async', 'await', 'return',
        'if', 'else', 'for', 'while', 'do', 'switch', 'case', 'break',
        'continue', 'try', 'catch', 'finally', 'throw', 'new', 'class',
        'extends', 'import', 'export', 'from', 'default', 'typeof',
        'instanceof', 'this', 'super', 'in', 'of'
    ];
    
    keywords.forEach(keyword => {
        const regex = new RegExp(`\\b(${keyword})\\b`, 'g');
        code = code.replace(regex, '<span class="keyword">$1</span>');
    });
    
    // 高亮布尔值和特殊值
    code = code.replace(/\b(true|false|null|undefined)\b/g, '<span class="boolean">$1</span>');
    
    // 高亮方法调用
    code = code.replace(/\.(\w+)(\s*\()/g, '.<span class="method">$1</span>$2');
    
    // 高亮函数调用
    code = code.replace(/\b(\w+)(\s*\()/g, '<span class="function">$1</span>$2');
    
    // 高亮属性访问
    code = code.replace(/\.(\w+)(?!\s*[\(<])/g, '.<span class="property">$1</span>');
    
    return code;
}

function formatCode(code) {
    const lines = code.split('\n');
    let indentLevel = 0;
    const indentSize = 2;
    
    return lines.map(line => {
        const trimmed = line.trim();
        if (!trimmed) return '';
        
        // 减少缩进
        if (trimmed.match(/^[\}\]\)]/)) {
            indentLevel = Math.max(0, indentLevel - 1);
        }
        
        const indented = ' '.repeat(indentLevel * indentSize) + trimmed;
        
        // 增加缩进
        if (trimmed.match(/[\{\[\(]\s*$/) || 
            trimmed.match(/=>\s*$/) ||
            trimmed.match(/:\s*$/) && !trimmed.match(/case\s+.*:/)) {
            indentLevel++;
        }
        
        return indented;
    }).join('\n');
}

// 模块加载函数
async function loadModule(moduleName) {
    try {
        // 检查是否是本地文件协议
        if (window.location.protocol === 'file:') {
            // 对于本地文件，显示提示信息
            return `<div class="error">
                <h3>本地文件访问限制</h3>
                <p>由于浏览器安全限制，无法在 file:// 协议下动态加载模块。</p>
                <p>请使用以下方式之一访问文档：</p>
                <ul style="margin: 10px 0; padding-left: 20px;">
                    <li>使用本地 HTTP 服务器（如 Python: <code>python -m http.server</code>）</li>
                    <li>使用 Live Server 扩展</li>
                    <li>部署到 Web 服务器</li>
                </ul>
                <p>或者直接查看单独的模块文件：<code>modules/${moduleName}.html</code></p>
            </div>`;
        }
        
        const response = await fetch(`modules/${moduleName}.html`);
        if (!response.ok) {
            throw new Error(`Failed to load module: ${moduleName} (${response.status})`);
        }
        return await response.text();
    } catch (error) {
        console.error('Error loading module:', error);
        return `<div class="error">
            <h3>模块加载失败</h3>
            <p>无法加载模块: ${moduleName}</p>
            <p>错误信息: ${error.message}</p>
            <p>请检查文件路径或使用 HTTP 服务器访问文档。</p>
        </div>`;
    }
}

// 显示模块内容
async function showModule(moduleName) {
    const contentBody = document.querySelector('.content-body');
    const loadingHtml = '<div class="loading">Loading...</div>';
    
    contentBody.innerHTML = loadingHtml;
    
    try {
        const moduleContent = await loadModule(moduleName);
        contentBody.innerHTML = moduleContent;
        
        // 重新应用语法高亮
        highlightCode();
    } catch (error) {
        contentBody.innerHTML = `<div class="error">Error loading module: ${error.message}</div>`;
    }
}

// 导航功能
document.addEventListener('DOMContentLoaded', function() {
    // 应用语法高亮
    highlightCode();
    
    const navLinks = document.querySelectorAll('.nav-link');
    const sections = document.querySelectorAll('.content-section');

    navLinks.forEach(link => {
        link.addEventListener('click', async function(e) {
            e.preventDefault();
            
            const targetId = this.getAttribute('href').substring(1);
            
            // 更新导航状态
            navLinks.forEach(l => l.classList.remove('active'));
            this.classList.add('active');
            
            // 如果是静态内容，显示对应section
            const targetSection = document.getElementById(targetId);
            if (targetSection) {
                sections.forEach(section => {
                    section.classList.remove('active');
                });
                targetSection.classList.add('active');
                highlightCode();
            } else {
                // 动态加载模块
                await showModule(targetId);
            }
        });
    });
});