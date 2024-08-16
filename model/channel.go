package model

type Channel struct {
	Id      string `db:"id" json:"id"`
	Name    string `db:"name" json:"name"`
	OwnerId string `db:"ownerId" json:"ownerId"`
}

type DBChannel struct {
	Id       string `db:"id" json:"id"`
	Name     string `db:"name" json:"name"`
	OwnerId  string `db:"ownerId" json:"ownerId"`
	Password string `db:"password" json:"password"`
}

func (c *Channel) GetName() string {
	return c.Name
}

func (c *Channel) GetId() string {
	return c.Id
}
