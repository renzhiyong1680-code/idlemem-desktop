// Command xiancun-tray 是闲存云 Agent 的桌面外壳：Windows 系统托盘 / macOS 菜单栏。
//
// 它把 idlemem-core 核心库装配成一个常驻托盘进程，节点主人可以：
//   - 一眼看到贡献状态（状态 / 容量 / KV 端口 / 心跳上报地址）；
//   - 随时暂停 / 恢复贡献（手动开关，优先于内存压力判定）；
//   - 一键退出（退出时优雅关闭 KV 端口并停止心跳）。
//
// 构建：
//
//	# Windows（需要 CGO 工具链，如 WinLibs gcc）
//	CGO_ENABLED=1 go build -o xiancun-tray.exe ./cmd/xiancun-tray
//	# macOS（在 mac 上构建）
//	CGO_ENABLED=1 go build -o xiancun-tray ./cmd/xiancun-tray
//
// 用法：
//
//	xiancun-tray -server http://127.0.0.1:8080 -kv :16380
//	xiancun-tray -config tray.json
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	_ "embed"
	"os"
	"time"

	"github.com/getlantern/systray"
	"github.com/xiancuncloud/idlemem-core/agent"
	"github.com/xiancuncloud/idlemem-core/governor"
	"github.com/xiancuncloud/idlemem-core/heartbeat"
	"github.com/xiancuncloud/idlemem-core/kvsrv"
	"github.com/xiancuncloud/idlemem-core/kvstore"
	"github.com/xiancuncloud/idlemem-core/meminfo"
	"github.com/xiancuncloud/idlemem-core/nodeid"
)

//go:embed assets/icon.png
var iconData []byte

// version 由构建时注入（-ldflags "-X main.version=v0.1.0"），用于运维核对节点上跑的是哪一版。
var version = "dev"

// config 与 cmd/xiancun-agent 保持一致的运行配置。
type config struct {
	Server     string        `json:"server"`
	KVAddr     string        `json:"kv_addr"`
	NodeID     string        `json:"node_id"`
	Interval   time.Duration `json:"interval"`
	KVMaxBytes uint64        `json:"kv_max_bytes"`
}

func (c *config) applyDefaults() {
	if c.Server == "" {
		c.Server = "http://127.0.0.1:8080"
	}
	if c.KVAddr == "" {
		c.KVAddr = ":16380"
	}
	if c.Interval <= 0 {
		c.Interval = 10 * time.Second
	}
}

// trayApp 托盘应用状态。
type trayApp struct {
	cfg    config
	core   *agent.Agent
	kvSrv  *kvsrv.Server
	ctx    context.Context
	cancel context.CancelFunc
	logger *log.Logger

	// 菜单项
	statusItem *systray.MenuItem
	kvItem     *systray.MenuItem
	serverItem *systray.MenuItem
	toggleItem *systray.MenuItem
	quitItem   *systray.MenuItem
	paused     bool
}

func main() {
	logger := log.New(os.Stdout, "[tray] ", log.LstdFlags|log.Lmsgprefix)

	var (
		cfgPath  = flag.String("config", "", "config json file path (optional)")
		server   = flag.String("server", "", "control plane API address, e.g. http://127.0.0.1:8080")
		kvAddr   = flag.String("kv", "", "KV service listen address, e.g. :16380")
		nodeID   = flag.String("id", "", "node id (default: auto-generated stable id)")
		interval = flag.Duration("interval", 0, "heartbeat interval (default 10s)")
		showVersion = flag.Bool("version", false, "print the version and exit")
	)
	flag.Parse()

	if *showVersion {
		fmt.Println("xiancun-tray " + version)
		return
	}

	cfg := config{}
	if *cfgPath != "" {
		if err := loadConfig(*cfgPath, &cfg); err != nil {
			logger.Fatalf("load config: %v", err)
		}
	}
	if *server != "" {
		cfg.Server = *server
	}
	if *kvAddr != "" {
		cfg.KVAddr = *kvAddr
	}
	if *nodeID != "" {
		cfg.NodeID = *nodeID
	}
	if *interval > 0 {
		cfg.Interval = *interval
	}
	cfg.applyDefaults()

	app := &trayApp{cfg: cfg, logger: logger}
	app.run()
}

func (a *trayApp) run() {
	// systray.Run 阻塞直到 Quit；onReady 内完成所有装配与启动。
	systray.Run(func() { a.onReady() }, func() { a.onExit() })
}

