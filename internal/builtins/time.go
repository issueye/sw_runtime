package builtins

import (
	"fmt"
	"sync"
	"time"

	"github.com/dop251/goja"
)

// TimeModule 时间模块
type TimeModule struct {
	vm       *goja.Runtime
	tickers  map[int]*Ticker
	mutex    sync.RWMutex
	tickerId int
}

// Ticker 定时器
type Ticker struct {
	id       int
	ticker   *time.Ticker
	callback goja.Callable
	vm       *goja.Runtime
	stopped  bool
	mutex    sync.RWMutex
}

// NewTimeModule 创建时间模块
func NewTimeModule(vm *goja.Runtime) *TimeModule {
	return &TimeModule{
		vm:      vm,
		tickers: make(map[int]*Ticker),
	}
}

// GetModule 获取时间模块对象
func (t *TimeModule) GetModule() *goja.Object {
	obj := t.vm.NewObject()

	// 获取当前时间
	obj.Set("now", t.now)
	obj.Set("nowUnix", t.nowUnix)
	obj.Set("nowUnixMilli", t.nowUnixMilli)
	obj.Set("nowUnixNano", t.nowUnixNano)

	// 时间解析和格式化
	obj.Set("parse", t.parse)
	obj.Set("format", t.format)

	// 延迟执行
	obj.Set("sleep", t.sleep)
	obj.Set("sleepMillis", t.sleepMillis)

	// 时间计算
	obj.Set("add", t.add)
	obj.Set("addDays", t.addDays)
	obj.Set("addHours", t.addHours)
	obj.Set("addMinutes", t.addMinutes)
	obj.Set("addSeconds", t.addSeconds)

	// 时间比较
	obj.Set("isBefore", t.isBefore)
	obj.Set("isAfter", t.isAfter)
	obj.Set("diff", t.diff)

	// 时区处理
	obj.Set("utc", t.utc)
	obj.Set("local", t.local)
	obj.Set("inLocation", t.inLocation)

	// 时间组件获取
	obj.Set("getYear", t.getYear)
	obj.Set("getMonth", t.getMonth)
	obj.Set("getDay", t.getDay)
	obj.Set("getHour", t.getHour)
	obj.Set("getMinute", t.getMinute)
	obj.Set("getSecond", t.getSecond)
	obj.Set("getWeekday", t.getWeekday)

	// 时间创建
	obj.Set("create", t.create)
	obj.Set("fromUnix", t.fromUnix)
	obj.Set("fromUnixMilli", t.fromUnixMilli)

	// Ticker 定时器
	obj.Set("setInterval", t.setInterval)
	obj.Set("clearInterval", t.clearInterval)
	obj.Set("createTicker", t.createTicker)

	// 常用格式常量
	formats := t.vm.NewObject()
	formats.Set("RFC3339", time.RFC3339)
	formats.Set("RFC3339Nano", time.RFC3339Nano)
	formats.Set("RFC822", time.RFC822)
	formats.Set("RFC1123", time.RFC1123)
	formats.Set("ANSIC", time.ANSIC)
	formats.Set("UnixDate", time.UnixDate)
	formats.Set("RubyDate", time.RubyDate)
	formats.Set("Kitchen", time.Kitchen)
	formats.Set("DateTime", "2006-01-02 15:04:05")
	formats.Set("Date", "2006-01-02")
	formats.Set("Time", "15:04:05")
	obj.Set("FORMAT", formats)

	// 时间单位常量
	units := t.vm.NewObject()
	units.Set("NANOSECOND", int64(time.Nanosecond))
	units.Set("MICROSECOND", int64(time.Microsecond))
	units.Set("MILLISECOND", int64(time.Millisecond))
	units.Set("SECOND", int64(time.Second))
	units.Set("MINUTE", int64(time.Minute))
	units.Set("HOUR", int64(time.Hour))
	obj.Set("UNIT", units)

	return obj
}

// now 获取当前时间（ISO 8601 格式）
func (t *TimeModule) now(call goja.FunctionCall) goja.Value {
	return t.vm.ToValue(time.Now().Format(time.RFC3339))
}

