package test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"sw_runtime/internal/runtime"
)

func TestIntegrationBasicApp(t *testing.T) {
	// 创建一个简单的应用程序
	tempDir := t.TempDir()
	appFile := filepath.Join(tempDir, "app.ts")

	appContent := `
		interface Config {
			name: string;
			version: string;
			debug: boolean;
		}
		
		const config: Config = {
			name: "Test App",
			version: "1.0.0",
			debug: true
		};
		
		class Application {
			private config: Config;
			private startTime: number;
			
			constructor(config: Config) {
				this.config = config;
				this.startTime = Date.now();
			}
			
			start(): void {
				console.log('Starting application:', this.config.name);
				console.log('Version:', this.config.version);
				console.log('Debug mode:', this.config.debug);
			}
			
			getUptime(): number {
				return Date.now() - this.startTime;
			}
			
			getInfo(): Config {
				return { ...this.config };
			}
		}
		
		const app = new Application(config);
		app.start();
		
		// 导出给测试使用
		global.testApp = app;
		global.appStarted = true;
	`

	err := os.WriteFile(appFile, []byte(appContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create app file: %v", err)
	}

	runner := runtime.New()
	err = runner.RunFile(appFile)
	if err != nil {
		t.Fatalf("Failed to run integration app: %v", err)
	}

	// 验证应用启动
	started := runner.GetValue("appStarted")
	if !started.ToBoolean() {
		t.Fatal("Application did not start correctly")
	}

	// 验证应用实例
	app := runner.GetValue("testApp")
	if app == nil {
		t.Fatal("Application instance not found")
	}
}

func TestIntegrationAsyncOperations(t *testing.T) {
	runner := runtime.New()

	code := `
		let asyncResults = {
			timeoutCompleted: false,
			intervalCount: 0,
			promiseResolved: false,
			allCompleted: false
		};
		
		// 测试 setTimeout
		setTimeout(() => {
			asyncResults.timeoutCompleted = true;
			console.log('Timeout completed');
			
			// 检查是否所有异步操作都完成
			checkCompletion();
		}, 50);
		
		// 测试 setInterval
		let intervalId = setInterval(() => {
			asyncResults.intervalCount++;
			console.log('Interval count:', asyncResults.intervalCount);
			
			if (asyncResults.intervalCount >= 3) {
				clearInterval(intervalId);
				checkCompletion();
			}
		}, 30);
		
		// 测试 Promise
		new Promise((resolve) => {
			setTimeout(() => {
				resolve('Promise resolved');
			}, 40);
		}).then((result) => {
			asyncResults.promiseResolved = true;
			console.log('Promise result:', result);
			checkCompletion();
		});
		
		function checkCompletion() {
			if (asyncResults.timeoutCompleted && 
				asyncResults.intervalCount >= 3 && 
				asyncResults.promiseResolved) {
				asyncResults.allCompleted = true;
				console.log('All async operations completed');
			}
		}
		
		global.asyncResults = asyncResults;
	`

	err := runner.RunCode(code)
	if err != nil {
		t.Fatalf("Failed to run async integration test: %v", err)
	}

	// 等待所有异步操作完成
	time.Sleep(200 * time.Millisecond)

	results := runner.GetValue("asyncResults")
	if results == nil {
		t.Fatal("Async results not found")
	}
}

func TestIntegrationModuleInteraction(t *testing.T) {
	// 创建多个相互依赖的模块
	tempDir := t.TempDir()

	// 工具模块
	utilsFile := filepath.Join(tempDir, "utils.ts")
	utilsContent := `
		export interface Logger {
			log(message: string): void;
			error(message: string): void;
		}
		
		export class ConsoleLogger implements Logger {
			log(message: string): void {
				console.log('[LOG]', message);
			}
			
			error(message: string): void {
				console.error('[ERROR]', message);
			}
		}
		
		export function formatDate(date: Date): string {
			return date.toISOString().split('T')[0];
		}
		
		export const VERSION = '1.0.0';
	`

	// 服务模块
	serviceFile := filepath.Join(tempDir, "service.ts")
	serviceContent := `
		const utils = require('./utils.ts');
		
		export class DataService {
			private logger: utils.Logger;
			private data: any[] = [];
			
			constructor(logger: utils.Logger) {
				this.logger = logger;
				this.logger.log('DataService initialized');
			}
			
			addData(item: any): void {
				this.data.push({
					...item,
					timestamp: utils.formatDate(new Date()),
					version: utils.VERSION
				});
				this.logger.log('Data added: ' + JSON.stringify(item));
			}
			
			getData(): any[] {
				return [...this.data];
			}
			
			getCount(): number {
				return this.data.length;
			}
		}
	`

	// 主应用模块
	mainFile := filepath.Join(tempDir, "main.ts")
	mainContent := `
		const utils = require('./utils.ts');
		const service = require('./service.ts');
		
		// 创建日志器
		const logger = new utils.ConsoleLogger();
		
		// 创建服务
		const dataService = new service.DataService(logger);
		
		// 添加一些数据
		dataService.addData({ name: 'Item 1', value: 100 });
		dataService.addData({ name: 'Item 2', value: 200 });
		
		// 获取结果
		const data = dataService.getData();
		const count = dataService.getCount();
		
		logger.log('Total items: ' + count);
		logger.log('Data: ' + JSON.stringify(data));
		
		// 导出给测试使用
		global.integrationResults = {
			dataCount: count,
			hasData: data.length > 0,
			firstItem: data[0],
			serviceWorking: true
		};
	`

	// 写入文件
	files := map[string]string{
		utilsFile:   utilsContent,
		serviceFile: serviceContent,
		mainFile:    mainContent,
	}

	for file, content := range files {
		err := os.WriteFile(file, []byte(content), 0644)
		if err != nil {
			t.Fatalf("Failed to create file %s: %v", file, err)
		}
	}

	// 切换到临时目录
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(tempDir)

	runner := runtime.New()
	err := runner.RunFile("main.ts")
	if err != nil {
		t.Fatalf("Failed to run module integration test: %v", err)
	}

	results := runner.GetValue("integrationResults")
	if results == nil {
		t.Fatal("Integration results not found")
	}
}

func TestIntegrationErrorRecovery(t *testing.T) {
	runner := runtime.New()

	code := `
		let errorResults = {
			normalCodeExecuted: false,
			errorCaught: false,
			recoverySuccessful: false
		};
		
		// 正常代码
		errorResults.normalCodeExecuted = true;
		console.log('Normal code executed');
		
		// 尝试执行可能出错的代码
		try {
			// 这会抛出错误
			throw new Error('Test error');
		} catch (e) {
			errorResults.errorCaught = true;
			console.log('Error caught:', e.message);
			
			// 错误恢复
			setTimeout(() => {
				errorResults.recoverySuccessful = true;
				console.log('Recovery completed');
			}, 50);
		}
		
		global.errorResults = errorResults;
	`

	err := runner.RunCode(code)
	if err != nil {
		t.Fatalf("Failed to run error recovery test: %v", err)
	}

	// 等待恢复完成
	time.Sleep(100 * time.Millisecond)

	results := runner.GetValue("errorResults")
	if results == nil {
		t.Fatal("Error recovery results not found")
	}
}

func TestIntegrationComplexApplication(t *testing.T) {
	// 创建一个复杂的应用程序示例
	tempDir := t.TempDir()
	appFile := filepath.Join(tempDir, "complex-app.ts")

	appContent := `
		// 导入内置模块
		const path = require('path');
		const crypto = require('crypto');
		
		interface Task {
			id: string;
			name: string;
			status: 'pending' | 'running' | 'completed' | 'failed';
			createdAt: Date;
			completedAt?: Date;
		}
		
		class TaskManager {
			private tasks: Map<string, Task> = new Map();
			private runningTasks: Set<string> = new Set();
			
			createTask(name: string): string {
				const id = this.generateId();
				const task: Task = {
					id,
					name,
					status: 'pending',
					createdAt: new Date()
				};
				
				this.tasks.set(id, task);
				console.log('Task created:', id, name);
				return id;
			}
			
			async runTask(id: string): Promise<void> {
				const task = this.tasks.get(id);
				if (!task) {
					throw new Error('Task not found: ' + id);
				}
				
				if (this.runningTasks.has(id)) {
					throw new Error('Task already running: ' + id);
				}
				
				task.status = 'running';
				this.runningTasks.add(id);
				console.log('Task started:', id);
				
				return new Promise((resolve, reject) => {
					setTimeout(() => {
						try {
							// 模拟任务执行
							if (Math.random() > 0.2) { // 80% 成功率
								task.status = 'completed';
								task.completedAt = new Date();
								console.log('Task completed:', id);
							} else {
								task.status = 'failed';
								console.log('Task failed:', id);
							}
							
							this.runningTasks.delete(id);
							resolve();
						} catch (error) {
							task.status = 'failed';
							this.runningTasks.delete(id);
							reject(error);
						}
					}, Math.random() * 100 + 50); // 50-150ms 执行时间
				});
			}
			
			getTask(id: string): Task | undefined {
				return this.tasks.get(id);
			}
			
			getAllTasks(): Task[] {
				return Array.from(this.tasks.values());
			}
			
			getTasksByStatus(status: Task['status']): Task[] {
				return this.getAllTasks().filter(task => task.status === status);
			}
			
			private generateId(): string {
				return crypto.md5(Date.now().toString() + Math.random().toString());
			}
		}
		
		// 创建任务管理器
		const taskManager = new TaskManager();
		
		// 创建一些任务
		const taskIds = [
			taskManager.createTask('Process data'),
			taskManager.createTask('Send notifications'),
			taskManager.createTask('Generate report'),
			taskManager.createTask('Cleanup temp files')
		];
		
		// 异步运行所有任务
		Promise.all(taskIds.map(id => taskManager.runTask(id)))
			.then(() => {
				console.log('All tasks completed');
				
				const completedTasks = taskManager.getTasksByStatus('completed');
				const failedTasks = taskManager.getTasksByStatus('failed');
				
				global.complexAppResults = {
					totalTasks: taskIds.length,
					completedCount: completedTasks.length,
					failedCount: failedTasks.length,
					allTasksProcessed: completedTasks.length + failedTasks.length === taskIds.length
				};
			})
			.catch((error) => {
				console.error('Task execution error:', error);
				global.complexAppResults = {
					error: error.message
				};
			});
	`

	err := os.WriteFile(appFile, []byte(appContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create complex app file: %v", err)
	}

	runner := runtime.New()
	err = runner.RunFile(appFile)
	if err != nil {
		t.Fatalf("Failed to run complex application: %v", err)
	}

	// 等待所有异步任务完成
	time.Sleep(500 * time.Millisecond)

	results := runner.GetValue("complexAppResults")
	if results == nil {
		t.Fatal("Complex application results not found")
	}
}
