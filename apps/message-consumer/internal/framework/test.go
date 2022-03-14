package framework

import "kloudlite.io/apps/message-consumer/internal/app"

var DataConfigCreate = &app.Message{
	Action:       "create",
	ResourceId:   "proj-3ytyfo-fegigkuxign8nyqztzhsp8pjz-kl",
	ResourceType: "config",
	Metadata: map[string]string{
		"name": "my-real-config-1",
	},
}

var DataConfigDelete = &app.Message{
	Action:       "delete",
	ResourceId:    "proj-3ytyfo-fegigkuxign8nyqztzhsp8pjz-kl",
	ResourceType: "config",
	Metadata: map[string]string{
		"name": "my-real-config-1",
	},
}

var DataSecretCreate = &app.Message{
	Action:       "create",
	ResourceId:    "proj-3ytyfo-fegigkuxign8nyqztzhsp8pjz-kl",
	ResourceType: "secret",
	Metadata: map[string]string{
		"name": "my-secret-1",
	},
}

var DataSecretDelete = &app.Message{
	Action:       "delete",
	ResourceId:    "proj-3ytyfo-fegigkuxign8nyqztzhsp8pjz-kl",
	ResourceType: "secret",
	Metadata: map[string]string{
		"name": "my-secret-1",
	},
}
