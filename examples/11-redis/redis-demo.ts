// redis-demo.ts - Redis 客户端功能演示

const { redis } = require('db');

console.log('=== Redis 客户端功能演示 ===');

// 注意：这个演示需要本地运行 Redis 服务器
// 如果没有 Redis 服务器，这些操作会失败

try {
    // 1. 创建 Redis 客户端连接
    console.log('\n1. 创建 Redis 连接:');
    const client = redis.createClient({
        host: 'localhost',
        port: 6379,
        db: 0
    });

    console.log('Redis 客户端创建成功');

    // 2. 基本字符串操作
    console.log('\n2. 字符串操作:');
    
    // 设置值
    client.set('test:string', 'Hello Redis!')
        .then((result: string) => {
            console.log('SET 结果:', result);
            return client.get('test:string');
        })
        .then((value: string) => {
            console.log('GET 结果:', value);
            return client.exists('test:string');
        })
        .then((exists: boolean) => {
            console.log('键是否存在:', exists);
            return client.expire('test:string', 60);
        })
        .then((success: boolean) => {
            console.log('设置过期时间:', success);
            return client.ttl('test:string');
        })
        .then((ttl: number) => {
            console.log('剩余生存时间:', ttl, '秒');
        })
        .catch((error: Error) => {
            console.error('字符串操作错误:', error.message);
        });

    // 3. JSON 数据操作
    console.log('\n3. JSON 数据操作:');
    setTimeout(() => {
        const userData = {
            id: 1,
            name: 'John Doe',
            email: 'john@example.com',
            age: 30,
            active: true
        };

        client.setJSON('user:1', userData, 300) // 5分钟过期
            .then((result: string) => {
                console.log('JSON SET 结果:', result);
                return client.getJSON('user:1');
            })
            .then((data: any) => {
                console.log('JSON GET 结果:', data);
                console.log('用户名:', data.name);
                console.log('邮箱:', data.email);
            })
            .catch((error: Error) => {
                console.error('JSON 操作错误:', error.message);
            });
    }, 1000);

    // 4. 哈希操作
    console.log('\n4. 哈希操作:');
    setTimeout(() => {
        client.hset('user:profile:1', 'name', 'Alice')
            .then((result: number) => {
                console.log('HSET 结果:', result);
                return client.hset('user:profile:1', 'age', '25');
            })
            .then((result: number) => {
                return client.hget('user:profile:1', 'name');
            })
            .then((name: string) => {
                console.log('用户名:', name);
                return client.hgetall('user:profile:1');
            })
            .then((profile: any) => {
                console.log('完整用户资料:', profile);
            })
            .catch((error: Error) => {
                console.error('哈希操作错误:', error.message);
            });
    }, 2000);

    // 5. 列表操作
    console.log('\n5. 列表操作:');
    setTimeout(() => {
        client.lpush('tasks', 'task1', 'task2', 'task3')
            .then((length: number) => {
                console.log('列表长度:', length);
                return client.rpush('tasks', 'task4', 'task5');
            })
            .then((length: number) => {
                console.log('添加后列表长度:', length);
                return client.lrange('tasks', 0, -1);
            })
            .then((tasks: string[]) => {
                console.log('所有任务:', tasks);
                return client.lpop('tasks');
            })
            .then((task: string) => {
                console.log('弹出的任务:', task);
                return client.llen('tasks');
            })
            .then((length: number) => {
                console.log('剩余任务数:', length);
            })
            .catch((error: Error) => {
                console.error('列表操作错误:', error.message);
            });
    }, 3000);

    // 6. 集合操作
    console.log('\n6. 集合操作:');
    setTimeout(() => {
        client.sadd('tags', 'javascript', 'typescript', 'redis', 'nodejs')
            .then((added: number) => {
                console.log('添加的标签数:', added);
                return client.sismember('tags', 'javascript');
            })
            .then((isMember: boolean) => {
                console.log('javascript 是否在集合中:', isMember);
                return client.smembers('tags');
            })
            .then((members: string[]) => {
                console.log('所有标签:', members);
                return client.srem('tags', 'nodejs');
            })
            .then((removed: number) => {
                console.log('移除的标签数:', removed);
                return client.smembers('tags');
            })
            .then((members: string[]) => {
                console.log('剩余标签:', members);
            })
            .catch((error: Error) => {
                console.error('集合操作错误:', error.message);
            });
    }, 4000);

    // 7. 有序集合操作
    console.log('\n7. 有序集合操作:');
    setTimeout(() => {
        client.zadd('scores', 100, 'alice')
            .then((added: number) => {
                return client.zadd('scores', 85, 'bob');
            })
            .then((added: number) => {
                return client.zadd('scores', 92, 'charlie');
            })
            .then((added: number) => {
                return client.zrange('scores', 0, -1);
            })
            .then((members: string[]) => {
                console.log('按分数排序的成员:', members);
                return client.zscore('scores', 'alice');
            })
            .then((score: number) => {
                console.log('Alice 的分数:', score);
                return client.zrank('scores', 'alice');
            })
            .then((rank: number) => {
                console.log('Alice 的排名:', rank);
            })
            .catch((error: Error) => {
                console.error('有序集合操作错误:', error.message);
            });
    }, 5000);

    // 8. 通用操作
    console.log('\n8. 通用操作:');
    setTimeout(() => {
        client.ping()
            .then((result: string) => {
                console.log('PING 结果:', result);
                return client.keys('test:*');
            })
            .then((keys: string[]) => {
                console.log('匹配 test:* 的键:', keys);
            })
            .catch((error: Error) => {
                console.error('通用操作错误:', error.message);
            });
    }, 6000);

    // 9. 清理和关闭
    setTimeout(() => {
        console.log('\n9. 清理操作:');
        client.del('test:string', 'user:1', 'user:profile:1', 'tasks', 'tags', 'scores')
            .then((deleted: number) => {
                console.log('删除的键数量:', deleted);
                console.log('演示完成，连接保持打开状态');
            })
            .catch((error: Error) => {
                console.error('清理操作错误:', error.message);
            });
    }, 7000);

} catch (error) {
    console.error('Redis 连接失败:', (error as Error).message);
    console.log('请确保 Redis 服务器正在运行在 localhost:6379');
    console.log('启动 Redis 服务器的命令: redis-server');
}

export { redis };

// 如果在 CommonJS 环境中运行
if (typeof module !== 'undefined' && module.exports) {
    module.exports = { redis };
} else {
    // 在全局作用域中，不需要导出
}