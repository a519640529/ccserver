package main

import "container/list"

// 更新头像
type IHead interface {
	UpdateHead(snId, head int32)
}

// 更新昵称
type IName interface {
	UpdateName(snId int32, name string)
}

// 更新头像框
type IHeadOutline interface {
	UpdateHeadOutline(snId, outline int32)
}

// 观察者
type PlayerSubject struct {
	headList    *list.List
	nameList    *list.List
	outlineList *list.List
}

func (u *PlayerSubject) UpdateHead(snId int32, head int32) {
	for e := u.headList.Front(); e != nil; e = e.Next() {
		if o, ok := e.Value.(IHead); ok {
			o.UpdateHead(snId, head)
		}
	}
}

func (u *PlayerSubject) AttachHead(obj IHead) {
	for e := u.headList.Front(); e != nil; e = e.Next() {
		if e.Value == obj {
			return
		}
	}
	u.headList.PushBack(obj)
}

func (u *PlayerSubject) UpdateName(snId int32, name string) {
	for e := u.nameList.Front(); e != nil; e = e.Next() {
		if o, ok := e.Value.(IName); ok {
			o.UpdateName(snId, name)
		}
	}
}

func (u *PlayerSubject) AttachName(obj IName) {
	for e := u.nameList.Front(); e != nil; e = e.Next() {
		if e.Value == obj {
			return
		}
	}
	u.nameList.PushBack(obj)
}

func (u *PlayerSubject) UpdateHeadOutline(snId, outline int32) {
	for e := u.outlineList.Front(); e != nil; e = e.Next() {
		if o, ok := e.Value.(IHeadOutline); ok {
			o.UpdateHeadOutline(snId, outline)
		}
	}
}

func (u *PlayerSubject) AttachHeadOutline(obj IHeadOutline) {
	for e := u.outlineList.Front(); e != nil; e = e.Next() {
		if e.Value == obj {
			return
		}
	}
	u.outlineList.PushBack(obj)
}

// 单例
var PlayerSubjectSign = &PlayerSubject{
	headList:    list.New(),
	nameList:    list.New(),
	outlineList: list.New(),
}
