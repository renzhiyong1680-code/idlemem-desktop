# 桌面托盘部署（Windows）

一键安装 `xiancun-tray` 到当前用户环境，可选开机自启：

```powershell
# 复制托盘到 %LOCALAPPDATA%\XianCun，生成 tray.json，注册开机自启并立即启动
powershell -ExecutionPolicy Bypass -File .\install-tray.ps1 -Binary .\xiancun-tray.exe

# 指向生产控制面，且不注册自启
powershell -ExecutionPolicy Bypass -File .\install-tray.ps1 `
    -Binary .\xiancun-tray.exe -Server http://1.2.3.4:8080 -KV :16380 -NoAutostart
```

## 安装内容

| 项目 | 位置 |
|---|---|
| 可执行文件 | `%LOCALAPPDATA%\XianCun\xiancun-tray.exe` |
| 运行配置 | `%LOCALAPPDATA%\XianCun\tray.json` |
| 开机自启 | `HKCU\Software\Microsoft\Windows\CurrentVersion\Run\XianCunTray` |

## 卸载

```powershell
Remove-ItemProperty -Path 'HKCU:\Software\Microsoft\Windows\CurrentVersion\Run' -Name 'XianCunTray'
Remove-Item -Recurse -Force "$env:LOCALAPPDATA\XianCun"
```

## 手工配置

不运行脚本也可直接使用：

```powershell
# 生成配置（interval 为纳秒，10000000000 = 10s）
'{"server":"http://127.0.0.1:8080","kv_addr":":16380","interval":10000000000}' |
    Set-Content -Path "$env:LOCALAPPDATA\XianCun\tray.json" -Encoding UTF8
```

## 说明

- 安装仅写入当前用户目录与当前用户注册表，无需管理员权限；
- 开机自启使用 `-config` 指向生成的 `tray.json`，改动配置无需重装；
- 托盘进程退出不卸载自启项（节点主人可能临时退出，开机后仍会拉起）。
