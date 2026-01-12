// http-demo.ts - HTTP 客户端功能演示

const http = require('http/client');

console.log('=== HTTP 客户端功能演示 ===');

// 1. 基本 GET 请求
console.log('\n1. 基本 GET 请求:');
http.get('https://jsonplaceholder.typicode.com/posts/1')
    .then((response: any) => {
        console.log('状态码:', response.status);
        console.log('状态文本:', response.statusText);
        console.log('响应数据:', response.data);
        console.log('Content-Type:', response.headers['Content-Type']);
    })
    .catch((error: Error) => {
        console.error('GET 请求失败:', error.message);
    });

// 2. 带参数的 GET 请求
console.log('\n2. 带参数的 GET 请求:');
setTimeout(() => {
    http.get('https://jsonplaceholder.typicode.com/posts', {
        params: {
            userId: '1'
        }
    })
        .then((response: any) => {
            console.log('用户1的文章数量:', response.data.length);
            console.log('第一篇文章标题:', response.data[0]?.title);
        })
        .catch((error: Error) => {
            console.error('带参数 GET 请求失败:', error.message);
        });
}, 1000);

// 3. POST 请求
console.log('\n3. POST 请求:');
setTimeout(() => {
    http.post('https://jsonplaceholder.typicode.com/posts', {
        data: {
            title: 'SW Runtime HTTP Test',
            body: 'This is a test post from SW Runtime',
            userId: 1
        },
        headers: {
            'Content-Type': 'application/json'
        }
    })
        .then((response: any) => {
            console.log('POST 响应状态:', response.status);
            console.log('创建的文章ID:', response.data.id);
            console.log('文章标题:', response.data.title);
        })
        .catch((error: Error) => {
            console.error('POST 请求失败:', error.message);
        });
}, 2000);

// 4. PUT 请求
console.log('\n4. PUT 请求:');
setTimeout(() => {
    http.put('https://jsonplaceholder.typicode.com/posts/1', {
        data: {
            id: 1,
            title: 'Updated Title',
            body: 'Updated body content',
            userId: 1
        }
    })
        .then((response: any) => {
            console.log('PUT 响应状态:', response.status);
            console.log('更新后的标题:', response.data.title);
        })
        .catch((error: Error) => {
            console.error('PUT 请求失败:', error.message);
        });
}, 3000);

// 5. DELETE 请求
console.log('\n5. DELETE 请求:');
setTimeout(() => {
    http.delete('https://jsonplaceholder.typicode.com/posts/1')
        .then((response: any) => {
            console.log('DELETE 响应状态:', response.status);
            console.log('删除成功');
        })
        .catch((error: Error) => {
            console.error('DELETE 请求失败:', error.message);
        });
}, 4000);

// 6. 自定义客户端
console.log('\n6. 自定义客户端:');
setTimeout(() => {
    const customClient = http.createClient({
        timeout: 10 // 10秒超时
    });

    customClient.get('https://jsonplaceholder.typicode.com/users')
        .then((response: any) => {
            console.log('用户列表长度:', response.data.length);
            console.log('第一个用户:', response.data[0]?.name);
        })
        .catch((error: Error) => {
            console.error('自定义客户端请求失败:', error.message);
        });
}, 5000);

// 7. 错误处理演示
console.log('\n7. 错误处理演示:');
setTimeout(() => {
    http.get('https://nonexistent-domain-12345.com/api')
        .then((response: any) => {
            console.log('不应该到达这里');
        })
        .catch((error: Error) => {
            console.log('预期的错误:', error.message);
        });
}, 6000);

// 8. 状态码常量使用
console.log('\n8. 状态码常量:');
console.log('HTTP 200 OK:', http.STATUS_CODES.OK);
console.log('HTTP 404 Not Found:', http.STATUS_CODES.NOT_FOUND);
console.log('HTTP 500 Internal Server Error:', http.STATUS_CODES.INTERNAL_SERVER_ERROR);

export { http };

// 如果在 CommonJS 环境中运行
if (typeof module !== 'undefined' && module.exports) {
    module.exports = { http };
} else {
    // 在全局作用域中，不需要导出
}