func (a *trayApp) onReady() {
	a.logger.Printf("node id: %s", a.nodeID())
	a.logger.Printf("starting: server=%s kv=%s interval=%s", a.cfg.Server, a.cfg.KVAddr, a.cfg.Interval)

	// 装配核心库。
	prov := meminfo.SystemProvider{}
	rep := heartbeat.NewHTTPReporter(a.cfg.Server + "/api/v1/nodes/heartbeat")
	core, err := agent.New(agent.Config{
		NodeID:         a.nodeID(),
		Version:        agent.Version,
		MemProvider:    prov,
		GovernorCfg:    governor.DefaultConfig(),
		Reporter:       rep,
		StoreOpts:      kvstore.Options{MaxBytes: a.cfg.KVMaxBytes},
		ReportInterval: a.cfg.Interval,
		KVAddr:         a.cfg.KVAddr,
	})
	if err != nil {
		a.logger.Fatalf("create agent core: %v", err)
	}
	a.core = core

	kvSrv, err := kvsrv.New(a.cfg.KVAddr, core, a.logger)
	if err != nil {
		a.logger.Fatalf("create kv service: %v", err)
	}
	a.kvSrv = kvSrv
	go kvSrv.Serve()

	a.ctx, a.cancel = context.WithCancel(context.Background())
	if err := core.Start(a.ctx); err != nil {
		a.logger.Fatalf("start agent: %v", err)
	}
	a.logger.Printf("agent started")

	// 托盘 UI。
	systray.SetIcon(iconData)
	systray.SetTitle("闲存云 Agent")
	systray.SetTooltip("闲存云 Agent " + version + " - 闲置内存贡献节点")

	a.statusItem = systray.AddMenuItem("状态：启动中", "当前贡献状态")
	a.statusItem.Disable()
	a.kvItem = systray.AddMenuItem("KV 端口："+a.cfg.KVAddr, "Redis 协议缓存端口")
	a.kvItem.Disable()
	a.serverItem = systray.AddMenuItem("控制面："+a.cfg.Server, "心跳上报地址")
	a.serverItem.Disable()
	systray.AddSeparator()
	a.toggleItem = systray.AddMenuItem("暂停贡献", "暂停/恢复贡献（手动开关）")
	a.quitItem = systray.AddMenuItem("退出", "停止贡献并退出")

	// 状态刷新循环。
	go a.refreshLoop()
}

func (a *trayApp) onExit() {
	a.logger.Printf("shutting down...")
	if a.kvSrv != nil {
		a.kvSrv.Close()
	}
	if a.core != nil {
		a.core.Stop()
	}
	if a.cancel != nil {
		a.cancel()
	}
}

// refreshLoop 每 2 秒刷新状态文本并处理暂停/恢复与退出。
func (a *trayApp) refreshLoop() {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-a.ctx.Done():
			return
		case <-ticker.C:
			a.refresh()
		case <-a.toggleItem.ClickedCh:
			a.togglePause()
		case <-a.quitItem.ClickedCh:
			systray.Quit()
			return
		}
	}
}

func (a *trayApp) refresh() {
	if a.core == nil {
		return
	}
	st := a.core.Store()
	status := string(a.core.Status())
	capGB := fmtGB(a.core.Capacity())
	usedGB := fmtGB(st.UsedBytes())
	keys := st.Len()
	a.statusItem.SetTitle("状态：" + status + "　贡献 " + capGB + "　占用 " + usedGB + "　keys " + itoa(keys))
	if a.paused {
		a.toggleItem.SetTitle("恢复贡献")
	} else {
		a.toggleItem.SetTitle("暂停贡献")
	}
}

func (a *trayApp) togglePause() {
	a.paused = !a.paused
	a.core.Governor().SetEnabled(!a.paused)
	a.logger.Printf("contribution %s", map[bool]string{true: "paused", false: "resumed"}[a.paused])
	a.refresh()
}

func (a *trayApp) nodeID() string {
	if a.cfg.NodeID != "" {
		return a.cfg.NodeID
	}
	return nodeid.String()
}

// ---------- 工具 ----------

// loadConfig 从 JSON 文件加载配置。
func loadConfig(path string, cfg *config) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, cfg)
}

// fmtGB 把字节格式化为可读的 GB 字符串（保留 1 位小数）。
func fmtGB(b uint64) string {
	return fmt.Sprintf("%.1fGB", float64(b)/(1024*1024*1024))
}

// itoa 最小整数转字符串。
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}
