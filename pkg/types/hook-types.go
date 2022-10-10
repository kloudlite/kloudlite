package types

type HttpHook struct {
	Body        []byte            `json:"body"`
	Headers     map[string]string `json:"headers"`
	Url         string            `json:"url"`
	QueryParams []byte            `json:"queryParams,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

type GitHttpHook struct {
	HttpHook    `json:",inline"`
	GitProvider string
}
