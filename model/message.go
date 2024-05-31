package model

type Message struct {
	id        string
	ownerId   string
	createdAt string
	body      string
}

func (m *Message) GetId() string {
	return m.id
}

func (m *Message) GetOwnerId() string {
	return m.ownerId
}

func (m *Message) GetBody() string {
	return m.body
}
