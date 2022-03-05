package domain

type User struct {
	Name     string `json:"name",bson:"name,required"`
	Email    string `json:"email"`
	Password string `json:"password", bson:"password"`
	Avatar   string `json:"avatar",bson:"avatar"`

	ProviderGithub struct {
		TokenId string `json:"token_id",bson:"token_id"`
		Avatar  string `json:"avatar",bson:"avatar"`
	} `json:"provider_github",bson:"provider_github"`

	ProviderGitlab struct {
		TokenId string `json:"token_id",bson:"token_id"`
		Avatar  string `json:"avatar",bson:"avatar"`
	} `json:"provider_github",bson:"provider_github"`

	ProviderGoogle struct {
		TokenId string `json:"token_id",bson:"token_id"`
		Avatar  string `json:"avatar",bson:"avatar"`
	} `json:"provider_github", bson:"provider_github"`

	Invite string `json:"invite",bson:"invite"`

	Verified bool `json:"verified",bson:"verified"`

	Metadata bool `json:"metadata",bson:"metadata"`
}

type UserDomain interface {
	Create(user *User) error
	Get(id string) (*User, error)
	GetByEmail(email string) (*User, error)
	GetByInvite(invite string) (*User, error)
	GetByProvider(provider string, tokenId string) (*User, error)
	Update(user *User) error
	Delete(id string) error
}
