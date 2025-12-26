package builtins

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/dop251/goja"
	"github.com/go-redis/redis/v8"
)

// RedisModule Redis 客户端模块
type RedisModule struct {
	vm      *goja.Runtime
	clients map[string]*redis.Client
}

// NewRedisModule 创建 Redis 模块
func NewRedisModule(vm *goja.Runtime) *RedisModule {
	return &RedisModule{
		vm:      vm,
		clients: make(map[string]*redis.Client),
	}
}

// GetModule 获取 Redis 模块对象
func (r *RedisModule) GetModule() *goja.Object {
	obj := r.vm.NewObject()

	// 创建连接
	obj.Set("createClient", r.createClient)
	obj.Set("connect", r.connect)

	return obj
}

// RedisConfig Redis 连接配置
type RedisConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Password string `json:"password"`
	DB       int    `json:"db"`
	Name     string `json:"name"`
}

// parseRedisConfig 解析 Redis 配置
func (r *RedisModule) parseRedisConfig(args []goja.Value) *RedisConfig {
	config := &RedisConfig{
		Host: "localhost",
		Port: 6379,
		DB:   0,
		Name: "default",
	}

	if len(args) > 0 && args[0] != goja.Undefined() {
		configObj := args[0].ToObject(r.vm)
		if configObj != nil {
			if host := configObj.Get("host"); host != nil && host != goja.Undefined() {
				config.Host = host.String()
			}
			if port := configObj.Get("port"); port != nil && port != goja.Undefined() {
				config.Port = int(port.ToInteger())
			}
			if password := configObj.Get("password"); password != nil && password != goja.Undefined() {
				config.Password = password.String()
			}
			if db := configObj.Get("db"); db != nil && db != goja.Undefined() {
				config.DB = int(db.ToInteger())
			}
			if name := configObj.Get("name"); name != nil && name != goja.Undefined() {
				config.Name = name.String()
			}
		}
	}

	return config
}

// createClient 创建 Redis 客户端
func (r *RedisModule) createClient(call goja.FunctionCall) goja.Value {
	config := r.parseRedisConfig(call.Arguments)

	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", config.Host, config.Port),
		Password: config.Password,
		DB:       config.DB,
	})

	// 测试连接（非阻塞）
	promise, resolve, reject := r.vm.NewPromise()

	go func() {
		ctx := context.Background()
		_, err := client.Ping(ctx).Result()
		if err != nil {
			client.Close()
			reject(r.vm.NewGoError(fmt.Errorf("failed to connect to Redis: %w", err)))
		} else {
			// 存储客户端
			r.clients[config.Name] = client
			// 创建客户端对象
			clientObj := r.createClientObject(client)
			resolve(clientObj)
		}
	}()

	return r.vm.ToValue(promise)
}

// connect 连接到 Redis（别名）
func (r *RedisModule) connect(call goja.FunctionCall) goja.Value {
	return r.createClient(call)
}

