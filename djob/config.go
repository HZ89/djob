package djob

import (
	"encoding/base64"
	"errors"
	"flag"
	"fmt"
	"github.com/spf13/viper"
	"local/djob/cmd"
	"net"
	"os"
	"path/filepath"
)

type Config struct {
	Server            bool              // start as server or just agent
	Region            string            // agent region
	Nodename          string            //serf node name
	Tags              map[string]string // agent tags used to filter agent
	SerfBindIp        string            // serf system bind address ip
	SerfBindPort      int
	SerfAdvertiseIp   string // serf system advertise address ip used for agent behind a firewall
	SerfAdvertisePort int
	SerfJoin          []string
	APIBindIp         string // HTTP API PORT just effect when server is true
	APIBindPort       int
	JobStore          string   // key value storage type etd/zookeeper etc.
	JobStoreServers   []string // k/v storage server list [ip:port]
	JobStoreKeyspace  string   // keyspace in the storage
	encryptKey        string   // serf encrypt key
	LogLevel          string   // info debug error
	RpcTls            bool     // grpc enable tls or not
	RpcCAfile         string   // tls ca file used in agent
	RpcKeyFile        string   // key file used in server
	RpcCertFile       string   // cert file used in server
	RpcBindIp         string   // grcp bind addr ip
	RpcBindPort       int
	RpcAdcertiseIP    string // sames to serf advertise addr
	RpcAdcertisePort  int
	SerfSnapshotPath  string //serf use this path to save snapshot of joined server
}

const (
	DefaultRegion       string = "MARS"
	DefaultSerfPort     int    = 8998
	DefaultHttpPort     int    = 8088
	DefaultRPCPort      int    = 7979
	DefaultSnapshotPath string = "./snapshot"
	DefaultConfigFile   string = "./conf/djob.yml"
	DefaultPidFile      string = "./djob.pid"
	DefaultLogFile      string = "./djob.log"
	DefaultKeySpeace    string = "djob"
)

//func init() {
//	viper.SetConfigFile("djob")
//	viper.AddConfigPath("./config")
//	viper.SetConfigType("yaml")
//}

func NewConfig(args []string) (*Config, error) {
	runRoot, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		return nil, err
	}

	cmdFlags := flag.NewFlagSet("agent", flag.ContinueOnError)
	cmdFlags.String("config", filepath.Join(runRoot, DefaultConfigFile), "config file path")
	cmdFlags.String("pid", filepath.Join(runRoot, DefaultPidFile), "pid file path")
	cmdFlags.String("logfile", filepath.Join(runRoot, DefaultLogFile), "log file path")

	if err := cmdFlags.Parse(args); err != nil {
		return nil, err
	}
	viper.SetConfigFile(string(cmdFlags.Lookup("config").Value))
	viper.SetDefault("serf_bind_addr", fmt.Sprintf("%s:%d", "0.0.0.0", DefaultSerfPort))
	viper.SetDefault("http_api_addr", fmt.Sprintf("%s:%d", "0.0.0.0", DefaultHttpPort))
	viper.SetDefault("rpc_bind_addr", fmt.Sprintf("%s:%d", "0.0.0.0", DefaultRPCPort))
	viper.SetDefault("job_store", "etcd")
	viper.SetDefault("job_store_keyspeace", DefaultKeySpeace)
	viper.SetDefault("log_level", "info")
	viper.SetDefault("rpc_tls", false)
	viper.SetDefault("region", DefaultRegion)
	viper.SetDefault("server", false)
	viper.SetDefault("serf_snapshot_dir", filepath.Join(runRoot, DefaultSnapshotPath))

	return ReadConfig()
}

