package main

import (
	"bytes"
	"fmt"
	"os"
	"strings"

	"github.com/kloudlite/operator/pkg/aws"
	fn "github.com/kloudlite/operator/pkg/functions"
)

func toConstantName(name string) string {
	return strings.ToUpper(strings.ReplaceAll(name, "-", "_"))
}

func main() {
	accessKey := os.Getenv("ACCESS_KEY")
	secretKey := os.Getenv("SECRET_KEY")
	sess, err := aws.NewSessionWithStaticCreds(accessKey, secretKey)
	if err != nil {
		panic(err)
	}

	sess.Config.Region = fn.New("ap-south-1")

	regions, err := aws.ListAllRegions(sess)
	if err != nil {
		panic(err)
	}

	buff := new(bytes.Buffer)

	buff.WriteString("package v1\n")
	buff.WriteString("type AwsRegion string\n")
	buff.WriteString("const (\n")
	for i := range regions {
		buff.WriteString(fmt.Sprintf("AwsRegion_%s AwsRegion = %q\n", toConstantName(regions[i]), regions[i]))
	}
	buff.WriteString(")\n")

	m := map[string][]string{}

	buff.WriteString("type AwsAZ string\n")
	buff.WriteString("const (\n")
	for i := range regions {
		buff.WriteString("\n")
		sess.Config.Region = fn.New(regions[i])
		az, err := aws.ListAllAvailabilityZones(sess)
		if err != nil {
			panic(err)
		}
		names := make([]string, len(az))
		for i := range az {
			names[i] = az[i].Name
			buff.WriteString(fmt.Sprintf("AwsAZ_%s AwsAZ = %q\n", toConstantName(names[i]), names[i]))
		}
		m[regions[i]] = names
	}
	buff.WriteString(")\n")

	buff.WriteString("var AwsRegionToAZs = map[AwsRegion][]AwsAZ{\n")
	for k, v := range m {
		azs := make([]string, len(v))
		for i := range v {
			azs[i] = fmt.Sprintf("AwsAZ_%s", toConstantName(v[i]))
		}
		buff.WriteString(fmt.Sprintf("AwsRegion_%s: []AwsAZ{%s},\n", toConstantName(k), strings.Join(azs, ",")))
	}
	buff.WriteString("}\n")

	fmt.Printf("%s", buff.Bytes())
}
