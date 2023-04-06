package api

import (
	"fmt"
	"net/http"
	"time"

	"github.com/idealeak/goserver/core/basic"
	"github.com/idealeak/goserver/core/logger"
)

var ApiDefaultTimeout = time.Second * 30

type HandlerWrapper func(*WebApiEvent, []byte) bool

type WebApiEvent struct {
	req      *http.Request
	path     string
	rawQuery string
	body     []byte
	h        HandlerWrapper
	res      chan []byte
}

func (this *WebApiEvent) Done(o *basic.Object) error {
	defer o.ProcessSeqnum()
	this.h(this, this.body)
	return nil
}

func (this *WebApiEvent) Response(data []byte) {
	this.res <- data
}

func webApiResponse(rw http.ResponseWriter, data []byte) bool {
	dataLen := len(data)
	rw.Header().Set("Content-Length", fmt.Sprintf("%v", dataLen))
	rw.WriteHeader(http.StatusOK)
	pos := 0
	for pos < dataLen {
		writeLen, err := rw.Write(data[pos:])
		if err != nil {
			logger.Logger.Info("webApiResponse SendData error:", err, " data=", string(data[:]), " pos=", pos, " writelen=", writeLen, " dataLen=", dataLen)
			return false
		}
		pos += writeLen
	}

	return true
}