// nowUnix 获取当前 Unix 时间戳（秒）
func (t *TimeModule) nowUnix(call goja.FunctionCall) goja.Value {
	return t.vm.ToValue(time.Now().Unix())
}

// nowUnixMilli 获取当前 Unix 时间戳（毫秒）
func (t *TimeModule) nowUnixMilli(call goja.FunctionCall) goja.Value {
	return t.vm.ToValue(time.Now().UnixMilli())
}

// nowUnixNano 获取当前 Unix 时间戳（纳秒）
func (t *TimeModule) nowUnixNano(call goja.FunctionCall) goja.Value {
	return t.vm.ToValue(time.Now().UnixNano())
}

// parse 解析时间字符串
func (t *TimeModule) parse(call goja.FunctionCall) goja.Value {
	if len(call.Arguments) < 1 {
		panic(t.vm.NewTypeError("parse requires at least 1 argument"))
	}

	timeStr := call.Arguments[0].String()
	layout := time.RFC3339

	if len(call.Arguments) > 1 {
		layout = call.Arguments[1].String()
	}

	parsedTime, err := time.Parse(layout, timeStr)
	if err != nil {
		panic(t.vm.NewGoError(err))
	}

	result := t.vm.NewObject()
	result.Set("unix", parsedTime.Unix())
	result.Set("unixMilli", parsedTime.UnixMilli())
	result.Set("unixNano", parsedTime.UnixNano())
	result.Set("iso", parsedTime.Format(time.RFC3339))
	result.Set("year", parsedTime.Year())
	result.Set("month", int(parsedTime.Month()))
	result.Set("day", parsedTime.Day())
	result.Set("hour", parsedTime.Hour())
	result.Set("minute", parsedTime.Minute())
	result.Set("second", parsedTime.Second())
	result.Set("weekday", parsedTime.Weekday().String())

	return result
}

// format 格式化时间戳
func (t *TimeModule) format(call goja.FunctionCall) goja.Value {
	if len(call.Arguments) < 1 {
		panic(t.vm.NewTypeError("format requires at least 1 argument"))
	}

	timestamp := call.Arguments[0].ToInteger()
	layout := time.RFC3339

	if len(call.Arguments) > 1 {
		layout = call.Arguments[1].String()
	}

	timeObj := time.Unix(timestamp, 0)
	return t.vm.ToValue(timeObj.Format(layout))
}

// sleep 延迟执行（秒）
func (t *TimeModule) sleep(call goja.FunctionCall) goja.Value {
	if len(call.Arguments) < 1 {
		panic(t.vm.NewTypeError("sleep requires seconds argument"))
	}

	seconds := call.Arguments[0].ToFloat()

	promise, resolve, _ := t.vm.NewPromise()

	go func() {
		time.Sleep(time.Duration(seconds * float64(time.Second)))
		resolve(goja.Undefined())
	}()

	return t.vm.ToValue(promise)
}

// sleepMillis 延迟执行（毫秒）
func (t *TimeModule) sleepMillis(call goja.FunctionCall) goja.Value {
	if len(call.Arguments) < 1 {
		panic(t.vm.NewTypeError("sleepMillis requires milliseconds argument"))
	}

	millis := call.Arguments[0].ToInteger()

	promise, resolve, _ := t.vm.NewPromise()

	go func() {
		time.Sleep(time.Duration(millis) * time.Millisecond)
		resolve(goja.Undefined())
	}()

	return t.vm.ToValue(promise)
}

// add 添加时间间隔
func (t *TimeModule) add(call goja.FunctionCall) goja.Value {
	if len(call.Arguments) < 2 {
		panic(t.vm.NewTypeError("add requires timestamp and duration arguments"))
	}

	timestamp := call.Arguments[0].ToInteger()
	duration := call.Arguments[1].ToInteger()

	timeObj := time.Unix(timestamp, 0)
	newTime := timeObj.Add(time.Duration(duration))

	return t.vm.ToValue(newTime.Unix())
}

