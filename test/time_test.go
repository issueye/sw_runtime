package test

import (
	"testing"
	"time"

	"sw_runtime/internal/runtime"
)

// TestTimeModule 测试时间模块基础功能
func TestTimeModule(t *testing.T) {
	runner := runtime.NewOrPanic()

	script := `
		const time = require('time');

		if (!time) {
			throw new Error('Time module not loaded');
		}

		console.log('Time module loaded successfully');
	`

	err := runner.RunCode(script)
	if err != nil {
		t.Fatalf("Script execution failed: %v", err)
	}
}

// TestTimeNow 测试获取当前时间
func TestTimeNow(t *testing.T) {
	runner := runtime.NewOrPanic()

	script := `
		const time = require('time');

		const now = time.now();
		if (!now || typeof now !== 'string') {
			throw new Error('now() should return a string');
		}

		const nowUnix = time.nowUnix();
		if (typeof nowUnix !== 'number') {
			throw new Error('nowUnix() should return a number');
		}

		const nowUnixMilli = time.nowUnixMilli();
		if (typeof nowUnixMilli !== 'number') {
			throw new Error('nowUnixMilli() should return a number');
		}

		console.log('Time now test passed');
	`

	err := runner.RunCode(script)
	if err != nil {
		t.Fatalf("Script execution failed: %v", err)
	}
}

// TestTimeFormat 测试时间格式化
func TestTimeFormat(t *testing.T) {
	runner := runtime.NewOrPanic()

	script := `
		const time = require('time');

		const timestamp = time.nowUnix();
		const formatted = time.format(timestamp, time.FORMAT.DateTime);

		if (!formatted || typeof formatted !== 'string') {
			throw new Error('format() should return a string');
		}

		console.log('Formatted time:', formatted);
		console.log('Time format test passed');
	`

	err := runner.RunCode(script)
	if err != nil {
		t.Fatalf("Script execution failed: %v", err)
	}
}

// TestTimeParse 测试时间解析
func TestTimeParse(t *testing.T) {
	runner := runtime.NewOrPanic()

	script := `
		const time = require('time');

		const parsed = time.parse('2024-12-27T15:30:00Z');

		if (!parsed) {
			throw new Error('parse() should return an object');
		}

		if (parsed.year !== 2024) {
			throw new Error('Year should be 2024');
		}

		if (parsed.month !== 12) {
			throw new Error('Month should be 12');
		}

		if (parsed.day !== 27) {
			throw new Error('Day should be 27');
		}

		console.log('Time parse test passed');
	`

	err := runner.RunCode(script)
	if err != nil {
		t.Fatalf("Script execution failed: %v", err)
	}
}

// TestTimeAdd 测试时间计算
func TestTimeAdd(t *testing.T) {
	runner := runtime.NewOrPanic()

	script := `
		const time = require('time');

		const now = time.nowUnix();

		const tomorrow = time.addDays(now, 1);
		if (tomorrow <= now) {
			throw new Error('addDays should increase timestamp');
		}

		const nextHour = time.addHours(now, 1);
		if (nextHour <= now) {
			throw new Error('addHours should increase timestamp');
		}

		const nextMinute = time.addMinutes(now, 1);
		if (nextMinute <= now) {
			throw new Error('addMinutes should increase timestamp');
		}

		const nextSecond = time.addSeconds(now, 1);
		if (nextSecond <= now) {
			throw new Error('addSeconds should increase timestamp');
		}

		console.log('Time add test passed');
	`

	err := runner.RunCode(script)
	if err != nil {
		t.Fatalf("Script execution failed: %v", err)
	}
}

// TestTimeComparison 测试时间比较
func TestTimeComparison(t *testing.T) {
	runner := runtime.NewOrPanic()

	script := `
		const time = require('time');

		const time1 = time.nowUnix();
		const time2 = time.addHours(time1, 1);

		if (!time.isBefore(time1, time2)) {
			throw new Error('time1 should be before time2');
		}

		if (!time.isAfter(time2, time1)) {
			throw new Error('time2 should be after time1');
		}

		console.log('Time comparison test passed');
	`

	err := runner.RunCode(script)
	if err != nil {
		t.Fatalf("Script execution failed: %v", err)
	}
}

