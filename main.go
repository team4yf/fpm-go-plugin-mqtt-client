package main

import (
	"github.com/team4yf/yf-fpm-server-go/fpm"

	_ "github.com/team4yf/fpm-go-plugin-mqtt-client/plugin"
)

func main() {

	app := fpm.New()
	app.Init()
	app.Execute("mqttclient.subscribe", &fpm.BizParam{
		"topics": "foo.bar",
	})

	app.Subscribe("#mqtt/receive", func(topic string, data interface{}) {
		app.Logger.Debugf("topic: %s, payload: %+v", topic, data)
	})
	app.Execute("mqttclient.publish", &fpm.BizParam{
		"topic":   "foo.push",
		"payload": ([]byte)(`{"test":1}`),
	})

	app.Run()

}
