//Package plugin 开发的插件
package plugin

import (
	MQTT "github.com/eclipse/paho.mqtt.golang"
	"github.com/google/uuid"
	"github.com/team4yf/yf-fpm-server-go/fpm"
	"github.com/team4yf/yf-fpm-server-go/pkg/log"
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
	Host     string
	User     string
	Pass     string
	ClientID string
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
		setting := &mqttSetting{
			Options: &MQTT.ClientOptions{},
		}
		if err := app.FetchConfig("mqtt", &setting); err != nil {
			panic(err)
		}

		log.Debugf("Mqtt Config : %+v", setting)
		clientID := setting.ClientID + GenUUID()
		setting.Options.AddBroker("tcp://" + setting.Host)
		setting.Options.SetClientID(clientID)
		setting.Options.SetUsername(setting.User)
		setting.Options.SetPassword(setting.Pass)

		mq := newMQTTPubSub(setting)
		app.Publish("#mqtt/connected", map[string]interface{}{
			"topic":   "mqtt/connected",
			"payload": clientID,
		})
		handler := func(topic, payload interface{}) {
			messsage := map[string]interface{}{
				"topic":   topic,
				"payload": payload,
			}
			app.Publish("#mqtt/receive", messsage)
		}
		bizModule := make(fpm.BizModule, 0)

		bizModule["subscribe"] = func(param *fpm.BizParam) (data interface{}, err error) {
			topics := make([]string, 0)
			switch (*param)["topics"].(type) {
			case string:
				topics = append(topics, (*param)["topics"].(string))
			case []string:
				topics = (*param)["topics"].([]string)
			case []interface{}:
				for _, t := range (*param)["topics"].([]interface{}) {
					topics = append(topics, t.(string))
				}
			}
			for _, t := range topics {
				mq.Subscribe(t, handler)
			}
			data = 1
			return
		}
		bizModule["publish"] = func(param *fpm.BizParam) (data interface{}, err error) {
			topic := (*param)["topic"].(string)
			payload := (*param)["payload"].([]byte)
			mq.Publish(topic, payload)
			data = 1
			return
		}
		app.AddBizModule("mqttclient", &bizModule)

	})
}
