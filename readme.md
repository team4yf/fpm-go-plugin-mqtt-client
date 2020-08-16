## fpm-go-plugin-mqtt-client

mqtt client 的插件

#### config

```json
{
    "mqtt": {
        "host": "yourhost:1883",
        "pass": "*****",
        "user": "*****"
    }
}

```

#### import

` import _ "github.com/team4yf/fpm-go-plugin-mqtt-client/plugin" `

#### useage

```
//执行订阅的函数
app.Execute("mqttclient.subscribe", &fpm.BizParam{
    "topics": "$s2d/+/ipc/demo/execute",
})
//通过订阅系统消息处理业务
app.Subscribe("$s2d/111/ipc/demo/execute", func (topic string, data interface{} ){
    //data 通常是 byte[] 类型，可以转成 string 或者 map
    body := (string)(data.([]byte))
    log.Debugf("data: %+v", body)
})
//执行发布消息的函数
app.Execute("mqttclient.publish", &fpm.BizParam{
    "topic": "$s2d/111/ipc/demo/feedback",
    "payload": ([]byte)(`{"test":1}`),
})
```