// createClientObject 创建客户端对象
func (r *RedisModule) createClientObject(client *redis.Client) goja.Value {
	clientObj := r.vm.NewObject()

	// 字符串操作
	clientObj.Set("set", r.createSetMethod(client))
	clientObj.Set("get", r.createGetMethod(client))
	clientObj.Set("del", r.createDelMethod(client))
	clientObj.Set("exists", r.createExistsMethod(client))
	clientObj.Set("expire", r.createExpireMethod(client))
	clientObj.Set("ttl", r.createTTLMethod(client))

	// 哈希操作
	clientObj.Set("hset", r.createHSetMethod(client))
	clientObj.Set("hget", r.createHGetMethod(client))
	clientObj.Set("hgetall", r.createHGetAllMethod(client))
	clientObj.Set("hdel", r.createHDelMethod(client))
	clientObj.Set("hexists", r.createHExistsMethod(client))

	// 列表操作
	clientObj.Set("lpush", r.createLPushMethod(client))
	clientObj.Set("rpush", r.createRPushMethod(client))
	clientObj.Set("lpop", r.createLPopMethod(client))
	clientObj.Set("rpop", r.createRPopMethod(client))
	clientObj.Set("llen", r.createLLenMethod(client))
	clientObj.Set("lrange", r.createLRangeMethod(client))

	// 集合操作
	clientObj.Set("sadd", r.createSAddMethod(client))
	clientObj.Set("srem", r.createSRemMethod(client))
	clientObj.Set("smembers", r.createSMembersMethod(client))
	clientObj.Set("sismember", r.createSIsMemberMethod(client))

	// 有序集合操作
	clientObj.Set("zadd", r.createZAddMethod(client))
	clientObj.Set("zrem", r.createZRemMethod(client))
	clientObj.Set("zrange", r.createZRangeMethod(client))
	clientObj.Set("zrank", r.createZRankMethod(client))
	clientObj.Set("zscore", r.createZScoreMethod(client))

	// 通用操作
	clientObj.Set("ping", r.createPingMethod(client))
	clientObj.Set("flushdb", r.createFlushDBMethod(client))
	clientObj.Set("keys", r.createKeysMethod(client))
	clientObj.Set("info", r.createInfoMethod(client))

	// JSON 操作辅助方法
	clientObj.Set("setJSON", r.createSetJSONMethod(client))
	clientObj.Set("getJSON", r.createGetJSONMethod(client))

	// 关闭连接
	clientObj.Set("quit", r.createQuitMethod(client))

	return clientObj
}

// 字符串操作方法

func (r *RedisModule) createSetMethod(client *redis.Client) func(goja.FunctionCall) goja.Value {
	return func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 2 {
			panic(r.vm.NewTypeError("set requires key and value"))
		}

		key := call.Arguments[0].String()
		value := call.Arguments[1].String()
		expiration := time.Duration(0)

		if len(call.Arguments) > 2 {
			exp := call.Arguments[2].ToInteger()
			expiration = time.Duration(exp) * time.Second
		}

		promise, resolve, reject := r.vm.NewPromise()

		go func() {
			ctx := context.Background()
			err := client.Set(ctx, key, value, expiration).Err()
			if err != nil {
				reject(r.vm.NewGoError(err))
			} else {
				resolve(r.vm.ToValue("OK"))
			}
		}()

		return r.vm.ToValue(promise)
	}
}

func (r *RedisModule) createGetMethod(client *redis.Client) func(goja.FunctionCall) goja.Value {
	return func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 1 {
			panic(r.vm.NewTypeError("get requires key"))
		}

		key := call.Arguments[0].String()
		promise, resolve, reject := r.vm.NewPromise()

		go func() {
			ctx := context.Background()
			val, err := client.Get(ctx, key).Result()
			if err == redis.Nil {
				resolve(goja.Null())
			} else if err != nil {
				reject(r.vm.NewGoError(err))
			} else {
				resolve(r.vm.ToValue(val))
			}
		}()

		return r.vm.ToValue(promise)
	}
}

func (r *RedisModule) createDelMethod(client *redis.Client) func(goja.FunctionCall) goja.Value {
	return func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 1 {
			panic(r.vm.NewTypeError("del requires at least one key"))
		}

		keys := make([]string, len(call.Arguments))
		for i, arg := range call.Arguments {
			keys[i] = arg.String()
		}

		promise, resolve, reject := r.vm.NewPromise()

		go func() {
			ctx := context.Background()
			deleted, err := client.Del(ctx, keys...).Result()
			if err != nil {
				reject(r.vm.NewGoError(err))
			} else {
				resolve(r.vm.ToValue(deleted))
			}
		}()

		return r.vm.ToValue(promise)
	}
}

func (r *RedisModule) createExistsMethod(client *redis.Client) func(goja.FunctionCall) goja.Value {
	return func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 1 {
			panic(r.vm.NewTypeError("exists requires key"))
		}

		key := call.Arguments[0].String()
		promise, resolve, reject := r.vm.NewPromise()

		go func() {
			ctx := context.Background()
			exists, err := client.Exists(ctx, key).Result()
			if err != nil {
				reject(r.vm.NewGoError(err))
			} else {
				resolve(r.vm.ToValue(exists > 0))
			}
		}()

		return r.vm.ToValue(promise)
	}
}