// addDays 添加天数
func (t *TimeModule) addDays(call goja.FunctionCall) goja.Value {
	if len(call.Arguments) < 2 {
		panic(t.vm.NewTypeError("addDays requires timestamp and days arguments"))
	}

	timestamp := call.Arguments[0].ToInteger()
	days := call.Arguments[1].ToInteger()

	timeObj := time.Unix(timestamp, 0)
	newTime := timeObj.AddDate(0, 0, int(days))

	return t.vm.ToValue(newTime.Unix())
}

// addHours 添加小时
func (t *TimeModule) addHours(call goja.FunctionCall) goja.Value {
	if len(call.Arguments) < 2 {
		panic(t.vm.NewTypeError("addHours requires timestamp and hours arguments"))
	}

	timestamp := call.Arguments[0].ToInteger()
	hours := call.Arguments[1].ToInteger()

	timeObj := time.Unix(timestamp, 0)
	newTime := timeObj.Add(time.Duration(hours) * time.Hour)

	return t.vm.ToValue(newTime.Unix())
}

// addMinutes 添加分钟
func (t *TimeModule) addMinutes(call goja.FunctionCall) goja.Value {
	if len(call.Arguments) < 2 {
		panic(t.vm.NewTypeError("addMinutes requires timestamp and minutes arguments"))
	}

	timestamp := call.Arguments[0].ToInteger()
	minutes := call.Arguments[1].ToInteger()

	timeObj := time.Unix(timestamp, 0)
	newTime := timeObj.Add(time.Duration(minutes) * time.Minute)

	return t.vm.ToValue(newTime.Unix())
}

// addSeconds 添加秒数
func (t *TimeModule) addSeconds(call goja.FunctionCall) goja.Value {
	if len(call.Arguments) < 2 {
		panic(t.vm.NewTypeError("addSeconds requires timestamp and seconds arguments"))
	}

	timestamp := call.Arguments[0].ToInteger()
	seconds := call.Arguments[1].ToInteger()

	timeObj := time.Unix(timestamp, 0)
	newTime := timeObj.Add(time.Duration(seconds) * time.Second)

	return t.vm.ToValue(newTime.Unix())
}

// isBefore 判断时间是否在另一个时间之前
func (t *TimeModule) isBefore(call goja.FunctionCall) goja.Value {
	if len(call.Arguments) < 2 {
		panic(t.vm.NewTypeError("isBefore requires two timestamp arguments"))
	}

	time1 := call.Arguments[0].ToInteger()
	time2 := call.Arguments[1].ToInteger()

	return t.vm.ToValue(time1 < time2)
}

// isAfter 判断时间是否在另一个时间之后
func (t *TimeModule) isAfter(call goja.FunctionCall) goja.Value {
	if len(call.Arguments) < 2 {
		panic(t.vm.NewTypeError("isAfter requires two timestamp arguments"))
	}

	time1 := call.Arguments[0].ToInteger()
	time2 := call.Arguments[1].ToInteger()

	return t.vm.ToValue(time1 > time2)
}

// diff 计算时间差（返回秒数）
func (t *TimeModule) diff(call goja.FunctionCall) goja.Value {
	if len(call.Arguments) < 2 {
		panic(t.vm.NewTypeError("diff requires two timestamp arguments"))
	}

	time1 := call.Arguments[0].ToInteger()
	time2 := call.Arguments[1].ToInteger()

	diff := time1 - time2

	result := t.vm.NewObject()
	result.Set("seconds", diff)
	result.Set("minutes", diff/60)
	result.Set("hours", diff/3600)
	result.Set("days", diff/86400)

	return result
}

// utc 转换为 UTC 时区
func (t *TimeModule) utc(call goja.FunctionCall) goja.Value {
	if len(call.Arguments) < 1 {
		panic(t.vm.NewTypeError("utc requires timestamp argument"))
	}

	timestamp := call.Arguments[0].ToInteger()
	timeObj := time.Unix(timestamp, 0).UTC()

	return t.vm.ToValue(timeObj.Format(time.RFC3339))
}

