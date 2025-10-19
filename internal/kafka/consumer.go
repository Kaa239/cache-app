package kafka

import (
	"cache-app/internal/model"
	"cache-app/internal/service"
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/IBM/sarama"
)

type Consumer struct {
	service  service.OrderService
	consumer sarama.Consumer
	topic    string
}

func NewConsumer(brokers []string, topic string, service service.OrderService) (*Consumer, error) {
	config := sarama.NewConfig()
	config.Consumer.Return.Errors = true

	consumer, err := sarama.NewConsumer(brokers, config)
	if err != nil {
		return nil, err
	}

	return &Consumer{
		service:  service,
		consumer: consumer,
		topic:    topic,
	}, nil
}

func (c *Consumer) Start() {
	partitions, err := c.consumer.Partitions(c.topic)
	if err != nil {
		log.Fatalf("Failed to get partitions: %v", err)
	}

	for _, partition := range partitions {
		go c.consumePartition(partition)
	}
}

func (c *Consumer) consumePartition(partition int32) {
	pc, err := c.consumer.ConsumePartition(c.topic, partition, sarama.OffsetNewest)
	if err != nil {
		log.Printf("Failed to start consumer for partition %d: %v", partition, err)
		return
	}
	defer pc.Close()

	for {
		select {
		case msg := <-pc.Messages():
			c.processMessage(msg)
		case err := <-pc.Errors():
			log.Printf("Error consuming partition %d: %v", partition, err)
		}
	}
}

func (c *Consumer) processMessage(msg *sarama.ConsumerMessage) {
	var order model.Order
	if err := json.Unmarshal(msg.Value, &order); err != nil {
		log.Printf("Failed to unmarshal order: %v", err)
		return
	}

	// Валидация базовых полей
	if order.OrderUID == "" {
		log.Printf("Invalid order: missing order_uid")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := c.service.CreateOrder(ctx, &order); err != nil {
		log.Printf("Failed to create order %s: %v", order.OrderUID, err)
	} else {
		log.Printf("Order %s processed successfully", order.OrderUID)
	}
}

func (c *Consumer) Close() error {
	return c.consumer.Close()
}
