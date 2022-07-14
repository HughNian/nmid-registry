package cluster

import (
	"bytes"
	"fmt"
	"go.etcd.io/etcd/api/v3/etcdserverpb"
	"gopkg.in/yaml.v2"
	"nmid-registry/pkg/loger"
	"nmid-registry/pkg/option"
	"nmid-registry/pkg/utils"
	"os"
	"sort"
	"strings"
	"sync"
)

const (
	MembersFilename       = "members.yaml"
	MembersBackupFilename = "members.bak.yaml"

	StatusMemberPrefix = "/status/members/"
	StatusMemberFormat = "/status/members/%s" // +memberName
)

type (
	Members struct {
		sync.RWMutex

		Options *option.Options

		lastData             []byte
		dataFile, backupFile string

		ClusterMembers *membersArr
		KnownMembers   *membersArr
	}

	membersArr []*member

	member struct {
		ID      uint64 `yaml:"id"`
		Name    string `yaml:"name"`
		PeerUrl string `yaml:"peerUrl"`
	}

	MemberStatus struct {
		Options option.Options `yaml:"options"`

		// RFC3339 format
		LastHeartbeatTime string `yaml:"lastHeartbeatTime"`

		LastDefragTime string `yaml:"lastDefragTime,omitempty"`

		//only if it's cluster status is master.
		CStatus *ClusterStatus `yaml:"etcd,omitempty"`
	}
)

func NewMembers(opt *option.Options) (*Members, error) {
	mem := &Members{
		Options:        opt,
		ClusterMembers: &membersArr{},
		KnownMembers:   &membersArr{},
	}
	mem.initializeMembers(opt)
	err := mem.loadFileData()
	if nil != err {
		return nil, err
	}

	return mem, nil
}

//initializeMembers adds first member to ClusterMembers and all members to KnownMembers.
func (m *Members) initializeMembers(opt *option.Options) {
	initMA := make(membersArr, 0)
	if opt.ClusterRole == "master" && len(opt.ClusterInitialAdvertisePeerURLs) > 0 {
		initMA = append(initMA, &member{
			Name:    opt.Name,
			PeerUrl: opt.ClusterInitialAdvertisePeerURLs[0],
		})
	}
	m.ClusterMembers.update(initMA)

	//Add all members to list of known members
	if len(opt.ClusterJoinURLs) > 0 {
		for _, peerUrl := range opt.ClusterJoinURLs {
			initMA = append(initMA, &member{
				PeerUrl: peerUrl,
			})
		}
	}
	m.KnownMembers.update(initMA)
}

func (m *Members) GetSelf() *member {
	m.Lock()
	defer m.Unlock()

	cm := m.ClusterMembers.GetMemberByName(m.Options.Name)
	if cm != nil {
		return cm
	}

	km := m.KnownMembers.GetMemberByName(m.Options.Name)
	if km != nil {
		return km
	}

	if m.Options.ClusterRole == `slave` {
		loger.Loger.Errorf("slave role error")
	}

	peerURL := ""
	if len(m.Options.ClusterInitialAdvertisePeerURLs) != 0 {
		peerURL = m.Options.ClusterInitialAdvertisePeerURLs[0]
	}

	return &member{
		Name:    m.Options.Name,
		PeerUrl: peerURL,
	}
}

func (m *Members) ClusterMembersLen() int {
	m.RLock()
	defer m.RUnlock()
	return m.ClusterMembers.Len()
}

func (m *Members) InitCluster2String() string {
	m.RLock()
	defer m.RUnlock()

	return m.ClusterMembers.initCluster2String()
}

func (m *Members) KnownPeerUrls() []string {
	m.RLock()
	defer m.RUnlock()

	return m.KnownMembers.peerUrls()
}

func (m *Members) selfWithoutID() *member {
	s := m.GetSelf()
	s.ID = 0
	return s
}