// local 转换为本地时区
func (t *TimeModule) local(call goja.FunctionCall) goja.Value {
	if len(call.Arguments) < 1 {
		panic(t.vm.NewTypeError("local requires timestamp argument"))
	}

	timestamp := call.Arguments[0].ToInteger()
	timeObj := time.Unix(timestamp, 0).Local()

	return t.vm.ToValue(timeObj.Format(time.RFC3339))
}

// inLocation 转换到指定时区
func (t *TimeModule) inLocation(call goja.FunctionCall) goja.Value {
	if len(call.Arguments) < 2 {
		panic(t.vm.NewTypeError("inLocation requires timestamp and location arguments"))
	}

	timestamp := call.Arguments[0].ToInteger()
	locationName := call.Arguments[1].String()

	loc, err := time.LoadLocation(locationName)
	if err != nil {
		panic(t.vm.NewGoError(fmt.Errorf("invalid location: %v", err)))
	}

	timeObj := time.Unix(timestamp, 0).In(loc)
	return t.vm.ToValue(timeObj.Format(time.RFC3339))
}

// getYear 获取年份
func (t *TimeModule) getYear(call goja.FunctionCall) goja.Value {
	if len(call.Arguments) < 1 {
		panic(t.vm.NewTypeError("getYear requires timestamp argument"))
	}

	timestamp := call.Arguments[0].ToInteger()
	timeObj := time.Unix(timestamp, 0)

	return t.vm.ToValue(timeObj.Year())
}

// getMonth 获取月份（1-12）
func (t *TimeModule) getMonth(call goja.FunctionCall) goja.Value {
	if len(call.Arguments) < 1 {
		panic(t.vm.NewTypeError("getMonth requires timestamp argument"))
	}

	timestamp := call.Arguments[0].ToInteger()
	timeObj := time.Unix(timestamp, 0)

	return t.vm.ToValue(int(timeObj.Month()))
}

// getDay 获取日期（1-31）
func (t *TimeModule) getDay(call goja.FunctionCall) goja.Value {
	if len(call.Arguments) < 1 {
		panic(t.vm.NewTypeError("getDay requires timestamp argument"))
	}

	timestamp := call.Arguments[0].ToInteger()
	timeObj := time.Unix(timestamp, 0)

	return t.vm.ToValue(timeObj.Day())
}

// getHour 获取小时（0-23）
func (t *TimeModule) getHour(call goja.FunctionCall) goja.Value {
	if len(call.Arguments) < 1 {
		panic(t.vm.NewTypeError("getHour requires timestamp argument"))
	}

	timestamp := call.Arguments[0].ToInteger()
	timeObj := time.Unix(timestamp, 0)

	return t.vm.ToValue(timeObj.Hour())
}

// getMinute 获取分钟（0-59）
func (t *TimeModule) getMinute(call goja.FunctionCall) goja.Value {
	if len(call.Arguments) < 1 {
		panic(t.vm.NewTypeError("getMinute requires timestamp argument"))
	}

	timestamp := call.Arguments[0].ToInteger()
	timeObj := time.Unix(timestamp, 0)

	return t.vm.ToValue(timeObj.Minute())
}

// getSecond 获取秒数（0-59）
func (t *TimeModule) getSecond(call goja.FunctionCall) goja.Value {
	if len(call.Arguments) < 1 {
		panic(t.vm.NewTypeError("getSecond requires timestamp argument"))
	}

	timestamp := call.Arguments[0].ToInteger()
	timeObj := time.Unix(timestamp, 0)

	return t.vm.ToValue(timeObj.Second())
}

// getWeekday 获取星期几
func (t *TimeModule) getWeekday(call goja.FunctionCall) goja.Value {
	if len(call.Arguments) < 1 {
		panic(t.vm.NewTypeError("getWeekday requires timestamp argument"))
	}

	timestamp := call.Arguments[0].ToInteger()
	timeObj := time.Unix(timestamp, 0)

	result := t.vm.NewObject()
	result.Set("number", int(timeObj.Weekday()))
	result.Set("name", timeObj.Weekday().String())

	return result
}

