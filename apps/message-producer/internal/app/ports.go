package app

type Json map[string]interface{}

type Messenger struct {
	SendMessage func (topic string, key string, message Json) error
}
