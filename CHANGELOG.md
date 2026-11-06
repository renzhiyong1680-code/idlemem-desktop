# 更新日志

本项目遵循 [语义化版本](https://semver.org/lang/zh-CN/)。

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
