package main

/*
Плагин для работы с Clickhouse.
Плагин сохраняет пакет в jsonb поле point у заданной в настройках таблице.

Раздел настроек, которые должны отвечают в конфиге для подключения плагина:

[store]
plugin = "clickhouse.so"
host = "localhost:9000"
user = "postgres"
password = "postgres"
database = "receiver"
table = "points"
batch_len = "50000"
*/

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/amoskaliov/egts-protocol/cli/receiver/storage"
	log "github.com/sirupsen/logrus"
)

type ClickhouseConnector struct {
	connection    clickhouse.Conn
	config        map[string]string
	batch         []*storage.NavRecord
	max_batch_len int
	query         string
}

func (c *ClickhouseConnector) Init(cfg map[string]string) error {
	var (
		err error
	)
	if cfg == nil {
		return fmt.Errorf("Некорректная ссылка на конфигурацию")
	}
	c.config = cfg
	c.max_batch_len, err = strconv.Atoi(c.config["batch_len"])
	c.batch = nil
	c.query = fmt.Sprintf("INSERT INTO %s.%s", c.config["database"], c.config["table"])

	if err != nil {
		return fmt.Errorf("Неверно задан параметр batch_len: %v", err)
	}

	c.connection, err = clickhouse.Open(&clickhouse.Options{
		Addr: []string{c.config["host"]},
		Auth: clickhouse.Auth{
			Database: c.config["database"],
			Username: c.config["user"],
			Password: c.config["password"],
		},
		//Debug:           true,
		DialTimeout:     time.Second,
		MaxOpenConns:    10,
		MaxIdleConns:    5,
		ConnMaxLifetime: time.Hour,
	})

	if err != nil {
		return fmt.Errorf("Ошибка подключения к Clickhouse: %v", err)
	}

	return err
}

func (c *ClickhouseConnector) Save(msg *storage.NavRecord) error {
	c.batch = append(c.batch, msg)

	if len(c.batch) >= c.max_batch_len {
		log.Debugf("Пакет готов к отправке. Строк: %v", len(c.batch))
		ctx := context.Background()

		ch_batch, err := c.connection.PrepareBatch(ctx, c.query)
		if err != nil {
			c.batch = nil
			return fmt.Errorf("Ошибка подготовки транзакции: %v", err)
		}

		for _, element := range c.batch {
			err = ch_batch.Append(
				element.Client,
				element.PacketID,
				element.NavigationTimestamp,
				element.ReceivedTimestamp,
				element.Latitude,
				element.Longitude,
				element.Speed,
				element.Course,
			)
			if err != nil {
				c.batch = nil
				return fmt.Errorf("Ошибка добавления элемента в транзакцию: %v", err)
			}
		}

		err = ch_batch.Send()
		if err != nil {
			c.batch = nil
			return fmt.Errorf("Ошибка выполнения транзакции: %v", err)
		}

		c.batch = nil

	}

	return nil
}

func (c *ClickhouseConnector) Close() error {
	return c.connection.Close()
}

var Connector ClickhouseConnector
