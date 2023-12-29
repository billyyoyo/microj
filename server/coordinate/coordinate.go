package coordinate

import (
	"fmt"
	"github.com/billyyoyo/microj/app"
	"github.com/billyyoyo/microj/config"
	"github.com/billyyoyo/microj/logger"
	"github.com/hashicorp/raft"
	raftboltdb "github.com/hashicorp/raft-boltdb"
	"io"
	"net"
	"path/filepath"
	"time"
)

const (
	maxPool = 3

	tcpTimeout = 10 * time.Second

	raftSnapShotRetain = 2

	raftLogCacheSize = 512
)

type ConfigRaft struct {
	Id        string       `yaml:"id"`
	Address   string       `yaml:"address"`
	Port      int          `yaml:"port"`
	VolumeDir string       `yaml:"volumeDir"`
	Mode      string       `yaml:"mode"`
	Nodes     []ConfigNode `yaml:"nodes" mapstructure:"nodes"`
}

func (r ConfigRaft) getRaftServers() (list []raft.Server) {
	for _, s := range r.Nodes {
		list = append(list, s.getRaftServer())
	}
	return
}

type ConfigNode struct {
	Id         string `yaml:"id"`
	Addr       string `yaml:"addr"`
	RaftPort   int    `yaml:"raftPort"`
	ServerPort int    `yaml:"serverPort"`
}

func (n ConfigNode) getRaftServer() raft.Server {
	return raft.Server{
		ID:      raft.ServerID(n.Id),
		Address: raft.ServerAddress(fmt.Sprintf("%s:%d", n.Addr, n.RaftPort)),
	}
}

func (n ConfigNode) getServerAddr() string {
	return fmt.Sprintf("http://%s:%d", n.Addr, n.ServerPort)
}

type CoordinateServer struct {
	Port    int
	Config  ConfigRaft
	Raft    *raft.Raft
	handler StateHandler
}

func NewServer() *CoordinateServer {
	srv := CoordinateServer{}
	srv.Init()
	return &srv
}

func (d *CoordinateServer) Persist(sink raft.SnapshotSink) error {
	defer func() {
		if err := sink.Close(); err != nil {
			logger.Error("snapshot save failed", err)
		}
	}()
	bs, err := d.handler.GetApplicationData()
	if err != nil {
		logger.Error("snapshot application data", err)
		return err
	}
	_, err = sink.Write(bs)
	if err != nil {
		logger.Error("save snapshot failed", err)
		return err
	}
	return nil
}

func (d *CoordinateServer) Release() {
}

func (d *CoordinateServer) Apply(log *raft.Log) interface{} {
	if log.Type != raft.LogCommand {
		return nil
	}
	return d.handler.CommandApplied(log.Data)
}

func (d *CoordinateServer) Snapshot() (raft.FSMSnapshot, error) {
	return d, nil
}

func (d *CoordinateServer) Restore(reader io.ReadCloser) error {
	defer func() {
		if err := reader.Close(); err != nil {
			logger.Err(err)
		}
	}()
	bs, err := io.ReadAll(reader)
	if err != nil {
		logger.Error("restore snapshot failed", err)
		return err
	}
	return d.handler.SnapshotRestored(bs)
}

func (d *CoordinateServer) Init() {
	d.Port = app.Port()
	conf := ConfigRaft{}
	err := config.Scan("raft", &conf)
	if err != nil {
		logger.Fatal("raft config load failed", err)
	}
	d.Config = conf
}

func (d *CoordinateServer) Run() {
	// raft数据传输地址
	var raftBinAddr = fmt.Sprintf("%s:%d", d.Config.Address, d.Config.Port)

	raftConf := raft.DefaultConfig()
	raftConf.LocalID = raft.ServerID(d.Config.Id)
	raftConf.SnapshotThreshold = 1024
	raftConf.Logger = NewLogger()
	notifyChan := make(chan bool)
	raftConf.NotifyCh = notifyChan
	go func(ch chan bool) {
		for {
			select {
			case change := <-ch:
				d.handler.LeaderChanged(change)
			}
		}
	}(notifyChan)

	// 封装状态机存储
	store, err := raftboltdb.NewBoltStore(filepath.Join(d.Config.VolumeDir, "raft.dataRepo"))
	if err != nil {
		logger.Fatal("can not create a raft store", err)
		return
	}

	// 封装日志存储 这里在接口与存储之间又加了一个缓存  应该是避免高并发问题
	// Wrap the store in a LogCache to improve performance.
	cacheStore, err := raft.NewLogCache(raftLogCacheSize, store)
	if err != nil {
		logger.Fatal("can not create log cache store", err)
		return
	}

	// 封装快照存储
	snapshotStore, err := raft.NewFileSnapshotStoreWithLogger(d.Config.VolumeDir, raftSnapShotRetain, NewLogger())
	if err != nil {
		logger.Fatal("can not create snapshot store", err)
		return
	}

	// 长连接 服务端
	tcpAddr, err := net.ResolveTCPAddr("tcp", raftBinAddr)
	if err != nil {
		logger.Fatal("create tcp server address failed", err)
		return
	}

	// 创建传输通道
	transport, err := raft.NewTCPTransportWithLogger(raftBinAddr, tcpAddr, maxPool, tcpTimeout, NewLogger())
	if err != nil {
		logger.Fatal("create tcp transport failed", err)
		return
	}

	// raft核心
	d.Raft, err = raft.NewRaft(raftConf, d, cacheStore, store, snapshotStore, transport)
	if err != nil {
		logger.Fatal("create raft logic server failed", err)
		return
	}
	d.handler.InitFinished(d.Config, d.Raft)

	// mode : cluster / leader /follower
	if d.Config.Mode == "cluster" {
		if !findLeader(d.Config.Id, d.Config.Nodes) {
			configuration := raft.Configuration{
				Servers: d.Config.getRaftServers(),
			}

			// 启动leader  先启动起来再说  后面接口调用互相加voter
			d.Raft.BootstrapCluster(configuration)
			logger.Info("I am the boss")
		} else {
			logger.Info("I am waiting for boss")
		}
	} else if d.Config.Mode == "leader" {
		configuration := raft.Configuration{
			Servers: []raft.Server{
				{
					ID:      raft.ServerID(d.Config.Id),
					Address: transport.LocalAddr(),
				},
			},
		}
		d.Raft.BootstrapCluster(configuration)
	}
}

func (d *CoordinateServer) Stop() {
	d.Raft.Shutdown()
}

func (d *CoordinateServer) WithHandler(h StateHandler) *CoordinateServer {
	d.handler = h
	return d
}

type StateHandler interface {
	LeaderChanged(isLeader bool)
	CommandApplied(data []byte) interface{}
	GetApplicationData() ([]byte, error)
	SnapshotRestored(data []byte) error
	InitFinished(c ConfigRaft, r *raft.Raft)
}
