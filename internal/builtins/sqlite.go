package builtins

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/dop251/goja"
	_ "modernc.org/sqlite"
)

// SQLiteModule SQLite 数据库模块
type SQLiteModule struct {
	vm        *goja.Runtime
	databases map[string]*sql.DB
}

// NewSQLiteModule 创建 SQLite 模块
func NewSQLiteModule(vm *goja.Runtime) *SQLiteModule {
	return &SQLiteModule{
		vm:        vm,
		databases: make(map[string]*sql.DB),
	}
}

// GetModule 获取 SQLite 模块对象
func (s *SQLiteModule) GetModule() *goja.Object {
	obj := s.vm.NewObject()

	// 数据库连接
	obj.Set("open", s.open)
	obj.Set("connect", s.open) // 别名

	// 工具函数
	obj.Set("version", s.version)

	return obj
}

// SQLiteConfig SQLite 连接配置
type SQLiteConfig struct {
	Database string `json:"database"`
	Mode     string `json:"mode"`
	Cache    string `json:"cache"`
	Name     string `json:"name"`
}

// parseSQLiteConfig 解析 SQLite 配置
func (s *SQLiteModule) parseSQLiteConfig(args []goja.Value) *SQLiteConfig {
	config := &SQLiteConfig{
		Database: ":memory:",
		Mode:     "rwc",
		Cache:    "shared",
		Name:     "default",
	}

	if len(args) > 0 && args[0] != goja.Undefined() {
		if args[0].ExportType().Kind().String() == "string" {
			// 如果第一个参数是字符串，作为数据库路径
			config.Database = args[0].String()
		} else {
			// 否则作为配置对象
			configObj := args[0].ToObject(s.vm)
			if configObj != nil {
				if database := configObj.Get("database"); database != nil && database != goja.Undefined() {
					config.Database = database.String()
				}
				if mode := configObj.Get("mode"); mode != nil && mode != goja.Undefined() {
					config.Mode = mode.String()
				}
				if cache := configObj.Get("cache"); cache != nil && cache != goja.Undefined() {
					config.Cache = cache.String()
				}
				if name := configObj.Get("name"); name != nil && name != goja.Undefined() {
					config.Name = name.String()
				}
			}
		}
	}

	return config
}

// open 打开数据库连接
func (s *SQLiteModule) open(call goja.FunctionCall) goja.Value {
	config := s.parseSQLiteConfig(call.Arguments)

	// 构建连接字符串
	var dsn string
	if config.Database == ":memory:" {
		dsn = ":memory:"
	} else {
		// 确保目录存在
		if dir := filepath.Dir(config.Database); dir != "." {
			os.MkdirAll(dir, 0755)
		}

		dsn = config.Database
		if config.Mode != "" || config.Cache != "" {
			params := []string{}
			if config.Mode != "" {
				params = append(params, "mode="+config.Mode)
			}
			if config.Cache != "" {
				params = append(params, "cache="+config.Cache)
			}
			if len(params) > 0 {
				dsn += "?" + strings.Join(params, "&")
			}
		}
	}

	promise, resolve, reject := s.vm.NewPromise()

	go func() {
		db, err := sql.Open("sqlite", dsn)
		if err != nil {
			reject(s.vm.NewGoError(fmt.Errorf("failed to open database: %w", err)))
			return
		}

		// 测试连接
		if err := db.Ping(); err != nil {
			db.Close()
			reject(s.vm.NewGoError(fmt.Errorf("failed to ping database: %w", err)))
			return
		}

		// 存储数据库连接
		s.databases[config.Name] = db

		// 创建数据库对象
		dbObj := s.createDatabaseObject(db, config.Database)
		resolve(dbObj)
	}()

	return s.vm.ToValue(promise)
}

