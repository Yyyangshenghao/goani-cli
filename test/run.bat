@echo off
echo === goani-cli 测试套件 ===
echo.

echo [1/2] 运行核心功能测试...
go run test/main.go
if %errorlevel% neq 0 exit /b %errorlevel%
echo.

echo [2/2] 运行播放器测试...
go run test/player.go
if %errorlevel% neq 0 exit /b %errorlevel%
echo.

echo === 全部测试通过 ===
