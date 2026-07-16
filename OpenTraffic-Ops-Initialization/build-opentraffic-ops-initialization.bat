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

echo [4/6] Copying frontend dist, binary packages and opentraffic-control tar packages to backend embed directory...
mkdir "backend\pkg\static\dist" 2>nul
xcopy /e /i /q "frontend\dist\*" "backend\pkg\static\dist\"

for %%a in (amd64 arm64 loong64) do (
    set "SRC_TAR=..\opentraffic-control-linux-%%a.tar"
    set "DST_TAR=backend\pkg\assets\images\opentraffic-control-linux-%%a.tar"
    if exist "!SRC_TAR!" (
        xcopy /y /q "!SRC_TAR!" "backend\pkg\assets\images\"
    ) else if not exist "!DST_TAR!" (
        echo [WARN] opentraffic-control-linux-%%a.tar not found in project root or embed dir, skipping copy
    )
)

:: 拷贝龙芯 Python 环境包与 control 默认配置（如存在）
set "SRC_PY315=..\py315-loong.tar.gz"
set "DST_PY315=backend\pkg\assets\images\py315-loong.tar.gz"
if exist "!SRC_PY315!" (
    xcopy /y /q "!SRC_PY315!" "backend\pkg\assets\images\"
) else if not exist "!DST_PY315!" (
    echo [WARN] py315-loong.tar.gz not found, skipping copy
)

set "SRC_MQ_CONFIG=..\mq_config.json"
set "DST_MQ_CONFIG=backend\pkg\assets\images\mq_config.json"
if exist "!SRC_MQ_CONFIG!" (
    xcopy /y /q "!SRC_MQ_CONFIG!" "backend\pkg\assets\images\"
) else if not exist "!DST_MQ_CONFIG!" (
    echo [WARN] mq_config.json not found, skipping copy
)

for %%a in (amd64 arm64 loong64) do (
    set "SRC_OPS=..\OpenTraffic-Ops\backend\opentraffic-ops-linux-%%a"
    set "DST_OPS=backend\pkg\assets\images\opentraffic-ops-linux-%%a"
    if exist "!SRC_OPS!" (
        xcopy /y /q "!SRC_OPS!" "backend\pkg\assets\images\"
    ) else if not exist "!DST_OPS!" (
        echo [WARN] opentraffic-ops-linux-%%a not found, skipping copy
    )

    set "SRC_PROXY=..\OpenTraffic-Ops\proxy\dist\opentraffic-ops-proxy-linux-%%a"
    set "DST_PROXY=backend\pkg\assets\images\opentraffic-ops-proxy-linux-%%a"
    if exist "!SRC_PROXY!" (
        xcopy /y /q "!SRC_PROXY!" "backend\pkg\assets\images\"
    ) else if not exist "!DST_PROXY!" (
        echo [WARN] opentraffic-ops-proxy-linux-%%a not found, skipping copy
    )
)
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