func (r *RedisModule) createExpireMethod(client *redis.Client) func(goja.FunctionCall) goja.Value {
	return func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 2 {
			panic(r.vm.NewTypeError("expire requires key and seconds"))
		}

		key := call.Arguments[0].String()
		seconds := call.Arguments[1].ToInteger()
		promise, resolve, reject := r.vm.NewPromise()

		go func() {
			ctx := context.Background()
			success, err := client.Expire(ctx, key, time.Duration(seconds)*time.Second).Result()
			if err != nil {
				reject(r.vm.NewGoError(err))
			} else {
				resolve(r.vm.ToValue(success))
			}
		}()

		return r.vm.ToValue(promise)
	}
}

func (r *RedisModule) createTTLMethod(client *redis.Client) func(goja.FunctionCall) goja.Value {
	return func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 1 {
			panic(r.vm.NewTypeError("ttl requires key"))
		}

		key := call.Arguments[0].String()
		promise, resolve, reject := r.vm.NewPromise()

		go func() {
			ctx := context.Background()
			ttl, err := client.TTL(ctx, key).Result()
			if err != nil {
				reject(r.vm.NewGoError(err))
			} else {
				resolve(r.vm.ToValue(int64(ttl.Seconds())))
			}
		}()

		return r.vm.ToValue(promise)
	}
}

// 哈希操作方法

func (r *RedisModule) createHSetMethod(client *redis.Client) func(goja.FunctionCall) goja.Value {
	return func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 3 {
			panic(r.vm.NewTypeError("hset requires key, field, and value"))
		}

		key := call.Arguments[0].String()
		field := call.Arguments[1].String()
		value := call.Arguments[2].String()
		promise, resolve, reject := r.vm.NewPromise()

		go func() {
			ctx := context.Background()
			result, err := client.HSet(ctx, key, field, value).Result()
			if err != nil {
				reject(r.vm.NewGoError(err))
			} else {
				resolve(r.vm.ToValue(result))
			}
		}()

		return r.vm.ToValue(promise)
	}
}

func (r *RedisModule) createHGetMethod(client *redis.Client) func(goja.FunctionCall) goja.Value {
	return func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 2 {
			panic(r.vm.NewTypeError("hget requires key and field"))
		}

		key := call.Arguments[0].String()
		field := call.Arguments[1].String()
		promise, resolve, reject := r.vm.NewPromise()

		go func() {
			ctx := context.Background()
			val, err := client.HGet(ctx, key, field).Result()
			if err == redis.Nil {
				resolve(goja.Null())
			} else if err != nil {
				reject(r.vm.NewGoError(err))
			} else {
				resolve(r.vm.ToValue(val))
			}
		}()

		return r.vm.ToValue(promise)
	}
}

func (r *RedisModule) createHGetAllMethod(client *redis.Client) func(goja.FunctionCall) goja.Value {
	return func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 1 {
			panic(r.vm.NewTypeError("hgetall requires key"))
		}

		key := call.Arguments[0].String()
		promise, resolve, reject := r.vm.NewPromise()

		go func() {
			ctx := context.Background()
			result, err := client.HGetAll(ctx, key).Result()
			if err != nil {
				reject(r.vm.NewGoError(err))
			} else {
				resolve(r.vm.ToValue(result))
			}
		}()

		return r.vm.ToValue(promise)
	}
}

func (r *RedisModule) createHDelMethod(client *redis.Client) func(goja.FunctionCall) goja.Value {
	return func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 2 {
			panic(r.vm.NewTypeError("hdel requires key and field"))
		}

		key := call.Arguments[0].String()
		fields := make([]string, len(call.Arguments)-1)
		for i := 1; i < len(call.Arguments); i++ {
			fields[i-1] = call.Arguments[i].String()
		}

		promise, resolve, reject := r.vm.NewPromise()

		go func() {
			ctx := context.Background()
			result, err := client.HDel(ctx, key, fields...).Result()
			if err != nil {
				reject(r.vm.NewGoError(err))
			} else {
				resolve(r.vm.ToValue(result))
			}
		}()

		return r.vm.ToValue(promise)
	}
}

