// Ticker 定时器示例
const time = require('time');

console.log('=== Ticker Demo ===\n');

// 1. 使用 setInterval（类似 JavaScript）
console.log('--- 测试 1: setInterval 基础使用 ---');
let count = 0;

const intervalId = time.setInterval(() => {
    count++;
    console.log(`[setInterval] 执行次数: ${count}, 时间: ${time.now()}`);
    
    if (count >= 5) {
        console.log('[setInterval] 达到 5 次，停止定时器');
        time.clearInterval(intervalId);
    }
}, 1000); // 每 1000 毫秒执行一次

console.log(`定时器 ID: ${intervalId}\n`);

// 2. 使用 createTicker 创建 Ticker 对象
console.log('\n--- 测试 2: createTicker 高级用法 ---');

const ticker = time.createTicker(500); // 500 毫秒间隔
let tickCount = 0;

ticker.tick(() => {
    tickCount++;
    console.log(`[Ticker] Tick ${tickCount} - ${time.format(time.nowUnix(), time.FORMAT.Time)}`);
    
    if (tickCount >= 8) {
        console.log('[Ticker] 达到 8 次，停止 ticker');
        ticker.stop();
    }
});

// 3. 倒计时示例
console.log('\n--- 测试 3: 倒计时 ---');

let countdown = 10;
const countdownTimer = time.setInterval(() => {
    console.log(`倒计时: ${countdown} 秒`);
    countdown--;
    
    if (countdown < 0) {
        console.log('⏰ 倒计时结束！');
        time.clearInterval(countdownTimer);
    }
}, 1000);

// 4. 多个定时器同时运行
console.log('\n--- 测试 4: 多个定时器 ---');

const timer1 = time.setInterval(() => {
    console.log('🟢 快速定时器 (200ms)');
}, 200);

const timer2 = time.setInterval(() => {
    console.log('🔵 中速定时器 (500ms)');
}, 500);

const timer3 = time.setInterval(() => {
    console.log('🔴 慢速定时器 (1000ms)');
}, 1000);

// 5 秒后停止所有定时器
time.sleep(5).then(() => {
    console.log('\n--- 停止所有定时器 ---');
    time.clearInterval(timer1);
    time.clearInterval(timer2);
    time.clearInterval(timer3);
    console.log('所有定时器已停止');
});

// 5. 使用 Ticker 的 reset 功能
console.log('\n--- 测试 5: Ticker Reset ---');

const resetTicker = time.createTicker(1000);
let resetCount = 0;

resetTicker.tick(() => {
    resetCount++;
    console.log(`[Reset Ticker] Count: ${resetCount}`);
    
    if (resetCount === 3) {
        console.log('  → 加速！重置为 300ms');
        resetTicker.reset(300);
    }
    
    if (resetCount >= 10) {
        resetTicker.stop();
    }
});

// 6. 性能监控示例
console.log('\n--- 测试 6: 性能监控 ---');

let operationCount = 0;
const startTime = time.nowUnixMilli();

const monitorTimer = time.setInterval(() => {
    operationCount++;
    const elapsed = time.nowUnixMilli() - startTime;
    const rate = (operationCount / elapsed * 1000).toFixed(2);
    
    console.log(`操作次数: ${operationCount}, 耗时: ${elapsed}ms, 速率: ${rate} ops/s`);
    
    if (operationCount >= 20) {
        time.clearInterval(monitorTimer);
        console.log('性能监控结束');
    }
}, 100);

// 7. 周期性任务调度
console.log('\n--- 测试 7: 任务调度器 ---');

class TaskScheduler {
    constructor() {
        this.tasks = [];
        this.timerId = null;
    }
    
    addTask(name, interval, callback) {
        const task = {
            name: name,
            interval: interval,
            callback: callback,
            lastRun: 0,
            runCount: 0
        };
        this.tasks.push(task);
        console.log(`✓ 添加任务: ${name} (间隔: ${interval}ms)`);
    }
    
    start() {
        console.log('启动调度器...');
        this.timerId = time.setInterval(() => {
            const now = time.nowUnixMilli();
            
            this.tasks.forEach(task => {
                if (now - task.lastRun >= task.interval) {
                    task.lastRun = now;
                    task.runCount++;
                    console.log(`  [${task.name}] 执行第 ${task.runCount} 次`);
                    task.callback();
                }
            });
        }, 50); // 50ms 检查间隔
    }
    
    stop() {
        if (this.timerId) {
            time.clearInterval(this.timerId);
            console.log('调度器已停止');
        }
    }
}

const scheduler = new TaskScheduler();

scheduler.addTask('数据同步', 1000, () => {
    // 模拟数据同步
});

scheduler.addTask('健康检查', 2000, () => {
    // 模拟健康检查
});

scheduler.addTask('日志清理', 3000, () => {
    // 模拟日志清理
});

scheduler.start();

// 10 秒后停止调度器
time.sleep(10).then(() => {
    scheduler.stop();
});

// 8. 时间同步示例
console.log('\n--- 测试 8: 时间同步 ---');

let syncCount = 0;
const syncTimer = time.setInterval(() => {
    syncCount++;
    const currentTime = time.now();
    console.log(`[时间同步] ${syncCount} - ${currentTime}`);
    
    if (syncCount >= 5) {
        time.clearInterval(syncTimer);
    }
}, 2000);

// 9. 心跳检测
console.log('\n--- 测试 9: 心跳检测 ---');

class HeartbeatMonitor {
    constructor(interval) {
        this.interval = interval;
        this.ticker = null;
        this.beatCount = 0;
        this.lastBeat = time.nowUnix();
    }
    
    start() {
        console.log(`心跳监控启动 (间隔: ${this.interval}ms)`);
        this.ticker = time.createTicker(this.interval);
        
        this.ticker.tick(() => {
            this.beatCount++;
            const now = time.nowUnix();
            const timeSinceLastBeat = now - this.lastBeat;
            this.lastBeat = now;
            
            console.log(`💓 心跳 #${this.beatCount} - 间隔: ${timeSinceLastBeat}s`);
            
            if (this.beatCount >= 6) {
                this.stop();
            }
        });
    }
    
    stop() {
        if (this.ticker) {
            this.ticker.stop();
            console.log('❌ 心跳监控已停止');
        }
    }
}

const heartbeat = new HeartbeatMonitor(800);
heartbeat.start();

console.log('\n✨ 所有定时器示例已启动！');
console.log('提示: 程序将运行一段时间以展示各种定时器功能\n');