// version 获取 SQLite 版本
func (s *SQLiteModule) version(call goja.FunctionCall) goja.Value {
	promise, resolve, reject := s.vm.NewPromise()

	go func() {
		db, err := sql.Open("sqlite", ":memory:")
		if err != nil {
			reject(s.vm.NewGoError(err))
			return
		}
		defer db.Close()

		var version string
		err = db.QueryRow("SELECT sqlite_version()").Scan(&version)
		if err != nil {
			reject(s.vm.NewGoError(err))
		} else {
			resolve(s.vm.ToValue(version))
		}
	}()

	return s.vm.ToValue(promise)
}

// createDatabaseObject 创建数据库对象
func (s *SQLiteModule) createDatabaseObject(db *sql.DB, dbPath string) goja.Value {
	dbObj := s.vm.NewObject()

	// 基本信息
	dbObj.Set("path", dbPath)

	// 执行方法
	dbObj.Set("exec", s.createExecMethod(db))
	dbObj.Set("run", s.createRunMethod(db))
	dbObj.Set("get", s.createGetMethod(db))
	dbObj.Set("all", s.createAllMethod(db))

	// 事务方法
	dbObj.Set("transaction", s.createTransactionMethod(db))
	dbObj.Set("begin", s.createBeginMethod(db))

	// 预处理语句
	dbObj.Set("prepare", s.createPrepareMethod(db))

	// 数据库信息
	dbObj.Set("tables", s.createTablesMethod(db))
	dbObj.Set("schema", s.createSchemaMethod(db))

	// 关闭连接
	dbObj.Set("close", s.createCloseMethod(db))

	return dbObj
}

// createExecMethod 创建 exec 方法（执行不返回结果的 SQL）
func (s *SQLiteModule) createExecMethod(db *sql.DB) func(goja.FunctionCall) goja.Value {
	return func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 1 {
			panic(s.vm.NewTypeError("exec requires SQL statement"))
		}

		sqlStmt := call.Arguments[0].String()
		var args []interface{}

		// 解析参数
		if len(call.Arguments) > 1 {
			if call.Arguments[1].ExportType().Kind().String() == "slice" {
				// 参数数组
				argsArray := call.Arguments[1].Export().([]interface{})
				args = argsArray
			} else {
				// 单个参数或多个参数
				for i := 1; i < len(call.Arguments); i++ {
					args = append(args, call.Arguments[i].Export())
				}
			}
		}

		promise, resolve, reject := s.vm.NewPromise()

		go func() {
			result, err := db.Exec(sqlStmt, args...)
			if err != nil {
				reject(s.vm.NewGoError(err))
				return
			}

			// 构建结果对象
			resultObj := s.vm.NewObject()

			if lastInsertId, err := result.LastInsertId(); err == nil {
				resultObj.Set("lastInsertId", lastInsertId)
			}

			if rowsAffected, err := result.RowsAffected(); err == nil {
				resultObj.Set("rowsAffected", rowsAffected)
			}

			resolve(resultObj)
		}()

		return s.vm.ToValue(promise)
	}
}

// createRunMethod 创建 run 方法（exec 的别名）
func (s *SQLiteModule) createRunMethod(db *sql.DB) func(goja.FunctionCall) goja.Value {
	return s.createExecMethod(db)
}

// createGetMethod 创建 get 方法（查询单行）
func (s *SQLiteModule) createGetMethod(db *sql.DB) func(goja.FunctionCall) goja.Value {
	return func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 1 {
			panic(s.vm.NewTypeError("get requires SQL statement"))
		}

		sqlStmt := call.Arguments[0].String()
		var args []interface{}

		// 解析参数
		if len(call.Arguments) > 1 {
			if call.Arguments[1].ExportType().Kind().String() == "slice" {
				argsArray := call.Arguments[1].Export().([]interface{})
				args = argsArray
			} else {
				for i := 1; i < len(call.Arguments); i++ {
					args = append(args, call.Arguments[i].Export())
				}
			}
		}

		promise, resolve, reject := s.vm.NewPromise()

		go func() {
			rows, err := db.Query(sqlStmt, args...)
			if err != nil {
				reject(s.vm.NewGoError(err))
				return
			}
			defer rows.Close()

			columns, err := rows.Columns()
			if err != nil {
				reject(s.vm.NewGoError(err))
				return
			}

			if !rows.Next() {
				resolve(goja.Null())
				return
			}

			// 创建扫描目标
			values := make([]interface{}, len(columns))
			valuePtrs := make([]interface{}, len(columns))
			for i := range values {
				valuePtrs[i] = &values[i]
			}

			if err := rows.Scan(valuePtrs...); err != nil {
				reject(s.vm.NewGoError(err))
				return
			}

			// 构建结果对象
			result := make(map[string]interface{})
			for i, col := range columns {
				val := values[i]
				if b, ok := val.([]byte); ok {
					result[col] = string(b)
				} else {
					result[col] = val
				}
			}

			resolve(s.vm.ToValue(result))
		}()

		return s.vm.ToValue(promise)
	}
}

