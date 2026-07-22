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

echo [4/6] Copying frontend dist, binary packages, opentraffic-control tar packages and opentraffic-perception tar package to backend embed directory...
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

:: 拷贝龙芯 Python 环境包（如存在）
set "SRC_ENV_LOONG=..\trafficlight-loong64.tar.gz"
set "DST_ENV_LOONG=backend\pkg\assets\images\trafficlight-loong64.tar.gz"
if exist "!SRC_ENV_LOONG!" (
    xcopy /y /q "!SRC_ENV_LOONG!" "backend\pkg\assets\images\"
) else if not exist "!DST_ENV_LOONG!" (
    echo [WARN] trafficlight-loong64.tar.gz not found, skipping copy
)

:: 拷贝 ARM Python 环境包（如存在）
set "SRC_ENV_ARM=..\trafficlight-arm64.tar.gz"
set "DST_ENV_ARM=backend\pkg\assets\images\trafficlight-arm64.tar.gz"
if exist "!SRC_ENV_ARM!" (
    xcopy /y /q "!SRC_ENV_ARM!" "backend\pkg\assets\images\"
) else if not exist "!DST_ENV_ARM!" (
    echo [WARN] trafficlight-arm64.tar.gz not found, skipping copy
)

:: 拷贝 x86/amd64 Python 环境包（如存在）
set "SRC_ENV_AMD64=..\trafficlight-amd64.tar.gz"
set "DST_ENV_AMD64=backend\pkg\assets\images\trafficlight-amd64.tar.gz"
if exist "!SRC_ENV_AMD64!" (
    xcopy /y /q "!SRC_ENV_AMD64!" "backend\pkg\assets\images\"
) else if not exist "!DST_ENV_AMD64!" (
    echo [WARN] trafficlight-amd64.tar.gz not found, skipping copy
)

:: 拷贝 opentraffic-perception x86/amd64 算法包与默认配置（如存在）
set "SRC_PERCEPTION=..\opentraffic-perception-linux-amd64.tar"
set "DST_PERCEPTION=backend\pkg\assets\images\opentraffic-perception-linux-amd64.tar"
if exist "!SRC_PERCEPTION!" (
    xcopy /y /q "!SRC_PERCEPTION!" "backend\pkg\assets\images\"
) else if not exist "!DST_PERCEPTION!" (
    echo [WARN] opentraffic-perception-linux-amd64.tar not found, skipping copy
)

:: 拷贝 opentraffic-perception ARM aarch64 算法包（如存在）
set "SRC_PERCEPTION_ARM=..\opentraffic-perception-linux-arm64.tar"
set "DST_PERCEPTION_ARM=backend\pkg\assets\images\opentraffic-perception-linux-arm64.tar"
if exist "!SRC_PERCEPTION_ARM!" (
    xcopy /y /q "!SRC_PERCEPTION_ARM!" "backend\pkg\assets\images\"
) else if not exist "!DST_PERCEPTION_ARM!" (
    echo [WARN] opentraffic-perception-linux-arm64.tar not found, skipping copy
)

set "SRC_PERCEPTION_CONFIG=..\opentraffic-perception-config.json"
set "DST_PERCEPTION_CONFIG=backend\pkg\assets\images\opentraffic-perception-config.json"
if exist "!SRC_PERCEPTION_CONFIG!" (
    xcopy /y /q "!SRC_PERCEPTION_CONFIG!" "backend\pkg\assets\images\"
) else if not exist "!DST_PERCEPTION_CONFIG!" (
    echo [WARN] opentraffic-perception-config.json not found, skipping copy
)

:: 拷贝 opentraffic-control 默认配置（如存在）。远程文件仍为 config/mq_config.json，嵌入目录使用带服务前缀的名称避免混淆
set "SRC_MQ_CONFIG=..\mq_config.json"
set "DST_MQ_CONFIG=backend\pkg\assets\images\opentraffic-control-config.json"
if exist "!SRC_MQ_CONFIG!" (
    copy /y "!SRC_MQ_CONFIG!" "!DST_MQ_CONFIG!" >nul
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