func (r *RedisModule) createHExistsMethod(client *redis.Client) func(goja.FunctionCall) goja.Value {
	return func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 2 {
			panic(r.vm.NewTypeError("hexists requires key and field"))
		}

		key := call.Arguments[0].String()
		field := call.Arguments[1].String()
		promise, resolve, reject := r.vm.NewPromise()

		go func() {
			ctx := context.Background()
			exists, err := client.HExists(ctx, key, field).Result()
			if err != nil {
				reject(r.vm.NewGoError(err))
			} else {
				resolve(r.vm.ToValue(exists))
			}
		}()

		return r.vm.ToValue(promise)
	}
}

// JSON 辅助方法

func (r *RedisModule) createSetJSONMethod(client *redis.Client) func(goja.FunctionCall) goja.Value {
	return func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 2 {
			panic(r.vm.NewTypeError("setJSON requires key and value"))
		}

		key := call.Arguments[0].String()
		value := call.Arguments[1].Export()
		expiration := time.Duration(0)

		if len(call.Arguments) > 2 {
			exp := call.Arguments[2].ToInteger()
			expiration = time.Duration(exp) * time.Second
		}

		promise, resolve, reject := r.vm.NewPromise()

		go func() {
			jsonData, err := json.Marshal(value)
			if err != nil {
				reject(r.vm.NewGoError(err))
				return
			}

			ctx := context.Background()
			err = client.Set(ctx, key, string(jsonData), expiration).Err()
			if err != nil {
				reject(r.vm.NewGoError(err))
			} else {
				resolve(r.vm.ToValue("OK"))
			}
		}()

		return r.vm.ToValue(promise)
	}
}

func (r *RedisModule) createGetJSONMethod(client *redis.Client) func(goja.FunctionCall) goja.Value {
	return func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 1 {
			panic(r.vm.NewTypeError("getJSON requires key"))
		}

		key := call.Arguments[0].String()
		promise, resolve, reject := r.vm.NewPromise()

		go func() {
			ctx := context.Background()
			val, err := client.Get(ctx, key).Result()
			if err == redis.Nil {
				resolve(goja.Null())
			} else if err != nil {
				reject(r.vm.NewGoError(err))
			} else {
				var jsonData interface{}
				if err := json.Unmarshal([]byte(val), &jsonData); err != nil {
					reject(r.vm.NewGoError(err))
				} else {
					resolve(r.vm.ToValue(jsonData))
				}
			}
		}()

		return r.vm.ToValue(promise)
	}
}

// 通用操作方法

func (r *RedisModule) createPingMethod(client *redis.Client) func(goja.FunctionCall) goja.Value {
	return func(call goja.FunctionCall) goja.Value {
		promise, resolve, reject := r.vm.NewPromise()

		go func() {
			ctx := context.Background()
			result, err := client.Ping(ctx).Result()
			if err != nil {
				reject(r.vm.NewGoError(err))
			} else {
				resolve(r.vm.ToValue(result))
			}
		}()

		return r.vm.ToValue(promise)
	}
}

func (r *RedisModule) createKeysMethod(client *redis.Client) func(goja.FunctionCall) goja.Value {
	return func(call goja.FunctionCall) goja.Value {
		pattern := "*"
		if len(call.Arguments) > 0 {
			pattern = call.Arguments[0].String()
		}

		promise, resolve, reject := r.vm.NewPromise()

		go func() {
			ctx := context.Background()
			keys, err := client.Keys(ctx, pattern).Result()
			if err != nil {
				reject(r.vm.NewGoError(err))
			} else {
				resolve(r.vm.ToValue(keys))
			}
		}()

		return r.vm.ToValue(promise)
	}
}

