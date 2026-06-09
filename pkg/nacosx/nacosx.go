package nacosx

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/nacos-group/nacos-sdk-go/v2/clients"
	"github.com/nacos-group/nacos-sdk-go/v2/clients/config_client"
	"github.com/nacos-group/nacos-sdk-go/v2/common/constant"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
)

// Config controls optional Nacos configuration-center integration.
type Config struct {
	Enabled     bool   `mapstructure:"enabled" yaml:"enabled"`
	ServerAddr  string `mapstructure:"server_addr" yaml:"server_addr"`
	ServerPort  uint64 `mapstructure:"server_port" yaml:"server_port"`
	NamespaceID string `mapstructure:"namespace_id" yaml:"namespace_id"`
	Group       string `mapstructure:"group" yaml:"group"`
	DataID      string `mapstructure:"data_id" yaml:"data_id"`
	Username    string `mapstructure:"username" yaml:"username"`
	Password    string `mapstructure:"password" yaml:"password"`
	TimeoutMs   uint64 `mapstructure:"timeout_ms" yaml:"timeout_ms"`
	LogDir      string `mapstructure:"log_dir" yaml:"log_dir"`
	CacheDir    string `mapstructure:"cache_dir" yaml:"cache_dir"`
	LogLevel    string `mapstructure:"log_level" yaml:"log_level"`
	FailFast    bool   `mapstructure:"fail_fast" yaml:"fail_fast"`
	Watch       bool   `mapstructure:"watch" yaml:"watch"`
}

// DefaultConfig returns local-development defaults that keep Nacos disabled.
func DefaultConfig() Config {
	return Config{
		Enabled:    false,
		ServerAddr: "127.0.0.1",
		ServerPort: 8848,
		Group:      "DEFAULT_GROUP",
		DataID:     "xiaolanshu.yaml",
		TimeoutMs:  5000,
		LogDir:     "logs/nacos",
		CacheDir:   "data/nacos/cache",
		LogLevel:   "info",
	}
}

// WithDefaults fills empty optional fields without forcing Nacos on.
func WithDefaults(cfg Config) Config {
	defaults := DefaultConfig()
	if cfg.ServerAddr == "" {
		cfg.ServerAddr = defaults.ServerAddr
	}
	if cfg.ServerPort == 0 {
		cfg.ServerPort = defaults.ServerPort
	}
	if cfg.Group == "" {
		cfg.Group = defaults.Group
	}
	if cfg.DataID == "" {
		cfg.DataID = defaults.DataID
	}
	if cfg.TimeoutMs == 0 {
		cfg.TimeoutMs = defaults.TimeoutMs
	}
	if cfg.LogDir == "" {
		cfg.LogDir = defaults.LogDir
	}
	if cfg.CacheDir == "" {
		cfg.CacheDir = defaults.CacheDir
	}
	if cfg.LogLevel == "" {
		cfg.LogLevel = defaults.LogLevel
	}
	return cfg
}

// LoadConfig fetches one YAML config document from Nacos.
func LoadConfig(cfg Config) (string, error) {
	cfg = WithDefaults(cfg)
	if !cfg.Enabled {
		return "", nil
	}
	if strings.TrimSpace(cfg.DataID) == "" {
		return "", errors.New("nacos data_id is required")
	}
	if strings.TrimSpace(cfg.Group) == "" {
		return "", errors.New("nacos group is required")
	}

	client, err := NewConfigClient(cfg)
	if err != nil {
		return "", err
	}
	return client.GetConfig(vo.ConfigParam{
		DataId: cfg.DataID,
		Group:  cfg.Group,
	})
}

// NewConfigClient creates an official Nacos config client from framework config.
func NewConfigClient(cfg Config) (config_client.IConfigClient, error) {
	cfg = WithDefaults(cfg)
	return clients.NewConfigClient(
		vo.NacosClientParam{
			ClientConfig: &constant.ClientConfig{
				NamespaceId:         cfg.NamespaceID,
				TimeoutMs:           cfg.TimeoutMs,
				Username:            cfg.Username,
				Password:            cfg.Password,
				LogDir:              cfg.LogDir,
				CacheDir:            cfg.CacheDir,
				LogLevel:            cfg.LogLevel,
				NotLoadCacheAtStart: true,
			},
			ServerConfigs: []constant.ServerConfig{
				{
					IpAddr: cfg.ServerAddr,
					Port:   cfg.ServerPort,
				},
			},
		},
	)
}

// RegisterGRPCService 注册 gRPC 服务到 Nacos
func RegisterGRPCService(cfg Config, serviceName string, ip string, port int) (string, error) {
	client, err := clients.CreateNamingClient(map[string]interface{}{
		"clientConfig": &constant.ClientConfig{
			NamespaceId:         cfg.NamespaceID,
			TimeoutMs:           cfg.TimeoutMs,
			Username:            cfg.Username,
			Password:            cfg.Password,
			LogDir:              cfg.LogDir,
			CacheDir:            cfg.CacheDir,
			LogLevel:            cfg.LogLevel,
			NotLoadCacheAtStart: true,
		},
		"serverConfigs": []constant.ServerConfig{
			{
				IpAddr: cfg.ServerAddr,
				Port:   cfg.ServerPort,
			},
		}})
	if err != nil {
		fmt.Println(err)
		return "nacos注册服务连接失败", err
	}
	success, err := client.RegisterInstance(vo.RegisterInstanceParam{
		Ip:          ip,
		Port:        uint64(port),
		ServiceName: serviceName,
		GroupName:   cfg.Group,
		Ephemeral:   true, // 临时实例（自动心跳）
	})
	if err != nil || !success {
		fmt.Println(err)
		return "nacos注册服务失败-2", err
	}
	fmt.Println("nacos：注册微服务-成功")
	return "", nil
}

// GetServiceAddr 获取单个服务信息
func GetServiceAddr(cfg Config, serviceName string) (string, string, int) {
	client, err := clients.CreateNamingClient(map[string]interface{}{
		"clientConfig": &constant.ClientConfig{
			NamespaceId:         cfg.NamespaceID,
			TimeoutMs:           cfg.TimeoutMs,
			Username:            cfg.Username,
			Password:            cfg.Password,
			LogDir:              cfg.LogDir,
			CacheDir:            cfg.CacheDir,
			LogLevel:            cfg.LogLevel,
			NotLoadCacheAtStart: true,
		},
		"serverConfigs": []constant.ServerConfig{
			{
				IpAddr: cfg.ServerAddr,
				Port:   cfg.ServerPort,
			},
		}})
	if err != nil {
		fmt.Println(err)
		return "nacos连接失败", "", 0
	}

	instances, err := client.GetService(vo.GetServiceParam{
		ServiceName: serviceName,
		GroupName:   cfg.Group,
	})
	if err != nil {
		fmt.Println(err)
		return "获取实例列表失败", "", 0
	}
	jsonData1, _ := json.Marshal(instances)
	fmt.Println("服务信息：", string(jsonData1))

	jsonData2, _ := json.Marshal(instances.Hosts)
	fmt.Println("服务信息-ip：", string(jsonData2))

	instance := instances.Hosts[0]
	str := fmt.Sprintf("获取成功，%v:%v", instance.Ip, int(instance.Port))
	fmt.Println(str)

	//ip, _ := strconv.Atoi(instance.Ip)
	return "", instance.Ip, int(instance.Port)
}
