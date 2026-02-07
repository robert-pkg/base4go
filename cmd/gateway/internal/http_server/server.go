package http_server

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/robert-pkg/base4go/cmd/gateway/internal/config"
	"github.com/robert-pkg/base4go/log"
)

func NewServer(svrCfg *config.Server) *Server {
	return &Server{
		Port:      svrCfg.Port,
		ApiPrefix: svrCfg.ApiPrefix,
		exit:      make(chan chan error)}
}

type Server struct {
	Port      int
	ApiPrefix string

	exit chan chan error

	sync.RWMutex
	// marks the serve as started
	started bool
	httpSvr *http.Server
}

func (svr *Server) Start() (err error) {
	svr.RLock()
	if svr.started {
		svr.RUnlock()
		return nil
	}
	svr.RUnlock()

	log.Infof("start, port=%d", svr.Port)
	if err = svr.startHTTPServer(svr.Port); err != nil {
		return
	}

	go func() {
		ch := <-svr.exit

		// stop the grpc server
		exitGrpcCh := make(chan bool)

		go func() {
			close_ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
			defer cancel()

			// server.Shutdown 方法是个阻塞方法，一旦执行之后，它会阻塞当前 Goroutine，并且在所有连接请求都结束之后，才继续往后执行。
			if httpErr := svr.httpSvr.Shutdown(close_ctx); httpErr != nil {
				log.Error("Shutdown gracefully fail.", "err", httpErr)
			}

			close(exitGrpcCh)
		}()

		// 等待grpc关闭
		<-exitGrpcCh

		// close transport
		ch <- nil
	}()

	// mark the server as started
	svr.Lock()
	svr.started = true
	svr.Unlock()

	return nil

}

func (svr *Server) Stop() {

	svr.RLock()
	if !svr.started {
		svr.RUnlock()
		return
	}
	svr.RUnlock()

	log.Info("Server Stop enter...")

	ch := make(chan error)
	svr.exit <- ch

	err := <-ch
	svr.Lock()
	svr.started = false
	svr.Unlock()

	if err != nil {
		log.Error("server stop error", "err", err)
	}

	log.Info("Server Stop leave...")
}
