package cluster

import (
	"fmt"
	"go.etcd.io/etcd/server/v3/embed"
	"nmid-registry/pkg/option"
	"nmid-registry/pkg/utils"
	"path/filepath"
)

const (
	SnapshotNum      = 5000
	MaxRequestSize   = 8 * 1024 * 1024        // 8MB
	QuotaBackendSize = 8 * 1024 * 1024 * 1024 // 8GB
	MaxTxnOps        = 10240
	LogFileName      = "etcd_server.log"
)

func CreateEtcdConfig(opt *option.Options) (*embed.Config, error) {
	config := embed.NewConfig()

	clientUrls, err := option.ParseUrls(opt.Cluster.ListenClientURLs)
	if err != nil {
		return nil, err
	}
	peerUrls, err := option.ParseUrls(opt.Cluster.ListenPeerURLs)
	if err != nil {
		return nil, err
	}
	adClientUrls, err := option.ParseUrls(opt.Cluster.AdvertiseClientURLs)
	if err != nil {
		return nil, err
	}
	adPeerURLs, err := option.ParseUrls(opt.Cluster.InitialAdvertisePeerURLs)
	if err != nil {
		return nil, err
	}

	config.EnableV2 = false
	config.Name = opt.Name
	config.Dir = opt.AbsDataDir
	config.InitialClusterToken = opt.ClusterName
	config.LCUrls = clientUrls
	config.ACUrls = adClientUrls
	config.LPUrls = peerUrls
	config.APUrls = adPeerURLs
	config.AutoCompactionMode = embed.CompactorModeRevision
	config.AutoCompactionRetention = "10"
	config.QuotaBackendBytes = QuotaBackendSize
	config.MaxTxnOps = MaxTxnOps
	config.MaxRequestBytes = MaxRequestSize
	config.SnapshotCount = SnapshotNum
	config.Logger = "zap"
	config.LogOutputs = []string{utils.GOOSPath(filepath.Join(opt.AbsLogDir, LogFileName))}
	config.ClusterState = embed.ClusterStateFlagNew
	if opt.Cluster.StateFlag == "existing" {
		config.ClusterState = embed.ClusterStateFlagExisting
	}
	config.InitialCluster = opt.InitialCluster2String()

	return config, nil
}

func CreateEtcdConfigAddMember(opt *option.Options, members *Members) (*embed.Config, error) {
	config := embed.NewConfig()

	clientUrls, err := option.ParseUrls(opt.Cluster.ListenClientURLs)
	if err != nil {
		return nil, err
	}
	peerUrls, err := option.ParseUrls(opt.Cluster.ListenPeerURLs)
	if err != nil {
		return nil, err
	}
	adClientUrls, err := option.ParseUrls(opt.Cluster.AdvertiseClientURLs)
	if err != nil {
		return nil, err
	}
	adPeerURLs, err := option.ParseUrls(opt.Cluster.InitialAdvertisePeerURLs)
	if err != nil {
		return nil, err
	}

	config.EnableV2 = false
	config.Name = opt.Name
	config.Dir = opt.AbsDataDir
	config.InitialClusterToken = opt.ClusterName
	config.LCUrls = clientUrls
	config.ACUrls = adClientUrls
	config.LPUrls = peerUrls
	config.APUrls = adPeerURLs
	config.AutoCompactionMode = embed.CompactorModeRevision
	config.AutoCompactionRetention = "10"
	config.QuotaBackendBytes = QuotaBackendSize
	config.MaxTxnOps = MaxTxnOps
	config.MaxRequestBytes = MaxRequestSize
	config.SnapshotCount = SnapshotNum
	config.Logger = "zap"
	config.LogOutputs = []string{utils.GOOSPath(filepath.Join(opt.AbsLogDir, LogFileName))}

	if len(opt.ClusterJoinURLs) == 0 {
		if members.ClusterMembersLen() == 1 && utils.IsDirEmpty(opt.AbsDataDir) {
			config.ClusterState = embed.ClusterStateFlagNew
		}
	} else if members.ClusterMembersLen() == 1 {
		return nil, fmt.Errorf("join mode with only one cluster member: %v", *members.ClusterMembers)
	}
	config.InitialCluster = members.InitCluster2String()

	return config, nil
}
