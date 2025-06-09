package aws

import (
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	fn "github.com/kloudlite/operator/pkg/functions"
)

func ListAllRegions(sess *session.Session) ([]string, error) {
	svc := ec2.New(sess)
	dro, err := svc.DescribeRegions(nil)
	if err != nil {
		return nil, err
	}

	regions := make([]string, 0, len(dro.Regions))

	for i := range dro.Regions {
		if dro.Regions[i] != nil {
			regions = append(regions, *dro.Regions[i].RegionName)
		}
	}

	return regions, nil
}

type AvailabilityZone struct {
	Name   string
	ZoneId string
}

// only for region defined in AWS Session
func ListAllAvailabilityZones(sess *session.Session) ([]AvailabilityZone, error) {
	svc := ec2.New(sess)
	out, err := svc.DescribeAvailabilityZones(&ec2.DescribeAvailabilityZonesInput{
		Filters: []*ec2.Filter{
			{
				Name:   fn.New("opt-in-status"),
				Values: []*string{fn.New("opt-in-not-required")},
			},
		},
	})
	if err != nil {
		return nil, err
	}

	azs := make([]AvailabilityZone, 0, len(out.AvailabilityZones))

	for i := range out.AvailabilityZones {
		if out.AvailabilityZones[i] != nil {
			azs = append(azs, AvailabilityZone{
				Name:   *out.AvailabilityZones[i].ZoneName,
				ZoneId: *out.AvailabilityZones[i].ZoneId,
			})
		}
	}

	return azs, nil
}
