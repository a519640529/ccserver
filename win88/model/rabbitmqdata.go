package model

type RabbitMQData struct {
	MQName string
	Data   interface{}
}

func NewRabbitMQData(mqName string, data interface{}) *RabbitMQData {
	log := &RabbitMQData{
		MQName: mqName,
		Data:   data,
	}
	return log
}
