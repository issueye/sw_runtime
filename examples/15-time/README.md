# 时间模块示例

本目录包含 time 时间处理模块的功能演示。

## 文件说明

- **time-demo.js** - 时间模块基础演示
- **time-advanced-demo.js** - 时间模块高级应用示例
- **ticker-demo.js** - Ticker 定时器功能演示

## 功能特点

### 时间获取
- 当前时间（ISO 8601 格式）
- Unix 时间戳（秒、毫秒、纳秒）
- 时区转换（UTC、本地时区、自定义时区）

### 时间解析和格式化
- 解析时间字符串
- 多种格式支持（RFC3339、RFC1123、自定义格式）
- 时间戳格式化
- 人性化时间显示（刚刚、几分钟前）

### 时间计算
- 添加/减少时间（天、小时、分钟、秒）
- 时间差计算
- 日期范围计算
- 工作日计算

### 时间比较
- 判断时间先后
- 时间戳验证
- 过期时间检查

### 延迟执行
- 秒级延迟（sleep）
- 毫秒级延迟（sleepMillis）
- Promise 支持

### Ticker 定时器
- setInterval - 周期性执行（类似 JavaScript）
- clearInterval - 清除定时器
- createTicker - 创建 Ticker 对象
- 支持动态重置间隔
- 多定时器并发支持

### 时间组件
- 获取年、月、日
- 获取时、分、秒
- 获取星期几
- 创建自定义时间

## 运行示例

```bash
# 基础示例
sw_runtime run examples/15-time/time-demo.js

# 高级示例
sw_runtime run examples/15-time/time-advanced-demo.js

# Ticker 定时器示例
sw_runtime run examples/15-time/ticker-demo.js
```

## 示例代码

### 获取当前时间
```javascript
const time = require('time');

// 不同格式的当前时间
console.log(time.now());           // ISO 8601 格式
console.log(time.nowUnix());       // Unix 时间戳（秒）
console.log(time.nowUnixMilli());  // Unix 时间戳（毫秒）
```

### 时间格式化
```javascript
const timestamp = time.nowUnix();

// 使用预定义格式
console.log(time.format(timestamp, time.FORMAT.DateTime));  // 2024-12-27 15:30:00
console.log(time.format(timestamp, time.FORMAT.Date));      // 2024-12-27
console.log(time.format(timestamp, time.FORMAT.RFC3339));   // 2024-12-27T15:30:00Z
```

### 时间解析
```javascript
// 解析 ISO 8601 格式
const parsed = time.parse('2024-12-27T15:30:00Z');
console.log(parsed.year);    // 2024
console.log(parsed.month);   // 12
console.log(parsed.day);     // 27

// 解析自定义格式
const custom = time.parse('2024-12-27 15:30:00', time.FORMAT.DateTime);
```

### 时间计算
```javascript
const now = time.nowUnix();

// 添加时间
const tomorrow = time.addDays(now, 1);
const nextHour = time.addHours(now, 1);
const in30Min = time.addMinutes(now, 30);
const in60Sec = time.addSeconds(now, 60);

// 计算时间差
const diff = time.diff(tomorrow, now);
console.log(diff.days);     // 1
console.log(diff.hours);    // 24
console.log(diff.minutes);  // 1440
```

### 时间比较
```javascript
const time1 = time.nowUnix();
const time2 = time.addHours(time1, 1);

console.log(time.isBefore(time1, time2));  // true
console.log(time.isAfter(time1, time2));   // false
```

### 延迟执行
```javascript
// 延迟 2 秒
time.sleep(2).then(() => {
    console.log('2 秒后执行');
});

// 延迟 500 毫秒
time.sleepMillis(500).then(() => {
    console.log('500 毫秒后执行');
});

// 链式延迟
time.sleep(1)
    .then(() => time.sleep(1))
    .then(() => console.log('2秒后完成'));
```

### Ticker 定时器
```javascript
// setInterval 用法（类似 JavaScript）
const timerId = time.setInterval(() => {
    console.log('每秒执行');
}, 1000);

// 清除定时器
time.clearInterval(timerId);

// createTicker 用法
const ticker = time.createTicker(500);

ticker.tick(() => {
    console.log('Tick!');
});

// 停止 ticker
ticker.stop();

// 重置间隔
ticker.reset(1000);
```

### 获取时间组件
```javascript
const timestamp = time.nowUnix();

console.log(time.getYear(timestamp));    // 2024
console.log(time.getMonth(timestamp));   // 12
console.log(time.getDay(timestamp));     // 27
console.log(time.getHour(timestamp));    // 15
console.log(time.getMinute(timestamp));  // 30
console.log(time.getSecond(timestamp));  // 0

const weekday = time.getWeekday(timestamp);
console.log(weekday.name);    // "Friday"
console.log(weekday.number);  // 5
```

### 创建时间
```javascript
// 创建指定日期时间
const christmas = time.create(2024, 12, 25, 18, 30, 0);
console.log(time.format(christmas, time.FORMAT.DateTime));

// 从时间戳创建
const timeObj = time.fromUnix(1703692800);
console.log(timeObj.iso);  // ISO 格式
console.log(timeObj.year); // 年份
```

### 时区转换
```javascript
const now = time.nowUnix();

// 转换为 UTC
const utc = time.utc(now);
console.log(utc);

// 转换为本地时区
const local = time.local(now);
console.log(local);

// 转换到指定时区
const tokyo = time.inLocation(now, 'Asia/Tokyo');
console.log(tokyo);
```

## 实际应用场景

### 1. 定时任务调度
```javascript
function scheduleTask(taskName, delaySeconds) {
    console.log(`调度任务: ${taskName}`);
    return time.sleep(delaySeconds).then(() => {
        console.log(`执行任务: ${taskName}`);
    });
}

scheduleTask('数据备份', 60);
```

### 2. 性能计时
```javascript
const startTime = time.nowUnixMilli();

// 执行一些操作...

const endTime = time.nowUnixMilli();
const duration = endTime - startTime;
console.log('耗时:', duration, '毫秒');
```

### 3. 倒计时
```javascript
function countdown(targetTimestamp) {
    const now = time.nowUnix();
    const diff = time.diff(targetTimestamp, now);
    
    return {
        days: diff.days,
        hours: diff.hours % 24,
        minutes: diff.minutes % 60,
        seconds: diff.seconds % 60
    };
}
```

### 4. 相对时间显示
```javascript
function timeAgo(timestamp) {
    const now = time.nowUnix();
    const diff = now - timestamp;
    
    if (diff < 60) return '刚刚';
    if (diff < 3600) return `${Math.floor(diff / 60)} 分钟前`;
    if (diff < 86400) return `${Math.floor(diff / 3600)} 小时前`;
    return time.format(timestamp, time.FORMAT.Date);
}
```

## 格式常量

```javascript
time.FORMAT.RFC3339      // "2006-01-02T15:04:05Z07:00"
time.FORMAT.DateTime     // "2006-01-02 15:04:05"
time.FORMAT.Date         // "2006-01-02"
time.FORMAT.Time         // "15:04:05"
time.FORMAT.RFC1123      // "Mon, 02 Jan 2006 15:04:05 MST"
time.FORMAT.Kitchen      // "3:04PM"
```

## 时间单位常量

```javascript
time.UNIT.NANOSECOND     // 纳秒
time.UNIT.MICROSECOND    // 微秒
time.UNIT.MILLISECOND    // 毫秒
time.UNIT.SECOND         // 秒
time.UNIT.MINUTE         // 分钟
time.UNIT.HOUR           // 小时
```
