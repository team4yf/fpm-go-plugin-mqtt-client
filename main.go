package main

import (
	"github.com/team4yf/yf-fpm-server-go/fpm"

	"github.com/team4yf/yf-fpm-server-go/pkg/log"
	_ "github.com/team4yf/fpm-go-plugin-mqtt-client/plugin"
)

func main() {

	app := fpm.New()
	app.Init()
	app.Execute("mqttclient.subscribe", &fpm.BizParam{
		"topics": "$s2d/+/ipc/demo/execute",
	})

	app.Subscribe("$s2d/111/ipc/demo/execute", func (topic string, data interface{} ){
		body := (string)(data.([]byte))
		log.Debugf("data: %+v", body)
	})
	app.Execute("mqttclient.publish", &fpm.BizParam{
		"topic": "$s2d/111/ipc/demo/feedback",
		"payload": ([]byte)(`{"test":1}`),
	})

	app.Run(":9999")

}
