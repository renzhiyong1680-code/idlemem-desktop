# 更新日志

本项目遵循 [语义化版本](https://semver.org/lang/zh-CN/)。

## [v0.1.2] - 2026-11-10

> 本版**没有托盘程序源码变更**：`main.go` 与 `v0.1.1` 一致，产物行为相同。改动集中在 CI（macOS job 不再把整轮判红、产物名带真实架构）与 README。

### 新增

- README 新增「已知限制」：macOS 只做构建验证未做真机交互、macOS 产物只有 arm64（CI 构建机为 Apple Silicon，Intel Mac 需自行编译）、产物无代码签名、托盘不持久化状态
- `CLA.md`：贡献者许可协议（中英双语十节），贡献者保留著作权，仅授予项目方在开源与商业授权两种模式下的使用许可
- `CONTRIBUTING.md`、`SECURITY.md`、`.github/ISSUE_TEMPLATE/`：贡献流程、漏洞报告通道与 Issue 模板
- README 顶部增加三仓导航、项目介绍入口与 Release 直链
- README 新增「下载后怎么打开」：说明 `v0.1.1` 二进制未做代码签名，给出 macOS（`xattr -d com.apple.quarantine` / 右键打开 / 隐私与安全性放行）与 Windows（SmartScreen「仍要运行」）的处理步骤
- `docs/assets/architecture.svg`：三仓协作架构图，README 新增「三仓如何协作」段并引用该图
- README「三仓如何协作」补一句 AI 基站（AI-RAN）落地位（托盘装在边缘服务器上，登记闲置内存），并链接 core 仓的 `ai-ran-architecture.svg`
- `ci.yml`：`build-macos` 加 `continue-on-error: true`（无 macOS 真机验证，不让整轮判红），新增 `Report target arch` 步骤把 `go env GOARCH` 写进 job output 与产物名

### 修复

- README「已知限制」把"macOS 产物架构以 CI 日志为准"改为实测结论：Release 产物为 `xiancun-tray-darwin_arm64`（CI 构建机 Apple Silicon），Intel Mac 需自行编译

### 变更

- macOS 产物名由硬编码的 `xiancun-tray-darwin_amd64` 改为 `xiancun-tray-darwin_<arch>`，架构取自 `go env GOARCH`，避免在 arm64 构建机上产出叫 amd64 的二进制
- 核心库依赖由 `idlemem-core v0.1.0` 升级为 `v0.1.1`（两版代码一致，仅许可证变更）

### 修复

- CI 的 `build-macos` job 在 GitHub 托管 macOS runner 排队超时被平台取消，会把整轮 CI 判成失败（同一次运行的 `build-windows` 十个步骤全部成功）。现在该 job 标记 `continue-on-error: true`，单独失败不再染红整轮

## [v0.1.1] - 2026-11-06

仅许可证变更，代码与 `v0.1.0` 完全一致。

### 变更

- 许可证由 MIT 变更为 AGPL-3.0（GNU Affero General Public License v3.0），保留后续商业授权选项；
  `v0.1.0` 及更早版本仍按 MIT 发布，该授权不可撤销。
- 新增 `COMMERCIAL.md`：商业授权的适用场景、授权形式与联系方式

## [v0.1.0] - 2026-11-06

首个公开版本：闲存云 Agent 的桌面外壳，Windows 系统托盘 / macOS 菜单栏。

### 功能

- 托盘/菜单栏常驻，一眼看到贡献状态：状态、贡献容量、实际占用、keys、KV 端口、心跳上报地址
- 暂停 / 恢复贡献（优先于内存压力判定），退出时优雅关闭 KV 端口并停止心跳
- `-version`：版本由构建时注入并显示在托盘提示里，便于核对节点上跑的是哪一版
- `deploy/install-tray.ps1`：Windows 安装 + 开机自启（注册表 Run 键、配置写入、启动验证）

### 构建

- 依赖 systray，必须 `CGO_ENABLED=1`：
  - Windows 需要 C 工具链（如 MinGW）
  - macOS 需在 mac 上用系统 clang 构建
- CI：Windows 与 macOS 各构建一次（macOS 未在真机人工验证，仅由 CI 编译证明）

### 已知限制

- macOS 菜单栏仅在代码层面就绪，未做真机交互验证
- 依赖 `../idlemem-core` 同级目录（本地联合开发形态），正式发布按版本号依赖
