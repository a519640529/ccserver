package rpc

import (
	"errors"
	"io"
	"net/rpc"
	"sync/atomic"
	"time"

	"github.com/idealeak/goserver/core/utils"
)

var (
	ErrRPClientNoConn      = errors.New("RPClient no connect.")
	ErrRPClientShutdown    = errors.New("RPClient is shutdown.")
	ErrRPClientCallTimeout = errors.New("RPClient call timeout.")
)

type RPClient struct {
	client       *rpc.Client
	reconnInterv time.Duration
	net          string
	addr         string
	connecting   int32
	ready        bool
	shutdown     bool
}

func NewRPClient(net, addr string, d time.Duration) *RPClient {
	if d < time.Second {
		d = time.Second
	}

	c := &RPClient{
		net:          net,
		addr:         addr,
		ready:        false,
		shutdown:     false,
		reconnInterv: d,
	}
	c.init()
	return c
}

func (this *RPClient) IsReady() bool {
	return this.ready
}

func (this *RPClient) IsConnecting() bool {
	return atomic.LoadInt32(&this.connecting) == 1
}

func (this *RPClient) IsShutdown() bool {
	return this.shutdown
}

func (this *RPClient) init() {
}

func (this *RPClient) Start() error {
	if atomic.CompareAndSwapInt32(&this.connecting, 0, 1) {
		this.dialRoutine()
	}
	return nil
}

func (this *RPClient) Stop() error {
	if !this.shutdown {
		this.shutdown = true
		if this.client != nil {
			return this.client.Close()
		}
	}
	return nil
}

func (this *RPClient) Call(serviceMethod string, args interface{}, reply interface{}) error {
	return this.CallWithTimeout(serviceMethod, args, reply, time.Hour*24)
}

func (this *RPClient) CallWithTimeout(serviceMethod string, args interface{}, reply interface{}, d time.Duration) error {
	if this.client == nil {
		return ErrRPClientNoConn
	}

	if d <= time.Second {
		d = time.Second
	}

	start := time.Now()
	var err error
	call := this.client.Go(serviceMethod, args, reply, make(chan *rpc.Call, 1))
	select {
	case <-time.After(d):
		return ErrRPClientCallTimeout
	case call = <-call.Done:
		err = call.Error
	}
	if err != nil && (err == rpc.ErrShutdown || err == io.ErrUnexpectedEOF) {
		var dailed chan struct{}
		if atomic.CompareAndSwapInt32(&this.connecting, 0, 1) {
			dailed = make(chan struct{})
			_ = this.client.Close()
			go utils.CatchPanic(func() {
				this.dialRoutine()
				dailed <- struct{}{}
			})
		}

		dd := d - time.Now().Sub(start)
		if dd <= 0 {
			return ErrRPClientCallTimeout
		}

		select {
		case <-time.After(dd):
			return ErrRPClientCallTimeout
		case <-dailed:
			dd = d - time.Now().Sub(start)
			if dd <= 0 {
				return ErrRPClientCallTimeout
			}
			call = this.client.Go(serviceMethod, args, reply, make(chan *rpc.Call, 1))
			dd = d - time.Now().Sub(start)
			if dd <= 0 {
				return ErrRPClientCallTimeout
			}
			select {
			case <-time.After(dd):
				return ErrRPClientCallTimeout
			case call = <-call.Done:
				err = call.Error
			}
			return err
		}
	}
	return err
}

func (this *RPClient) dialRoutine() {
	defer func() {
		atomic.StoreInt32(&this.connecting, 0)
	}()
	dial := func() error {
		client, err := rpc.DialHTTP(this.net, this.addr)
		if err == nil {
			if this.shutdown {
				client.Close()
				return ErrRPClientShutdown
			}
			this.client = client
			this.ready = true
			return nil
		}
		return err
	}

	err := dial()
	if err != nil {
		for !this.shutdown {
			select {
			case <-time.After(this.reconnInterv):
				err = dial()
				if err == nil {
					return
				}
			}
		}
	}
}
