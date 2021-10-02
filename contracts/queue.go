package contracts

type QueueManager interface {
	SendMessage(message interface{}) error
	CloseConnection()
	CloseChannel()
}
