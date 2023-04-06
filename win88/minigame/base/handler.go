package base

import (
	"fmt"
	"reflect"

	"github.com/idealeak/goserver/core/netlib"
)

var handlers = make(map[int]Handler)

type Handler interface {
	Process(s *netlib.Session, packetid int, data interface{}, scene *Scene, p *Player) error
}

type HandlerWrapper func(s *netlib.Session, packetid int, data interface{}, scene *Scene, p *Player) error

func (hw HandlerWrapper) Process(s *netlib.Session, packetid int, data interface{}, scene *Scene, p *Player) error {
	return hw(s, packetid, data, scene, p)
}

func RegisterHandler(packetId int, h Handler) {
	if _, ok := handlers[packetId]; ok {
		panic(fmt.Sprintf("repeate register handler: %v Handler type=%v", packetId, reflect.TypeOf(h)))
	}

	handlers[packetId] = h
}

func GetHandler(packetId int) Handler {
	if h, ok := handlers[packetId]; ok {
		return h
	}

	return nil
}
