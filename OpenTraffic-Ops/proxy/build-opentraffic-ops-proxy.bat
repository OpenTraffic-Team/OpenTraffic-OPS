@echo off
setlocal enabledelayedexpansion

cd /d "%~dp0"

echo =========================================
echo  OpenTraffic Ops Proxy - Linux Cross Build
echo =========================================
echo.

echo [1/4] Checking Go environment...
go version > nul 2>&1
if errorlevel 1 (
    echo [ERROR] Go not found. Please install Go and add it to PATH.
    pause
    exit /b 1
)
for /f "tokens=3" %%a in ('go version') do set "GOVER=%%a"
echo       Go version: %GOVER%
echo.

echo [2/4] Cleaning old builds...
if exist "dist\opentraffic-ops-proxy-linux-*" (
    del /q "dist\opentraffic-ops-proxy-linux-*"
)
if not exist "dist" mkdir dist
echo       [OK] Cleaned
echo.

set GOOS=linux
set CGO_ENABLED=0
set "BUILD_FAILED=0"

set "BUILD_TIME=%date:~0,4%-%date:~5,2%-%date:~8,2%_%time:~0,2%:%time:~3,2%:%time:~6,2%"
set "BUILD_TIME=%BUILD_TIME: =0%"
set "LDFLAGS=-s -w -X main.buildTime=%BUILD_TIME% -X main.goVersion=%GOVER%"

echo [3/4] Building Go binaries for Linux...
echo.

:: amd64
echo   --^> Building amd64 ...
set GOARCH=amd64
go build -ldflags "%LDFLAGS%" -o dist\opentraffic-ops-proxy-linux-amd64 .
if errorlevel 1 (
    echo       [FAIL] amd64 build failed
    set "BUILD_FAILED=1"
) else (
    echo       [OK] opentraffic-ops-proxy-linux-amd64
)

:: arm64
echo   --^> Building arm64 ...
set GOARCH=arm64
go build -ldflags "%LDFLAGS%" -o dist\opentraffic-ops-proxy-linux-arm64 .
if errorlevel 1 (
    echo       [FAIL] arm64 build failed
    set "BUILD_FAILED=1"
) else (
    echo       [OK] opentraffic-ops-proxy-linux-arm64
)

:: loong64 (LoongArch, 龙芯 3A5000+/3C5000+/3D5000+)
echo   --^> Building loong64 ...
set GOARCH=loong64
go build -ldflags "%LDFLAGS%" -o dist\opentraffic-ops-proxy-linux-loong64 .
if errorlevel 1 (
    echo       [FAIL] loong64 build failed
    set "BUILD_FAILED=1"
) else (
    echo       [OK] opentraffic-ops-proxy-linux-loong64
)

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
echo Output binaries in dist\:
for %%f in (dist\opentraffic-ops-proxy-linux-*) do (
    echo   - %%~nxf
)
echo.
echo Deploy example:
echo   # Upload to your Linux server
echo   scp dist\opentraffic-ops-proxy-linux-amd64 user@host:/opt/opentraffic-ops-proxy/
echo   # On server:
echo   chmod +x /opt/opentraffic-ops-proxy/opentraffic-ops-proxy-linux-amd64
echo   ./opentraffic-ops-proxy-linux-amd64 -c config.json
echo.
pause