// TestTimeDiff 测试时间差计算
func TestTimeDiff(t *testing.T) {
	runner := runtime.NewOrPanic()

	script := `
		const time = require('time');

		const time1 = time.nowUnix();
		const time2 = time.addHours(time1, 2);

		const diff = time.diff(time2, time1);

		if (!diff) {
			throw new Error('diff() should return an object');
		}

		if (diff.hours !== 2) {
			throw new Error('Diff should be 2 hours');
		}

		console.log('Time diff test passed');
	`

	err := runner.RunCode(script)
	if err != nil {
		t.Fatalf("Script execution failed: %v", err)
	}
}

// TestTimeComponents 测试时间组件获取
func TestTimeComponents(t *testing.T) {
	runner := runtime.NewOrPanic()

	script := `
		const time = require('time');

		const timestamp = time.nowUnix();

		const year = time.getYear(timestamp);
		if (typeof year !== 'number' || year < 2024) {
			throw new Error('Invalid year');
		}

		const month = time.getMonth(timestamp);
		if (typeof month !== 'number' || month < 1 || month > 12) {
			throw new Error('Invalid month');
		}

		const day = time.getDay(timestamp);
		if (typeof day !== 'number' || day < 1 || day > 31) {
			throw new Error('Invalid day');
		}

		const weekday = time.getWeekday(timestamp);
		if (!weekday || !weekday.name || typeof weekday.number !== 'number') {
			throw new Error('Invalid weekday');
		}

		console.log('Time components test passed');
	`

	err := runner.RunCode(script)
	if err != nil {
		t.Fatalf("Script execution failed: %v", err)
	}
}

// TestTimeCreate 测试时间创建
func TestTimeCreate(t *testing.T) {
	runner := runtime.NewOrPanic()

	script := `
		const time = require('time');

		const customTime = time.create(2024, 12, 25, 18, 30, 0);

		if (typeof customTime !== 'number') {
			throw new Error('create() should return a timestamp');
		}

		const year = time.getYear(customTime);
		if (year !== 2024) {
			throw new Error('Year should be 2024');
		}

		const month = time.getMonth(customTime);
		if (month !== 12) {
			throw new Error('Month should be 12');
		}

		const day = time.getDay(customTime);
		if (day !== 25) {
			throw new Error('Day should be 25');
		}

		console.log('Time create test passed');
	`

	err := runner.RunCode(script)
	if err != nil {
		t.Fatalf("Script execution failed: %v", err)
	}
}

// TestTimeFromUnix 测试从时间戳创建
func TestTimeFromUnix(t *testing.T) {
	runner := runtime.NewOrPanic()

	script := `
		const time = require('time');

		const timeObj = time.fromUnix(1703692800);

		if (!timeObj || !timeObj.iso || !timeObj.year) {
			throw new Error('fromUnix() should return a time object');
		}

		console.log('Time fromUnix test passed');
	`

	err := runner.RunCode(script)
	if err != nil {
		t.Fatalf("Script execution failed: %v", err)
	}
}

// TestTimeSleep 测试延迟执行
func TestTimeSleep(t *testing.T) {
	runner := runtime.NewOrPanic()

	script := `
		const time = require('time');

		let executed = false;

		time.sleep(0.1).then(() => {
			executed = true;
			console.log('Sleep executed');
		});

		// 给一些时间让 Promise 执行
		setTimeout(() => {
			if (!executed) {
				console.log('Warning: sleep may not have executed yet');
			}
		}, 200);
	`

	err := runner.RunCode(script)
	if err != nil {
		t.Fatalf("Script execution failed: %v", err)
	}

	// 等待一下让异步操作完成
	time.Sleep(200 * time.Millisecond)
}

