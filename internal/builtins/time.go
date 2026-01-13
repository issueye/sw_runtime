package builtins

import (
	"fmt"
	"strconv"
	"strings"
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

// DayjsTime dayjs 风格的时间对象
type DayjsTime struct {
	vm *goja.Runtime
	t  time.Time
	id int // 内部时间对象ID
}

// dayjs 对象存储
type dayjsStore struct {
	times map[int]*DayjsTime
	mutex sync.RWMutex
	last  int
}

var globalDayjsStore = &dayjsStore{
	times: make(map[int]*DayjsTime),
}

func newDayjsStore() *dayjsStore {
	return globalDayjsStore
}

func (s *dayjsStore) add(d *DayjsTime) int {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.last++
	d.id = s.last
	s.times[s.last] = d
	return s.last
}

func (s *dayjsStore) get(id int) *DayjsTime {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.times[id]
}

func (s *dayjsStore) remove(id int) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	delete(s.times, id)
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

	// dayjs 工厂函数
	obj.Set("dayjs", t.dayjs)
	obj.Set("unix", t.dayjsUnix)

	// 时间解析和格式化
	obj.Set("parse", t.parse)
	obj.Set("format", t.format)
	obj.Set("toISOString", t.toISOString)

	// 延迟执行
	obj.Set("sleep", t.sleep)
	obj.Set("sleepMillis", t.sleepMillis)

	// 时间计算 (支持负数)
	obj.Set("add", t.add)
	obj.Set("subtract", t.subtract)
	obj.Set("addDays", t.addDays)
	obj.Set("addHours", t.addHours)
	obj.Set("addMinutes", t.addMinutes)
	obj.Set("addSeconds", t.addSeconds)
	obj.Set("subtractDays", t.subtractDays)
	obj.Set("subtractHours", t.subtractHours)
	obj.Set("subtractMinutes", t.subtractMinutes)
	obj.Set("subtractSeconds", t.subtractSeconds)

	// 时间边界
	obj.Set("startOf", t.startOf)
	obj.Set("endOf", t.endOf)

	// 时间比较
	obj.Set("isBefore", t.isBefore)
	obj.Set("isAfter", t.isAfter)
	obj.Set("isSame", t.isSame)
	obj.Set("diff", t.diff)

	// 时区处理
	obj.Set("utc", t.utc)
	obj.Set("local", t.local)
	obj.Set("inLocation", t.inLocation)

	// 时间组件获取 (独立函数)
	obj.Set("getYear", t.getYear)
	obj.Set("getMonth", t.getMonth)
	obj.Set("getDay", t.getDay)
	obj.Set("getHour", t.getHour)
	obj.Set("getMinute", t.getMinute)
	obj.Set("getSecond", t.getSecond)
	obj.Set("getWeekday", t.getWeekday)
	obj.Set("getDate", t.getDate)
	obj.Set("getValueOf", t.getValueOf)

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
	// dayjs 风格格式常量
	formats.Set("YYYY", "2006")
	formats.Set("YY", "06")
	formats.Set("M", "1")
	formats.Set("MM", "01")
	formats.Set("MMM", "Jan")
	formats.Set("MMMM", "January")
	formats.Set("D", "2")
	formats.Set("DD", "02")
	formats.Set("dddd", "Monday")
	formats.Set("ddd", "Mon")
	formats.Set("HH", "15")
	formats.Set("H", "15")
	formats.Set("hh", "03")
	formats.Set("h", "3")
	formats.Set("mm", "04")
	formats.Set("m", "4")
	formats.Set("ss", "05")
	formats.Set("s", "5")
	formats.Set("SSS", "000")
	formats.Set("Z", "-07:00")
	formats.Set("ZZ", "-0700")
	obj.Set("FORMAT", formats)

	// 时间单位常量
	units := t.vm.NewObject()
	units.Set("NANOSECOND", int64(time.Nanosecond))
	units.Set("MICROSECOND", int64(time.Microsecond))
	units.Set("MILLISECOND", int64(time.Millisecond))
	units.Set("SECOND", int64(time.Second))
	units.Set("MINUTE", int64(time.Minute))
	units.Set("HOUR", int64(time.Hour))
	// dayjs 风格单位 (字符串)
	units.Set("DATE", "date")
	units.Set("DAY", "day")
	units.Set("WEEK", "week")
	units.Set("MONTH", "month")
	units.Set("YEAR", "year")
	units.Set("HOUR_UNIT", "hour")
	units.Set("MINUTE_UNIT", "minute")
	units.Set("SECOND_UNIT", "second")
	units.Set("MILLISECOND_UNIT", "millisecond")
	obj.Set("UNIT", units)

	return obj
}

