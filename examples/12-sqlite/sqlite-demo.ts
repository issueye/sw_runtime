// SQLite 数据库演示
const sqlite = require('sqlite');

async function sqliteDemo() {
    console.log('=== SQLite 数据库演示 ===');

    try {
        // 1. 获取 SQLite 版本
        console.log('\n1. 获取 SQLite 版本:');
        const version = await sqlite.version();
        console.log('SQLite 版本:', version);

        // 2. 打开内存数据库
        console.log('\n2. 打开内存数据库:');
        const db = await sqlite.open(':memory:');
        console.log('数据库已打开');

        // 3. 创建表
        console.log('\n3. 创建用户表:');
        await db.exec(`
            CREATE TABLE users (
                id INTEGER PRIMARY KEY AUTOINCREMENT,
                name TEXT NOT NULL,
                email TEXT UNIQUE NOT NULL,
                age INTEGER,
                created_at DATETIME DEFAULT CURRENT_TIMESTAMP
            )
        `);
        console.log('用户表创建成功');

        // 4. 插入数据
        console.log('\n4. 插入用户数据:');
        const insertResult1 = await db.run(
            'INSERT INTO users (name, email, age) VALUES (?, ?, ?)',
            ['张三', 'zhangsan@example.com', 25]
        );
        console.log('插入用户1:', insertResult1);

        const insertResult2 = await db.run(
            'INSERT INTO users (name, email, age) VALUES (?, ?, ?)',
            ['李四', 'lisi@example.com', 30]
        );
        console.log('插入用户2:', insertResult2);

        // 5. 查询单个用户
        console.log('\n5. 查询单个用户:');
        const user = await db.get('SELECT * FROM users WHERE name = ?', ['张三']);
        console.log('查询结果:', user);

        // 6. 查询所有用户
        console.log('\n6. 查询所有用户:');
        const allUsers = await db.all('SELECT * FROM users ORDER BY id');
        console.log('所有用户:', allUsers);

        // 7. 使用事务
        console.log('\n7. 使用事务批量插入:');
        await db.transaction(async (tx) => {
            await tx.run('INSERT INTO users (name, email, age) VALUES (?, ?, ?)', 
                ['王五', 'wangwu@example.com', 28]);
            await tx.run('INSERT INTO users (name, email, age) VALUES (?, ?, ?)', 
                ['赵六', 'zhaoliu@example.com', 35]);
        });
        console.log('事务执行完成');

        // 8. 验证事务结果
        console.log('\n8. 验证事务结果:');
        const userCount = await db.get('SELECT COUNT(*) as count FROM users');
        console.log('用户总数:', userCount);

        // 9. 使用预处理语句
        console.log('\n9. 使用预处理语句:');
        const stmt = await db.prepare('SELECT * FROM users WHERE age > ?');
        const olderUsers = await stmt.all(25);
        console.log('年龄大于25的用户:', olderUsers);
        await stmt.close();

        // 10. 获取表信息
        console.log('\n10. 获取数据库表信息:');
        const tables = await db.tables();
        console.log('数据库表:', tables);

        const schema = await db.schema('users');
        console.log('用户表结构:', schema);

        // 11. 更新数据
        console.log('\n11. 更新用户数据:');
        const updateResult = await db.run(
            'UPDATE users SET age = ? WHERE name = ?',
            [26, '张三']
        );
        console.log('更新结果:', updateResult);

        // 12. 删除数据
        console.log('\n12. 删除用户数据:');
        const deleteResult = await db.run('DELETE FROM users WHERE name = ?', ['李四']);
        console.log('删除结果:', deleteResult);

        // 13. 最终查询
        console.log('\n13. 最终用户列表:');
        const finalUsers = await db.all('SELECT * FROM users ORDER BY id');
        console.log('最终用户:', finalUsers);

        // 14. 关闭数据库
        console.log('\n14. 关闭数据库:');
        await db.close();
        console.log('数据库已关闭');

        console.log('\n=== SQLite 演示完成 ===');

    } catch (error) {
        console.error('SQLite 演示出错:', error);
    }
}

// 文件数据库演示
async function fileDatabaseDemo() {
    console.log('\n=== 文件数据库演示 ===');

    try {
        // 打开文件数据库
        const db = await sqlite.open('./test.db');
        console.log('文件数据库已打开');

        // 创建产品表
        await db.exec(`
            CREATE TABLE IF NOT EXISTS products (
                id INTEGER PRIMARY KEY AUTOINCREMENT,
                name TEXT NOT NULL,
                price REAL NOT NULL,
                category TEXT,
                in_stock BOOLEAN DEFAULT 1
            )
        `);

        // 插入产品数据
        await db.run('INSERT OR REPLACE INTO products (name, price, category) VALUES (?, ?, ?)',
            ['笔记本电脑', 5999.99, '电子产品']);
        await db.run('INSERT OR REPLACE INTO products (name, price, category) VALUES (?, ?, ?)',
            ['无线鼠标', 99.99, '电子产品']);

        // 查询产品
        const products = await db.all('SELECT * FROM products');
        console.log('产品列表:', products);

        await db.close();
        console.log('文件数据库已关闭');

    } catch (error) {
        console.error('文件数据库演示出错:', error);
    }
}

// 运行演示
sqliteDemo().then(() => {
    return fileDatabaseDemo();
});