// TestTimeConstants 测试时间常量
func TestTimeConstants(t *testing.T) {
	runner := runtime.NewOrPanic()

	script := `
		const time = require('time');

		// 测试格式常量
		if (!time.FORMAT || !time.FORMAT.DateTime || !time.FORMAT.RFC3339) {
			throw new Error('Format constants not defined');
		}

		// 测试 dayjs 格式常量
		if (!time.FORMAT.YYYY || !time.FORMAT.MM || !time.FORMAT.DD) {
			throw new Error('Dayjs format constants not defined');
		}

		// 测试单位常量
		if (!time.UNIT || typeof time.UNIT.SECOND !== 'number') {
			throw new Error('Unit constants not defined');
		}

		console.log('Time constants test passed');
	`

	err := runner.RunCode(script)
	if err != nil {
		t.Fatalf("Script execution failed: %v", err)
	}
}

// BenchmarkTimeModule 性能测试
func BenchmarkTimeModule(b *testing.B) {
	runner := runtime.NewOrPanic()

	script := `
		const time = require('time');
		const now = time.nowUnix();
		time.format(now, time.FORMAT.DateTime);
	`

	for i := 0; i < b.N; i++ {
		err := runner.RunCode(script)
		if err != nil {
			b.Fatalf("Script execution failed: %v", err)
		}
	}
}

// TestTimeSetInterval 测试 setInterval
func TestTimeSetInterval(t *testing.T) {
	runner := runtime.NewOrPanic()

	script := `
		const time = require('time');

		let count = 0;
		const timerId = time.setInterval(() => {
			count++;
			console.log('Tick:', count);

			if (count >= 3) {
				time.clearInterval(timerId);
			}
		}, 100);

		if (typeof timerId !== 'number') {
			throw new Error('setInterval should return a timer ID');
		}

		console.log('setInterval test passed');
	`

	err := runner.RunCode(script)
	if err != nil {
		t.Fatalf("Script execution failed: %v", err)
	}

	// 等待定时器执行
	time.Sleep(400 * time.Millisecond)
}

// TestTimeClearInterval 测试 clearInterval
func TestTimeClearInterval(t *testing.T) {
	runner := runtime.NewOrPanic()

	script := `
		const time = require('time');

		let count = 0;
		const timerId = time.setInterval(() => {
			count++;
		}, 100);

		// 立即清除定时器
		time.clearInterval(timerId);

		// 等待一下
		setTimeout(() => {
			if (count > 0) {
				throw new Error('Timer should have been cleared');
			}
			console.log('clearInterval test passed');
		}, 300);
	`

	err := runner.RunCode(script)
	if err != nil {
		t.Fatalf("Script execution failed: %v", err)
	}

	time.Sleep(400 * time.Millisecond)
}

// TestTimeCreateTicker 测试 createTicker
func TestTimeCreateTicker(t *testing.T) {
	runner := runtime.NewOrPanic()

	script := `
		const time = require('time');

		const ticker = time.createTicker(100);

		if (!ticker) {
			throw new Error('createTicker should return a ticker object');
		}

		if (typeof ticker.tick !== 'function') {
			throw new Error('ticker should have tick method');
		}

		if (typeof ticker.stop !== 'function') {
			throw new Error('ticker should have stop method');
		}

		if (typeof ticker.reset !== 'function') {
			throw new Error('ticker should have reset method');
		}

		let tickCount = 0;
		ticker.tick(() => {
			tickCount++;
			console.log('Ticked:', tickCount);

			if (tickCount >= 3) {
				ticker.stop();
			}
		});

		console.log('createTicker test passed');
	`

	err := runner.RunCode(script)
	if err != nil {
		t.Fatalf("Script execution failed: %v", err)
	}

	time.Sleep(400 * time.Millisecond)
}

// TestTickerReset 测试 ticker reset
func TestTickerReset(t *testing.T) {
	runner := runtime.NewOrPanic()

	script := `
		const time = require('time');

		const ticker = time.createTicker(200);
		let resetCalled = false;

		ticker.tick(() => {
			if (!resetCalled) {
				resetCalled = true;
				ticker.reset(50); // 加速
				console.log('Ticker reset to 50ms');
			}
		});

		setTimeout(() => {
			ticker.stop();
			console.log('Ticker reset test passed');
		}, 500);
	`

	err := runner.RunCode(script)
	if err != nil {
		t.Fatalf("Script execution failed: %v", err)
	}

	time.Sleep(600 * time.Millisecond)
}

