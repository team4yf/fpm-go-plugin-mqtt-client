//Package plugin 开发的插件
package plugin

import (
	"github.com/team4yf/yf-fpm-server-go/fpm"
	"github.com/team4yf/yf-fpm-server-go/pkg/log"
	"github.com/google/uuid"
	MQTT "github.com/eclipse/paho.mqtt.golang"
)
//PubSub 定义接口
// 主要包含发布和订阅
type pubSub interface {
	Publish(topic string, payload []byte)
	Subscribe(topic string, handler func(topic, payload interface{}))
}

type mqttSetting struct {
	Options  *MQTT.ClientOptions
	Qos      byte
	Retained bool
}

//mqttPS 定义MQTT 的结构体
// 包含一个 MQTT 的客户端和一些配置信息
type mqttPS struct {
	mClient MQTT.Client
	config  *mqttSetting
}

//NewMQTTPubSub 构建实例的函数,用于返回一个MQTT的对象,通过 PubSub 接口返回
func newMQTTPubSub(c *mqttSetting) pubSub {
	client := MQTT.NewClient(c.Options)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}
	instance := &mqttPS{
		mClient: client,
		config:  c,
	}
	return instance
}

//Publish 实现Publish函数
func (m *mqttPS) Publish(topic string, payload []byte) {
	token := m.mClient.Publish(topic, m.config.Qos, m.config.Retained, payload)
	token.Wait()
}

//Subscribe 实现Subscribe
func (m *mqttPS) Subscribe(topic string, handler func(topic, payload interface{})) {
	m.mClient.Subscribe(topic, m.config.Qos, func(_ MQTT.Client, message MQTT.Message) {
		handler(message.Topic(), message.Payload())
	})
}

// GenUUID 生成随机字符串，eg: 76d27e8c-a80e-48c8-ad20-e5562e0f67e4
func GenUUID() string {
	u, _ := uuid.NewRandom()
	return u.String()
}
func init() {
	fpm.Register(func(app *fpm.Fpm) {
		// 配置 MQTT 客户端
		if !app.HasConfig("mqtt") {
			panic("mqtt config node required")
		}
		mqttConfig := app.GetConfig("mqtt").(map[string]interface{})
		log.Debugf("Mqtt Config : %+v", mqttConfig)
		
		setting := &mqttSetting{
			Options:  &MQTT.ClientOptions{},
			Retained: false,
			Qos:      (byte)(0),
		}
		setting.Options.AddBroker("tcp://" + mqttConfig["host"].(string))
		setting.Options.SetClientID("iot-device-" + GenUUID())
		setting.Options.SetUsername(mqttConfig["user"].(string))
		setting.Options.SetPassword(mqttConfig["pass"].(string))
	
		mq := newMQTTPubSub(setting)
		log.Debugf("mqtt client inited!")

		bizModule := make(fpm.BizModule, 0)
		bizModule["subscribe"] = func(param *fpm.BizParam) (data interface{}, err error) {
			topics:= (*param)["topics"].(string)
			mq.Subscribe(topics, func(topic, payload interface{}){
				app.Publish(topic.(string), payload)
			})
			data = 1
			return
		}
		bizModule["publish"] = func(param *fpm.BizParam) (data interface{}, err error) {
			topic:= (*param)["topic"].(string)
			payload:= (*param)["payload"].([]byte)
			mq.Publish(topic, payload)
			data = 1
			return
		}
		app.AddBizModule("mqttclient", &bizModule)

	})
}
