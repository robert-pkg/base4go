package http_server

import (
	"fmt"
	"net/http"
	"os"

	"github.com/chenjiandongx/ginprom"
	"github.com/gin-gonic/gin"
	"github.com/robert-pkg/base4go/log"
)

func (svr *Server) startHTTPServer(nPort int) error {

	r := gin.New()

	r.Use(gin.Recovery(),
		ginprom.PromMiddleware(nil), // use prometheus metrics exporter middleware.
	)

	svr.configHandle(r)

	svr.httpSvr = &http.Server{
		Addr:    fmt.Sprintf(":%d", nPort),
		Handler: r, // r 实现了 server.Handler 接口
	}

	go func() {
		if err := svr.httpSvr.ListenAndServe(); err != nil {
			if err == http.ErrServerClosed {
				log.Info("http server closed")
				return
			}

			log.Error("http服务器启动失败. ListenAndServe fail", "err", err)
			os.Exit(1)
		}
	}()

	return nil

}
