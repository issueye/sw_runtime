package test

import (
	"testing"

	"sw_runtime/internal/runtime"
)

func BenchmarkRunnerBasicExecution(b *testing.B) {
	code := `
		let result = 0;
		for (let i = 0; i < 1000; i++) {
			result += i;
		}
	`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		runner := runtime.NewOrPanic()
		err := runner.RunCode(code)
		if err != nil {
			b.Fatalf("Failed to run code: %v", err)
		}
	}
}

func BenchmarkRunnerTypeScriptCompilation(b *testing.B) {
	tsCode := `
		interface Calculator {
			add(a: number, b: number): number;
			multiply(a: number, b: number): number;
		}

		class BasicCalculator implements Calculator {
			add(a: number, b: number): number {
				return a + b;
			}

			multiply(a: number, b: number): number {
				return a * b;
			}
		}

		const calc = new BasicCalculator();
		const result = calc.add(calc.multiply(5, 10), 25);
	`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		runner := runtime.NewOrPanic()
		err := runner.RunCode(tsCode)
		if err != nil {
			b.Fatalf("Failed to run TypeScript code: %v", err)
		}
	}
}

func BenchmarkRunnerModuleLoading(b *testing.B) {
	code := `
		const path = require('path');
		const fs = require('fs');
		const crypto = require('crypto');
		
		const result = path.join('test', 'file.txt');
	`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		runner := runtime.NewOrPanic()
		err := runner.RunCode(code)
		if err != nil {
			b.Fatalf("Failed to run module loading: %v", err)
		}
	}
}

func BenchmarkRunnerAsyncOperations(b *testing.B) {
	code := `
		let completed = false;
		setTimeout(() => {
			completed = true;
		}, 1);
	`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		runner := runtime.NewOrPanic()
		err := runner.RunCode(code)
		if err != nil {
			b.Fatalf("Failed to run async operations: %v", err)
		}
	}
}

func BenchmarkRunnerPromiseExecution(b *testing.B) {
	code := `
		let promiseResult = null;
		
		new Promise((resolve) => {
			resolve("benchmark result");
		}).then((result) => {
			promiseResult = result;
		});
	`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		runner := runtime.NewOrPanic()
		err := runner.RunCode(code)
		if err != nil {
			b.Fatalf("Failed to run promise execution: %v", err)
		}
	}
}

func BenchmarkRunnerComplexCalculation(b *testing.B) {
	code := `
		function fibonacci(n) {
			if (n <= 1) return n;
			return fibonacci(n - 1) + fibonacci(n - 2);
		}
		
		function isPrime(n) {
			if (n <= 1) return false;
			if (n <= 3) return true;
			if (n % 2 === 0 || n % 3 === 0) return false;
			
			for (let i = 5; i * i <= n; i += 6) {
				if (n % i === 0 || n % (i + 2) === 0) return false;
			}
			return true;
		}
		
		let results = [];
		for (let i = 1; i <= 20; i++) {
			results.push({
				number: i,
				fibonacci: fibonacci(i),
				isPrime: isPrime(i)
			});
		}
	`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		runner := runtime.NewOrPanic()
		err := runner.RunCode(code)
		if err != nil {
			b.Fatalf("Failed to run complex calculation: %v", err)
		}
	}
}

func BenchmarkRunnerObjectManipulation(b *testing.B) {
	code := `
		const data = [];
		
		for (let i = 0; i < 1000; i++) {
			data.push({
				id: i,
				name: 'Item ' + i,
				value: Math.random() * 100,
				tags: ['tag1', 'tag2', 'tag3'],
				metadata: {
					created: new Date(),
					active: i % 2 === 0
				}
			});
		}
		
		const filtered = data
			.filter(item => item.metadata.active)
			.map(item => ({
				id: item.id,
				name: item.name,
				value: Math.round(item.value)
			}))
			.sort((a, b) => b.value - a.value)
			.slice(0, 10);
	`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		runner := runtime.NewOrPanic()
		err := runner.RunCode(code)
		if err != nil {
			b.Fatalf("Failed to run object manipulation: %v", err)
		}
	}
}

func BenchmarkRunnerStringOperations(b *testing.B) {
	code := `
		const text = "Lorem ipsum dolor sit amet, consectetur adipiscing elit. " +
					"Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua.";
		
		let results = [];
		
		for (let i = 0; i < 100; i++) {
			const processed = text
				.toLowerCase()
				.split(' ')
				.filter(word => word.length > 3)
				.map(word => word.charAt(0).toUpperCase() + word.slice(1))
				.join('-');
			
			results.push(processed);
		}
	`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		runner := runtime.NewOrPanic()
		err := runner.RunCode(code)
		if err != nil {
			b.Fatalf("Failed to run string operations: %v", err)
		}
	}
}

func BenchmarkRunnerMemoryUsage(b *testing.B) {
	code := `
		let largeArray = [];
		for (let i = 0; i < 10000; i++) {
			largeArray.push({
				index: i,
				data: 'x'.repeat(100),
				nested: {
					values: Array.from({length: 10}, (_, j) => i * j)
				}
			});
		}
		
		// 清理
		largeArray = null;
	`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		runner := runtime.NewOrPanic()
		err := runner.RunCode(code)
		if err != nil {
			b.Fatalf("Failed to run memory usage test: %v", err)
		}
	}
}

func BenchmarkRunnerConcurrentOperations(b *testing.B) {
	code := `
		let promises = [];
		
		for (let i = 0; i < 10; i++) {
			promises.push(new Promise((resolve) => {
				setTimeout(() => {
					resolve(i * i);
				}, Math.random() * 10);
			}));
		}
		
		Promise.all(promises).then((results) => {
			const sum = results.reduce((acc, val) => acc + val, 0);
		});
	`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		runner := runtime.NewOrPanic()
		err := runner.RunCode(code)
		if err != nil {
			b.Fatalf("Failed to run concurrent operations: %v", err)
		}
	}
}
