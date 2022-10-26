package option

import (
	"fmt"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	yaml "gopkg.in/yaml.v2"
	"net"
	"net/url"
	"nmid-registry/pkg/utils"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	VERSION = "0.0.1"
)

type Options struct {
	flags   *pflag.FlagSet
	viper   *viper.Viper
	yamlStr string

	//from command line only.
	ShowVersion     bool   `yaml:"-"`
	ShowHelp        bool   `yaml:"-"`
	ShowConfig      bool   `yaml:"-"`
	ConfigFile      string `yaml:"-"`
	ForceNewCluster bool   `yaml:"-"`
	SignalUpgrade   bool   `yaml:"-"`

	// meta
	WriteOnly                bool              `yaml:"write-only"`
	Name                     string            `yaml:"name" env:"NMIDR_NAME"`
	Labels                   map[string]string `yaml:"labels" env:"NMIDR_LABELS"`
	ApiAddr                  string            `yaml:"api-addr"`
	DisableAccessLog         bool              `yaml:"disable-access-log"`
	InitialObjectConfigFiles []string          `yaml:"initial-object-config-files"`
	ApiTimeout               time.Duration     `yaml:"api-timeout"`
	ApiReadTimeout           time.Duration     `yaml:"api-read-timeout"`
	ApiWriteTimeout          time.Duration     `yaml:"api-write-timeout"`

	//cluster options
	UseStandEtcd                    bool           `yaml:"use-stand-etcd"`
	ClusterDebug                    bool           `yaml:"cluster-debug"`
	ClusterName                     string         `yaml:"cluster-name"`
	ClusterRole                     string         `yaml:"cluster-role"`
	ClusterRequestTimeout           string         `yaml:"cluster-request-timeout"`
	ClusterListenClientUrls         []string       `yaml:"cluster-listen-client-Urls"`
	ClusterListenPeerUrls           []string       `yaml:"cluster-listen-peer-Urls"`
	ClusterAdvertiseClientUrls      []string       `yaml:"cluster-advertise-client-Urls"`
	ClusterInitialAdvertisePeerUrls []string       `yaml:"cluster-initial-advertise-peer-Urls"`
	ClusterJoinUrls                 []string       `yaml:"cluster-join-Urls"`
	Cluster                         ClusterOptions `yaml:"cluster"`

	// path
	HomeDir   string `yaml:"home-dir"`
	DataDir   string `yaml:"data-dir"`
	WALDir    string `yaml:"wal-dir"`
	LogDir    string `yaml:"log-dir"`
	MemberDir string `yaml:"member-dir"`

	// items below in advance
	AbsHomeDir   string `yaml:"-"`
	AbsDataDir   string `yaml:"-"`
	AbsWALDir    string `yaml:"-"`
	AbsLogDir    string `yaml:"-"`
	AbsMemberDir string `yaml:"-"`
}

// ClusterOptions defines the cluster members.
type ClusterOptions struct {
	ListenPeerUrls           []string          `yaml:"listen-peer-Urls"`
	ListenClientUrls         []string          `yaml:"listen-client-Urls"`
	AdvertiseClientUrls      []string          `yaml:"advertise-client-Urls"`
	InitialAdvertisePeerUrls []string          `yaml:"initial-advertise-peer-Urls"`
	InitialCluster           map[string]string `yaml:"initial-cluster"`
	StateFlag                string            `yaml:"state-flag"`
	MasterListenPeerUrls     []string          `yaml:"master-listen-peer-Urls"`
	MaxCallSendMsgSize       int               `yaml:"max-call-send-msg-size"`
}

