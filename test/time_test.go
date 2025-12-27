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
			tickerCount++;
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
