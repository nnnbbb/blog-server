# ============================================
# deploy.ps1
# 一键编译并部署 Go 项目到 Linux 服务器
# ============================================

# === 基本配置 ===
$GOOS = "linux"
$GOARCH = "amd64"
$BINARY_NAME = "blog-server"           # 可执行文件名
$MAIN_FILE = "main.go"              # 程序入口
$REMOTE_USER = "root"               # SSH 用户
$REMOTE_PATH = "/data/blog-server"  # 部署路径
$PRIVATE_KEY = "~/.ssh/id_rsa"      # SSH 私钥路径

# === Step 0: 检查 SERVER_LIST ===
if (-not $Global:SERVER_LIST -or $Global:SERVER_LIST.Count -eq 0) {
    Write-Host "❌ 未检测到全局变量 `SERVER_LIST`" -ForegroundColor Red"
    Write-Host ""
    Write-Host "请先在 PowerShell 全局配置文件中添加，例如：" -ForegroundColor Yellow
    Write-Host 'notepad $PROFILE'
    Write-Host ""
    Write-Host "示例内容：" -ForegroundColor Cyan
    Write-Host '$Global:SERVER_LIST = @('
    Write-Host '    @{ Name = "生产服务器"; IP = "8.8.8.8" },'
    Write-Host '    @{ Name = "测试服务器"; IP = "127.0.0.1" }'
    Write-Host ')' -ForegroundColor Cyan
    exit 1
}

# === Step 1: 选择目标服务器 ===
Write-Host "可部署的服务器列表：" -ForegroundColor Cyan
for ($i = 0; $i -lt $Global:SERVER_LIST.Count; $i++) {
    $server = $Global:SERVER_LIST[$i]
    Write-Host "[$i] $($server.Name) ($($server.IP))"
}

$choice = Read-Host "请输入要部署的编号 (默认 0)"
if ([string]::IsNullOrWhiteSpace($choice)) {
    $choice = 0
}

if ($choice -notmatch '^\d+$' -or [int]$choice -ge $Global:SERVER_LIST.Count) {
    Write-Host "❌ 无效的选择" -ForegroundColor Red
    exit 1
}

$server = $Global:SERVER_LIST[$choice]
$REMOTE_HOST = $server.IP
Write-Host "🚀 即将部署到：$($server.Name) ($REMOTE_HOST)" -ForegroundColor Green

# === Step 2: 编译 Linux 二进制 ===
Write-Host "==> 编译 Go 程序为 Linux 二进制文件..." -ForegroundColor Cyan
$env:GOOS = $GOOS
$env:GOARCH = $GOARCH

go build -trimpath -o $BINARY_NAME $MAIN_FILE
if ($LASTEXITCODE -ne 0) {
    Write-Host "❌ 编译失败" -ForegroundColor Red
    exit 1
}
Write-Host "✅ 编译成功: $BINARY_NAME" -ForegroundColor Green

# === Step 3: 删除服务器上旧文件 ===
Write-Host "==> 删除远程旧文件..." -ForegroundColor Cyan
$removeCmd = "rm -f $REMOTE_PATH/$BINARY_NAME"
ssh -i $PRIVATE_KEY -o StrictHostKeyChecking=no $REMOTE_USER@$REMOTE_HOST "$removeCmd"

if ($LASTEXITCODE -eq 0) {
    Write-Host "✅ 已删除远程旧文件" -ForegroundColor Green
} else {
    Write-Host "⚠️ 可能旧文件不存在" -ForegroundColor Yellow
}

# === Step 4: 上传新文件 ===
Write-Host "==> 上传新二进制到服务器..." -ForegroundColor Cyan
$uploadCmd = "scp -i $PRIVATE_KEY -o StrictHostKeyChecking=no $BINARY_NAME ${REMOTE_USER}@${REMOTE_HOST}:${REMOTE_PATH}/"
Invoke-Expression $uploadCmd

if ($LASTEXITCODE -ne 0) {
    Write-Host "❌ 上传失败" -ForegroundColor Red
    exit 1
}
Write-Host "✅ 上传成功" -ForegroundColor Green

# === Step 5: 远程启动服务 ===
Write-Host "==> 执行远程部署命令..." -ForegroundColor Cyan

$deployCmd = @"
cd $REMOTE_PATH
export GIN_MODE=release
chmod +x $REMOTE_PATH/$BINARY_NAME
fuser -k 8080/tcp || true
nohup $REMOTE_PATH/$BINARY_NAME > $REMOTE_PATH/blog-server.log 2>&1 &
"@

ssh -i $PRIVATE_KEY -o StrictHostKeyChecking=no $REMOTE_USER@$REMOTE_HOST "$deployCmd"

if ($LASTEXITCODE -ne 0) {
    Write-Host "❌ 远程执行失败" -ForegroundColor Red
    exit 1
}

Write-Host "✅ 部署完成并已启动服务" -ForegroundColor Green

# === Step 6: 清理本地编译产物 ===
Remove-Item $BINARY_NAME -Force
Write-Host "🧹 已清理本地编译文件" -ForegroundColor Gray