// createAllMethod 创建 all 方法（查询多行）
func (s *SQLiteModule) createAllMethod(db *sql.DB) func(goja.FunctionCall) goja.Value {
	return func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 1 {
			panic(s.vm.NewTypeError("all requires SQL statement"))
		}

		sqlStmt := call.Arguments[0].String()
		var args []interface{}

		// 解析参数
		if len(call.Arguments) > 1 {
			if call.Arguments[1].ExportType().Kind().String() == "slice" {
				argsArray := call.Arguments[1].Export().([]interface{})
				args = argsArray
			} else {
				for i := 1; i < len(call.Arguments); i++ {
					args = append(args, call.Arguments[i].Export())
				}
			}
		}

		promise, resolve, reject := s.vm.NewPromise()

		go func() {
			rows, err := db.Query(sqlStmt, args...)
			if err != nil {
				reject(s.vm.NewGoError(err))
				return
			}
			defer rows.Close()

			columns, err := rows.Columns()
			if err != nil {
				reject(s.vm.NewGoError(err))
				return
			}

			var results []map[string]interface{}

			for rows.Next() {
				// 创建扫描目标
				values := make([]interface{}, len(columns))
				valuePtrs := make([]interface{}, len(columns))
				for i := range values {
					valuePtrs[i] = &values[i]
				}

				if err := rows.Scan(valuePtrs...); err != nil {
					reject(s.vm.NewGoError(err))
					return
				}

				// 构建结果对象
				result := make(map[string]interface{})
				for i, col := range columns {
					val := values[i]
					if b, ok := val.([]byte); ok {
						result[col] = string(b)
					} else {
						result[col] = val
					}
				}
				results = append(results, result)
			}

			if err := rows.Err(); err != nil {
				reject(s.vm.NewGoError(err))
				return
			}

			resolve(s.vm.ToValue(results))
		}()

		return s.vm.ToValue(promise)
	}
}

// createTransactionMethod 创建事务方法
func (s *SQLiteModule) createTransactionMethod(db *sql.DB) func(goja.FunctionCall) goja.Value {
	return func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 1 {
			panic(s.vm.NewTypeError("transaction requires a function"))
		}

		fn, ok := goja.AssertFunction(call.Arguments[0])
		if !ok {
			panic(s.vm.NewTypeError("transaction requires a function"))
		}

		promise, resolve, reject := s.vm.NewPromise()

		go func() {
			tx, err := db.Begin()
			if err != nil {
				reject(s.vm.NewGoError(err))
				return
			}

			// 创建事务对象
			txObj := s.createTransactionObject(tx)

			// 执行事务函数
			result, err := fn(goja.Undefined(), txObj)
			if err != nil {
				tx.Rollback()
				reject(s.vm.NewGoError(err))
				return
			}

			// 如果返回 Promise，等待其完成
			if promise, ok := result.Export().(*goja.Promise); ok {
				switch promise.State() {
				case goja.PromiseStateFulfilled:
					if err := tx.Commit(); err != nil {
						reject(s.vm.NewGoError(err))
					} else {
						resolve(promise.Result())
					}
				case goja.PromiseStateRejected:
					tx.Rollback()
					reject(promise.Result())
				default:
					// Promise 仍在 pending 状态，这里简化处理
					if err := tx.Commit(); err != nil {
						reject(s.vm.NewGoError(err))
					} else {
						resolve(result)
					}
				}
			} else {
				// 非 Promise 结果，直接提交
				if err := tx.Commit(); err != nil {
					reject(s.vm.NewGoError(err))
				} else {
					resolve(result)
				}
			}
		}()

		return s.vm.ToValue(promise)
	}
}