// create 创建时间
func (t *TimeModule) create(call goja.FunctionCall) goja.Value {
	if len(call.Arguments) < 3 {
		panic(t.vm.NewTypeError("create requires year, month, day arguments"))
	}

	year := int(call.Arguments[0].ToInteger())
	month := int(call.Arguments[1].ToInteger())
	day := int(call.Arguments[2].ToInteger())

	hour := 0
	minute := 0
	second := 0

	if len(call.Arguments) > 3 {
		hour = int(call.Arguments[3].ToInteger())
	}
	if len(call.Arguments) > 4 {
		minute = int(call.Arguments[4].ToInteger())
	}
	if len(call.Arguments) > 5 {
		second = int(call.Arguments[5].ToInteger())
	}

	timeObj := time.Date(year, time.Month(month), day, hour, minute, second, 0, time.Local)
	return t.vm.ToValue(timeObj.Unix())
}

// fromUnix 从 Unix 时间戳创建时间对象
func (t *TimeModule) fromUnix(call goja.FunctionCall) goja.Value {
	if len(call.Arguments) < 1 {
		panic(t.vm.NewTypeError("fromUnix requires timestamp argument"))
	}

	timestamp := call.Arguments[0].ToInteger()
	timeObj := time.Unix(timestamp, 0)

	result := t.vm.NewObject()
	result.Set("unix", timeObj.Unix())
	result.Set("iso", timeObj.Format(time.RFC3339))
	result.Set("year", timeObj.Year())
	result.Set("month", int(timeObj.Month()))
	result.Set("day", timeObj.Day())
	result.Set("hour", timeObj.Hour())
	result.Set("minute", timeObj.Minute())
	result.Set("second", timeObj.Second())
	result.Set("weekday", timeObj.Weekday().String())

	return result
}

// fromUnixMilli 从毫秒时间戳创建时间对象
func (t *TimeModule) fromUnixMilli(call goja.FunctionCall) goja.Value {
	if len(call.Arguments) < 1 {
		panic(t.vm.NewTypeError("fromUnixMilli requires timestamp argument"))
	}

	millis := call.Arguments[0].ToInteger()
	timeObj := time.UnixMilli(millis)

	result := t.vm.NewObject()
	result.Set("unix", timeObj.Unix())
	result.Set("unixMilli", timeObj.UnixMilli())
	result.Set("iso", timeObj.Format(time.RFC3339))
	result.Set("year", timeObj.Year())
	result.Set("month", int(timeObj.Month()))
	result.Set("day", timeObj.Day())
	result.Set("hour", timeObj.Hour())
	result.Set("minute", timeObj.Minute())
	result.Set("second", timeObj.Second())
	result.Set("weekday", timeObj.Weekday().String())

	return result
}

// setInterval 设置定时器（类似 JavaScript 的 setInterval）
func (t *TimeModule) setInterval(call goja.FunctionCall) goja.Value {
	if len(call.Arguments) < 2 {
		panic(t.vm.NewTypeError("setInterval requires callback and interval arguments"))
	}

	callback, ok := goja.AssertFunction(call.Arguments[0])
	if !ok {
		panic(t.vm.NewTypeError("First argument must be a function"))
	}

	interval := call.Arguments[1].ToInteger()
	if interval <= 0 {
		panic(t.vm.NewTypeError("Interval must be greater than 0"))
	}

	t.mutex.Lock()
	t.tickerId++
	tickerId := t.tickerId
	t.mutex.Unlock()

	ticker := &Ticker{
		id:       tickerId,
		ticker:   time.NewTicker(time.Duration(interval) * time.Millisecond),
		callback: callback,
		vm:       t.vm,
	}

	t.mutex.Lock()
	t.tickers[tickerId] = ticker
	t.mutex.Unlock()

	// 启动 ticker
	go func() {
		for {
			ticker.mutex.RLock()
			if ticker.stopped {
				ticker.mutex.RUnlock()
				break
			}
			ticker.mutex.RUnlock()

			select {
			case <-ticker.ticker.C:
				ticker.mutex.RLock()
				if !ticker.stopped {
					_, err := ticker.callback(goja.Undefined())
					if err != nil {
						fmt.Printf("Ticker callback error: %v\n", err)
					}
				}
				ticker.mutex.RUnlock()
			}
		}
	}()

	return t.vm.ToValue(tickerId)
}