func (r *RedisModule) createFlushDBMethod(client *redis.Client) func(goja.FunctionCall) goja.Value {
	return func(call goja.FunctionCall) goja.Value {
		promise, resolve, reject := r.vm.NewPromise()

		go func() {
			ctx := context.Background()
			result, err := client.FlushDB(ctx).Result()
			if err != nil {
				reject(r.vm.NewGoError(err))
			} else {
				resolve(r.vm.ToValue(result))
			}
		}()

		return r.vm.ToValue(promise)
	}
}

func (r *RedisModule) createInfoMethod(client *redis.Client) func(goja.FunctionCall) goja.Value {
	return func(call goja.FunctionCall) goja.Value {
		promise, resolve, reject := r.vm.NewPromise()

		go func() {
			ctx := context.Background()
			info, err := client.Info(ctx).Result()
			if err != nil {
				reject(r.vm.NewGoError(err))
			} else {
				resolve(r.vm.ToValue(info))
			}
		}()

		return r.vm.ToValue(promise)
	}
}

func (r *RedisModule) createQuitMethod(client *redis.Client) func(goja.FunctionCall) goja.Value {
	return func(call goja.FunctionCall) goja.Value {
		promise, resolve, reject := r.vm.NewPromise()

		go func() {
			err := client.Close()
			if err != nil {
				reject(r.vm.NewGoError(err))
			} else {
				resolve(r.vm.ToValue("OK"))
			}
		}()

		return r.vm.ToValue(promise)
	}
}

// 简化的列表、集合、有序集合操作（这里只实现几个关键方法作为示例）

func (r *RedisModule) createLPushMethod(client *redis.Client) func(goja.FunctionCall) goja.Value {
	return func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 2 {
			panic(r.vm.NewTypeError("lpush requires key and value"))
		}

		key := call.Arguments[0].String()
		values := make([]interface{}, len(call.Arguments)-1)
		for i := 1; i < len(call.Arguments); i++ {
			values[i-1] = call.Arguments[i].String()
		}

		promise, resolve, reject := r.vm.NewPromise()

		go func() {
			ctx := context.Background()
			result, err := client.LPush(ctx, key, values...).Result()
			if err != nil {
				reject(r.vm.NewGoError(err))
			} else {
				resolve(r.vm.ToValue(result))
			}
		}()

		return r.vm.ToValue(promise)
	}
}

func (r *RedisModule) createRPushMethod(client *redis.Client) func(goja.FunctionCall) goja.Value {
	return func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 2 {
			panic(r.vm.NewTypeError("rpush requires key and value"))
		}

		key := call.Arguments[0].String()
		values := make([]interface{}, len(call.Arguments)-1)
		for i := 1; i < len(call.Arguments); i++ {
			values[i-1] = call.Arguments[i].String()
		}

		promise, resolve, reject := r.vm.NewPromise()

		go func() {
			ctx := context.Background()
			result, err := client.RPush(ctx, key, values...).Result()
			if err != nil {
				reject(r.vm.NewGoError(err))
			} else {
				resolve(r.vm.ToValue(result))
			}
		}()

		return r.vm.ToValue(promise)
	}
}

func (r *RedisModule) createLPopMethod(client *redis.Client) func(goja.FunctionCall) goja.Value {
	return func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 1 {
			panic(r.vm.NewTypeError("lpop requires key"))
		}

		key := call.Arguments[0].String()
		promise, resolve, reject := r.vm.NewPromise()

		go func() {
			ctx := context.Background()
			val, err := client.LPop(ctx, key).Result()
			if err == redis.Nil {
				resolve(goja.Null())
			} else if err != nil {
				reject(r.vm.NewGoError(err))
			} else {
				resolve(r.vm.ToValue(val))
			}
		}()

		return r.vm.ToValue(promise)
	}
}

func (r *RedisModule) createRPopMethod(client *redis.Client) func(goja.FunctionCall) goja.Value {
	return func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 1 {
			panic(r.vm.NewTypeError("rpop requires key"))
		}

		key := call.Arguments[0].String()
		promise, resolve, reject := r.vm.NewPromise()

		go func() {
			ctx := context.Background()
			val, err := client.RPop(ctx, key).Result()
			if err == redis.Nil {
				resolve(goja.Null())
			} else if err != nil {
				reject(r.vm.NewGoError(err))
			} else {
				resolve(r.vm.ToValue(val))
			}
		}()

		return r.vm.ToValue(promise)
	}
}

