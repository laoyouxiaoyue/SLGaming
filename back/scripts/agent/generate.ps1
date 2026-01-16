# Agent Proto 生成脚本
# 使用方法: 在项目根目录执行 .\back\scripts\agent\generate.ps1

Write-Host "开始生成 Agent proto 文件..." -ForegroundColor Green

# 切换到 proto 文件目录
$protoDir = "back\rpc\agent"
$outputDir = "back\services\agent"

if (-not (Test-Path $protoDir)) {
    Write-Host "错误: 找不到 $protoDir 目录" -ForegroundColor Red
    exit 1
}

# 查找 goctl 路径
$goctlPath = $null

# 方法1: 尝试从 PATH 中查找
$goctlCmd = Get-Command goctl -ErrorAction SilentlyContinue
if ($goctlCmd) {
    $goctlPath = $goctlCmd.Path
    Write-Host "从 PATH 中找到 goctl: $goctlPath" -ForegroundColor Cyan
} else {
    # 方法2: 从 GOPATH 中查找
    $gopath = go env GOPATH
    if ($gopath) {
        $goctlCandidate = Join-Path $gopath "bin\goctl.exe"
        if (Test-Path $goctlCandidate) {
            $goctlPath = $goctlCandidate
            Write-Host "从 GOPATH 中找到 goctl: $goctlPath" -ForegroundColor Cyan
        }
    }
    
    # 方法3: 从用户目录查找
    if (-not $goctlPath) {
        $userGoBin = Join-Path $env:USERPROFILE "go\bin\goctl.exe"
        if (Test-Path $userGoBin) {
            $goctlPath = $userGoBin
            Write-Host "从用户目录找到 goctl: $goctlPath" -ForegroundColor Cyan
        }
    }
}

if (-not $goctlPath) {
    Write-Host "错误: 未找到 goctl 命令" -ForegroundColor Red
    Write-Host "请先安装 goctl: go install github.com/zeromicro/go-zero/tools/goctl@latest" -ForegroundColor Yellow
    Write-Host "安装后，请确保 goctl 在 PATH 中，或重启终端" -ForegroundColor Yellow
    exit 1
}

# 切换到 proto 目录
Push-Location $protoDir

try {
    # 执行生成命令
    Write-Host "执行生成命令..." -ForegroundColor Cyan
    Write-Host "使用 goctl: $goctlPath" -ForegroundColor Gray
    
    $cmdArgs = @(
        "rpc",
        "protoc",
        "agent.proto",
        "--go_out=../../services/agent",
        "--go-grpc_out=../../services/agent",
        "--zrpc_out=../../services/agent"
    )
    
    & $goctlPath $cmdArgs
    
    if ($LASTEXITCODE -eq 0) {
        Write-Host "生成成功！" -ForegroundColor Green
        Write-Host "生成的文件位置: $outputDir\agent\" -ForegroundColor Cyan
    } else {
        Write-Host "生成失败，退出码: $LASTEXITCODE" -ForegroundColor Red
        exit $LASTEXITCODE
    }
} finally {
    Pop-Location
}
