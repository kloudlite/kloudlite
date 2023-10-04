package entities

type GitlabWebhookId int
type GithubWebhookId int64

type GitlabGroup struct {
	Id        string `json:"id"`
	FullName  string `json:"full_name"`
	AvatarUrl string `json:"avatar_url"`
}
