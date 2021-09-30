package contracts

type LoggerContract interface {
	Info(message string)
	Warn(message string)
	Error(message string)
	AddFields(fields map[string]interface{})
	ClearField(key string)
}
