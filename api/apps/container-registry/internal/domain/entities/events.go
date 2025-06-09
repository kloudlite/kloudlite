package entities

type RepoReference struct {
	MediaType string `json:"mediaType"`
	Size      int    `json:"size"`
	Digest    string `json:"digest"`
}

type Event struct {
	ID        string `json:"id"`
	Timestamp string `json:"timestamp"`
	Action    string `json:"action"`
	Target    struct {
		MediaType  string          `json:"mediaType"`
		Size       int             `json:"size"`
		Digest     string          `json:"digest"`
		Length     int             `json:"length"`
		Repository string          `json:"repository"`
		URL        string          `json:"url"`
		Tag        string          `json:"tag"`
		References []RepoReference `json:"references"`
	} `json:"target"`
	Request struct {
		ID        string `json:"id"`
		Addr      string `json:"addr"`
		Host      string `json:"host"`
		Method    string `json:"method"`
		Useragent string `json:"useragent"`
	} `json:"request"`
	Actor struct {
		Name string `json:"name"`
	} `json:"actor"`
	Source struct {
		Addr       string `json:"addr"`
		InstanceID string `json:"instanceID"`
	} `json:"source"`
}

type EventMessage struct {
	Events []Event
}
