package aws

import (
	"fmt"
	"net"
	"reflect"
	"sort"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/route53"
)

type Route53 interface {
	UpdateRecord(site string, aRecords []string, hostedZone string) error
	getRecord(site string, hostedZone *string, zoneId *string) (*route53.ResourceRecordSet, error)
	DeleteRecord(site string, hostedZone *string, zoneId *string) error
}

type aws_route53 struct {
	svc *route53.Route53
}

func IsStringSliceEqual(a, b []string) bool {
	sort.Strings(a)
	sort.Strings(b)
	if len(a) != len(b) {
		return false
	}

	return reflect.DeepEqual(a, b)
}

func getRoute53Client(accessKey, accessSecret string) (*route53.Route53, error) {
	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String("us-west-1"),
		Credentials: credentials.NewStaticCredentials(accessKey, accessSecret, ""),
	})
	if err != nil {
		return nil, err
	}

	return route53.New(sess), nil
}

func (a aws_route53) DeleteRecord(site string, hostedZone *string, zoneId *string) error {

	var err error
	if zoneId == nil {
		zoneId, err = a.getHid(hostedZone)
		if err != nil {
			return err
		}
	}

	var rSet *route53.ResourceRecordSet
	rSet, err = a.getRecord(site, nil, zoneId)
	if err != nil {
		return err
	}

	changeBatch := &route53.ChangeBatch{
		Changes: []*route53.Change{
			{
				Action:            aws.String(route53.ChangeActionDelete),
				ResourceRecordSet: rSet,
			},
		},
	}

	// Create the change input
	changeInput := &route53.ChangeResourceRecordSetsInput{
		HostedZoneId: zoneId,
		ChangeBatch:  changeBatch,
	}

	// Call the ChangeResourceRecordSets API
	_, err = a.svc.ChangeResourceRecordSets(changeInput)
	if err != nil {
		return err
	}

	return nil
}

func (a aws_route53) UpdateRecord(site string, aRecords []string, hostedZone string) error {
	zoneId, err := a.getHid(aws.String(hostedZone))
	if err != nil {
		return err
	}

	if len(aRecords) == 0 {
		if err := a.DeleteRecord(site, nil, zoneId); err != nil {
			return err
		}

		return nil
	}

	rRecords := []*route53.ResourceRecord{}
	for _, v := range aRecords {
		rRecords = append(rRecords, &route53.ResourceRecord{
			Value: aws.String(v),
		})
	}

	changeBatch := &route53.ChangeBatch{
		Changes: []*route53.Change{
			{
				Action: aws.String(route53.ChangeActionUpsert),
				ResourceRecordSet: &route53.ResourceRecordSet{
					Name:            aws.String(fmt.Sprintf("%s.", site)),
					Type:            aws.String(route53.RRTypeA),
					TTL:             aws.Int64(300),
					ResourceRecords: rRecords,
				},
			},
		},
	}

	// Create the change input
	changeInput := &route53.ChangeResourceRecordSetsInput{
		HostedZoneId: zoneId,
		ChangeBatch:  changeBatch,
	}

	// Call the ChangeResourceRecordSets API
	_, err = a.svc.ChangeResourceRecordSets(changeInput)
	if err != nil {
		return err
	}

	return nil
}

func (a aws_route53) getHid(hZone *string) (*string, error) {
	if hZone == nil {
		return nil, fmt.Errorf("hosted zone is required")
	}

	result, err := a.svc.ListHostedZones(nil)
	if err != nil {
		return nil, err
	}

	zoneId := ""
	// Print the hosted zones
	for _, zone := range result.HostedZones {
		if *zone.Name == fmt.Sprintf("%s.", *hZone) {
			zoneId = *zone.Id
			break
		}
	}

	if zoneId == "" {
		return nil, fmt.Errorf("no zones with hostedzone %s found", *hZone)
	}

	return aws.String(zoneId), nil
}

func (a aws_route53) getRecord(site string, hostedZone *string, zoneId *string) (*route53.ResourceRecordSet, error) {

	var err error
	if zoneId == nil {
		zoneId, err = a.getHid(hostedZone)
		if err != nil {
			return nil, err
		}
	}

	// fmt.Println(zoneId, "here")
	recordName := fmt.Sprintf("%s.", site)
	recordType := route53.RRTypeA

	_result, err := a.svc.ListResourceRecordSets(&route53.ListResourceRecordSetsInput{
		HostedZoneId:    zoneId,
		MaxItems:        aws.String("1"),
		StartRecordName: aws.String(recordName),
		StartRecordType: aws.String(recordType),
	})
	if err != nil {
		return nil, err
	}

	for _, recordSet := range _result.ResourceRecordSets {
		if *recordSet.Name == recordName && *recordSet.Type == recordType {
			return recordSet, nil
		}
	}

	return nil, fmt.Errorf("domain %s not found", site)
}

func GetARecordFromLive(host string) []string {
	addresses, err := net.LookupHost(host)
	if err != nil {
		return []string{}
	}

	return addresses
}

func NewAwsRoute53Client(accessKey, accessSec string) (Route53, error) {
	svc, err := getRoute53Client(accessKey, accessSec)
	if err != nil {
		return nil, err
	}

	a := aws_route53{
		svc: svc,
	}

	return a, nil
}
