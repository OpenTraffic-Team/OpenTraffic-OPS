@echo off
setlocal enabledelayedexpansion

cd /d "%~dp0"

echo =========================================
echo  RTM Monitor Platform - Linux Cross Build
echo =========================================
echo.

echo [1/6] Cleaning old embed directory and previous builds...
if exist "backend\pkg\static\dist" (
    rmdir /s /q "backend\pkg\static\dist"
)
if exist "backend\rtm-monitor-platform-linux-*" (
    del /q "backend\rtm-monitor-platform-linux-*"
)

echo [2/6] Installing frontend dependencies...
cd /d "%~dp0\frontend"

call npm install
if errorlevel 1 (
    echo [ERROR] Frontend npm install failed
    pause
    exit /b 1
)

echo [3/6] Building frontend (production)...
call npm run build:prod
if errorlevel 1 (
    echo [ERROR] Frontend build failed
    pause
    exit /b 1
)

cd /d "%~dp0"

echo [4/6] Copying frontend dist to backend embed directory...
mkdir "backend\pkg\static\dist" 2>nul
xcopy /e /i /q "frontend\dist\*" "backend\pkg\static\dist\"
set "XCOPY_ERR=%errorlevel%"
if %XCOPY_ERR% geq 2 (
    echo [ERROR] Failed to copy dist files ^(xcopy errorlevel=%XCOPY_ERR%^)
    pause
    exit /b 1
)

cd /d "%~dp0\backend"
set GOOS=linux
set CGO_ENABLED=0

echo [5/5] Building Go binaries for multiple architectures...
echo.

set "BUILD_FAILED=0"

:: amd64
echo   --> Building amd64 ...
set GOARCH=amd64
go build -ldflags "-s -w" -o rtm-monitor-platform-linux-amd64 cmd\server\main.go
if errorlevel 1 (
    echo       [FAIL] amd64 build failed
    set "BUILD_FAILED=1"
) else (
    echo       [OK] rtm-monitor-platform-linux-amd64
)

:: arm64
echo   --> Building arm64 ...
set GOARCH=arm64
go build -ldflags "-s -w" -o rtm-monitor-platform-linux-arm64 cmd\server\main.go
if errorlevel 1 (
    echo       [FAIL] arm64 build failed
    set "BUILD_FAILED=1"
) else (
    echo       [OK] rtm-monitor-platform-linux-arm64
)

cd /d "%~dp0"

echo.
echo =========================================
if "%BUILD_FAILED%"=="1" (
    echo  Build completed with ERRORS
    echo =========================================
    pause
    exit /b 1
)

echo  Build success!
echo =========================================
echo.
echo Output binaries in backend\:
for %%f in (backend\rtm-monitor-platform-linux-*) do (
    echo   - %%~nxf
)
echo.
echo Deploy example:
echo   # 先创建配置文件到 ~/.rtm-monitor-platform/config.yaml
echo   mkdir -p ~/.rtm-monitor-platform
echo   cp backend/configs/config.yaml ~/.rtm-monitor-platform/config.yaml
echo   chmod +x rtm-monitor-platform-linux-amd64
echo   ./rtm-monitor-platform-linux-amd64
echo.
pause