// TestMultipleTickers 测试多个定时器
func TestMultipleTickers(t *testing.T) {
	runner := runtime.NewOrPanic()

	script := `
		const time = require('time');

		const timer1 = time.setInterval(() => {
			console.log('Timer 1');
		}, 100);

		const timer2 = time.setInterval(() => {
			console.log('Timer 2');
		}, 150);

		const timer3 = time.setInterval(() => {
			console.log('Timer 3');
		}, 200);

		setTimeout(() => {
			time.clearInterval(timer1);
			time.clearInterval(timer2);
			time.clearInterval(timer3);
			console.log('All timers cleared');
		}, 500);
	`

	err := runner.RunCode(script)
	if err != nil {
		t.Fatalf("Script execution failed: %v", err)
	}

	time.Sleep(600 * time.Millisecond)
}

// ========== Dayjs 风格 API 测试 ==========

// TestDayjsBasic 测试 dayjs 基础用法
func TestDayjsBasic(t *testing.T) {
	runner := runtime.NewOrPanic()

	script := `
		const time = require('time');

		// 测试 dayjs() 创建当前时间
		const now = time.dayjs();
		if (!now || typeof now.format !== 'function') {
			throw new Error('dayjs() should return a time object with format method');
		}

		// 测试格式化
		const formatted = now.format('YYYY-MM-DD');
		if (!formatted || formatted.length !== 10) {
			throw new Error('format should return YYYY-MM-DD');
		}

		console.log('Dayjs basic test passed, formatted:', formatted);
	`

	err := runner.RunCode(script)
	if err != nil {
		t.Fatalf("Script execution failed: %v", err)
	}
}

// TestDayjsFromString 测试从字符串创建
func TestDayjsFromString(t *testing.T) {
	runner := runtime.NewOrPanic()

	script := `
		const time = require('time');

		// 测试从 ISO 字符串创建
		const d1 = time.dayjs('2024-12-25T10:30:00Z');
		if (!d1) {
			throw new Error('dayjs should parse string');
		}

		// 测试获取组件
		const year = d1.year();
		const month = d1.month();
		const date = d1.date();

		if (year !== 2024) {
			throw new Error('Year should be 2024');
		}
		if (month !== 11) { // dayjs month is 0-indexed
			throw new Error('Month should be 11 (December)');
		}
		if (date !== 25) {
			throw new Error('Date should be 25');
		}

		console.log('Dayjs from string test passed');
	`

	err := runner.RunCode(script)
	if err != nil {
		t.Fatalf("Script execution failed: %v", err)
	}
}

// TestDayjsUnix 测试从时间戳创建
func TestDayjsUnix(t *testing.T) {
	runner := runtime.NewOrPanic()

	script := `
		const time = require('time');

		// 测试从 Unix 时间戳创建
		const timestamp = 1703692800;
		const d = time.unix(timestamp);

		if (!d) {
			throw new Error('unix should return a dayjs object');
		}

		// 测试 unix() 方法返回秒时间戳
		const ts = d.unix();
		if (ts !== timestamp) {
			throw new Error('unix() should return the timestamp');
		}

		// 测试 valueOf() 返回毫秒时间戳
		const ms = d.valueOf();
		if (ms !== timestamp * 1000) {
			throw new Error('valueOf() should return milliseconds');
		}

		console.log('Dayjs unix test passed');
	`

	err := runner.RunCode(script)
	if err != nil {
		t.Fatalf("Script execution failed: %v", err)
	}
}

// TestDayjsChain 测试链式调用
func TestDayjsChain(t *testing.T) {
	runner := runtime.NewOrPanic()

	script := `
		const time = require('time');

		// 测试链式调用
		const result = time.dayjs()
			.add(1, 'day')
			.add(2, 'hour')
			.format('YYYY-MM-DD HH:mm:ss');

		if (!result || result.length < 19) {
			throw new Error('Chain should work');
		}

		console.log('Dayjs chain result:', result);
		console.log('Dayjs chain test passed');
	`

	err := runner.RunCode(script)
	if err != nil {
		t.Fatalf("Script execution failed: %v", err)
	}
}

