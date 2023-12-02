package models

import "gorm.io/gorm"

type User struct {
	gorm.Model
	Slug     string `gorm:"unique"`
	Username string `gorm:"unique"`
	Password string
	FullName string
	Role     string
}

type Member struct {
	ID       uint
	Slug     string
	Username string `gorm:"unique"`
	FullName string
	Role     string
}

type Tokens struct {
	AccessToken string `json:"accessToken"`
}
type ResponseToken struct {
	Tokens Tokens `json:"token"`
}

func (u *User) UserToMember() Member {
	return Member{
		ID:       u.ID,
		Slug:     u.Slug,
		Username: u.Username,
		FullName: u.FullName,
		Role:     u.Role,
	}
}

func (Member) TableName() string {
	return "users"
}
