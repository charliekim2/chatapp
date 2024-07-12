package model

type UserChannel struct {
	Id        string `db:"id" json:"id"`
	UserId    string `db:"userId" json:"userId"`
	ChannelId string `db:"channelId" json:"channelId"`
}