// TestDayjsAddSubtract 测试加减时间
func TestDayjsAddSubtract(t *testing.T) {
	runner := runtime.NewOrPanic()

	script := `
		const time = require('time');

		const d = time.dayjs('2024-01-15');

		// 加 1 周
		const d1 = d.add(1, 'week');
		if (d1.date() !== 22) {
			throw new Error('add(1, week) should add 7 days');
		}

		// 减 3 天
		const d2 = d.subtract(3, 'day');
		if (d2.date() !== 12) {
			throw new Error('subtract(3, day) should subtract 3 days');
		}

		// 加 1 个月
		const d3 = d.add(1, 'month');
		if (d3.month() !== 1) { // February = 1 (0-indexed)
			throw new Error('add(1, month) should increment month');
		}

		console.log('Dayjs add/subtract test passed');
	`

	err := runner.RunCode(script)
	if err != nil {
		t.Fatalf("Script execution failed: %v", err)
	}
}

// TestDayjsStartOfEndOf 测试 startOf/endOf
func TestDayjsStartOfEndOf(t *testing.T) {
	runner := runtime.NewOrPanic()

	script := `
		const time = require('time');

		const d = time.dayjs('2024-06-15T14:30:00');

		// startOf month
		const startOfMonth = d.startOf('month');
		if (startOfMonth.date() !== 1) {
			throw new Error('startOf(month) should return 1st day');
		}

		// endOf month
		const endOfMonth = d.endOf('month');
		if (endOfMonth.date() !== 30) {
			throw new Error('endOf(month) should return last day');
		}

		// startOf day
		const startOfDay = d.startOf('day');
		const hour = startOfDay.hour();
		const minute = startOfDay.minute();
		if (hour !== 0 || minute !== 0) {
			throw new Error('startOf(day) should return 00:00:00');
		}

		console.log('Dayjs startOf/endOf test passed');
	`

	err := runner.RunCode(script)
	if err != nil {
		t.Fatalf("Script execution failed: %v", err)
	}
}

// TestDayjsComparison 测试时间比较
func TestDayjsComparison(t *testing.T) {
	runner := runtime.NewOrPanic()

	script := `
		const time = require('time');

		const d1 = time.dayjs('2024-01-01');
		const d2 = time.dayjs('2024-01-15');

		// isBefore
		if (!d1.isBefore(d2)) {
			throw new Error('d1 should be before d2');
		}

		// isAfter
		if (!d2.isAfter(d1)) {
			throw new Error('d2 should be after d1');
		}

		// isSame
		const d3 = time.dayjs('2024-01-01');
		if (!d1.isSame(d3)) {
			throw new Error('d1 and d3 should be same');
		}

		console.log('Dayjs comparison test passed');
	`

	err := runner.RunCode(script)
	if err != nil {
		t.Fatalf("Script execution failed: %v", err)
	}
}

// TestDayjsFormatTokens 测试各种格式 tokens
func TestDayjsFormatTokens(t *testing.T) {
	runner := runtime.NewOrPanic()

	script := `
		const time = require('time');

		const d = time.dayjs('2024-12-25T14:30:45');

		// 测试各种格式
		const tests = [
			['YYYY', '2024'],
			['YY', '24'],
			['M', '12'],
			['MM', '12'],
			['D', '25'],
			['DD', '25'],
			['H', '14'],
			['HH', '14'],
			['m', '30'],
			['mm', '30'],
			['s', '45'],
			['ss', '45'],
		];

		for (const [fmt, expected] of tests) {
			const result = d.format(fmt);
			if (result !== expected) {
				throw new Error('format(' + fmt + ') should be ' + expected + ', got ' + result);
			}
		}

		// 复合格式
		const full = d.format('YYYY-MM-DD HH:mm:ss');
		if (full !== '2024-12-25 14:30:45') {
			throw new Error('Full format failed');
		}

		console.log('Dayjs format tokens test passed');
	`

	err := runner.RunCode(script)
	if err != nil {
		t.Fatalf("Script execution failed: %v", err)
	}
}

