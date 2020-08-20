module github.com/anton-dessiatov/sctf/tf

go 1.14

replace (
	github.com/Azure/go-autorest v11.1.2+incompatible => github.com/Azure/go-autorest v12.1.0+incompatible
	k8s.io/client-go => k8s.io/client-go v0.0.0-20190620085101-78d2af792bab
)

require (
	cloud.google.com/go v0.61.0 // indirect
	github.com/aws/aws-sdk-go v1.33.21 // indirect
	github.com/bmatcuk/doublestar v1.2.1 // indirect
	github.com/fatih/color v1.9.0 // indirect
	github.com/go-sql-driver/mysql v1.5.0
	github.com/hashicorp/go-hclog v0.10.0
	github.com/hashicorp/go-multierror v1.1.0 // indirect
	github.com/hashicorp/go-plugin v1.3.0
	github.com/hashicorp/go-version v1.2.1 // indirect
	github.com/hashicorp/hcl/v2 v2.6.0
	github.com/hashicorp/terraform v0.13.0
	github.com/jinzhu/gorm v1.9.16
	github.com/kr/text v0.2.0 // indirect
	github.com/mattn/go-colorable v0.1.7 // indirect
	github.com/mitchellh/cli v1.0.0
	github.com/mitchellh/colorstring v0.0.0-20190213212951-d06e56a500db
	github.com/mitchellh/go-testing-interface v1.0.4 // indirect
	github.com/niemeyer/pretty v0.0.0-20200227124842-a10e7caefd8e // indirect
	github.com/spf13/afero v1.2.2
	github.com/spf13/cobra v1.0.0
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/stretchr/testify v1.6.1 // indirect
	github.com/ulikunitz/xz v0.5.7 // indirect
	github.com/zclconf/go-cty v1.5.1
	gopkg.in/check.v1 v1.0.0-20200227125254-8fa46927fb4f // indirect
	gopkg.in/yaml.v2 v2.3.0
)
