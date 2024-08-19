package model

type User struct {
	Id       string `json:"id"`
	Username string `json:"username"`
	Name     string `json:"name"`
	Avatar   string `json:"avatar"`
}

type DBUser struct {
	Id              string `db:"id"`
	Name            string `form:"name" db:"username"`
	Password        string `form:"password" db:"password"`
	OldPassword     string `form:"oldPassword"`
	PasswordConfirm string `form:"passwordConfirm"`
}
