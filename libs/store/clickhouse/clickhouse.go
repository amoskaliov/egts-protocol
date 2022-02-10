package main

/*
Плагин для работы с Clickhouse.
Плагин сохраняет пакет в jsonb поле point у заданной в настройках таблице.

Раздел настроек, которые должны отвечают в конфиге для подключения плагина:

[store]
plugin = "clickhouse.so"
host = "localhost"
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
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
)

type ClickhouseConnector struct {
	connection        clickhouse.Conn
	ctx               context.Context
	config            map[string]string
	batch             driver.Batch
	current_batch_len int
	max_batch_len     int
	query             string
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
	c.current_batch_len = 0
	c.query = fmt.Sprintf("INSERT INTO %s.%s", c.config["database"], c.config["table"])

	if err != nil {
		return fmt.Errorf("Неверно задан параметр batch_len: %v", err)
	}

	c.ctx = context.Background()
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

	err = c.InitBatch()

	return err
}

func (c *ClickhouseConnector) InitBatch() error {
	var (
		err error
	)

	c.batch, err = c.connection.PrepareBatch(c.ctx, c.query)

	if err != nil {
		return fmt.Errorf("Ошибка инициализации транзакции: %v", err)
	}

	return err
}

func (c *ClickhouseConnector) Save(msg interface{ ToBytes() ([]byte, error) }) error {

	err := c.batch.AppendStruct(&msg)

	if err != nil {
		return fmt.Errorf("Ошибка добавления пакета в транзакцию: %v", err)
	}

	c.current_batch_len++

	if c.current_batch_len > c.max_batch_len {
		err = c.batch.Send()
		if err != nil {
			return fmt.Errorf("Ошибка выполнения транзакции: %v", err)
		}
		err = c.InitBatch()
		if err != nil {
			return fmt.Errorf("Ошибка выполнения транзакции: %v", err)
		}
		c.current_batch_len = 0
	}

	return nil
}

func (c *ClickhouseConnector) Close() error {
	return c.connection.Close()
}

var Connector ClickhouseConnector
