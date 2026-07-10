@echo off
setlocal enabledelayedexpansion

cd /d "%~dp0"

echo =========================================
echo  OpenTraffic Ops Init - Linux Cross Build
echo =========================================
echo.

echo [1/6] Cleaning old embed directory and previous builds...
if exist "backend\pkg\static\dist" (
    rmdir /s /q "backend\pkg\static\dist"
)
if exist "backend\opentraffic-ops-init-linux-*" (
    del /q "backend\opentraffic-ops-init-linux-*"
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
call npm run build
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

echo [5/6] Building Go binaries for multiple architectures...
echo.

set "BUILD_FAILED=0"

:: amd64
echo   --^> Building amd64 ...
set GOARCH=amd64
go build -ldflags "-s -w" -o opentraffic-ops-init-linux-amd64 cmd\server\main.go
if errorlevel 1 (
    echo       [FAIL] amd64 build failed
    set "BUILD_FAILED=1"
) else (
    echo       [OK] opentraffic-ops-init-linux-amd64
)

:: arm64
echo   --^> Building arm64 ...
set GOARCH=arm64
go build -ldflags "-s -w" -o opentraffic-ops-init-linux-arm64 cmd\server\main.go
if errorlevel 1 (
    echo       [FAIL] arm64 build failed
    set "BUILD_FAILED=1"
) else (
    echo       [OK] opentraffic-ops-init-linux-arm64
)

:: loong64 (LoongArch, 龙芯 3A5000+/3C5000+/3D5000+)
echo   --^> Building loong64 ...
set GOARCH=loong64
go build -ldflags "-s -w" -o opentraffic-ops-init-linux-loong64 cmd\server\main.go
if errorlevel 1 (
    echo       [FAIL] loong64 build failed
    set "BUILD_FAILED=1"
) else (
    echo       [OK] opentraffic-ops-init-linux-loong64
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
echo Output files in backend\:
for %%f in (backend\opentraffic-ops-init-linux-*) do (
    echo   - %%~nxf
)
echo.
echo Static files are embedded into each binary.
echo No nginx required - the binary serves frontend directly.
echo.
echo Deploy example:
echo   chmod +x opentraffic-ops-init-linux-amd64
echo   ./opentraffic-ops-init-linux-amd64
echo.
pause