// clearInterval 清除定时器
func (t *TimeModule) clearInterval(call goja.FunctionCall) goja.Value {
	if len(call.Arguments) < 1 {
		return goja.Undefined()
	}

	tickerId := int(call.Arguments[0].ToInteger())

	t.mutex.Lock()
	ticker, exists := t.tickers[tickerId]
	if exists {
		delete(t.tickers, tickerId)
	}
	t.mutex.Unlock()

	if exists && ticker != nil {
		ticker.mutex.Lock()
		if !ticker.stopped {
			ticker.stopped = true
			ticker.ticker.Stop()
		}
		ticker.mutex.Unlock()
	}

	return goja.Undefined()
}

// createTicker 创建 Ticker 对象
func (t *TimeModule) createTicker(call goja.FunctionCall) goja.Value {
	if len(call.Arguments) < 1 {
		panic(t.vm.NewTypeError("createTicker requires interval argument"))
	}

	interval := call.Arguments[0].ToInteger()
	if interval <= 0 {
		panic(t.vm.NewTypeError("Interval must be greater than 0"))
	}

	t.mutex.Lock()
	t.tickerId++
	tickerId := t.tickerId
	t.mutex.Unlock()

	ticker := &Ticker{
		id:     tickerId,
		ticker: time.NewTicker(time.Duration(interval) * time.Millisecond),
		vm:     t.vm,
	}

	t.mutex.Lock()
	t.tickers[tickerId] = ticker
	t.mutex.Unlock()

	// 创建 Ticker 对象
	tickerObj := t.vm.NewObject()
	tickerObj.Set("id", tickerId)

	// tick 方法 - 注册回调函数
	tickerObj.Set("tick", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 1 {
			panic(t.vm.NewTypeError("tick requires callback argument"))
		}

		callback, ok := goja.AssertFunction(call.Arguments[0])
		if !ok {
			panic(t.vm.NewTypeError("Argument must be a function"))
		}

		ticker.mutex.Lock()
		ticker.callback = callback
		ticker.mutex.Unlock()

		// 启动 ticker 循环
		go func() {
			for {
				ticker.mutex.RLock()
				if ticker.stopped {
					ticker.mutex.RUnlock()
					break
				}
				ticker.mutex.RUnlock()

				select {
				case <-ticker.ticker.C:
					ticker.mutex.RLock()
					if !ticker.stopped && ticker.callback != nil {
						_, err := ticker.callback(goja.Undefined())
						if err != nil {
							fmt.Printf("Ticker callback error: %v\n", err)
						}
					}
					ticker.mutex.RUnlock()
				}
			}
		}()

		return goja.Undefined()
	})

	// stop 方法 - 停止 ticker
	tickerObj.Set("stop", func(call goja.FunctionCall) goja.Value {
		ticker.mutex.Lock()
		defer ticker.mutex.Unlock()

		if !ticker.stopped {
			ticker.stopped = true
			ticker.ticker.Stop()

			// 从映射中删除
			t.mutex.Lock()
			delete(t.tickers, tickerId)
			t.mutex.Unlock()
		}

		return goja.Undefined()
	})

	// reset 方法 - 重置 ticker
	tickerObj.Set("reset", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 1 {
			panic(t.vm.NewTypeError("reset requires interval argument"))
		}

		newInterval := call.Arguments[0].ToInteger()
		if newInterval <= 0 {
			panic(t.vm.NewTypeError("Interval must be greater than 0"))
		}

		ticker.mutex.Lock()
		defer ticker.mutex.Unlock()

		if !ticker.stopped {
			ticker.ticker.Reset(time.Duration(newInterval) * time.Millisecond)
		}

		return goja.Undefined()
	})

	return tickerObj
}