// dayjs 创建 dayjs 风格的时间对象
func (t *TimeModule) dayjs(call goja.FunctionCall) goja.Value {
	var parseTime time.Time

	if len(call.Arguments) > 0 {
		timeStr := call.Arguments[0].String()
		layout := time.RFC3339
		parsed := false

		// 自动检测格式
		if strings.Contains(timeStr, "T") {
			if strings.HasSuffix(timeStr, "Z") {
				layout = "2006-01-02T15:04:05Z"
			} else if strings.Contains(timeStr, "+") || (len(timeStr) > 19 && (timeStr[19] == '+' || timeStr[19] == '-')) {
				// 有时区偏移
				layout = time.RFC3339
			} else {
				// 无时区信息，使用无时区格式
				layout = "2006-01-02T15:04:05"
			}
		} else if strings.Contains(timeStr, "-") && len(timeStr) == 10 {
			layout = "2006-01-02"
		} else if strings.Contains(timeStr, ":") && len(timeStr) == 8 {
			layout = "15:04:05"
		}

		if len(call.Arguments) > 1 {
			layout = call.Arguments[1].String()
		}

		var err error
		parseTime, err = time.Parse(layout, timeStr)
		if err == nil {
			parsed = true
		} else if !parsed {
			// 如果第一个格式解析失败，尝试其他常用格式
			layouts := []string{
				"2006-01-02T15:04:05Z07:00",
				"2006-01-02 15:04:05",
				"2006-01-02",
			}
			for _, l := range layouts {
				if l == layout {
					continue
				}
				parseTime, err = time.Parse(l, timeStr)
				if err == nil {
					parsed = true
					break
				}
			}
		}

		if !parsed {
			// 所有格式都失败，使用当前时间
			parseTime = time.Now()
		}
	} else {
		parseTime = time.Now()
	}

	dayjsObj := &DayjsTime{
		vm: t.vm,
		t:  parseTime,
	}

	id := newDayjsStore().add(dayjsObj)

	return createDayjsChain(t.vm, id)
}

