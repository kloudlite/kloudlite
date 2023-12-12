package common

import (
	"github.com/kloudlite/api/pkg/repos"
)

type AuthSession struct {
	repos.BaseEntity `json:",inline"`
	UserId           repos.ID `json:"user_id"`
	UserEmail        string   `json:"user_email"`
	UserName         string   `json:"user_name"`
	UserVerified     bool     `json:"user_verified"`
	LoginMethod      string   `json:"login_method"`
}

type Json map[string]any