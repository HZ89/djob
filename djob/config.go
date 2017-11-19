/*
 * Copyright (c) 2017.  Harrison Zhu <wcg6121@gmail.com>
 * This file is part of djob <https://github.com/HZ89/djob>.
 *
 * djob is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * djob is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with djob.  If not, see <http://www.gnu.org/licenses/>.
 */

package djob

import (
	"encoding/base64"
	"flag"
	"fmt"
	"net"
	"os"

	"github.com/spf13/viper"

	"github.com/HZ89/djob/errors"
	"github.com/HZ89/djob/log"
)

// agent config
type Config struct {
	Server            bool              // start as server or just agent
	LoadJobPolicy     loadJobPolicy     // the policy for loading jobs from the database at server startup
	Region            string            // agent region
	Nodename          string            //serf node name
	Tags              map[string]string // agent tags used to filter agent
	SerfBindIP        string            // serf system bind address ip
	SerfBindPort      int
	SerfAdvertiseIP   string // serf system advertise address ip used for agent behind a firewall
	SerfAdvertisePort int
	SerfJoin          []string
	APIBindIP         string // HTTP API PORT just effect when server is true
	APIBindPort       int
	JobStore          string   // key value storage type etd/zookeeper etc.
	JobStoreServers   []string // k/v storage server list [ip:port]
	JobStoreKeyspace  string   // keyspace in the storage
	encryptKey        string   // serf encrypt key
	LogLevel          string   // info debug error
	LogFile           string   // log file path
	RPCTls            bool     // grpc enable tls or not
	CAFile            string   // tls ca file used in agent
	KeyFile           string   // key file used in server
	CertFile          string   // cert file used in server
	RPCBindIP         string   // grcp bind addr ip
	RPCBindPort       int
	RPCAdcertiseIP    string // sames to serf advertise addr
	RPCAdcertisePort  int
	SerfSnapshotPath  string //serf use this path to save snapshot of joined server
	DSN               string
	APITokens         map[string]string
}

const (
	DefaultRegion       = "MARS"
	DefaultSerfPort     = 8998
	DefaultHttpPort     = 8088
	DefaultRPCPort      = 7979
	DefaultSnapshotPath = "./snapshot"
	DefaultConfigFile   = "./conf/djob.yml"
	DefaultPidFile      = "./djob.pid"
	DefaultLogFile      = "./djob.log"
	DefaultKeySpeace    = "djob"
	DefaultSQLPort      = 3306
	DefaultSQLHost      = "localhost"
	DefaultSQLUser      = "djob"
	DefaultDBName       = "djob"
)

type loadJobPolicy int

// load job policy
const (
	LOADNOTHING = 1 + iota // load nothing
	LOADOWN                // load all the jobs belong to this server
	LOADALL                // load all unassigned jobs
)

var policyValueName = map[int]string{
	1: "nothing",
	2: "own",
	3: "all",
}

var policyNameValue = map[string]int{
	"nothing": 1,
	"own":     2,
	"all":     3,
}

func (p loadJobPolicy) String() string {
	return policyValueName[int(p)]
}

func stringToLoadJobPolicy(k string) (loadJobPolicy, bool) {
	if v, ok := policyNameValue[k]; ok {
		return loadJobPolicy(v), true
	}
	return 0, false
}

// prepare the default value in the configure
func newConfig(args []string, version string) (*Config, error) {

	cmdFlags := flag.NewFlagSet("agent", flag.ContinueOnError)
	cmdFlags.String("config", DefaultConfigFile, "config file path")
	cmdFlags.String("pid", DefaultPidFile, "pid file path")
	cmdFlags.String("logfile", DefaultLogFile, "log file path")

	if err := cmdFlags.Parse(args); err != nil {
		return nil, err
	}
	viper.SetConfigFile(cmdFlags.Lookup("config").Value.String())
	viper.SetDefault("serf_bind_addr", fmt.Sprintf("%s:%d", "0.0.0.0", DefaultSerfPort))
	viper.SetDefault("http_api_addr", fmt.Sprintf("%s:%d", "0.0.0.0", DefaultHttpPort))
	viper.SetDefault("rpc_bind_addr", fmt.Sprintf("%s:%d", "0.0.0.0", DefaultRPCPort))
	viper.SetDefault("job_store", "etcd")
	viper.SetDefault("job_store_keyspeace", DefaultKeySpeace)
	viper.SetDefault("log_level", "info")
	viper.SetDefault("rpc_tls", false)
	viper.SetDefault("region", DefaultRegion)
	viper.SetDefault("server", false)
	viper.SetDefault("load_job_policy", fmt.Sprintf("%s", LOADOWN))
	viper.SetDefault("serf_snapshot_dir", DefaultSnapshotPath)
	viper.SetDefault("pid", cmdFlags.Lookup("pid").Value.String())
	viper.SetDefault("logfile", cmdFlags.Lookup("logfile").Value.String())
	viper.SetDefault("sql_port", DefaultSQLPort) // sql backend used to save execution
	viper.SetDefault("sql_host", DefaultSQLHost)
	viper.SetDefault("sql_user", DefaultSQLUser)
	viper.SetDefault("sql_dbname", DefaultDBName)

	return ReadConfig(version)
}

