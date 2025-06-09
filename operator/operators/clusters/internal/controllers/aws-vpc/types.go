package aws_vpc

type TFValues struct {
	AwsAccessKey  string `json:"aws_access_key"`
	AwsSecretKey  string `json:"aws_secret_key"`
	AwsRegion     string `json:"aws_region"`
	AwsAssumeRole struct {
		Enabled    bool   `json:"enabled"`
		RoleARN    string `json:"role_arn"`
		ExternalID string `json:"external_id"`
	} `json:"aws_assume_role"`
	VpcName       string `json:"vpc_name"`
	VpcCIDR       string `json:"vpc_cidr"`
	PublicSubnets []struct {
		AvailabilityZone string `json:"availability_zone"`
		CIDR             string `json:"cidr"`
	} `json:"public_subnets"`
	Tags map[string]string `json:"tags"`
}