// createBeginMethod 创建开始事务方法
func (s *SQLiteModule) createBeginMethod(db *sql.DB) func(goja.FunctionCall) goja.Value {
	return func(call goja.FunctionCall) goja.Value {
		promise, resolve, reject := s.vm.NewPromise()

		go func() {
			tx, err := db.Begin()
			if err != nil {
				reject(s.vm.NewGoError(err))
				return
			}

			txObj := s.createTransactionObject(tx)
			resolve(txObj)
		}()

		return s.vm.ToValue(promise)
	}
}

// createTransactionObject 创建事务对象
func (s *SQLiteModule) createTransactionObject(tx *sql.Tx) goja.Value {
	txObj := s.vm.NewObject()

	// 执行方法
	txObj.Set("exec", s.createTxExecMethod(tx))
	txObj.Set("run", s.createTxExecMethod(tx))
	txObj.Set("get", s.createTxGetMethod(tx))
	txObj.Set("all", s.createTxAllMethod(tx))

	// 事务控制
	txObj.Set("commit", s.createCommitMethod(tx))
	txObj.Set("rollback", s.createRollbackMethod(tx))

	return txObj
}

// createTxExecMethod 创建事务执行方法
func (s *SQLiteModule) createTxExecMethod(tx *sql.Tx) func(goja.FunctionCall) goja.Value {
	return func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 1 {
			panic(s.vm.NewTypeError("exec requires SQL statement"))
		}

		sqlStmt := call.Arguments[0].String()
		var args []interface{}

		if len(call.Arguments) > 1 {
			if call.Arguments[1].ExportType().Kind().String() == "slice" {
				argsArray := call.Arguments[1].Export().([]interface{})
				args = argsArray
			} else {
				for i := 1; i < len(call.Arguments); i++ {
					args = append(args, call.Arguments[i].Export())
				}
			}
		}

		promise, resolve, reject := s.vm.NewPromise()

		go func() {
			result, err := tx.Exec(sqlStmt, args...)
			if err != nil {
				reject(s.vm.NewGoError(err))
				return
			}

			resultObj := s.vm.NewObject()
			if lastInsertId, err := result.LastInsertId(); err == nil {
				resultObj.Set("lastInsertId", lastInsertId)
			}
			if rowsAffected, err := result.RowsAffected(); err == nil {
				resultObj.Set("rowsAffected", rowsAffected)
			}

			resolve(resultObj)
		}()

		return s.vm.ToValue(promise)
	}
}

// createTxGetMethod 创建事务查询单行方法
func (s *SQLiteModule) createTxGetMethod(tx *sql.Tx) func(goja.FunctionCall) goja.Value {
	return func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 1 {
			panic(s.vm.NewTypeError("get requires SQL statement"))
		}

		sqlStmt := call.Arguments[0].String()
		var args []interface{}

		if len(call.Arguments) > 1 {
			if call.Arguments[1].ExportType().Kind().String() == "slice" {
				argsArray := call.Arguments[1].Export().([]interface{})
				args = argsArray
			} else {
				for i := 1; i < len(call.Arguments); i++ {
					args = append(args, call.Arguments[i].Export())
				}
			}
		}

		promise, resolve, reject := s.vm.NewPromise()

		go func() {
			rows, err := tx.Query(sqlStmt, args...)
			if err != nil {
				reject(s.vm.NewGoError(err))
				return
			}
			defer rows.Close()

			columns, err := rows.Columns()
			if err != nil {
				reject(s.vm.NewGoError(err))
				return
			}

			if !rows.Next() {
				resolve(goja.Null())
				return
			}

			values := make([]interface{}, len(columns))
			valuePtrs := make([]interface{}, len(columns))
			for i := range values {
				valuePtrs[i] = &values[i]
			}

			if err := rows.Scan(valuePtrs...); err != nil {
				reject(s.vm.NewGoError(err))
				return
			}

			result := make(map[string]interface{})
			for i, col := range columns {
				val := values[i]
				if b, ok := val.([]byte); ok {
					result[col] = string(b)
				} else {
					result[col] = val
				}
			}

			resolve(s.vm.ToValue(result))
		}()

		return s.vm.ToValue(promise)
	}
}

