package domain

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/iam"
)

func (d *domain) ValidateProviderSecret(providerName string, accessKeyId, secretAccessKey string) error {
	switch providerName {
	case "aws":
		{
			return validateAwsKeys(accessKeyId, secretAccessKey)
		}
	default:
		{
			return fmt.Errorf("provider %s is not supported", providerName)
		}
	}

	return nil
}

func validateAwsKeys(accessKeyID, secretAccessKey string) error {
	requiredActions := []string{"iam:*", "s3:*", "ec2:*"}

	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String("us-west-2"),
		Credentials: credentials.NewStaticCredentials(accessKeyID, secretAccessKey, ""),
	})

	if err != nil {
		return fmt.Errorf("Error creating session: %s", err)
	}

	svc := iam.New(sess)
	userOutput, err := svc.GetUser(nil)

	if err != nil {
		return fmt.Errorf("Failed to get user: %s", err)
	}

	// Getting the policies attached to the user
	listUserPoliciesInput := &iam.ListAttachedUserPoliciesInput{
		UserName: userOutput.User.UserName,
	}

	policies, err := svc.ListAttachedUserPolicies(listUserPoliciesInput)

	if err != nil {
		return fmt.Errorf("Failed to list user policies: %s", err)
	}

	for _, attachedPolicy := range policies.AttachedPolicies {
		policyOutput, err := svc.GetPolicy(&iam.GetPolicyInput{
			PolicyArn: attachedPolicy.PolicyArn,
		})

		if err != nil {
			// fmt.Println("Failed to get policy:", err)
			continue
		}

		policyVersionOutput, err := svc.GetPolicyVersion(&iam.GetPolicyVersionInput{
			PolicyArn: policyOutput.Policy.Arn,
			VersionId: policyOutput.Policy.DefaultVersionId,
		})

		if err != nil {
			// fmt.Println("Failed to get policy version:", err)
			continue
		}

		s, err := url.QueryUnescape(*policyVersionOutput.PolicyVersion.Document)
		if err != nil {
			// fmt.Println("Failed to get policy version:", err)
			continue
		}

		b, err := validatePolicyJson(s, requiredActions)
		if err != nil {
			// fmt.Println("Permission Check Error:", err)
			continue
		}

		if b {
			return nil
		}
	}
	return fmt.Errorf("coudn't find the required permissions ( full access to [S3,EC2,IAM] on all resources )")

}

func validatePolicyJson(policyJSON string, actions []string) (bool, error) {
	type Statement struct {
		Action   []string `json:"Action"`
		Resource string   `json:"Resource"`
	}

	type Policy struct {
		Statement []Statement `json:"Statement"`
	}

	var policy Policy
	if err := json.Unmarshal([]byte(policyJSON), &policy); err != nil {
		return false, fmt.Errorf("Failed to unmarshal policy document: %v", err)
	}

	for _, s := range policy.Statement {
		if s.Resource != "*" {
			continue
		}
		found := false
		for _, v := range actions {
			f := false
			for _, v2 := range s.Action {
				if v == v2 {
					f = true
				}
			}
			if !f {
				found = false
				break
			}
			found = true
		}
		if found {
			return true, nil
		}
	}

	return false, fmt.Errorf("permissions not matched")
}

func contains(slice []string, value string) bool {
	for _, v := range slice {
		if v == value {
			return true
		}
	}
	return false
}