func (r *RedisModule) createLLenMethod(client *redis.Client) func(goja.FunctionCall) goja.Value {
	return func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 1 {
			panic(r.vm.NewTypeError("llen requires key"))
		}

		key := call.Arguments[0].String()
		promise, resolve, reject := r.vm.NewPromise()

		go func() {
			ctx := context.Background()
			length, err := client.LLen(ctx, key).Result()
			if err != nil {
				reject(r.vm.NewGoError(err))
			} else {
				resolve(r.vm.ToValue(length))
			}
		}()

		return r.vm.ToValue(promise)
	}
}

func (r *RedisModule) createLRangeMethod(client *redis.Client) func(goja.FunctionCall) goja.Value {
	return func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 3 {
			panic(r.vm.NewTypeError("lrange requires key, start, and stop"))
		}

		key := call.Arguments[0].String()
		start := call.Arguments[1].ToInteger()
		stop := call.Arguments[2].ToInteger()
		promise, resolve, reject := r.vm.NewPromise()

		go func() {
			ctx := context.Background()
			result, err := client.LRange(ctx, key, start, stop).Result()
			if err != nil {
				reject(r.vm.NewGoError(err))
			} else {
				resolve(r.vm.ToValue(result))
			}
		}()

		return r.vm.ToValue(promise)
	}
}

// 集合操作方法

func (r *RedisModule) createSAddMethod(client *redis.Client) func(goja.FunctionCall) goja.Value {
	return func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 2 {
			panic(r.vm.NewTypeError("sadd requires key and member"))
		}

		key := call.Arguments[0].String()
		members := make([]interface{}, len(call.Arguments)-1)
		for i := 1; i < len(call.Arguments); i++ {
			members[i-1] = call.Arguments[i].String()
		}

		promise, resolve, reject := r.vm.NewPromise()

		go func() {
			ctx := context.Background()
			result, err := client.SAdd(ctx, key, members...).Result()
			if err != nil {
				reject(r.vm.NewGoError(err))
			} else {
				resolve(r.vm.ToValue(result))
			}
		}()

		return r.vm.ToValue(promise)
	}
}

func (r *RedisModule) createSRemMethod(client *redis.Client) func(goja.FunctionCall) goja.Value {
	return func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 2 {
			panic(r.vm.NewTypeError("srem requires key and member"))
		}

		key := call.Arguments[0].String()
		members := make([]interface{}, len(call.Arguments)-1)
		for i := 1; i < len(call.Arguments); i++ {
			members[i-1] = call.Arguments[i].String()
		}

		promise, resolve, reject := r.vm.NewPromise()

		go func() {
			ctx := context.Background()
			result, err := client.SRem(ctx, key, members...).Result()
			if err != nil {
				reject(r.vm.NewGoError(err))
			} else {
				resolve(r.vm.ToValue(result))
			}
		}()

		return r.vm.ToValue(promise)
	}
}

func (r *RedisModule) createSMembersMethod(client *redis.Client) func(goja.FunctionCall) goja.Value {
	return func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 1 {
			panic(r.vm.NewTypeError("smembers requires key"))
		}

		key := call.Arguments[0].String()
		promise, resolve, reject := r.vm.NewPromise()

		go func() {
			ctx := context.Background()
			members, err := client.SMembers(ctx, key).Result()
			if err != nil {
				reject(r.vm.NewGoError(err))
			} else {
				resolve(r.vm.ToValue(members))
			}
		}()

		return r.vm.ToValue(promise)
	}
}

func (r *RedisModule) createSIsMemberMethod(client *redis.Client) func(goja.FunctionCall) goja.Value {
	return func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 2 {
			panic(r.vm.NewTypeError("sismember requires key and member"))
		}

		key := call.Arguments[0].String()
		member := call.Arguments[1].String()
		promise, resolve, reject := r.vm.NewPromise()

		go func() {
			ctx := context.Background()
			isMember, err := client.SIsMember(ctx, key, member).Result()
			if err != nil {
				reject(r.vm.NewGoError(err))
			} else {
				resolve(r.vm.ToValue(isMember))
			}
		}()

		return r.vm.ToValue(promise)
	}
}

