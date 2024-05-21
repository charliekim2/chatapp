package model

type Channel struct {
	Id   string `db:"id" json:"id"`
	Name string `db:"name" json:"name"`
}

func (c *Channel) GetName() string {
	return c.Name
}

func (c *Channel) GetId() string {
	return c.Id
}
