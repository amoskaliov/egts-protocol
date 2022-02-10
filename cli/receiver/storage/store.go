package storage

//Connector интерфейс для подключения внешних хранилищ
type Connector interface {
	// установка соединения с хранилищем
	Init(map[string]string) error

	// сохранение в хранилище
	Save(*NavRecord) error

	//закрытие соединения с хранилищем
	Close() error
}
