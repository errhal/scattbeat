package beater

import (
	"fmt"
	"time"
	"github.com/elastic/beats/libbeat/beat"
	"github.com/elastic/beats/libbeat/common"
	"github.com/elastic/beats/libbeat/logp"

	"github.com/errhal/scattbeat/config"
	"net"
	"encoding/json"
)

type Message struct {
	MessageType string `json:"messageType"`
	Query       string `json:"query"`
}

type Status struct {
	TotalConnectionsNumber   int64
	CurrentConnectionsNumber int64
}

// Scattbeat configuration.
type Scattbeat struct {
	done   chan struct{}
	config config.Config
	client beat.Client
}

// New creates an instance of scattbeat.
func New(b *beat.Beat, cfg *common.Config) (beat.Beater, error) {
	c := config.DefaultConfig
	if err := cfg.Unpack(&c); err != nil {
		return nil, fmt.Errorf("Error reading config file: %v", err)
	}

	bt := &Scattbeat{
		done:   make(chan struct{}),
		config: c,
	}
	return bt, nil
}

// Run starts scattbeat.
func (bt *Scattbeat) Run(b *beat.Beat) error {
	logp.Info("scattbeat is running! Hit CTRL-C to stop it.")

	var err error
	bt.client, err = b.Publisher.Connect()
	if err != nil {
		return err
	}

	ticker := time.NewTicker(bt.config.Period)
	counter := 1
	for {
		conn, err := net.Dial("tcp", "127.0.0.1:7000")
		if (err != nil) {
			println("Error during database connection")
			time.Sleep(1000)
			break
		}

		messageObject := Message{"query", "show status"}
		serializedMessage, _ := json.Marshal(messageObject)
		conn.Write([]byte(string(serializedMessage) + "\n"))
		print(string(serializedMessage) + "\n")

		buffer := make([]byte, 1024)
		n, _ := conn.Read(buffer)

		print(string(buffer))
		var status Status
		json.Unmarshal(buffer[:n], &status)

		conn.Close()

		select {
		case <-bt.done:
			return nil
		case <-ticker.C:
		}

		println("BEFORE SENT", status.TotalConnectionsNumber)

		event := beat.Event{
			Timestamp: time.Now(),
			Fields: common.MapStr{
				"type":                       b.Info.Name,
				"counter":                    counter,
				"total_connections_number":   status.TotalConnectionsNumber,
				"current_connections_number": status.CurrentConnectionsNumber,
			},
		}
		bt.client.Publish(event)
		logp.Info("Event sent")
		counter++
	}
	return err
}

// Stop stops scattbeat.
func (bt *Scattbeat) Stop() {
	bt.client.Close()
	close(bt.done)
}
