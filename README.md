# EGTS receiver

EGTS receiver server realization writen on Go. 

Library for implementation EGTS protocol that parsing binary packag based on 
[GOST R 54619 - 2011](./docs`/gost54619-2011.pdf) and 
[Order No. 285 of the Ministry of Transport of Russia dated July 31, 2012](./docs/mitransNo285.pdf). 
Describe fields you can find in these documents. 

More information you can read in [article](https://www.swe-notes.ru/post/protocol-egts/) (Russian).
 
Server save all navigation data from ```EGTS_SR_POS_DATA``` section. If packet have several records with 
```EGTS_SR_POS_DATA``` section, it saves all of them. 

Storage for data realized as plugins. Any plugin must have ```[store]``` section in configure file. 
Plugin interface will be described below.

If configure file has't section for a plugin (```[store]```), then packet will be print to stdout.

## Fork features

This project is forked from https://github.com/kuznetsovin/egts-protocol

Key differences:
1. If the `DIRH` field is set to 1, then `course = DIR + 256` (in the original project it's calculated as `course = DIR + 128`).
2. Also the `course` field now supports values more than 255 (as it can be up to 359).
3. The only supported external storage is Clickhouse DB.

## Install

```bash
git clone https://github.com/amoskaliov/egts-protocol
cd egts-protocol
make
```

## Run

```bash
./receiver -c config.toml
```

```config.toml``` - configure file

## Config format

```toml
host = "127.0.0.1"
port = "6000"
con_ttl = 10
log_level = "DEBUG"

[store]
plugin = "clickhouse.so"
host = "clickhouse:9000"
user = "receiver"
password = "CH_RCV_PASSWORD"
database = "db"
table = "table"
batch_len = "50000"
```

Parameters description:

- *host* - bind address  
- *port* - bind port 
- *conn_ttl* - if server not received data longer time in the parameter, then the connection is closed. 
- *log_level* - logging level

Clickhouse settings:

- *plugin* - path to `clickhouse.so`
- *host* - server address and port to Clickhouse native interface.
- *user* - username for connection.
- *password* - name of environment variable which contains the password for connection.
- *database* - database name.
- *table* - table name.
- *batch_len* - size of batch to be inserted.

## Database

Table example is below. Sensors data is missing.

```sql
CREATE TABLE IF NOT EXISTS telematics_service.queue (
    `client` UInt32,
    `packet_id` UInt32,
    `navigation_unix_time` Int64,
    `received_unix_time` Int64,
    `latitude` Float64,
    `longitude` Float64,
    `speed` UInt16,
    `course` UInt16
)
```