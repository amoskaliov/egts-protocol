package main

/*
Плагин для работы с RabbitMQ через amqp.

Раздел настроек, которые должны отвечать в конфиге для подключения плагина:

[store]
plugin = "rabbitmq.so"
host = "localhost"
port = "5672"
user = "guest"
password = "guest"
exchange = "receiver"
exchange_type = "topic"
*/

import (
	"fmt"

	"github.com/amoskaliov/egts-protocol/cli/receiver/storage"
	"github.com/streadway/amqp"
)

type RabbitMQConnector struct {
	connection *amqp.Connection
	channel    *amqp.Channel
	config     map[string]string
}

func (c *RabbitMQConnector) Init(cfg map[string]string) error {
	var (
		err error
	)
	if cfg == nil {
		return fmt.Errorf("Не корректная ссылка на конфигурацию")
	}

	c.config = cfg
	conStr := fmt.Sprintf("amqp://%s:%s@%s:%s/", c.config["user"], c.config["password"], c.config["host"], c.config["port"])
	if c.connection, err = amqp.Dial(conStr); err != nil {
		return fmt.Errorf("Ошибка установки соединеия RabbitMQ: %v", err)
	}

	if c.channel, err = c.connection.Channel(); err != nil {
		return fmt.Errorf("Ошибка открытия канала RabbitMQ: %v", err)
	}

	return err
}

func (c *RabbitMQConnector) Save(msg *storage.NavRecord) error {
	if msg == nil {
		return fmt.Errorf("Не корректная ссылка на пакет")
	}

	innerPkg, err := msg.ToBytes()
	if err != nil {
		return fmt.Errorf("Ошибка сериализации  пакета: %v", err)
	}

	if err = c.channel.Publish(
		c.config["exchange"],
		c.config["key"],
		false,
		false,
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        innerPkg,
		}); err != nil {
		return fmt.Errorf("Ошибка отправки сырого пакета в RabbitMQ: %v", err)
	}
	return nil
}

func (c *RabbitMQConnector) Close() error {
	var err error
	if c != nil {
		if c.channel != nil {
			if err = c.channel.Close(); err != nil {
				return err
			}
		}
		if c.connection != nil {
			if err = c.connection.Close(); err != nil {
				return err
			}
		}
	}
	return err
}

var Connector RabbitMQConnector
