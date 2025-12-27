// 时间模块高级示例 - 实际应用场景
const time = require('time');

console.log('=== Time Advanced Demo ===\n');

// 1. 定时任务调度
console.log('--- 定时任务调度 ---');

function scheduleTask(taskName, delaySeconds) {
    console.log(`[${time.now()}] 调度任务: ${taskName}, 延迟 ${delaySeconds} 秒`);
    
    return time.sleep(delaySeconds).then(() => {
        console.log(`[${time.now()}] 执行任务: ${taskName}`);
        return { task: taskName, completedAt: time.now() };
    });
}

scheduleTask('数据备份', 1).then(result => {
    console.log('任务完成:', result);
});

// 2. 性能计时器
console.log('\n--- 性能计时器 ---');

class PerformanceTimer {
    constructor() {
        this.startTime = null;
        this.endTime = null;
    }
    
    start() {
        this.startTime = time.nowUnixMilli();
        console.log('计时开始:', time.format(Math.floor(this.startTime / 1000), time.FORMAT.DateTime));
    }
    
    stop() {
        this.endTime = time.nowUnixMilli();
        const duration = this.endTime - this.startTime;
        console.log('计时结束:', time.format(Math.floor(this.endTime / 1000), time.FORMAT.DateTime));
        console.log('耗时:', duration, '毫秒');
        return duration;
    }
    
    reset() {
        this.startTime = null;
        this.endTime = null;
    }
}

const timer = new PerformanceTimer();
timer.start();

// 模拟一些操作
time.sleepMillis(100).then(() => {
    timer.stop();
});

// 3. 日期范围计算
console.log('\n--- 日期范围计算 ---');

function getDateRange(startDate, endDate) {
    const start = time.parse(startDate);
    const end = time.parse(endDate);
    
    const diff = time.diff(end.unix, start.unix);
    
    return {
        startDate: startDate,
        endDate: endDate,
        days: diff.days,
        hours: diff.hours,
        minutes: diff.minutes,
        seconds: diff.seconds
    };
}

const range = getDateRange('2024-12-01T00:00:00Z', '2024-12-31T23:59:59Z');
console.log('日期范围:', range);
console.log(`从 ${range.startDate} 到 ${range.endDate} 共 ${range.days} 天`);

// 4. 工作日计算（跳过周末）
console.log('\n--- 工作日计算 ---');

function addWorkdays(timestamp, days) {
    let current = timestamp;
    let addedDays = 0;
    
    while (addedDays < days) {
        current = time.addDays(current, 1);
        const weekday = time.getWeekday(current);
        
        // 跳过周六(6)和周日(0)
        if (weekday.number !== 0 && weekday.number !== 6) {
            addedDays++;
        }
    }
    
    return current;
}

const today = time.nowUnix();
const after5Workdays = addWorkdays(today, 5);
console.log('今天:', time.format(today, time.FORMAT.Date));
console.log('5个工作日后:', time.format(after5Workdays, time.FORMAT.Date));

// 5. 生日倒计时
console.log('\n--- 生日倒计时 ---');

function birthdayCountdown(birthdayMonth, birthdayDay) {
    const now = time.nowUnix();
    const currentYear = time.getYear(now);
    
    // 创建今年的生日
    let birthday = time.create(currentYear, birthdayMonth, birthdayDay, 0, 0, 0);
    
    // 如果今年的生日已过，计算明年的
    if (birthday < now) {
        birthday = time.create(currentYear + 1, birthdayMonth, birthdayDay, 0, 0, 0);
    }
    
    const diff = time.diff(birthday, now);
    
    return {
        date: time.format(birthday, time.FORMAT.Date),
        daysLeft: diff.days,
        hoursLeft: diff.hours % 24,
        minutesLeft: diff.minutes % 60
    };
}

