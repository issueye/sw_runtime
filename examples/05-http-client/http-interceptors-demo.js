// HTTP 拦截器示例 - 请求和响应拦截
const http = require('http');

console.log('=== HTTP Interceptors Demo ===\n');

// 1. 全局请求拦截器 - 自动添加认证 token
http.setRequestInterceptor((config) => {
    console.log('[请求拦截器] 添加 Authorization header');
    
    // 修改请求配置
    if (!config.headers) {
        config.headers = {};
    }
    config.headers['Authorization'] = 'Bearer my-secret-token';
    config.headers['X-Custom-Header'] = 'CustomValue';
    
    console.log('  URL:', config.url);
    console.log('  Headers:', config.headers);
    
    return config;
});

// 2. 全局响应拦截器 - 统一处理响应数据
http.setResponseInterceptor((response) => {
    console.log('[响应拦截器] 处理响应数据');
    console.log('  Status:', response.status);
    
    // 可以在这里统一处理响应，例如提取嵌套数据
    if (response.data && response.data.data) {
        console.log('  提取嵌套的 data 字段');
        response.data = response.data.data;
    }
    
    return response;
});

// 3. 测试全局拦截器
console.log('\n--- 测试 1: 使用全局拦截器 ---');
http.get('https://httpbin.org/headers')
    .then(response => {
        console.log('✅ 请求成功');
        console.log('Headers sent:', response.data.headers);
    })
    .catch(err => {
        console.error('❌ 请求失败:', err.message);
    });

// 4. 单个请求的拦截器
console.log('\n--- 测试 2: 单个请求拦截器 ---');
http.post('https://httpbin.org/post', {
    data: {
        username: 'john',
        password: 'secret123'
    },
    // beforeRequest: 在发送前修改请求
    beforeRequest: (config) => {
        console.log('[beforeRequest] 修改请求数据');
        
        // 修改请求参数
        config.params = { timestamp: Date.now() };
        
        // 修改请求头
        config.headers['X-Request-ID'] = Math.random().toString(36).substr(2, 9);
        
        console.log('  添加参数:', config.params);
        console.log('  添加请求头:', config.headers['X-Request-ID']);
        
        return config;
    },
    // afterResponse: 在接收后处理响应
    afterResponse: (response) => {
        console.log('[afterResponse] 处理响应');
        console.log('  原始状态:', response.status);
        
        // 可以修改响应数据
        if (response.data) {
            response.data.processed = true;
            response.data.processedAt = new Date().toISOString();
        }
        
        return response;
    }
}).then(response => {
    console.log('✅ 请求成功');
    console.log('Data:', response.data);
}).catch(err => {
    console.error('❌ 请求失败:', err.message);
});

// 5. transformRequest - 转换请求数据
console.log('\n--- 测试 3: transformRequest 转换请求数据 ---');
http.post('https://httpbin.org/post', {
    data: {
        items: [1, 2, 3, 4, 5]
    },
    transformRequest: (data) => {
        console.log('[transformRequest] 转换请求数据');
        console.log('  原始数据:', data);
        
        // 修改数据 - 计算总和
        const transformed = {
            ...data,
            sum: data.items.reduce((a, b) => a + b, 0),
            count: data.items.length
        };
        
        console.log('  转换后:', transformed);
        return transformed;
    }
}).then(response => {
    console.log('✅ 请求成功');
    console.log('发送的数据:', response.data.json);
}).catch(err => {
    console.error('❌ 请求失败:', err.message);
});

// 6. transformResponse - 转换响应数据
console.log('\n--- 测试 4: transformResponse 转换响应数据 ---');
http.get('https://httpbin.org/json', {
    transformResponse: (data) => {
        console.log('[transformResponse] 转换响应数据');
        
        // 提取特定字段
        if (data && data.slideshow) {
            console.log('  提取 slideshow 数据');
            return {
                title: data.slideshow.title,
                author: data.slideshow.author,
                slideCount: data.slideshow.slides ? data.slideshow.slides.length : 0
            };
        }
        
        return data;
    }
}).then(response => {
    console.log('✅ 请求成功');
    console.log('转换后的数据:', response.data);
}).catch(err => {
    console.error('❌ 请求失败:', err.message);
});

// 7. 组合使用多个拦截器
console.log('\n--- 测试 5: 组合使用多个拦截器 ---');
http.put('https://httpbin.org/put', {
    data: { message: 'Hello' },
    beforeRequest: (config) => {
        console.log('[beforeRequest] 步骤 1: 添加时间戳');
        config.headers['X-Timestamp'] = Date.now();
        return config;
    },
    transformRequest: (data) => {
        console.log('[transformRequest] 步骤 2: 转换数据格式');
        return {
            ...data,
            uppercase: data.message.toUpperCase()
        };
    },
    transformResponse: (data) => {
        console.log('[transformResponse] 步骤 3: 提取关键数据');
        return {
            originalJson: data.json,
            headers: data.headers
        };
    },
    afterResponse: (response) => {
        console.log('[afterResponse] 步骤 4: 添加元数据');
        response.data.meta = {
            processed: true,
            responseTime: new Date().toISOString()
        };
        return response;
    }
}).then(response => {
    console.log('✅ 请求成功');
    console.log('最终数据:', JSON.stringify(response.data, null, 2));
}).catch(err => {
    console.error('❌ 请求失败:', err.message);
});

// 8. 修改请求参数示例
console.log('\n--- 测试 6: 动态修改请求参数 ---');
http.get('https://httpbin.org/get', {
    params: {
        page: 1,
        size: 10
    },
    beforeRequest: (config) => {
        console.log('[beforeRequest] 动态修改参数');
        console.log('  原始参数:', config.params);
        
        // 添加额外参数
        config.params.sort = 'created_at';
        config.params.order = 'desc';
        config.params.filter = 'active';
        
        console.log('  修改后参数:', config.params);
        return config;
    }
}).then(response => {
    console.log('✅ 请求成功');
    console.log('请求 URL:', response.url);
    console.log('查询参数:', response.data.args);
}).catch(err => {
    console.error('❌ 请求失败:', err.message);
});

// 9. 修改响应头示例
console.log('\n--- 测试 7: 处理响应头 ---');
http.get('https://httpbin.org/response-headers?X-Custom=Value', {
    afterResponse: (response) => {
        console.log('[afterResponse] 处理响应头');
        console.log('  原始响应头:', response.headers);
        
        // 可以基于响应头做一些处理
        if (response.headers['Content-Type']) {
            console.log('  Content-Type:', response.headers['Content-Type']);
        }
        
        // 添加自定义标记
        response.headers['X-Processed'] = 'true';
        
        return response;
    }
}).then(response => {
    console.log('✅ 请求成功');
    console.log('响应头包含 X-Processed:', response.headers['X-Processed']);
}).catch(err => {
    console.error('❌ 请求失败:', err.message);
});

console.log('\n✨ 所有测试已启动，等待响应...\n');
