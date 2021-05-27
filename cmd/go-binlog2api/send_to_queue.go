package main

import (
	"encoding/json"
	"log"

	"github.com/cheshir/go-mq"
	scripts "github.com/gr4c2-2000/go-binlog2api/internal/go-binlog2api"
)

var msq mq.MQ
var mqst = false

func Produce(prd string, data interface{}) error {
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return err
	}
	msqa := messageQueue()
	producer, err := msqa.SyncProducer(prd)
	if err != nil {
		return err
	}
	err = producer.Produce(jsonBytes)
	if err != nil {
		return err
	}
	return nil
}

func messageQueue() mq.MQ {
	if mqst == false {
		conf := scripts.MqConfig()
		msq, err := mq.New(conf)
		if err != nil {
			log.Fatal("Failed to initialize message queue manager", err)
		}
		return msq
	}
	return msq
}

func close() {
	msq.Close()
}