func New() *Options {
	opt := &Options{
		flags: pflag.NewFlagSet(os.Args[0], pflag.ContinueOnError),
		viper: viper.New(),
	}

	opt.flags.BoolVarP(&opt.ShowVersion, "version", "v", false, "Print the version and exit.")
	opt.flags.BoolVarP(&opt.ShowHelp, "help", "h", false, "Print the helper message and exit.")
	opt.flags.BoolVarP(&opt.ShowConfig, "print-config", "c", false, "Print the configuration.")
	opt.flags.StringVarP(&opt.ConfigFile, "config-file", "f", "", "Load server configuration from a file(yaml format), other command line flags will be ignored if specified.")
	opt.flags.BoolVar(&opt.ForceNewCluster, "force-new-cluster", false, "Force to create a new one-member cluster.")
	opt.flags.BoolVar(&opt.SignalUpgrade, "signal-upgrade", false, "Send an upgrade signal to the server based on the local pid file, then exit. The original server will start a graceful upgrade after signal received.")
	opt.flags.StringVar(&opt.Name, "name", "nmidr-default-name", "Human-readable name for this member.")
	opt.flags.StringToStringVar(&opt.Labels, "labels", nil, "The labels for the instance of Nmid-registry.")
	opt.flags.BoolVar(&opt.UseStandEtcd, "use-stand-etcd", false, "Use standalone etcd instead of embedded .")
	addClusterVars(opt)
	opt.flags.StringVar(&opt.ApiAddr, "api-addr", "localhost:2381", "Address([host]:port) to listen on for administration traffic.")
	opt.flags.BoolVar(&opt.ClusterDebug, "cluster-debug", false, "Flag to set lowest log level from INFO downgrade DEBUG.")
	opt.flags.StringSliceVar(&opt.InitialObjectConfigFiles, "initial-object-config-files", nil, "List of configuration files for initial objects, these objects will be created at startup if not already exist.")

	opt.flags.StringVar(&opt.HomeDir, "home-dir", "./", "Path to the home directory.")
	opt.flags.StringVar(&opt.DataDir, "data-dir", "data", "Path to the data directory.")
	opt.flags.StringVar(&opt.WALDir, "wal-dir", "", "Path to the WAL directory.")
	opt.flags.StringVar(&opt.LogDir, "log-dir", "log", "Path to the log directory.")
	opt.flags.StringVar(&opt.MemberDir, "member-dir", "member", "Path to the member directory.")

	opt.viper.BindPFlags(opt.flags)

	return opt
}

func (opt *Options) YAML() string {
	return opt.yamlStr
}

