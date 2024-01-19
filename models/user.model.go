package models

import "gorm.io/gorm"

type User struct {
	gorm.Model
	Slug      string `gorm:"unique" form:"Slug" json:"Slug"`
	Username  string `gorm:"unique" form:"Username"`
	Password  string `form:"Password"`
	FullName  string `form:"FullName"`
	Role      string `form:"Role"`
	Tel       string `form:"Tel"`
	ImagePath string
}

type Member struct {
	ID        uint
	Slug      string `gorm:"unique"`
	Username  string `gorm:"unique"`
	FullName  string
	Role      string
	Tel       string
	ImagePath string
}

type Tokens struct {
	AccessToken string `json:"accessToken"`
}
type ResponseToken struct {
	Tokens Tokens `json:"token"`
}

func (u *User) UserToMember() Member {
	return Member{
		ID:        u.ID,
		Slug:      u.Slug,
		Username:  u.Username,
		FullName:  u.FullName,
		Role:      u.Role,
		ImagePath: u.ImagePath,
		Tel:       u.Tel,
	}
}

func (Member) TableName() string {
	return "users"
}

func (User) TableName() string {
	return "users"
}

const (
	ROLE_Admin = "admin"
	ROLE_User  = "user"
	ROLE_Tech  = "technician"
)
