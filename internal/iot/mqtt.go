package iot

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

// MQTTSender implements CommandSender using an MQTT broker.
// It publishes ON/OFF commands to a topic pattern like: <prefix>/<id>/cmd
// and (optionally) listens for status messages on: <prefix>/<id>/status
// Status messages are forwarded to the provided callback.
type MQTTSender struct {
	client   mqtt.Client
	prefix   string
	qos      byte
	retain   bool
	statusCb func(id int64, payload string)
	mu       sync.Mutex
}

// MQTTSenderOptions configures the MQTT sender.
type MQTTSenderOptions struct {
	Prefix         string
	QOS            byte
	Retain         bool
	ClientID       string
	Username       string
	Password       string
	CleanSession   bool
	StatusCallback func(id int64, payload string)
	ConnectTimeout time.Duration
}

// NewMQTTSender creates and connects a new MQTTSender.
// broker example: tcp://localhost:1883
func NewMQTTSender(broker string, opt MQTTSenderOptions) (*MQTTSender, error) {
	if opt.Prefix == "" {
		opt.Prefix = "ps"
	}
	if opt.ClientID == "" {
		opt.ClientID = fmt.Sprintf("heheswitch-%d", time.Now().UnixNano())
	}
	if opt.ConnectTimeout == 0 {
		opt.ConnectTimeout = 10 * time.Second
	}
	mopts := mqtt.NewClientOptions().AddBroker(broker).SetClientID(opt.ClientID).SetCleanSession(opt.CleanSession)
	if opt.Username != "" {
		mopts.SetUsername(opt.Username)
	}
	if opt.Password != "" {
		mopts.SetPassword(opt.Password)
	}
	sender := &MQTTSender{prefix: opt.Prefix, qos: opt.QOS, retain: opt.Retain, statusCb: opt.StatusCallback}
	mopts.SetOnConnectHandler(func(c mqtt.Client) {
		// subscribe to status
		topic := fmt.Sprintf("%s/+/status", sender.prefix)
		token := c.Subscribe(topic, sender.qos, sender.handleStatus)
		token.Wait()
		if token.Error() != nil {
			log.Printf("mqtt subscribe error: %v", token.Error())
		}
	})
	sender.client = mqtt.NewClient(mopts)
	token := sender.client.Connect()
	if !token.WaitTimeout(opt.ConnectTimeout) {
		return nil, fmt.Errorf("mqtt connect timeout")
	}
	if token.Error() != nil {
		return nil, token.Error()
	}
	return sender, nil
}

// Send publishes a command (e.g., ON / OFF) to the device topic.
func (m *MQTTSender) Send(consoleID int64, cmd string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.client == nil || !m.client.IsConnectionOpen() {
		return fmt.Errorf("mqtt not connected")
	}
	topic := fmt.Sprintf("%s/%d/cmd", m.prefix, consoleID)
	token := m.client.Publish(topic, m.qos, m.retain, cmd)
	token.Wait()
	return token.Error()
}

// IsConnected returns current connection state.
func (m *MQTTSender) IsConnected() bool { return m.client != nil && m.client.IsConnectionOpen() }

// Prefix returns topic prefix.
func (m *MQTTSender) Prefix() string { return m.prefix }

// Close disconnects the client if connected.
func (m *MQTTSender) Close() {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.client != nil && m.client.IsConnectionOpen() {
		m.client.Disconnect(250)
	}
}

func (m *MQTTSender) handleStatus(_ mqtt.Client, msg mqtt.Message) {
	// topic pattern: prefix/<id>/status
	parts := strings.Split(msg.Topic(), "/")
	if len(parts) < 3 {
		return
	}
	idStr := parts[len(parts)-2]
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return
	}
	if m.statusCb != nil {
		m.statusCb(id, string(msg.Payload()))
	}
}

// NewFromEnv builds an MQTTSender from environment variables.
// Required: MQTT_BROKER (e.g. tcp://localhost:1883)
// Optional: MQTT_PREFIX (default ps), MQTT_USERNAME, MQTT_PASSWORD, MQTT_CLIENT_ID
func NewFromEnv(statusCb func(id int64, payload string)) (*MQTTSender, error) {
	broker := os.Getenv("MQTT_BROKER")
	if broker == "" {
		return nil, fmt.Errorf("MQTT_BROKER not set")
	}
	prefix := os.Getenv("MQTT_PREFIX")
	cid := os.Getenv("MQTT_CLIENT_ID")
	user := os.Getenv("MQTT_USERNAME")
	pass := os.Getenv("MQTT_PASSWORD")
	return NewMQTTSender(broker, MQTTSenderOptions{Prefix: prefix, ClientID: cid, Username: user, Password: pass, QOS: 1, CleanSession: true, StatusCallback: statusCb})
}
