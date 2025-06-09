package types

type MsvcOutput struct {
	Username string `json:"USERNAME"`
	Password string `json:"PASSWORD"`
	Bucket   string `json:"BUCKET"`
	Org      string `json:"ORG"`
	Host     string `json:"HOST"`
	Token    string `json:"TOKEN"`
	Uri      string `json:"URI"`
}

type MresOutput struct {
	BucketName string `json:"BUCKET_NAME"`
	BucketId   string `json:"BUCKET_ID"`
	OrgId      string `json:"ORG_ID"`
	OrgName    string `json:"ORG_NAME"`
	Token      string `json:"TOKEN"`
	Uri        string `json:"URI"`
}
