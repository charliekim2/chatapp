package model

type Message struct {
	Id        string `db:"id" json:"id"`
	OwnerId   string `db:"ownerId" json:"ownerId"`
	CreatedAt string `db:"created" json:"created"`
	ChannelId string `db:"channelId" json:"channelId"`
	Body      string `db:"body" json:"body"`
}

type DBMessage struct {
	Id        string `db:"id" json:"id"`
	OwnerId   string `db:"ownerId" json:"ownerId"`
	ChannelId string `db:"channelId" json:"channelId"`
	Body      string `db:"body" json:"body"`
}

// TODO: refactor this out since all the fields are public anyway...
func (m *Message) GetId() string {
	return m.Id
}

func (m *Message) GetOwnerId() string {
	return m.OwnerId
}

func (m *Message) GetCreatedAt() string {
	return m.CreatedAt
}

func (m *Message) GetBody() string {
	return m.Body
}