// TestDayjsClone 测试克隆
func TestDayjsClone(t *testing.T) {
	runner := runtime.NewOrPanic()

	script := `
		const time = require('time');

		const d = time.dayjs('2024-01-01');
		const cloned = d.clone();

		// 修改克隆不应影响原对象
		const modified = cloned.add(1, 'day');

		const dDate = d.date();
		const modifiedDate = modified.date();

		// 克隆后的日期应该不同（克隆加1天后是2号，原对象还是1号）
		if (dDate === modifiedDate) {
			throw new Error('Clone should be independent');
		}

		console.log('Dayjs clone test passed');
	`

	err := runner.RunCode(script)
	if err != nil {
		t.Fatalf("Script execution failed: %v", err)
	}
}

// TestDayjsDiff 测试时间差
func TestDayjsDiff(t *testing.T) {
	runner := runtime.NewOrPanic()

	script := `
		const time = require('time');

		const d1 = time.dayjs('2024-01-01');
		const d2 = time.dayjs('2024-01-10');

		// 相差 9 天
		const diffDays = d2.diff(d1, 'day');
		if (diffDays !== 9) {
			throw new Error('diff in days should be 9, got ' + diffDays);
		}

		// 相差 216 小时
		const diffHours = d2.diff(d1, 'hour');
		if (diffHours !== 216) {
			throw new Error('diff in hours should be 216');
		}

		console.log('Dayjs diff test passed');
	`

	err := runner.RunCode(script)
	if err != nil {
		t.Fatalf("Script execution failed: %v", err)
	}
}

// TestDayjsUtils 测试其他工具方法
func TestDayjsUtils(t *testing.T) {
	runner := runtime.NewOrPanic()

	script := `
		const time = require('time');

		const d = time.dayjs('2024-03-15'); // 3 月 15 日

		// weekday (0=周日, 6=周六)
		const weekday = d.weekday();
		if (typeof weekday !== 'number') {
			throw new Error('weekday should return number');
		}

		// dayOfYear
		const dayOfYear = d.dayOfYear();
		if (dayOfYear !== 75) { // 3月15日是第75天
			throw new Error('dayOfYear should be 75, got ' + dayOfYear);
		}

		// isLeapYear
		const isLeap = d.isLeapYear();
		if (!isLeap) {
			throw new Error('2024 is leap year');
		}

		// daysInMonth
		const daysInMonth = d.daysInMonth();
		if (daysInMonth !== 31) {
			throw new Error('March should have 31 days');
		}

		console.log('Dayjs utils test passed');
	`

	err := runner.RunCode(script)
	if err != nil {
		t.Fatalf("Script execution failed: %v", err)
	}
}

// TestDayjsComplexChain 测试复杂链式调用
func TestDayjsComplexChain(t *testing.T) {
	runner := runtime.NewOrPanic()

	script := `
		const time = require('time');

		// 复杂链式调用示例
		const result = time.dayjs('2024-01-15')
			.add(1, 'year')
			.subtract(1, 'month')
			.startOf('month')
			.format('YYYY-MM-DD');

		console.log('Complex chain result:', result);

		// 验证结果
		// 2024-01-15 + 1年 = 2025-01-15
		// -1月 = 2024-12-15
		// startOf month = 2024-12-01
		if (result !== '2024-12-01') {
			throw new Error('Complex chain failed, expected 2024-12-01, got ' + result);
		}

		console.log('Dayjs complex chain test passed');
	`

	err := runner.RunCode(script)
	if err != nil {
		t.Fatalf("Script execution failed: %v", err)
	}
}

// TestDayjsSubtract 测试 subtract
func TestDayjsSubtract(t *testing.T) {
	runner := runtime.NewOrPanic()

	script := `
		const time = require('time');

		// 测试独立的 subtract 函数
		const now = time.nowUnix();
		const yesterday = time.subtractDays(now, 1);

		if (yesterday >= now) {
			throw new Error('subtractDays should return earlier timestamp');
		}

		// 测试 subtractHours, subtractMinutes, subtractSeconds
		const hourAgo = time.subtractHours(now, 1);
		const minuteAgo = time.subtractMinutes(now, 1);
		const secondAgo = time.subtractSeconds(now, 1);

		if (hourAgo >= now || minuteAgo >= now || secondAgo >= now) {
			throw new Error('subtract functions should return earlier timestamps');
		}

		console.log('Subtract functions test passed');
	`

	err := runner.RunCode(script)
	if err != nil {
		t.Fatalf("Script execution failed: %v", err)
	}
}

