package domain

import (
	"context"
)

type ContactUsData struct {
	Name         string `json:"name"`
	Email        string `json:"email"`
	MobileNumber int64  `json:"mobileNumber"`
	CompanyName  string `json:"CompanyName"`
	Country      string `json:"country"`
	Message      string `json:"message"`
}

type Domain interface {
	SendContactUsEmail(ctx context.Context, contactUsData *ContactUsData) error
}