func addClusterVars(opt *Options) {
	opt.flags.StringVar(&opt.ClusterName, "cluster-name", "eg-cluster-default-name", "Human-readable name for the new cluster, ignored while joining an existed cluster.")
	opt.flags.StringVar(&opt.ClusterRole, "cluster-role", "master", "Cluster role for this member (master, slave).")
	opt.flags.StringVar(&opt.ClusterRequestTimeout, "cluster-request-timeout", "10s", "Timeout to handle request in the cluster.")

	// Deprecated: Use 'Cluster connection configuration' instead.
	opt.flags.StringSliceVar(&opt.ClusterListenClientUrls, "cluster-listen-client-Urls", []string{"http://localhost:2379"}, "Deprecated. Use cluster.listen-client-Urls instead.")
	opt.flags.StringSliceVar(&opt.ClusterListenPeerUrls, "cluster-listen-peer-Urls", []string{"http://localhost:2380"}, "Deprecated. Use cluster.listen-peer-Urls instead.")
	opt.flags.StringSliceVar(&opt.ClusterAdvertiseClientUrls, "cluster-advertise-client-Urls", []string{"http://localhost:2379"}, "Deprecated. Use cluster.advertise-client-Urls instead.")
	opt.flags.StringSliceVar(&opt.ClusterInitialAdvertisePeerUrls, "cluster-initial-advertise-peer-Urls", []string{"http://localhost:2380"}, "Deprecated. Use cluster.initial-advertise-peer-Urls instead.")
	opt.flags.StringSliceVar(&opt.ClusterJoinUrls, "cluster-join-Urls", nil, "Deprecated. Use cluster.initial-cluster instead.")

	// Cluster connection configuration
	opt.flags.StringSliceVar(&opt.Cluster.ListenClientUrls, "listen-client-Urls", []string{"http://localhost:2379"}, "List of Urls to listen on for cluster client traffic.")
	opt.flags.StringSliceVar(&opt.Cluster.ListenPeerUrls, "listen-peer-Urls", []string{"http://localhost:2380"}, "List of Urls to listen on for cluster peer traffic.")
	opt.flags.StringSliceVar(&opt.Cluster.AdvertiseClientUrls, "advertise-client-Urls", []string{"http://localhost:2379"}, "List of this member's client Urls to advertise to the rest of the cluster.")
	opt.flags.StringSliceVar(&opt.Cluster.InitialAdvertisePeerUrls, "initial-advertise-peer-Urls", []string{"http://localhost:2380"}, "List of this member's peer Urls to advertise to the rest of the cluster.")
	opt.flags.StringToStringVarP(&opt.Cluster.InitialCluster, "initial-cluster", "", nil,
		"List of (member name, Url) pairs that will form the cluster. E.g. master-1=http://localhost:2380.")
	opt.flags.StringVar(&opt.Cluster.StateFlag, "state-flag", "new", "Cluster state (new, existing)")
	opt.flags.StringSliceVar(&opt.Cluster.MasterListenPeerUrls,
		"master-listen-peer-Urls",
		[]string{"http://localhost:2380"},
		"List of peer Urls of master members. Define this only, when cluster-role is secondary.")
	opt.flags.IntVar(&opt.Cluster.MaxCallSendMsgSize, "max-call-send-msg-size", 10*1024*1024, "Maximum size in bytes for cluster synchronization messages.")
}

func (opt *Options) Parse() (string, error) {
	err := opt.flags.Parse(os.Args[1:])
	if err != nil {
		return "", err
	}

	if opt.ShowVersion {
		return version(), nil
	}

	if opt.ShowHelp {
		return opt.flags.FlagUsages(), nil
	}

	opt.viper.AutomaticEnv()
	opt.viper.SetEnvPrefix("NR")
	opt.viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))

	if opt.ConfigFile != "" {
		opt.viper.SetConfigFile(opt.ConfigFile)
		opt.viper.SetConfigType("yaml")
		err := opt.viper.ReadInConfig()
		if err != nil {
			return "", fmt.Errorf("read config file %s failed: %v",
				opt.ConfigFile, err)
		}
	}

	// NOTE: Workaround because viper does not treat env vars the same as other config.
	// Reference: https://github.com/spf13/viper/issues/188#issuecomment-399518663
	for _, key := range opt.viper.AllKeys() {
		val := opt.viper.Get(key)
		// NOTE: We need to handle map[string]string
		// Reference: https://github.com/spf13/viper/issues/911
		if key == "labels" {
			val = opt.viper.GetStringMapString(key)
		}
		opt.viper.Set(key, val)
	}

	err = opt.viper.Unmarshal(opt, func(c *mapstructure.DecoderConfig) {
		c.TagName = "yaml"
	})
	if err != nil {
		return "", fmt.Errorf("yaml file unmarshal failed, please make sure you provide valid yaml file, %v", err)
	}

	if opt.UseStandEtcd {
		opt.ClusterRole = "slave" // when using external stand etcd, the cluster role cannot be "slave"
	}
	if opt.ClusterRole == "master" && len(opt.Cluster.InitialCluster) == 0 {
		opt.Cluster.InitialCluster = map[string]string{opt.Name: opt.Cluster.InitialAdvertisePeerUrls[0]}
	}

	err = opt.verification()
	if err != nil {
		return "", err
	}

	err = opt.initDir()
	if err != nil {
		return "", err
	}

	opt.adjust()

	buff, err := yaml.Marshal(opt)
	if err != nil {
		return "", fmt.Errorf("marshal config to yaml failed: %v", err)
	}
	opt.yamlStr = string(buff)

	if opt.ShowConfig {
		fmt.Printf("%s", opt.yamlStr)
	}

	return "", nil
}

