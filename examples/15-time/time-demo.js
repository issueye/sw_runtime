// 时间模块基础示例
const { time } = require('utils');

console.log('=== Time Module Demo ===\n');

// 1. 获取当前时间
console.log('--- 获取当前时间 ---');
console.log('ISO 格式:', time.now());
console.log('Unix 时间戳（秒）:', time.nowUnix());
console.log('Unix 时间戳（毫秒）:', time.nowUnixMilli());
console.log('Unix 时间戳（纳秒）:', time.nowUnixNano());

// 2. 时间格式化
console.log('\n--- 时间格式化 ---');
const timestamp = time.nowUnix();
console.log('RFC3339:', time.format(timestamp, time.FORMAT.RFC3339));
console.log('DateTime:', time.format(timestamp, time.FORMAT.DateTime));
console.log('Date:', time.format(timestamp, time.FORMAT.Date));
console.log('Time:', time.format(timestamp, time.FORMAT.Time));
console.log('RFC1123:', time.format(timestamp, time.FORMAT.RFC1123));
console.log('Kitchen:', time.format(timestamp, time.FORMAT.Kitchen));

// 3. 时间解析
console.log('\n--- 时间解析 ---');
const parsed = time.parse('2024-12-27T15:30:00Z');
console.log('解析结果:', parsed);
console.log('年份:', parsed.year);
console.log('月份:', parsed.month);
console.log('日期:', parsed.day);
console.log('小时:', parsed.hour);
console.log('分钟:', parsed.minute);
console.log('秒数:', parsed.second);
console.log('星期:', parsed.weekday);

// 4. 自定义格式解析
console.log('\n--- 自定义格式解析 ---');
const customParsed = time.parse('2024-12-27 15:30:00', time.FORMAT.DateTime);
console.log('自定义格式解析:', customParsed.iso);

// 5. 时间计算
console.log('\n--- 时间计算 ---');
const now = time.nowUnix();
console.log('当前时间戳:', now);

const tomorrow = time.addDays(now, 1);
console.log('明天:', time.format(tomorrow, time.FORMAT.DateTime));

const nextHour = time.addHours(now, 1);
console.log('一小时后:', time.format(nextHour, time.FORMAT.DateTime));

const in30Minutes = time.addMinutes(now, 30);
console.log('30分钟后:', time.format(in30Minutes, time.FORMAT.DateTime));

const in60Seconds = time.addSeconds(now, 60);
console.log('60秒后:', time.format(in60Seconds, time.FORMAT.DateTime));

// 6. 时间比较
console.log('\n--- 时间比较 ---');
const time1 = time.nowUnix();
const time2 = time.addHours(time1, 2);

console.log('时间1:', time.format(time1, time.FORMAT.DateTime));
console.log('时间2:', time.format(time2, time.FORMAT.DateTime));
console.log('时间1 在 时间2 之前?', time.isBefore(time1, time2));
console.log('时间1 在 时间2 之后?', time.isAfter(time1, time2));

// 7. 时间差计算
console.log('\n--- 时间差计算 ---');
const diff = time.diff(time2, time1);
console.log('时间差:');
console.log('  秒数:', diff.seconds);
console.log('  分钟:', diff.minutes);
console.log('  小时:', diff.hours);
console.log('  天数:', diff.days);

// 8. 时区转换
console.log('\n--- 时区转换 ---');
const utcTime = time.utc(now);
console.log('UTC 时间:', utcTime);

const localTime = time.local(now);
console.log('本地时间:', localTime);

// 9. 获取时间组件
console.log('\n--- 获取时间组件 ---');
console.log('年份:', time.getYear(now));
console.log('月份:', time.getMonth(now));
console.log('日期:', time.getDay(now));
console.log('小时:', time.getHour(now));
console.log('分钟:', time.getMinute(now));
console.log('秒数:', time.getSecond(now));

const weekday = time.getWeekday(now);
console.log('星期几:', weekday.name, '(' + weekday.number + ')');

// 10. 创建时间
console.log('\n--- 创建时间 ---');
const customTime = time.create(2024, 12, 25, 18, 30, 0);
console.log('圣诞节晚上6:30:', time.format(customTime, time.FORMAT.DateTime));

// 11. 从时间戳创建时间对象
console.log('\n--- 从时间戳创建 ---');
const timeObj = time.fromUnix(1703692800);
console.log('时间对象:', timeObj);
console.log('ISO 格式:', timeObj.iso);

// 12. 延迟执行示例
console.log('\n--- 延迟执行 ---');
console.log('开始延迟 2 秒...');
time.sleep(2).then(() => {
    console.log('2 秒延迟结束！');
    
    console.log('\n开始延迟 500 毫秒...');
    return time.sleepMillis(500);
}).then(() => {
    console.log('500 毫秒延迟结束！');
});

// 13. 时间单位常量
console.log('\n--- 时间单位常量 ---');
console.log('纳秒:', time.UNIT.NANOSECOND);
console.log('微秒:', time.UNIT.MICROSECOND);
console.log('毫秒:', time.UNIT.MILLISECOND);
console.log('秒:', time.UNIT.SECOND);
console.log('分钟:', time.UNIT.MINUTE);
console.log('小时:', time.UNIT.HOUR);

console.log('\n✨ 时间模块演示完成！');
