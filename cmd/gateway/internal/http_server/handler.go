package http_server

import (
	"github.com/gin-gonic/gin"
	"github.com/robert-pkg/base4go/cmd/gateway/internal/handler"
)

func (svr *Server) configHandle(r *gin.Engine) {
	r.POST(svr.ApiPrefix+"/*api", handler.ApiHandler)
}
