# 闲存云 Agent 桌面托盘 - Windows 安装脚本
#
# 用法（以管理员或普通用户身份均可，写入当前用户级目录与注册表）：
#   powershell -ExecutionPolicy Bypass -File .\install-tray.ps1 -Binary .\xiancun-tray.exe
#   powershell -ExecutionPolicy Bypass -File .\install-tray.ps1 -Binary .\xiancun-tray.exe -Server http://1.2.3.4:8080 -KV :16380
#
# 可选参数：
#   -Binary      托盘可执行文件路径（必填）
#   -Server      控制面 REST API 地址（默认 http://127.0.0.1:8080）
#   -KV          KV 服务监听地址（默认 :16380）
#   -NoAutostart 不注册开机自启（默认注册）
#
# 卸载：
#   Remove-ItemProperty -Path 'HKCU:\Software\Microsoft\Windows\CurrentVersion\Run' -Name 'XianCunTray'
#   然后删除 %LOCALAPPDATA%\XianCun 目录。

param(
    [Parameter(Mandatory = $true)][string]$Binary,
    [string]$Server = "http://127.0.0.1:8080",
    [string]$KV = ":16380",
    [switch]$NoAutostart
)

$ErrorActionPreference = "Stop"

$dir = Join-Path $env:LOCALAPPDATA "XianCun"
$exe = Join-Path $dir "xiancun-tray.exe"
$cfg = Join-Path $dir "tray.json"

# 1. 复制可执行文件。
New-Item -ItemType Directory -Force -Path $dir | Out-Null
if (-not (Test-Path -LiteralPath $Binary -PathType Leaf)) {
    throw "binary not found: $Binary"
}
Copy-Item -LiteralPath $Binary -Destination $exe -Force
Write-Host "[1/4] binary installed: $exe"

# 2. 生成运行配置（与 cmd/xiancun-tray 的 JSON 字段一致，interval=10s）。
$cfgObj = @{
    server   = $Server
    kv_addr  = $KV
    interval = 10000000000
}
$cfgText = $cfgObj | ConvertTo-Json -Compress
[IO.File]::WriteAllText($cfg, $cfgText, (New-Object Text.UTF8Encoding $false))
Write-Host "[2/4] config written: $cfg"

# 3. 注册开机自启（当前用户）。
if (-not $NoAutostart) {
    $runKey = "HKCU:\Software\Microsoft\Windows\CurrentVersion\Run"
    $value = "`"$exe`" -config `"$cfg`""
    Set-ItemProperty -Path $runKey -Name "XianCunTray" -Value $value
    Write-Host "[3/4] autostart registered (HKCU Run: XianCunTray)"
} else {
    Write-Host "[3/4] autostart skipped (-NoAutostart)"
}

# 4. 立即启动托盘。
Start-Process -FilePath $exe -ArgumentList @("-config", $cfg)
Write-Host "[4/4] tray started: $exe"
Write-Host "Done."
