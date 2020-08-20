package terra

type Credentials struct {
	AWS AWS `yaml:"aws"`
	GCP GCP `yaml:"gcp"`
}

type AWS struct {
	AccessKey            string `yaml:"access_key"`
	SecretKey            string `yaml:"secret_key"`
	AssumeRoleARN        string `yaml:"assume_role_arn"`
	AssumeRoleExternalID string `yaml:"assume_role_external_id"`
}

type GCP struct {
	Project string `yaml:"project"`
	JsonKey string `yaml:"json_key"`
}