// TestDayjsIsSame 测试 isSame
func TestDayjsIsSame(t *testing.T) {
	runner := runtime.NewOrPanic()

	script := `
		const time = require('time');

		// 测试独立的 isSame 函数
		const ts1 = time.nowUnix();
		const ts2 = time.addSeconds(ts1, 0);

		if (!time.isSame(ts1, ts2)) {
			throw new Error('isSame should return true for same timestamp');
		}

		const ts3 = time.addSeconds(ts1, 10);
		if (time.isSame(ts1, ts3)) {
			throw new Error('isSame should return false for different timestamps');
		}

		console.log('isSame function test passed');
	`

	err := runner.RunCode(script)
	if err != nil {
		t.Fatalf("Script execution failed: %v", err)
	}
}

// TestDayjsStartOfEndOfUnit 测试 startOf/endOf 单位
func TestDayjsStartOfEndOfUnit(t *testing.T) {
	runner := runtime.NewOrPanic()

	script := `
		const time = require('time');

		const ts = time.nowUnix();

		// 测试各种单位
		const units = ['year', 'month', 'week', 'day', 'hour', 'minute', 'second'];

		for (const unit of units) {
			const start = time.startOf(ts, unit);
			const end = time.endOf(ts, unit);

			if (!start || !end) {
				throw new Error('startOf/endOf should return timestamp for unit: ' + unit);
			}

			if (start >= end) {
				throw new Error('startOf should be before endOf for unit: ' + unit);
			}
		}

		console.log('startOf/endOf units test passed');
	`

	err := runner.RunCode(script)
	if err != nil {
		t.Fatalf("Script execution failed: %v", err)
	}
}

// TestDayjsToISOString 测试 toISOString
func TestDayjsToISOString(t *testing.T) {
	runner := runtime.NewOrPanic()

	script := `
		const time = require('time');

		const ts = time.nowUnix();
		const iso = time.toISOString(ts);

		// ISO 字符串应该包含 'T' 和 'Z'
		if (!iso.includes('T') || !iso.endsWith('Z')) {
			throw new Error('toISOString should return ISO format');
		}

		// dayjs 对象的 toISOString 方法
		const d = time.dayjs('2024-12-25T10:30:00Z');
		const dIso = d.toISOString();

		if (dIso !== '2024-12-25T10:30:00Z') {
			throw new Error('dayjs toISOString failed');
		}

		console.log('toISOString test passed');
	`

	err := runner.RunCode(script)
	if err != nil {
		t.Fatalf("Script execution failed: %v", err)
	}
}

// TestDayjsGetDate 测试 getDate 别名
func TestDayjsGetDate(t *testing.T) {
	runner := runtime.NewOrPanic()

	script := `
		const time = require('time');

		const ts = time.nowUnix();

		const day1 = time.getDay(ts);
		const day2 = time.getDate(ts);

		if (day1 !== day2) {
			throw new Error('getDay and getDate should return same value');
		}

		console.log('getDate alias test passed');
	`

	err := runner.RunCode(script)
	if err != nil {
		t.Fatalf("Script execution failed: %v", err)
	}
}

// TestDayjsGetValueOf 测试 getValueOf
func TestDayjsGetValueOf(t *testing.T) {
	runner := runtime.NewOrPanic()

	script := `
		const time = require('time');

		const ts = time.nowUnix();
		const ms = time.getValueOf(ts);

		// getValueOf 应该返回毫秒时间戳
		if (ms !== ts * 1000) {
			throw new Error('getValueOf should return milliseconds');
		}

		console.log('getValueOf test passed');
	`

	err := runner.RunCode(script)
	if err != nil {
		t.Fatalf("Script execution failed: %v", err)
	}
}