// dayjsUnix 从 Unix 时间戳创建 dayjs 对象
func (t *TimeModule) dayjsUnix(call goja.FunctionCall) goja.Value {
	if len(call.Arguments) < 1 {
		panic(t.vm.NewTypeError("unix requires timestamp argument"))
	}

	var parseTime time.Time
	milli := call.Arguments[0].ToInteger()

	if milli > 1e18 {
		// 纳秒时间戳
		parseTime = time.Unix(0, milli)
	} else if milli > 1e15 {
		// 毫秒时间戳
		parseTime = time.UnixMilli(milli)
	} else {
		// 秒时间戳
		parseTime = time.Unix(milli, 0)
	}

	dayjsObj := &DayjsTime{
		vm: t.vm,
		t:  parseTime,
	}

	id := newDayjsStore().add(dayjsObj)

	return createDayjsChain(t.vm, id)
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

// toISOString 转换为 ISO 字符串
func (t *TimeModule) toISOString(call goja.FunctionCall) goja.Value {
	if len(call.Arguments) < 1 {
		panic(t.vm.NewTypeError("toISOString requires timestamp argument"))
	}

	timestamp := call.Arguments[0].ToInteger()
	timeObj := time.Unix(timestamp, 0).UTC()

	return t.vm.ToValue(timeObj.Format("2006-01-02T15:04:05Z"))
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

// subtract 减去时间间隔
func (t *TimeModule) subtract(call goja.FunctionCall) goja.Value {
	if len(call.Arguments) < 2 {
		panic(t.vm.NewTypeError("subtract requires timestamp and duration arguments"))
	}

	timestamp := call.Arguments[0].ToInteger()
	duration := call.Arguments[1].ToInteger()

	timeObj := time.Unix(timestamp, 0)
	newTime := timeObj.Add(-time.Duration(duration))

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

// subtractDays 减去天数
func (t *TimeModule) subtractDays(call goja.FunctionCall) goja.Value {
	if len(call.Arguments) < 2 {
		panic(t.vm.NewTypeError("subtractDays requires timestamp and days arguments"))
	}

	timestamp := call.Arguments[0].ToInteger()
	days := call.Arguments[1].ToInteger()

	timeObj := time.Unix(timestamp, 0)
	newTime := timeObj.AddDate(0, 0, -int(days))

	return t.vm.ToValue(newTime.Unix())
}

// subtractHours 减去小时
func (t *TimeModule) subtractHours(call goja.FunctionCall) goja.Value {
	if len(call.Arguments) < 2 {
		panic(t.vm.NewTypeError("subtractHours requires timestamp and hours arguments"))
	}

	timestamp := call.Arguments[0].ToInteger()
	hours := call.Arguments[1].ToInteger()

	timeObj := time.Unix(timestamp, 0)
	newTime := timeObj.Add(-time.Duration(hours) * time.Hour)

	return t.vm.ToValue(newTime.Unix())
}

// subtractMinutes 减去分钟
func (t *TimeModule) subtractMinutes(call goja.FunctionCall) goja.Value {
	if len(call.Arguments) < 2 {
		panic(t.vm.NewTypeError("subtractMinutes requires timestamp and minutes arguments"))
	}

	timestamp := call.Arguments[0].ToInteger()
	minutes := call.Arguments[1].ToInteger()

	timeObj := time.Unix(timestamp, 0)
	newTime := timeObj.Add(-time.Duration(minutes) * time.Minute)

	return t.vm.ToValue(newTime.Unix())
}

// subtractSeconds 减去秒数
func (t *TimeModule) subtractSeconds(call goja.FunctionCall) goja.Value {
	if len(call.Arguments) < 2 {
		panic(t.vm.NewTypeError("subtractSeconds requires timestamp and seconds arguments"))
	}

	timestamp := call.Arguments[0].ToInteger()
	seconds := call.Arguments[1].ToInteger()

	timeObj := time.Unix(timestamp, 0)
	newTime := timeObj.Add(-time.Duration(seconds) * time.Second)

	return t.vm.ToValue(newTime.Unix())
}

// startOf 获取时间段的开始
func (t *TimeModule) startOf(call goja.FunctionCall) goja.Value {
	if len(call.Arguments) < 2 {
		panic(t.vm.NewTypeError("startOf requires timestamp and unit arguments"))
	}

	timestamp := call.Arguments[0].ToInteger()
	unit := call.Arguments[1].String()

	timeObj := time.Unix(timestamp, 0)

	switch strings.ToLower(unit) {
	case "year", "years":
		timeObj = time.Date(timeObj.Year(), 1, 1, 0, 0, 0, 0, timeObj.Location())
	case "month", "months":
		timeObj = time.Date(timeObj.Year(), timeObj.Month(), 1, 0, 0, 0, 0, timeObj.Location())
	case "week", "weeks":
		weekday := timeObj.Weekday()
		if weekday == time.Sunday {
			weekday = 7
		}
		timeObj = timeObj.AddDate(0, 0, -int(weekday-time.Monday))
		timeObj = time.Date(timeObj.Year(), timeObj.Month(), timeObj.Day(), 0, 0, 0, 0, timeObj.Location())
	case "day", "days":
		timeObj = time.Date(timeObj.Year(), timeObj.Month(), timeObj.Day(), 0, 0, 0, 0, timeObj.Location())
	case "hour", "hours":
		timeObj = time.Date(timeObj.Year(), timeObj.Month(), timeObj.Day(), timeObj.Hour(), 0, 0, 0, timeObj.Location())
	case "minute", "minutes":
		timeObj = time.Date(timeObj.Year(), timeObj.Month(), timeObj.Day(), timeObj.Hour(), timeObj.Minute(), 0, 0, timeObj.Location())
	case "second", "seconds":
		timeObj = time.Date(timeObj.Year(), timeObj.Month(), timeObj.Day(), timeObj.Hour(), timeObj.Minute(), timeObj.Second(), 0, timeObj.Location())
	default:
		panic(t.vm.NewTypeError("Invalid unit: " + unit))
	}

	// 返回带纳秒的时间戳（毫秒精度）
	return t.vm.ToValue(timeObj.UnixMilli())
}

// endOf 获取时间段的结束
func (t *TimeModule) endOf(call goja.FunctionCall) goja.Value {
	if len(call.Arguments) < 2 {
		panic(t.vm.NewTypeError("endOf requires timestamp and unit arguments"))
	}

	timestamp := call.Arguments[0].ToInteger()
	unit := call.Arguments[1].String()

	timeObj := time.Unix(timestamp, 0)

	switch strings.ToLower(unit) {
	case "year", "years":
		timeObj = time.Date(timeObj.Year(), 12, 31, 23, 59, 59, 999999999, timeObj.Location())
	case "month", "months":
		nextMonth := timeObj.AddDate(0, 1, 0)
		timeObj = time.Date(nextMonth.Year(), nextMonth.Month(), 1, 0, 0, 0, -1, timeObj.Location())
	case "week", "weeks":
		weekday := timeObj.Weekday()
		if weekday == time.Sunday {
			weekday = 7
		}
		timeObj = timeObj.AddDate(0, 0, 7-int(weekday))
		timeObj = time.Date(timeObj.Year(), timeObj.Month(), timeObj.Day(), 23, 59, 59, 999999999, timeObj.Location())
	case "day", "days":
		timeObj = time.Date(timeObj.Year(), timeObj.Month(), timeObj.Day(), 23, 59, 59, 999999999, timeObj.Location())
	case "hour", "hours":
		timeObj = time.Date(timeObj.Year(), timeObj.Month(), timeObj.Day(), timeObj.Hour(), 59, 59, 999999999, timeObj.Location())
	case "minute", "minutes":
		timeObj = time.Date(timeObj.Year(), timeObj.Month(), timeObj.Day(), timeObj.Hour(), timeObj.Minute(), 59, 999999999, timeObj.Location())
	case "second", "seconds":
		timeObj = time.Date(timeObj.Year(), timeObj.Month(), timeObj.Day(), timeObj.Hour(), timeObj.Minute(), timeObj.Second(), 999999999, timeObj.Location())
	default:
		panic(t.vm.NewTypeError("Invalid unit: " + unit))
	}

	// 返回带纳秒的时间戳（毫秒精度）
	return t.vm.ToValue(timeObj.UnixMilli())
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

// isSame 判断两个时间是否相同
func (t *TimeModule) isSame(call goja.FunctionCall) goja.Value {
	if len(call.Arguments) < 2 {
		panic(t.vm.NewTypeError("isSame requires two timestamp arguments"))
	}

	time1 := call.Arguments[0].ToInteger()
	time2 := call.Arguments[1].ToInteger()

	return t.vm.ToValue(time1 == time2)
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

// getDate 获取日期（别名）
func (t *TimeModule) getDate(call goja.FunctionCall) goja.Value {
	return t.getDay(call)
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

// getValueOf 获取毫秒时间戳
func (t *TimeModule) getValueOf(call goja.FunctionCall) goja.Value {
	if len(call.Arguments) < 1 {
		panic(t.vm.NewTypeError("getValueOf requires timestamp argument"))
	}

	timestamp := call.Arguments[0].ToInteger()
	timeObj := time.Unix(timestamp, 0)

	return t.vm.ToValue(timeObj.UnixMilli())
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

// createDayjsChain 创建 dayjs 链式对象
func createDayjsChain(vm *goja.Runtime, id int) *goja.Object {
	obj := vm.NewObject()

	// 获取时间对象
	dayjsTime := newDayjsStore().get(id)
	if dayjsTime == nil {
		panic(vm.NewTypeError("Invalid dayjs object ID"))
	}

	// valueOf - 返回毫秒时间戳
	obj.Set("valueOf", func(call goja.FunctionCall) goja.Value {
		return vm.ToValue(dayjsTime.t.UnixMilli())
	})

	// unix - 返回秒时间戳
	obj.Set("unix", func(call goja.FunctionCall) goja.Value {
		return vm.ToValue(dayjsTime.t.Unix())
	})

	// toISOString
	obj.Set("toISOString", func(call goja.FunctionCall) goja.Value {
		return vm.ToValue(dayjsTime.t.Format("2006-01-02T15:04:05Z"))
	})

	// format
	obj.Set("format", func(call goja.FunctionCall) goja.Value {
		layout := "YYYY-MM-DDTHH:mm:ssZ"
		if len(call.Arguments) > 0 {
			layout = call.Arguments[0].String()
		}
		return vm.ToValue(formatDayjs(dayjsTime.t, layout))
	})

	// toString
	obj.Set("toString", func(call goja.FunctionCall) goja.Value {
		return vm.ToValue(dayjsTime.t.Format("2006-01-02 15:04:05"))
	})

	// year
	obj.Set("year", func(call goja.FunctionCall) goja.Value {
		return vm.ToValue(dayjsTime.t.Year())
	})

	// month (0-indexed in dayjs)
	obj.Set("month", func(call goja.FunctionCall) goja.Value {
		return vm.ToValue(int(dayjsTime.t.Month()) - 1)
	})

	// date / day / D
	obj.Set("date", func(call goja.FunctionCall) goja.Value {
		return vm.ToValue(dayjsTime.t.Day())
	})
	obj.Set("day", func(call goja.FunctionCall) goja.Value {
		return vm.ToValue(dayjsTime.t.Weekday())
	})
	obj.Set("D", func(call goja.FunctionCall) goja.Value {
		return vm.ToValue(dayjsTime.t.Day())
	})

	// hour / h / H
	obj.Set("hour", func(call goja.FunctionCall) goja.Value {
		return vm.ToValue(dayjsTime.t.Hour())
	})
	obj.Set("H", func(call goja.FunctionCall) goja.Value {
		return vm.ToValue(dayjsTime.t.Hour())
	})

	// minute / m / mm
	obj.Set("minute", func(call goja.FunctionCall) goja.Value {
		return vm.ToValue(dayjsTime.t.Minute())
	})
	obj.Set("m", func(call goja.FunctionCall) goja.Value {
		return vm.ToValue(dayjsTime.t.Minute())
	})

	// second / s / ss
	obj.Set("second", func(call goja.FunctionCall) goja.Value {
		return vm.ToValue(dayjsTime.t.Second())
	})
	obj.Set("s", func(call goja.FunctionCall) goja.Value {
		return vm.ToValue(dayjsTime.t.Second())
	})

	// millisecond / ms / SSS
	obj.Set("millisecond", func(call goja.FunctionCall) goja.Value {
		return vm.ToValue(dayjsTime.t.Nanosecond() / 1000000)
	})
	obj.Set("ms", func(call goja.FunctionCall) goja.Value {
		return vm.ToValue(dayjsTime.t.Nanosecond() / 1000000)
	})

	// value - 返回毫秒时间戳
	obj.Set("value", func(call goja.FunctionCall) goja.Value {
		return vm.ToValue(dayjsTime.t.UnixMilli())
	})

	// clone
	obj.Set("clone", func(call goja.FunctionCall) goja.Value {
		newTime := &DayjsTime{
			vm: vm,
			t:  dayjsTime.t,
		}
		newId := newDayjsStore().add(newTime)
		return createDayjsChain(vm, newId)
	})

	// add
	obj.Set("add", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 2 {
			panic(vm.NewTypeError("add requires value and unit arguments"))
		}
		value := call.Arguments[0].ToInteger()
		unit := "second"
		if len(call.Arguments) > 1 {
			unit = call.Arguments[1].String()
		}

		newTime := dayjsTime.t
		switch strings.ToLower(unit) {
		case "year", "years":
			newTime = newTime.AddDate(int(value), 0, 0)
		case "month", "months":
			newTime = newTime.AddDate(0, int(value), 0)
		case "week", "weeks":
			newTime = newTime.AddDate(0, 0, int(value*7))
		case "day", "days":
			newTime = newTime.AddDate(0, 0, int(value))
		case "hour", "hours":
			newTime = newTime.Add(time.Duration(value) * time.Hour)
		case "minute", "minutes":
			newTime = newTime.Add(time.Duration(value) * time.Minute)
		case "second", "seconds":
			newTime = newTime.Add(time.Duration(value) * time.Second)
		case "millisecond", "milliseconds":
			newTime = newTime.Add(time.Duration(value) * time.Millisecond)
		}

		newDayjs := &DayjsTime{vm: vm, t: newTime}
		newId := newDayjsStore().add(newDayjs)
		return createDayjsChain(vm, newId)
	})

	// subtract
	obj.Set("subtract", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 2 {
			panic(vm.NewTypeError("subtract requires value and unit arguments"))
		}
		value := call.Arguments[0].ToInteger()
		unit := "second"
		if len(call.Arguments) > 1 {
			unit = call.Arguments[1].String()
		}

		newTime := dayjsTime.t
		switch strings.ToLower(unit) {
		case "year", "years":
			newTime = newTime.AddDate(-int(value), 0, 0)
		case "month", "months":
			newTime = newTime.AddDate(0, -int(value), 0)
		case "week", "weeks":
			newTime = newTime.AddDate(0, 0, -int(value*7))
		case "day", "days":
			newTime = newTime.AddDate(0, 0, -int(value))
		case "hour", "hours":
			newTime = newTime.Add(-time.Duration(value) * time.Hour)
		case "minute", "minutes":
			newTime = newTime.Add(-time.Duration(value) * time.Minute)
		case "second", "seconds":
			newTime = newTime.Add(-time.Duration(value) * time.Second)
		case "millisecond", "milliseconds":
			newTime = newTime.Add(-time.Duration(value) * time.Millisecond)
		}

		newDayjs := &DayjsTime{vm: vm, t: newTime}
		newId := newDayjsStore().add(newDayjs)
		return createDayjsChain(vm, newId)
	})

	// startOf
	obj.Set("startOf", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 1 {
			panic(vm.NewTypeError("startOf requires unit argument"))
		}
		unit := call.Arguments[0].String()

		newTime := dayjsTime.t
		switch strings.ToLower(unit) {
		case "year", "years":
			newTime = time.Date(newTime.Year(), 1, 1, 0, 0, 0, 0, newTime.Location())
		case "month", "months":
			newTime = time.Date(newTime.Year(), newTime.Month(), 1, 0, 0, 0, 0, newTime.Location())
		case "week", "weeks":
			weekday := newTime.Weekday()
			if weekday == time.Sunday {
				weekday = 7
			}
			newTime = newTime.AddDate(0, 0, -int(weekday-time.Monday))
			newTime = time.Date(newTime.Year(), newTime.Month(), newTime.Day(), 0, 0, 0, 0, newTime.Location())
		case "day", "days":
			newTime = time.Date(newTime.Year(), newTime.Month(), newTime.Day(), 0, 0, 0, 0, newTime.Location())
		case "hour", "hours":
			newTime = time.Date(newTime.Year(), newTime.Month(), newTime.Day(), newTime.Hour(), 0, 0, 0, newTime.Location())
		case "minute", "minutes":
			newTime = time.Date(newTime.Year(), newTime.Month(), newTime.Day(), newTime.Hour(), newTime.Minute(), 0, 0, newTime.Location())
		case "second", "seconds":
			newTime = time.Date(newTime.Year(), newTime.Month(), newTime.Day(), newTime.Hour(), newTime.Minute(), newTime.Second(), 0, newTime.Location())
		}

		newDayjs := &DayjsTime{vm: vm, t: newTime}
		newId := newDayjsStore().add(newDayjs)
		return createDayjsChain(vm, newId)
	})

	// endOf
	obj.Set("endOf", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 1 {
			panic(vm.NewTypeError("endOf requires unit argument"))
		}
		unit := call.Arguments[0].String()

		newTime := dayjsTime.t
		switch strings.ToLower(unit) {
		case "year", "years":
			newTime = time.Date(newTime.Year(), 12, 31, 23, 59, 59, 999999999, newTime.Location())
		case "month", "months":
			nextMonth := newTime.AddDate(0, 1, 0)
			newTime = time.Date(nextMonth.Year(), nextMonth.Month(), 1, 0, 0, 0, -1, newTime.Location())
		case "week", "weeks":
			weekday := newTime.Weekday()
			if weekday == time.Sunday {
				weekday = 7
			}
			newTime = newTime.AddDate(0, 0, 7-int(weekday))
			newTime = time.Date(newTime.Year(), newTime.Month(), newTime.Day(), 23, 59, 59, 999999999, newTime.Location())
		case "day", "days":
			newTime = time.Date(newTime.Year(), newTime.Month(), newTime.Day(), 23, 59, 59, 999999999, newTime.Location())
		case "hour", "hours":
			newTime = time.Date(newTime.Year(), newTime.Month(), newTime.Day(), newTime.Hour(), 59, 59, 999999999, newTime.Location())
		case "minute", "minutes":
			newTime = time.Date(newTime.Year(), newTime.Month(), newTime.Day(), newTime.Hour(), newTime.Minute(), 59, 999999999, newTime.Location())
		case "second", "seconds":
			newTime = time.Date(newTime.Year(), newTime.Month(), newTime.Day(), newTime.Hour(), newTime.Minute(), newTime.Second(), 999999999, newTime.Location())
		}

		newDayjs := &DayjsTime{vm: vm, t: newTime}
		newId := newDayjsStore().add(newDayjs)
		return createDayjsChain(vm, newId)
	})

	// isBefore
	obj.Set("isBefore", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 1 {
			panic(vm.NewTypeError("isBefore requires another dayjs argument"))
		}

		// 尝试从参数获取 _id 属性（dayjs 对象）
		var otherId int64 = -1
		if call.Arguments[0].Export() != nil {
			if obj, ok := call.Arguments[0].Export().(map[string]interface{}); ok {
				if id, ok := obj["_id"].(int64); ok {
					otherId = id
				}
			}
		}

		// 如果没找到，尝试作为时间戳
		if otherId < 0 {
			otherId = call.Arguments[0].ToInteger()
			return vm.ToValue(dayjsTime.t.Unix() < otherId)
		}

		otherTime := newDayjsStore().get(int(otherId))
		if otherTime == nil {
			return vm.ToValue(dayjsTime.t.Unix() < otherId)
		}
		return vm.ToValue(dayjsTime.t.Before(otherTime.t))
	})

	// isAfter
	obj.Set("isAfter", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 1 {
			panic(vm.NewTypeError("isAfter requires another dayjs argument"))
		}

		// 尝试从参数获取 _id 属性（dayjs 对象）
		var otherId int64 = -1
		if call.Arguments[0].Export() != nil {
			if obj, ok := call.Arguments[0].Export().(map[string]interface{}); ok {
				if id, ok := obj["_id"].(int64); ok {
					otherId = id
				}
			}
		}

		// 如果没找到，尝试作为时间戳
		if otherId < 0 {
			otherId = call.Arguments[0].ToInteger()
			return vm.ToValue(dayjsTime.t.Unix() > otherId)
		}

		otherTime := newDayjsStore().get(int(otherId))
		if otherTime == nil {
			return vm.ToValue(dayjsTime.t.Unix() > otherId)
		}
		return vm.ToValue(dayjsTime.t.After(otherTime.t))
	})

	// isSame
	obj.Set("isSame", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 1 {
			panic(vm.NewTypeError("isSame requires another dayjs argument"))
		}

		// 尝试从参数获取 _id 属性（dayjs 对象）
		var otherId int64 = -1
		if call.Arguments[0].Export() != nil {
			if obj, ok := call.Arguments[0].Export().(map[string]interface{}); ok {
				if id, ok := obj["_id"].(int64); ok {
					otherId = id
				}
			}
		}

		// 如果没找到，尝试作为时间戳
		if otherId < 0 {
			otherId = call.Arguments[0].ToInteger()
			return vm.ToValue(dayjsTime.t.Unix() == otherId)
		}

		otherTime := newDayjsStore().get(int(otherId))
		if otherTime == nil {
			return vm.ToValue(dayjsTime.t.Unix() == otherId)
		}
		return vm.ToValue(dayjsTime.t.Equal(otherTime.t))
	})

	// isBetween
	obj.Set("isBetween", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 2 {
			panic(vm.NewTypeError("isBetween requires start and end arguments"))
		}

		// 尝试从参数获取 _id 属性（dayjs 对象）
		var startId int64 = -1
		var endId int64 = -1

		if call.Arguments[0].Export() != nil {
			if obj, ok := call.Arguments[0].Export().(map[string]interface{}); ok {
				if id, ok := obj["_id"].(int64); ok {
					startId = id
				}
			}
		}
		if call.Arguments[1].Export() != nil {
			if obj, ok := call.Arguments[1].Export().(map[string]interface{}); ok {
				if id, ok := obj["_id"].(int64); ok {
					endId = id
				}
			}
		}

		startTime := newDayjsStore().get(int(startId))
		endTime := newDayjsStore().get(int(endId))
		if startTime == nil || endTime == nil {
			return vm.ToValue(false)
		}
		return vm.ToValue(dayjsTime.t.After(startTime.t) && dayjsTime.t.Before(endTime.t))
	})

	// diff
	obj.Set("diff", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 1 {
			panic(vm.NewTypeError("diff requires another dayjs argument"))
		}

		// 尝试从参数获取 _id 属性（dayjs 对象）
		var otherId int64 = -1
		if call.Arguments[0].Export() != nil {
			if obj, ok := call.Arguments[0].Export().(map[string]interface{}); ok {
				if id, ok := obj["_id"].(int64); ok {
					otherId = id
				}
			}
		}

		// 如果没找到，尝试作为时间戳
		if otherId < 0 {
			otherId = call.Arguments[0].ToInteger()
		}

		unit := "millisecond"
		if len(call.Arguments) > 1 {
			unit = call.Arguments[1].String()
		}

		otherTime := newDayjsStore().get(int(otherId))
		if otherTime == nil {
			return vm.ToValue(0)
		}

		diff := dayjsTime.t.Sub(otherTime.t)
		switch strings.ToLower(unit) {
		case "year", "years":
			return vm.ToValue(int(diff.Hours()) / (365 * 24))
		case "month", "months":
			return vm.ToValue(int(diff.Hours()) / (30 * 24))
		case "week", "weeks":
			return vm.ToValue(int(diff.Hours()) / (7 * 24))
		case "day", "days":
			return vm.ToValue(int(diff.Hours()) / 24)
		case "hour", "hours":
			return vm.ToValue(diff.Hours())
		case "minute", "minutes":
			return vm.ToValue(diff.Minutes())
		case "second", "seconds":
			return vm.ToValue(diff.Seconds())
		case "millisecond", "milliseconds":
			return vm.ToValue(diff.Milliseconds())
		default:
			return vm.ToValue(diff.Milliseconds())
		}
	})

	// weekday (0=Sunday, 6=Saturday like dayjs)
	obj.Set("weekday", func(call goja.FunctionCall) goja.Value {
		return vm.ToValue(int(dayjsTime.t.Weekday()))
	})

	// dayOfYear
	obj.Set("dayOfYear", func(call goja.FunctionCall) goja.Value {
		yearDay := dayjsTime.t.YearDay()
		return vm.ToValue(yearDay)
	})

	// isLeapYear
	obj.Set("isLeapYear", func(call goja.FunctionCall) goja.Value {
		year := dayjsTime.t.Year()
		return vm.ToValue((year%4 == 0 && year%100 != 0) || year%400 == 0)
	})

	// daysInMonth
	obj.Set("daysInMonth", func(call goja.FunctionCall) goja.Value {
		year := dayjsTime.t.Year()
		month := dayjsTime.t.Month()
		// 获取下个月的第一天，然后减一天就是这个月的最后一天
		nextMonth := time.Date(year, month+1, 1, 0, 0, 0, 0, dayjsTime.t.Location())
		lastDay := nextMonth.AddDate(0, 0, -1)
		return vm.ToValue(lastDay.Day())
	})

	// toJSON
	obj.Set("toJSON", func(call goja.FunctionCall) goja.Value {
		return vm.ToValue(dayjsTime.t.Format(time.RFC3339))
	})

	// JSON serialization helper
	obj.Set("_id", id)
	obj.Set("_time", dayjsTime.t.Unix())

	return obj
}

