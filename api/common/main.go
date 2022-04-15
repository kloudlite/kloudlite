package common

type AuthSession struct {
	Id           string `json:"id"`
	UserId       string `json:"userId"`
	UserEmail    string `json:"userEmail"`
	UserVerified bool   `json:"userVerified"`
	LoginMethod  string `json:"loginMethod"`
}