// createTxAllMethod 创建事务查询多行方法
func (s *SQLiteModule) createTxAllMethod(tx *sql.Tx) func(goja.FunctionCall) goja.Value {
	return func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 1 {
			panic(s.vm.NewTypeError("all requires SQL statement"))
		}

		sqlStmt := call.Arguments[0].String()
		var args []interface{}

		if len(call.Arguments) > 1 {
			if call.Arguments[1].ExportType().Kind().String() == "slice" {
				argsArray := call.Arguments[1].Export().([]interface{})
				args = argsArray
			} else {
				for i := 1; i < len(call.Arguments); i++ {
					args = append(args, call.Arguments[i].Export())
				}
			}
		}

		promise, resolve, reject := s.vm.NewPromise()

		go func() {
			rows, err := tx.Query(sqlStmt, args...)
			if err != nil {
				reject(s.vm.NewGoError(err))
				return
			}
			defer rows.Close()

			columns, err := rows.Columns()
			if err != nil {
				reject(s.vm.NewGoError(err))
				return
			}

			var results []map[string]interface{}

			for rows.Next() {
				values := make([]interface{}, len(columns))
				valuePtrs := make([]interface{}, len(columns))
				for i := range values {
					valuePtrs[i] = &values[i]
				}

				if err := rows.Scan(valuePtrs...); err != nil {
					reject(s.vm.NewGoError(err))
					return
				}

				result := make(map[string]interface{})
				for i, col := range columns {
					val := values[i]
					if b, ok := val.([]byte); ok {
						result[col] = string(b)
					} else {
						result[col] = val
					}
				}
				results = append(results, result)
			}

			if err := rows.Err(); err != nil {
				reject(s.vm.NewGoError(err))
				return
			}

			resolve(s.vm.ToValue(results))
		}()

		return s.vm.ToValue(promise)
	}
}

// createCommitMethod 创建提交方法
func (s *SQLiteModule) createCommitMethod(tx *sql.Tx) func(goja.FunctionCall) goja.Value {
	return func(call goja.FunctionCall) goja.Value {
		promise, resolve, reject := s.vm.NewPromise()

		go func() {
			if err := tx.Commit(); err != nil {
				reject(s.vm.NewGoError(err))
			} else {
				resolve(s.vm.ToValue("OK"))
			}
		}()

		return s.vm.ToValue(promise)
	}
}

// createRollbackMethod 创建回滚方法
func (s *SQLiteModule) createRollbackMethod(tx *sql.Tx) func(goja.FunctionCall) goja.Value {
	return func(call goja.FunctionCall) goja.Value {
		promise, resolve, reject := s.vm.NewPromise()

		go func() {
			if err := tx.Rollback(); err != nil {
				reject(s.vm.NewGoError(err))
			} else {
				resolve(s.vm.ToValue("OK"))
			}
		}()

		return s.vm.ToValue(promise)
	}
}

// createPrepareMethod 创建预处理语句方法
func (s *SQLiteModule) createPrepareMethod(db *sql.DB) func(goja.FunctionCall) goja.Value {
	return func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 1 {
			panic(s.vm.NewTypeError("prepare requires SQL statement"))
		}

		sqlStmt := call.Arguments[0].String()
		promise, resolve, reject := s.vm.NewPromise()

		go func() {
			stmt, err := db.Prepare(sqlStmt)
			if err != nil {
				reject(s.vm.NewGoError(err))
				return
			}

			stmtObj := s.createStatementObject(stmt)
			resolve(stmtObj)
		}()

		return s.vm.ToValue(promise)
	}
}