// formatDayjs 使用 dayjs 格式格式化时间
func formatDayjs(t time.Time, format string) string {
	result := format

	// 按长度从长到短替换，避免短 token 先替换导致问题
	// 4 字符
	result = strings.ReplaceAll(result, "MMMM", t.Month().String())
	result = strings.ReplaceAll(result, "dddd", t.Weekday().String())
	// 3 字符
	result = strings.ReplaceAll(result, "MMM", t.Month().String()[:3])
	result = strings.ReplaceAll(result, "ddd", t.Weekday().String()[:3])
	result = strings.ReplaceAll(result, "ddd", t.Weekday().String()[:3])
	// 2 字符 (先替换长的)
	result = strings.ReplaceAll(result, "YYYY", strconv.Itoa(t.Year()))
	result = strings.ReplaceAll(result, "MM", fmt.Sprintf("%02d", int(t.Month())))
	result = strings.ReplaceAll(result, "DD", fmt.Sprintf("%02d", t.Day()))
	result = strings.ReplaceAll(result, "HH", fmt.Sprintf("%02d", t.Hour()))
	result = strings.ReplaceAll(result, "hh", fmt.Sprintf("%02d", t.Hour()%12))
	result = strings.ReplaceAll(result, "mm", fmt.Sprintf("%02d", t.Minute()))
	result = strings.ReplaceAll(result, "ss", fmt.Sprintf("%02d", t.Second()))
	result = strings.ReplaceAll(result, "SSS", fmt.Sprintf("%03d", t.Nanosecond()/1000000))
	result = strings.ReplaceAll(result, "ZZ", t.Format("-0700"))
	result = strings.ReplaceAll(result, "YY", strconv.Itoa(t.Year())[2:])
	// 1 字符
	result = strings.ReplaceAll(result, "M", strconv.Itoa(int(t.Month())))
	result = strings.ReplaceAll(result, "D", strconv.Itoa(t.Day()))
	result = strings.ReplaceAll(result, "H", strconv.Itoa(t.Hour()))
	result = strings.ReplaceAll(result, "h", strconv.Itoa(t.Hour()%12))
	result = strings.ReplaceAll(result, "m", strconv.Itoa(t.Minute()))
	result = strings.ReplaceAll(result, "s", strconv.Itoa(t.Second()))
	result = strings.ReplaceAll(result, "Z", t.Format("-07:00"))

	// AM/PM (放在最后)
	if strings.Contains(result, "A") {
		if t.Hour() >= 12 {
			result = strings.ReplaceAll(result, "A", "PM")
			result = strings.ReplaceAll(result, "a", "pm")
		} else {
			result = strings.ReplaceAll(result, "A", "AM")
			result = strings.ReplaceAll(result, "a", "am")
		}
	}

	return result
}
