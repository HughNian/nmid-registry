package option

import (
	"fmt"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"net/url"
	"strings"
)

type Options struct {
	Flags   *pflag.FlagSet
	Viper   *viper.Viper
	YamlStr string

	// meta
	Name    string            `yaml:"name" env:"NMID_NAME"`
	Labels  map[string]string `yaml:"labels" env:"NMID_LABELS"`
	APIAddr string            `yaml:"api-addr"`

	//cluster options
	ClusterDebug                    bool           `yaml:"cluster-debug"`
	ClusterName                     string         `yaml:"cluster-name"`
	ClusterRole                     string         `yaml:"cluster-role"`
	ClusterRequestTimeout           string         `yaml:"cluster-request-timeout"`
	ClusterListenClientURLs         []string       `yaml:"cluster-listen-client-urls"`
	ClusterListenPeerURLs           []string       `yaml:"cluster-listen-peer-urls"`
	ClusterAdvertiseClientURLs      []string       `yaml:"cluster-advertise-client-urls"`
	ClusterInitialAdvertisePeerURLs []string       `yaml:"cluster-initial-advertise-peer-urls"`
	ClusterJoinURLs                 []string       `yaml:"cluster-join-urls"`
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
	ListenPeerURLs           []string          `yaml:"listen-peer-urls"`
	ListenClientURLs         []string          `yaml:"listen-client-urls"`
	AdvertiseClientURLs      []string          `yaml:"advertise-client-urls"`
	InitialAdvertisePeerURLs []string          `yaml:"initial-advertise-peer-urls"`
	InitialCluster           map[string]string `yaml:"initial-cluster"`
	StateFlag                string            `yaml:"state-flag"`
	SlaveListenPeerURLs      []string          `yaml:"slave-listen-peer-urls"`
	MaxCallSendMsgSize       int               `yaml:"max-call-send-msg-size"`
}

func (opt *Options) GetPeerUrls() []string {
	if opt.ClusterRole == "slave" {
		if len(opt.ClusterJoinURLs) != 0 {
			return opt.ClusterJoinURLs
		}
		return opt.Cluster.SlaveListenPeerURLs
	}

	peerUrls := make([]string, 0)
	for _, peerUrl := range opt.Cluster.InitialCluster {
		peerUrls = append(peerUrls, peerUrl)
	}
	return peerUrls
}

func (opt *Options) InitialCluster2String() string {
	ss := make([]string, 0)
	for name, peerURL := range opt.Cluster.InitialCluster {
		ss = append(ss, fmt.Sprintf("%s=%s", name, peerURL))
	}
	return strings.Join(ss, ",")
}

func (opt *Options) IsUseInitialCluster() bool {
	return len(opt.Cluster.InitialCluster) > 0
}

func ParseUrls(urlStr []string) ([]url.URL, error) {
	urls := make([]url.URL, len(urlStr))

	for i, urlval := range urlStr {
		parsedURL, err := url.Parse(urlval)
		if err != nil {
			return nil, fmt.Errorf(" %s: %v", urlval, err)
		}
		urls[i] = *parsedURL
	}

	return urls, nil
}