func ReadConfig() (*Config, error) {
	err := viper.ReadInConfig()
	if err != nil {
		return nil, err
	}
	tags := viper.GetStringMapString("tags")
	server := viper.GetBool("server")

	nodeName, err := os.Hostname()
	if err != nil {
		return nil, err
	}

	if server {
		tags["server"] = "true"
	}

	tags["version"] = cmd.VERSION
	tags["node"] = nodeName
	tags["region"] = viper.GetString("region")

	withTls := viper.GetBool("rpc_tls")
	keyFile := viper.GetString("rpc_key_file")
	certFile := viper.GetString("rpc_vert_file")
	caFile := viper.GetString("rpc_ca_file")
	if withTls {
		if server {
			if keyFile == "" || certFile == "" {
				return nil, errors.New("have no key file or cert file path.")
			}
		} else {
			if caFile == "" {
				return nil, errors.New("have no ca file path.")
			}
		}
	}

	serfBindip, serfBindport, err := splitNetAddr(viper.GetString("serf_bind_addr"), DefaultSerfPort)
	if err != nil {
		return nil, err
	}
	apiBindip, apiBindport, err := splitNetAddr(viper.GetString("http_api_addr"), DefaultHttpPort)
	if err != nil {
		return nil, err
	}
	rpcBindip, rpcBindport, err := splitNetAddr(viper.GetString("rpc_bind_addr"), DefaultRPCPort)
	if err != nil {
		return nil, err
	}

	serfAdip, serfAdport, err := handleAdvertise(viper.GetString("serf_advertise_addr"), fmt.Sprintf("%s:%d", serfBindip, serfBindport))
	if err != nil {
		return nil, err
	}

	rpcAdip, rpcAdport, err := handleAdvertise(viper.GetString("rpc_advertise_addr"), fmt.Sprintf("%s:%d", rpcBindip, rpcBindport))
	if err != nil {
		return nil, err
	}

	return &Config{
		Server:            server,
		Region:            tags["region"],
		Tags:              tags,
		SerfBindIp:        serfBindip,
		SerfBindPort:      serfBindport,
		APIBindIp:         apiBindip,
		APIBindPort:       apiBindport,
		RpcBindIp:         rpcBindip,
		RpcBindPort:       rpcBindport,
		SerfAdvertiseIp:   serfAdip,
		SerfAdvertisePort: serfAdport,
		RpcAdcertiseIP:    rpcAdip,
		RpcAdcertisePort:  rpcAdport,
		LogLevel:          viper.GetString("log_level"),
		RpcCAfile:         caFile,
		RpcCertFile:       certFile,
		RpcKeyFile:        keyFile,
		RpcTls:            withTls,
		JobStore:          viper.GetString("job_store"),
		JobStoreServers:   viper.GetStringSlice("job_store_servers"),
		JobStoreKeyspace:  viper.GetString("job_store_keyspeace"),
		SerfJoin:          viper.GetStringSlice("join"),
		Nodename:          nodeName,
		encryptKey:        viper.GetString("encrypt_key"),
		SerfSnapshotPath:  viper.GetString("serf_snapshot_dir"),
	}, nil
}

func (c *Config) EncryptKey() ([]byte, error) {
	return base64.StdEncoding.DecodeString(c.encryptKey)
}

// make sure advertise ip or port is not 0.0.0.0 or 0
func handleAdvertise(oaddr string, daddr string) (string, int, error) {
	if oaddr == "" {
		oaddr = daddr
	}
	addr, _ := net.ResolveTCPAddr("tcp", oaddr)
	if net.IPv6zero.Equal(addr.IP) {
		var privateIp net.IP
		var publicIp net.IP
		ifaces, err := net.Interfaces()
		if err != nil {
			return "", 0, err
		}

		for _, i := range ifaces {
			addrs, err := i.Addrs()
			if err != nil {
				return "", 0, err
			}
			for _, addr := range addrs {
				switch v := addr.(type) {
				case *net.IPAddr:
					if isPrivate(v.IP) {
						privateIp = v.IP
					} else {
						publicIp = v.IP
					}
				case *net.IPNet:
					if isPrivate(v.IP) {
						privateIp = v.IP
					} else {
						publicIp = v.IP
					}

				}
			}

		}
		if privateIp == nil {
			addr.IP = publicIp
		} else {
			addr.IP = privateIp
		}
	}
	return addr.IP.String(), addr.Port, nil

}

func isPrivate(x net.IP) bool {
	if x.To4() != nil {
		_, rfc1918_24BitBlock, _ := net.ParseCIDR("10.0.0.0/8")
		_, rfc1918_20BitBlock, _ := net.ParseCIDR("172.16.0.0/12")
		_, rtc1918_16BitBlock, _ := net.ParseCIDR("192.168.0.0/16")
		return rtc1918_16BitBlock.Contains(x) || rfc1918_20BitBlock.Contains(x) || rfc1918_24BitBlock.Contains(x)
	}
	_, rfc4193Block, _ := net.ParseCIDR("fd00::/8")
	return rfc4193Block.Contains(x)
}

// split ip:port if have no port, use the default port
func splitNetAddr(address string, defaultport int) (string, int, error) {
BEGIN:
	_, _, err := net.SplitHostPort(address)
	if es, ok := err.(*net.AddrError); ok && es.Err == "missing port in address" {
		address = fmt.Sprintf("%s:%d", address, defaultport)
		goto BEGIN
	} else {
		return "", 0, err
	}

	addr, err := net.ResolveTCPAddr("tcp", address)
	if err != nil {
		return "", 0, err
	}
	return addr.IP.String(), addr.Port, nil

}