func (opt *Options) GetPeerUrls() []string {
	if opt.ClusterRole == "slave" {
		if len(opt.ClusterJoinUrls) != 0 {
			return opt.ClusterJoinUrls
		}
		return opt.Cluster.MasterListenPeerUrls
	}

	peerUrls := make([]string, 0)
	for _, peerUrl := range opt.Cluster.InitialCluster {
		peerUrls = append(peerUrls, peerUrl)
	}
	return peerUrls
}

func (opt *Options) InitialCluster2String() string {
	ss := make([]string, 0)
	for name, peerUrl := range opt.Cluster.InitialCluster {
		ss = append(ss, fmt.Sprintf("%s=%s", name, peerUrl))
	}

	return strings.Join(ss, ",")
}

func (opt *Options) IsUseInitialCluster() bool {
	return len(opt.Cluster.InitialCluster) > 0
}

func (opt *Options) GetFirstAdvertiseClientUrl() (string, error) {
	if opt.IsUseInitialCluster() {
		if len(opt.Cluster.AdvertiseClientUrls) == 0 {
			return "", fmt.Errorf("cluster.advertise-client-Urls is empty")
		}
		return opt.Cluster.AdvertiseClientUrls[0], nil
	}
	if len(opt.ClusterAdvertiseClientUrls) == 0 {
		return "", fmt.Errorf("cluster-advertise-client-Urls is empty")
	}
	return opt.ClusterAdvertiseClientUrls[0], nil
}

func (opt *Options) verification() error {
	if opt.ClusterName == "" {
		return fmt.Errorf("cluster name empty")
	} else if err := utils.ValidateName(opt.ClusterName); err != nil {
		return err
	}

	if len(opt.ClusterJoinUrls) != 0 {
		if _, err := ParseUrls(opt.ClusterJoinUrls); err != nil {
			return fmt.Errorf("invalid cluster-join-Urls %v", err)
		}
	}

	switch opt.ClusterRole {
	case "slave":
		if opt.ForceNewCluster {
			return fmt.Errorf("slave got force-new-cluster")
		}
		if len(opt.Cluster.MasterListenPeerUrls) == 0 && len(opt.ClusterJoinUrls) == 0 {
			return fmt.Errorf("slave got empty cluster.slave-listen-peer-urls and cluster-join-urls entries")
		}
	case "master":
		argumentsToValidate := map[string][]string{
			"cluster-listen-client-urls":          opt.ClusterListenClientUrls,
			"cluster-listen-peer-urls":            opt.ClusterListenPeerUrls,
			"cluster-advertise-client-urls":       opt.ClusterAdvertiseClientUrls,
			"cluster-initial-advertise-peer-urls": opt.ClusterInitialAdvertisePeerUrls,
		}

		if opt.IsUseInitialCluster() {
			argumentsToValidate = map[string][]string{
				"listen-client-urls":          opt.Cluster.ListenClientUrls,
				"listen-peer-urls":            opt.Cluster.ListenPeerUrls,
				"advertise-client-urls":       opt.Cluster.AdvertiseClientUrls,
				"initial-advertise-peer-urls": opt.Cluster.InitialAdvertisePeerUrls,
			}
			initialClusterUrls := make([]string, 0, len(opt.Cluster.InitialCluster))
			for _, value := range opt.Cluster.InitialCluster {
				initialClusterUrls = append(initialClusterUrls, value)
			}
			if _, err := ParseUrls(initialClusterUrls); err != nil {
				return fmt.Errorf("invalid initial-cluster: %v", err)
			}
		}
		for arg, urls := range argumentsToValidate {
			if len(urls) == 0 {
				return fmt.Errorf("empty %s", arg)
			}
			if _, err := ParseUrls(urls); err != nil {
				return fmt.Errorf("invalid %s: %v", arg, err)
			}
		}
	default:
		return fmt.Errorf("invalid cluster-role supported roles are master/slave")
	}

	_, err := time.ParseDuration(opt.ClusterRequestTimeout)
	if err != nil {
		return fmt.Errorf("invalid cluster-request-timeout %v", err)
	}

	_, _, err = net.SplitHostPort(opt.ApiAddr)
	if err != nil {
		return fmt.Errorf("invalid api-addr %v", err)
	}
	if err != nil {
		return fmt.Errorf("invalid api-url %v", err)
	}

	// dirs
	if opt.HomeDir == "" {
		return fmt.Errorf("empty home-dir")
	}
	if opt.DataDir == "" {
		return fmt.Errorf("empty data-dir")
	}
	if opt.LogDir == "" {
		return fmt.Errorf("empty log-dir")
	}
	if !opt.IsUseInitialCluster() && opt.MemberDir == "" {
		return fmt.Errorf("empty member-dir")
	}

	// meta
	if opt.Name == "" {
		name, err := utils.GetMemberName(opt.ApiAddr)
		if err != nil {
			return err
		}
		opt.Name = name
	}
	if err := utils.ValidateName(opt.Name); err != nil {
		return err
	}

	return nil
}