const countdown = birthdayCountdown(12, 25); // 圣诞节
console.log('距离生日还有:');
console.log(`  日期: ${countdown.date}`);
console.log(`  剩余: ${countdown.daysLeft} 天 ${countdown.hoursLeft} 小时 ${countdown.minutesLeft} 分钟`);

// 6. 时间格式化工具
console.log('\n--- 时间格式化工具 ---');

class TimeFormatter {
    static toHumanReadable(timestamp) {
        const now = time.nowUnix();
        const diff = now - timestamp;
        
        if (diff < 60) {
            return '刚刚';
        } else if (diff < 3600) {
            const minutes = Math.floor(diff / 60);
            return `${minutes} 分钟前`;
        } else if (diff < 86400) {
            const hours = Math.floor(diff / 3600);
            return `${hours} 小时前`;
        } else if (diff < 604800) {
            const days = Math.floor(diff / 86400);
            return `${days} 天前`;
        } else {
            return time.format(timestamp, time.FORMAT.Date);
        }
    }
    
    static toDuration(seconds) {
        const days = Math.floor(seconds / 86400);
        const hours = Math.floor((seconds % 86400) / 3600);
        const minutes = Math.floor((seconds % 3600) / 60);
        const secs = seconds % 60;
        
        const parts = [];
        if (days > 0) parts.push(`${days}天`);
        if (hours > 0) parts.push(`${hours}小时`);
        if (minutes > 0) parts.push(`${minutes}分钟`);
        if (secs > 0) parts.push(`${secs}秒`);
        
        return parts.join(' ');
    }
}

const recentTime = time.addSeconds(time.nowUnix(), -300); // 5分钟前
console.log('相对时间:', TimeFormatter.toHumanReadable(recentTime));
console.log('时长格式化:', TimeFormatter.toDuration(3665)); // 1小时1分5秒

// 7. 时间戳验证
console.log('\n--- 时间戳验证 ---');

function isValidTimestamp(timestamp) {
    const year = time.getYear(timestamp);
    return year >= 1970 && year <= 2100;
}

function isExpired(timestamp, expirationSeconds) {
    const now = time.nowUnix();
    return now - timestamp > expirationSeconds;
}

const testTimestamp = time.nowUnix();
console.log('时间戳有效?', isValidTimestamp(testTimestamp));

const oldTimestamp = time.addSeconds(time.nowUnix(), -7200); // 2小时前
console.log('2小时前的时间戳已过期(1小时)?', isExpired(oldTimestamp, 3600));

// 8. 批量时间处理
console.log('\n--- 批量时间处理 ---');

function processBatchTimestamps(timestamps) {
    return timestamps.map(ts => ({
        original: ts,
        formatted: time.format(ts, time.FORMAT.DateTime),
        year: time.getYear(ts),
        month: time.getMonth(ts),
        day: time.getDay(ts),
        weekday: time.getWeekday(ts).name
    }));
}

const now = time.nowUnix();
const timestamps = [
    now,
    time.addDays(now, 1),
    time.addDays(now, 7),
    time.addDays(now, 30)
];

const processed = processBatchTimestamps(timestamps);
console.log('批量处理结果:');
processed.forEach((item, index) => {
    console.log(`  [${index}] ${item.formatted} (${item.weekday})`);
});

// 9. 时区转换示例
console.log('\n--- 时区转换 ---');

const currentTime = time.nowUnix();
console.log('本地时间:', time.format(currentTime, time.FORMAT.DateTime));
console.log('UTC 时间:', time.utc(currentTime));

// 10. 延迟执行链
console.log('\n--- 延迟执行链 ---');

console.log('开始执行延迟链...');
time.sleepMillis(500)
    .then(() => {
        console.log('步骤 1 完成 (500ms)');
        return time.sleepMillis(500);
    })
    .then(() => {
        console.log('步骤 2 完成 (500ms)');
        return time.sleepMillis(500);
    })
    .then(() => {
        console.log('步骤 3 完成 (500ms)');
        console.log('所有步骤完成！');
    });

console.log('\n✨ 高级示例演示完成！');
