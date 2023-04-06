package main

import "github.com/idealeak/goserver/core/netlib"

var CustomGroupMgrSington = &CustomGroupMgr{
	groups: make(map[string]map[int64]*netlib.Session),
}

type CustomGroupMgr struct {
	groups map[string]map[int64]*netlib.Session
}

func (this *CustomGroupMgr) AddToGroup(groupTags []string, sid int64, s *netlib.Session) {
	for _, tag := range groupTags {
		group, ok := this.groups[tag]
		if !ok {
			group = make(map[int64]*netlib.Session)
			this.groups[tag] = group
		}
		group[sid] = s
	}
}

func (this *CustomGroupMgr) DelFromGroup(groupTags []string, sid int64) {
	for _, tag := range groupTags {
		if group, ok := this.groups[tag]; ok {
			delete(group, sid)
		}
	}
}

func (this *CustomGroupMgr) Broadcast(tags []string, raw []byte) {
	for _, tag := range tags {
		if group, ok := this.groups[tag]; ok {
			for _, s := range group {
				s.Send(0, raw)
			}
		}
	}
}