func (opt *Options) initDir() error {
	abs, isAbs, clean, join := filepath.Abs, filepath.IsAbs, filepath.Clean, filepath.Join
	if isAbs(opt.HomeDir) {
		opt.AbsHomeDir = clean(opt.HomeDir)
	} else {
		absHomeDir, err := abs(opt.HomeDir)
		if err != nil {
			return err
		}
		opt.AbsHomeDir = absHomeDir
	}

	type dirItem struct {
		dir    string
		absDir *string
	}
	table := []dirItem{
		{dir: opt.DataDir, absDir: &opt.AbsDataDir},
		{dir: opt.WALDir, absDir: &opt.AbsWALDir},
		{dir: opt.LogDir, absDir: &opt.AbsLogDir},
		{dir: opt.MemberDir, absDir: &opt.AbsMemberDir},
	}
	for _, di := range table {
		if di.dir == "" {
			continue
		}
		if filepath.IsAbs(di.dir) {
			*di.absDir = clean(di.dir)
		} else {
			*di.absDir = clean(join(opt.AbsHomeDir, di.dir))
		}
	}

	return nil
}

func (opt *Options) adjust() {
	if opt.ClusterRole != "master" || opt.IsUseInitialCluster() {
		return
	}
	if len(opt.ClusterJoinUrls) == 0 {
		return
	}

	joinURL := opt.ClusterJoinUrls[0]

	for _, peerURL := range opt.ClusterInitialAdvertisePeerUrls {
		if strings.EqualFold(joinURL, peerURL) {
			fmt.Printf("cluster-join-urls %v changed to empty because it tries to join itself\n",
				opt.ClusterJoinUrls)
			opt.ClusterJoinUrls = nil
		}
	}
}

func ParseUrls(UrlStr []string) ([]url.URL, error) {
	Urls := make([]url.URL, len(UrlStr))

	for i, Urlval := range UrlStr {
		parsedUrl, err := url.Parse(Urlval)
		if err != nil {
			return nil, fmt.Errorf(" %s: %v", Urlval, err)
		}
		Urls[i] = *parsedUrl
	}

	return Urls, nil
}

func version() string {
	return fmt.Sprintf("nmid-registry version:%s", VERSION)
}

func AboutMe() string {
	return fmt.Sprintf(`Copyright© 2021 - %d Nmid-Registry(http://www.niansong.top). Nmid-Registry is Nmid Register Center All rights reserved.Apache License 2.0.`, time.Now().Year())
}
