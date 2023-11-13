package models

import (
	"fmt"

	"gorm.io/gorm"
)

type Project struct {
	gorm.Model
	Slug        string
	Title       string
	Subtitle    string
	Description string
}

func (p *Project) String() string {
	return fmt.Sprintf("Project{ID: %d, Slug: %s, Title: %s, Subtitle: %s, Description: %s}",
		p.ID, p.Slug, p.Title, p.Subtitle, p.Description)
}
