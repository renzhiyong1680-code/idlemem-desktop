# idlemem-desktop

闲存云 Agent 的**桌面外壳**仓库：Windows 系统托盘 + macOS 菜单栏应用。

它把 [idlemem-core](https://github.com/xiancuncloud/idlemem-core) 核心库装配成一个常驻托盘进程，节点主人无需命令行即可管理自己的"闲置内存贡献节点"。

## 功能

- **状态一瞥**：托盘菜单实时显示贡献状态（在线/暂停）、贡献容量、KV 占用、key 数、KV 端口与控制面地址；
- **手动暂停/恢复贡献**：一键开关，优先级高于内存压力判定（对核心库 `governor.SetEnabled` 的调用）；
- **优雅退出**：关闭 KV 端口并停止心跳上报；
- **跨平台同一份代码**：`getlantern/systray` 驱动，Windows/macOS 共用 `cmd/xiancun-tray`。

## 构建

```bash
# Windows（需要 CGO 工具链，如 WinLibs gcc）
CGO_ENABLED=1 go build -o xiancun-tray.exe ./cmd/xiancun-tray

# macOS（在 mac 上构建）
CGO_ENABLED=1 go build -o xiancun-tray ./cmd/xiancun-tray
```

> 说明：systray 依赖平台原生 GUI API（Windows Win32 / macOS Cocoa），因此必须开启 CGO，且 macOS 版需在 mac 上编译。CLI 外壳 `xiancun-agent` 仍是零依赖纯标准库，可在任意平台交叉编译。

## 运行

```bash
# 默认连 http://127.0.0.1:8080，KV 端口 :16380
xiancun-tray.exe

# 指定控制面与 KV 端口
xiancun-tray.exe -server http://1.2.3.4:8080 -kv :16380 -interval 10s

# 用 JSON 配置文件（字段与 cmd/xiancun-agent 一致）
xiancun-tray.exe -config tray.json
```

## 托盘菜单

```
闲存云 Agent            （图标：蓝底"闲"字）
├─ 状态：online　贡献 12.6GB　占用 0.1GB　keys 3
├─ KV 端口：:16380
├─ 控制面：http://127.0.0.1:8080
├─────────────────────
├─ 暂停贡献 / 恢复贡献
└─ 退出
```

## 与三端外壳的关系

| 外壳 | 形态 | 仓库 |
|---|---|---|
| CLI | 命令行 / systemd 守护 | `idlemem-core/cmd/xiancun-agent` |
| Windows 托盘 | 系统托盘 | 本仓库 `cmd/xiancun-tray` |
| macOS 菜单栏 | 菜单栏图标 | 本仓库 `cmd/xiancun-tray` |

## 架构

```
┌─────────────────────────────────────────────┐
│  xiancun-tray（本仓库）                       │
│   systray 菜单/图标 + 配置加载 + 生命周期装配   │
├─────────────────────────────────────────────┤
│  idlemem-core（replace 到本地）              │
│   agent（整合入口）+ kvsrv（RESP 端口）        │
│   + governor（手动暂停开关）+ heartbeat       │
└─────────────────────────────────────────────┘
```

## 开发

```bash
git clone git@github.com:xiancuncloud/idlemem-desktop.git
# 与 idlemem-core 同级目录，go.mod 已 replace 到 ../idlemem-core
go mod tidy
go vet ./...
```
