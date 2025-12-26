@echo off
echo Running SW Runtime Test Suite...
echo.

echo ================================
echo Running Unit Tests
echo ================================
go test ./test/ -v

echo.
echo ================================
echo Running Benchmark Tests
echo ================================
go test -bench=BenchmarkRunner ./test/benchmark_test.go -benchtime=1s

echo.
echo ================================
echo Generating Coverage Report
echo ================================
go test -coverprofile=coverage.out ./test/
go tool cover -html=coverage.out -o test/coverage.html

echo.
echo ================================
echo Test Reports Generated
echo ================================
echo - Markdown Report: test/TEST_REPORT.md
echo - JSON Report: test/test_results.json
echo - HTML Report: test/test_report.html
echo - Coverage Report: test/coverage.html

echo.
echo Test suite completed!
echo Open test/test_report.html in your browser to view the detailed report.
pause