// 有序集合操作方法

func (r *RedisModule) createZAddMethod(client *redis.Client) func(goja.FunctionCall) goja.Value {
	return func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 3 {
			panic(r.vm.NewTypeError("zadd requires key, score, and member"))
		}

		key := call.Arguments[0].String()
		score, _ := strconv.ParseFloat(call.Arguments[1].String(), 64)
		member := call.Arguments[2].String()
		promise, resolve, reject := r.vm.NewPromise()

		go func() {
			ctx := context.Background()
			result, err := client.ZAdd(ctx, key, &redis.Z{Score: score, Member: member}).Result()
			if err != nil {
				reject(r.vm.NewGoError(err))
			} else {
				resolve(r.vm.ToValue(result))
			}
		}()

		return r.vm.ToValue(promise)
	}
}

func (r *RedisModule) createZRemMethod(client *redis.Client) func(goja.FunctionCall) goja.Value {
	return func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 2 {
			panic(r.vm.NewTypeError("zrem requires key and member"))
		}

		key := call.Arguments[0].String()
		members := make([]interface{}, len(call.Arguments)-1)
		for i := 1; i < len(call.Arguments); i++ {
			members[i-1] = call.Arguments[i].String()
		}

		promise, resolve, reject := r.vm.NewPromise()

		go func() {
			ctx := context.Background()
			result, err := client.ZRem(ctx, key, members...).Result()
			if err != nil {
				reject(r.vm.NewGoError(err))
			} else {
				resolve(r.vm.ToValue(result))
			}
		}()

		return r.vm.ToValue(promise)
	}
}

func (r *RedisModule) createZRangeMethod(client *redis.Client) func(goja.FunctionCall) goja.Value {
	return func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 3 {
			panic(r.vm.NewTypeError("zrange requires key, start, and stop"))
		}

		key := call.Arguments[0].String()
		start := call.Arguments[1].ToInteger()
		stop := call.Arguments[2].ToInteger()
		promise, resolve, reject := r.vm.NewPromise()

		go func() {
			ctx := context.Background()
			result, err := client.ZRange(ctx, key, start, stop).Result()
			if err != nil {
				reject(r.vm.NewGoError(err))
			} else {
				resolve(r.vm.ToValue(result))
			}
		}()

		return r.vm.ToValue(promise)
	}
}

func (r *RedisModule) createZRankMethod(client *redis.Client) func(goja.FunctionCall) goja.Value {
	return func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 2 {
			panic(r.vm.NewTypeError("zrank requires key and member"))
		}

		key := call.Arguments[0].String()
		member := call.Arguments[1].String()
		promise, resolve, reject := r.vm.NewPromise()

		go func() {
			ctx := context.Background()
			rank, err := client.ZRank(ctx, key, member).Result()
			if err == redis.Nil {
				resolve(goja.Null())
			} else if err != nil {
				reject(r.vm.NewGoError(err))
			} else {
				resolve(r.vm.ToValue(rank))
			}
		}()

		return r.vm.ToValue(promise)
	}
}

func (r *RedisModule) createZScoreMethod(client *redis.Client) func(goja.FunctionCall) goja.Value {
	return func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 2 {
			panic(r.vm.NewTypeError("zscore requires key and member"))
		}

		key := call.Arguments[0].String()
		member := call.Arguments[1].String()
		promise, resolve, reject := r.vm.NewPromise()

		go func() {
			ctx := context.Background()
			score, err := client.ZScore(ctx, key, member).Result()
			if err == redis.Nil {
				resolve(goja.Null())
			} else if err != nil {
				reject(r.vm.NewGoError(err))
			} else {
				resolve(r.vm.ToValue(score))
			}
		}()

		return r.vm.ToValue(promise)
	}
}