// createStatementObject 创建预处理语句对象
func (s *SQLiteModule) createStatementObject(stmt *sql.Stmt) goja.Value {
	stmtObj := s.vm.NewObject()

	stmtObj.Set("exec", s.createStmtExecMethod(stmt))
	stmtObj.Set("run", s.createStmtExecMethod(stmt))
	stmtObj.Set("get", s.createStmtGetMethod(stmt))
	stmtObj.Set("all", s.createStmtAllMethod(stmt))
	stmtObj.Set("close", s.createStmtCloseMethod(stmt))

	return stmtObj
}

// createStmtExecMethod 创建预处理语句执行方法
func (s *SQLiteModule) createStmtExecMethod(stmt *sql.Stmt) func(goja.FunctionCall) goja.Value {
	return func(call goja.FunctionCall) goja.Value {
		var args []interface{}
		for _, arg := range call.Arguments {
			args = append(args, arg.Export())
		}

		promise, resolve, reject := s.vm.NewPromise()

		go func() {
			result, err := stmt.Exec(args...)
			if err != nil {
				reject(s.vm.NewGoError(err))
				return
			}

			resultObj := s.vm.NewObject()
			if lastInsertId, err := result.LastInsertId(); err == nil {
				resultObj.Set("lastInsertId", lastInsertId)
			}
			if rowsAffected, err := result.RowsAffected(); err == nil {
				resultObj.Set("rowsAffected", rowsAffected)
			}

			resolve(resultObj)
		}()

		return s.vm.ToValue(promise)
	}
}

// createStmtGetMethod 创建预处理语句查询单行方法
func (s *SQLiteModule) createStmtGetMethod(stmt *sql.Stmt) func(goja.FunctionCall) goja.Value {
	return func(call goja.FunctionCall) goja.Value {
		var args []interface{}
		for _, arg := range call.Arguments {
			args = append(args, arg.Export())
		}

		promise, resolve, reject := s.vm.NewPromise()

		go func() {
			rows, err := stmt.Query(args...)
			if err != nil {
				reject(s.vm.NewGoError(err))
				return
			}
			defer rows.Close()

			columns, err := rows.Columns()
			if err != nil {
				reject(s.vm.NewGoError(err))
				return
			}

			if !rows.Next() {
				resolve(goja.Null())
				return
			}

			values := make([]interface{}, len(columns))
			valuePtrs := make([]interface{}, len(columns))
			for i := range values {
				valuePtrs[i] = &values[i]
			}

			if err := rows.Scan(valuePtrs...); err != nil {
				reject(s.vm.NewGoError(err))
				return
			}

			result := make(map[string]interface{})
			for i, col := range columns {
				val := values[i]
				if b, ok := val.([]byte); ok {
					result[col] = string(b)
				} else {
					result[col] = val
				}
			}

			resolve(s.vm.ToValue(result))
		}()

		return s.vm.ToValue(promise)
	}
}

// createStmtAllMethod 创建预处理语句查询多行方法
func (s *SQLiteModule) createStmtAllMethod(stmt *sql.Stmt) func(goja.FunctionCall) goja.Value {
	return func(call goja.FunctionCall) goja.Value {
		var args []interface{}
		for _, arg := range call.Arguments {
			args = append(args, arg.Export())
		}

		promise, resolve, reject := s.vm.NewPromise()

		go func() {
			rows, err := stmt.Query(args...)
			if err != nil {
				reject(s.vm.NewGoError(err))
				return
			}
			defer rows.Close()

			columns, err := rows.Columns()
			if err != nil {
				reject(s.vm.NewGoError(err))
				return
			}

			var results []map[string]interface{}

			for rows.Next() {
				values := make([]interface{}, len(columns))
				valuePtrs := make([]interface{}, len(columns))
				for i := range values {
					valuePtrs[i] = &values[i]
				}

				if err := rows.Scan(valuePtrs...); err != nil {
					reject(s.vm.NewGoError(err))
					return
				}

				result := make(map[string]interface{})
				for i, col := range columns {
					val := values[i]
					if b, ok := val.([]byte); ok {
						result[col] = string(b)
					} else {
						result[col] = val
					}
				}
				results = append(results, result)
			}

			if err := rows.Err(); err != nil {
				reject(s.vm.NewGoError(err))
				return
			}

			resolve(s.vm.ToValue(results))
		}()

		return s.vm.ToValue(promise)
	}
}