// read the configuration file
func ReadConfig(version string) (*Config, error) {
	err := viper.ReadInConfig()
	if err != nil {
		return nil, err
	}
	tags := viper.GetStringMapString("tags")
	server := viper.GetBool("server")
	nodeName := viper.GetString("node")
	if nodeName == "" {
		nodeName, err = os.Hostname()
		if err != nil {
			return nil, err
		}
	}

	tokens := make(map[string]string)
	var policyType loadJobPolicy

	if server {
		var ok bool
		loadpolicy := viper.GetString("load_job_policy")
		policyType, ok = stringToLoadJobPolicy(loadpolicy)
		if !ok {
			policyType = LOADOWN
		}
		tags["server"] = "true"
		tokens = viper.GetStringMapString("tokens")
		if len(tokens) == 0 {
			tokens = map[string]string{
				"defualt": "djob-token",
			}
		}
	}

	tags["version"] = version
	tags["node"] = nodeName
	tags["region"] = viper.GetString("region")

	withTls := viper.GetBool("rpc_tls")
	keyFile := viper.GetString("rpc_key_file")
	certFile := viper.GetString("rpc_cert_file")
	caFile := viper.GetString("rpc_ca_file")
	if withTls {
		if server {
			if keyFile == "" || certFile == "" {
				return nil, errors.ErrMissKeyFile
			}
		} else {
			if caFile == "" {
				return nil, errors.ErrMissCaFile
			}
		}
	}
	// init logger
	log.InitLogger(viper.GetString("log_level"), nodeName, viper.GetString("logfile"))

	SerfBindIP, serfBindport, err := splitNetAddr(viper.GetString("serf_bind_addr"), DefaultSerfPort)
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

	serfAdip, serfAdport, err := handleAdvertise(viper.GetString("serf_advertise_addr"), fmt.Sprintf("%s:%d", SerfBindIP, serfBindport))
	if err != nil {
		return nil, err
	}

	rpcAdip, rpcAdport, err := handleAdvertise(viper.GetString("rpc_advertise_addr"), fmt.Sprintf("%s:%d", rpcBindip, rpcBindport))
	if err != nil {
		return nil, err
	}
	tags["RPCADIP"] = rpcAdip
	tags["RPCADPORT"] = fmt.Sprintf("%d", rpcAdport)

	sqlPasswd := viper.GetString("sql_password")
	unixSocket := viper.GetString("sql_unix")
	sqlAddr := fmt.Sprintf("%s:%d", viper.GetString("sql_host"), viper.GetInt("sql_port"))
	dbName := viper.GetString("sql_dbname")
	dsn := viper.GetString("sql_user")

	if sqlPasswd != "" {
		dsn = dsn + ":" + sqlPasswd
	}
	if unixSocket != "" {
		dsn = dsn + "@unix(" + unixSocket + ")"
	} else {
		dsn = dsn + "@tcp(" + sqlAddr + ")"
	}
	dsn = dsn + "/" + dbName + "?parseTime=true&loc=Local&autocommit=1"

	return &Config{
		Server:            server,
		LoadJobPolicy:     policyType,
		Region:            tags["region"],
		Tags:              tags,
		SerfBindIP:        SerfBindIP,
		SerfBindPort:      serfBindport,
		APIBindIP:         apiBindip,
		APIBindPort:       apiBindport,
		RPCBindIP:         rpcBindip,
		RPCBindPort:       rpcBindport,
		SerfAdvertiseIP:   serfAdip,
		SerfAdvertisePort: serfAdport,
		RPCAdcertiseIP:    rpcAdip,
		RPCAdcertisePort:  rpcAdport,
		LogLevel:          viper.GetString("log_level"),
		LogFile:           viper.GetString("logfile"),
		CAFile:            caFile,
		CertFile:          certFile,
		KeyFile:           keyFile,
		RPCTls:            withTls,
		JobStore:          viper.GetString("job_store"),
		JobStoreServers:   viper.GetStringSlice("job_store_servers"),
		JobStoreKeyspace:  viper.GetString("job_store_keyspeace"),
		SerfJoin:          viper.GetStringSlice("join"),
		Nodename:          nodeName,
		encryptKey:        viper.GetString("encrypt_key"),
		SerfSnapshotPath:  viper.GetString("serf_snapshot_dir"),
		DSN:               dsn,
		APITokens:         tokens,
	}, nil
}

// base64 encode encrypt key
func (c *Config) EncryptKey() ([]byte, error) {
	return base64.StdEncoding.DecodeString(c.encryptKey)
}

// make sure advertise ip or port is not 0.0.0.0 or 0
func handleAdvertise(oaddr string, daddr string) (string, int, error) {
	if oaddr == "" {
		oaddr = daddr
	}
	addr, err := net.ResolveTCPAddr("tcp", oaddr)
	// if the gived ip port is invalid, get it from the local port
	if err != nil {
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

// if ip is a private network address, return true
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
	}
	if err != nil {
		return "", 0, err
	}

	addr, err := net.ResolveTCPAddr("tcp", address)
	if err != nil {
		return "", 0, err
	}
	return addr.IP.String(), addr.Port, nil

}
