package cluster

import (
	"fmt"
	"nmid-registry/pkg/loger"
	"nmid-registry/pkg/option"
	"strings"
	"sync"
)

type (
	Members struct {
		sync.RWMutex

		Options        *option.Options
		ClusterMembers *membersArr
		KnownMembers   *membersArr
	}

	membersArr []*member

	member struct {
		ID      uint64 `yaml:"id"`
		Name    string `yaml:"name"`
		PeerUrl string `yaml:"peerUrl"`
	}
)

func NewMembers(opt *option.Options) (*Members, error) {
	mem := &Members{
		Options:        opt,
		ClusterMembers: &membersArr{},
		KnownMembers:   &membersArr{},
	}

	return mem, nil
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
