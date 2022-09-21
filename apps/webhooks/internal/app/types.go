package app

type HttpHook struct {
	Body        []byte            `json:"body"`
	Headers     map[string]string `json:"headers"`
	Url         string            `json:"url"`
	QueryParams []byte            `json:"queryParams"`
}
