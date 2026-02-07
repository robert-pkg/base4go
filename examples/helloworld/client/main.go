package main

import (
	"context"
	"encoding/json"
	"os"

	"github.com/robert-pkg/base4go/app"
	"github.com/robert-pkg/base4go/log"
	grpc_client "github.com/robert-pkg/base4go/rpc/client/grpc_client"
)

func main() {
	a := app.New()
	if err := a.Init(); err != nil {
		os.Exit(1)
	}

	cm := grpc_client.GetClientMgr()

	greeterClient, err := cm.GetClient("consul://" + "Greeter")
	if err != nil {
		panic(err)
	}

	if true {
		var req struct {
			Name string `json:"name"`
		}

		req.Name = "tom"
		reqData, _ := json.Marshal(req)

		var out []byte
		err := greeterClient.Invoke(context.Background(), "/api.Greeter/SayHello", reqData, &out)
		if err != nil {

			log.Errorf("err: %v", err)

		} else {
			log.Infof("success. ret: %v", string(out))

		}

	}

	if true {

	}

}