func (m *Members) loadFileData() error {
	if !utils.FileExist(m.dataFile) {
		return nil
	}

	data, err := os.ReadFile(m.dataFile)
	if nil != err {
		return err
	}

	loadMembers := &Members{}
	err = yaml.Unmarshal(data, loadMembers)
	if nil != err {
		return err
	}

	m.ClusterMembers.update(*loadMembers.ClusterMembers)
	m.KnownMembers.update(*loadMembers.KnownMembers)

	return nil
}

func (m *Members) storeFileData() {
	data, err := yaml.Marshal(m)
	if err != nil {
		loger.Loger.Errorf("store file data get yaml of %#v failed: %v", m.KnownMembers, err)
	}
	if bytes.Equal(m.lastData, data) {
		return
	}

	//backup datafile
	if utils.FileExist(m.dataFile) {
		err := os.Rename(m.dataFile, m.backupFile)
		if err != nil {
			loger.Loger.Errorf("rename %s to %s failed: %v", m.dataFile, m.backupFile, err)
			return
		}
	}

	err = os.WriteFile(m.dataFile, data, 0o644)
	if err != nil {
		loger.Loger.Errorf("write file %s failed: %v", m.dataFile, err)
	} else {
		m.lastData = data
		loger.Loger.Infof("store clusterMembers: %s", m.ClusterMembers)
		loger.Loger.Infof("store knownMembers  : %s", m.KnownMembers)
	}
}

func (m *Members) UpdateClusterMembers(pbMembers []*etcdserverpb.Member) {
	m.Lock()
	defer m.Unlock()

	oldSelfID := m.GetSelf().ID
	ma := pbMembers2MembersArr(pbMembers)
	ma.update(membersArr{m.selfWithoutID()})
	m.ClusterMembers.replace(ma)

	selfID := m.GetSelf().ID
	if selfID != oldSelfID {
		loger.Loger.Infof("self ID changed from %x to %x", oldSelfID, selfID)
		//m.selfIDChanged = true
	}

	m.KnownMembers.update(*m.ClusterMembers)

	m.storeFileData()
}

func (ma membersArr) Len() int           { return len(ma) }
func (ma membersArr) Swap(i, j int)      { ma[i], ma[j] = ma[j], ma[i] }
func (ma membersArr) Less(i, j int) bool { return ma[i].Name < ma[j].Name }

func (ma membersArr) peerUrls() []string {
	ss := make([]string, 0)
	for _, m := range ma {
		ss = append(ss, m.PeerUrl)
	}
	return ss
}

func (ma membersArr) initCluster2String() string {
	ss := make([]string, 0)
	for _, m := range ma {
		if m.Name != "" {
			ss = append(ss, fmt.Sprintf("%s=%s", m.Name, m.PeerUrl))
		}
	}
	return strings.Join(ss, ",")
}

func (ma *membersArr) GetMemberByName(name string) *member {
	if name == "" {
		return nil
	}

	for _, m := range *ma {
		if m.Name == name {
			return &member{
				ID:      m.ID,
				Name:    m.Name,
				PeerUrl: m.PeerUrl,
			}
		}
	}

	return nil
}

func (ma *membersArr) update(updateMembers membersArr) {
	for _, updateMember := range updateMembers {
		var found bool
		if updateMember.PeerUrl == "" {
			continue
		}

		found = false
		for _, m := range *ma {
			if m.PeerUrl == updateMember.PeerUrl {
				found = true
				if len(updateMember.Name) > 0 {
					m.Name = updateMember.Name
				}

				if updateMember.ID != 0 {
					m.ID = updateMember.ID
				}
			}

			if !found {
				*ma = append(*ma, updateMember)
			}
		}
	}

	sort.Sort(*ma)
}

func (ma *membersArr) replace(replaceMembers membersArr) {
	*ma = replaceMembers
}

func pbMembers2MembersArr(pbMembers []*etcdserverpb.Member) membersArr {
	ms := make(membersArr, 0)

	for _, pbMember := range pbMembers {
		var peerUrl string
		if len(pbMember.PeerURLs) > 0 {
			peerUrl = pbMember.PeerURLs[0]
		}
		ms = append(ms, &member{
			ID:      pbMember.ID,
			Name:    pbMember.Name,
			PeerUrl: peerUrl,
		})
	}

	return ms
}