// createStmtCloseMethod 创建预处理语句关闭方法
func (s *SQLiteModule) createStmtCloseMethod(stmt *sql.Stmt) func(goja.FunctionCall) goja.Value {
	return func(call goja.FunctionCall) goja.Value {
		promise, resolve, reject := s.vm.NewPromise()

		go func() {
			if err := stmt.Close(); err != nil {
				reject(s.vm.NewGoError(err))
			} else {
				resolve(s.vm.ToValue("OK"))
			}
		}()

		return s.vm.ToValue(promise)
	}
}

// createTablesMethod 创建获取表列表方法
func (s *SQLiteModule) createTablesMethod(db *sql.DB) func(goja.FunctionCall) goja.Value {
	return func(call goja.FunctionCall) goja.Value {
		promise, resolve, reject := s.vm.NewPromise()

		go func() {
			rows, err := db.Query("SELECT name FROM sqlite_master WHERE type='table' ORDER BY name")
			if err != nil {
				reject(s.vm.NewGoError(err))
				return
			}
			defer rows.Close()

			var tables []string
			for rows.Next() {
				var tableName string
				if err := rows.Scan(&tableName); err != nil {
					reject(s.vm.NewGoError(err))
					return
				}
				tables = append(tables, tableName)
			}

			if err := rows.Err(); err != nil {
				reject(s.vm.NewGoError(err))
				return
			}

			resolve(s.vm.ToValue(tables))
		}()

		return s.vm.ToValue(promise)
	}
}

// createSchemaMethod 创建获取表结构方法
func (s *SQLiteModule) createSchemaMethod(db *sql.DB) func(goja.FunctionCall) goja.Value {
	return func(call goja.FunctionCall) goja.Value {
		var tableName string
		if len(call.Arguments) > 0 {
			tableName = call.Arguments[0].String()
		}

		promise, resolve, reject := s.vm.NewPromise()

		go func() {
			var query string
			if tableName != "" {
				query = "SELECT sql FROM sqlite_master WHERE type='table' AND name=?"
			} else {
				query = "SELECT name, sql FROM sqlite_master WHERE type='table' ORDER BY name"
			}

			var rows *sql.Rows
			var err error

			if tableName != "" {
				rows, err = db.Query(query, tableName)
			} else {
				rows, err = db.Query(query)
			}

			if err != nil {
				reject(s.vm.NewGoError(err))
				return
			}
			defer rows.Close()

			if tableName != "" {
				// 单个表的结构
				if rows.Next() {
					var sql string
					if err := rows.Scan(&sql); err != nil {
						reject(s.vm.NewGoError(err))
						return
					}
					resolve(s.vm.ToValue(sql))
				} else {
					resolve(goja.Null())
				}
			} else {
				// 所有表的结构
				schemas := make(map[string]interface{})
				for rows.Next() {
					var name, sql string
					if err := rows.Scan(&name, &sql); err != nil {
						reject(s.vm.NewGoError(err))
						return
					}
					schemas[name] = sql
				}
				resolve(s.vm.ToValue(schemas))
			}

			if err := rows.Err(); err != nil {
				reject(s.vm.NewGoError(err))
				return
			}
		}()

		return s.vm.ToValue(promise)
	}
}

// createCloseMethod 创建关闭数据库方法
func (s *SQLiteModule) createCloseMethod(db *sql.DB) func(goja.FunctionCall) goja.Value {
	return func(call goja.FunctionCall) goja.Value {
		promise, resolve, reject := s.vm.NewPromise()

		go func() {
			if err := db.Close(); err != nil {
				reject(s.vm.NewGoError(err))
			} else {
				resolve(s.vm.ToValue("OK"))
			}
		}()

		return s.vm.ToValue(promise)
	}
}
