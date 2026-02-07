package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/robert-pkg/base4go/cmd/gateway/internal/config"
	"github.com/robert-pkg/base4go/log"
)

func ApiHandler(c *gin.Context) {
	var r apiRequest
	r.Handle(c)
}

type apiRequest struct {
}

func (apiReq *apiRequest) Handle(c *gin.Context) {

	addr := c.Param("api")

	// consul://Greeter/Echo
	scheme, service, method, ok := config.GetApiMapping(addr)
	if !ok {
		c.String(http.StatusNotFound, "api not found")
		return
	}

	target := scheme + "://" + service
	fullMethod := "/api." + service + "/" + method

	client, err := g_ClientMgr.GetClient(target)
	if err != nil {
		log.Errorf("get grpc client fail. target=%s, err=%v", target, err)
		c.String(http.StatusInternalServerError, "Internal Server Error")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err = client.Invoke(ctx, fullMethod, nil, nil); err != nil {
		log.Errorf("get grpc client fail. target=%s, err=%v", target, err)
		c.String(http.StatusInternalServerError, "Internal Server Error")
		return
	}

	c.String(http.StatusOK, "xxx")
	return
}
