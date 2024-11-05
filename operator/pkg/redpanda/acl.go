package redpanda

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/kloudlite/operator/pkg/errors"
	fn "github.com/kloudlite/operator/pkg/functions"
	exec2 "k8s.io/utils/exec"
)

type AdminClient interface {
	CreateUser(username, password string) error
	DeleteUser(username string) error
	UserExists(username string) (bool, error)
	CreateTopic(topicName string, partitionCount int) error
	DeleteTopic(topicName string) error
	TopicExists(topicName string) (bool, error)
	AllowUserOnTopics(username string, allowedOperations string, topicNames ...string) error
}

type adminCli struct {
	kafkaBrokers  string
	adminEndpoint string
	saslAuthFlags string
	adminAuthOpts *AdminAuthOpts
}

func exitCode(err error) int {
	if exitErr, ok := err.(exec2.ExitError); ok {
		return exitErr.ExitStatus()
	}
	return 17
}

func (a *adminCli) withAuthnIfAvailable() string {
	if a.adminAuthOpts != nil {
		return fmt.Sprintf("--user %q --password %q", a.adminAuthOpts.Username, a.adminAuthOpts.Password)
	}
	return ""
}

func (a *adminCli) UserExists(username string) (bool, error) {
	err, stdout, stderr := fn.Exec(
		fmt.Sprintf(
			"rpk acl user list %s --api-urls %s | grep -i %s", a.withAuthnIfAvailable(), a.adminEndpoint, username,
		),
	)

	if err != nil {
		if e, ok := err.(*exec.ExitError); ok {
			if len(stderr.String()) == 0 {
				return false, nil
			}
			return false, errors.NewEf(e, string(e.Stderr))
		}
		return false, err
	}

	foundUsername := strings.TrimSpace(stdout.String())

	if len(foundUsername) != len(username) || foundUsername != username {
		return false, nil
	}

	return true, nil
}

func (a *adminCli) TopicExists(topicName string) (bool, error) {
	err, _, stderr := fn.Exec(
		fmt.Sprintf(
			"rpk topic list --brokers %s %s | grep -i %s", a.kafkaBrokers, a.withAuthnIfAvailable(), topicName,
		),
	)
	if err != nil {
		if len(stderr.String()) == 0 {
			return false, nil
		}
		return false, errors.NewEf(err, stderr.String())
	}
	return true, nil
}

func (a *adminCli) CreateUser(username, password string) error {
	err, _, stderr := fn.Exec(fmt.Sprintf("rpk acl user create %s -p %s --api-urls %s", username, password, a.adminEndpoint))
	if err != nil {
		return errors.NewEf(err, stderr.String())
	}
	return nil
}

func (a *adminCli) DeleteUser(username string) error {
	err, _, stderr := fn.Exec(
		fmt.Sprintf(
			"rpk acl user delete %s  --api-urls %s %s", username, a.adminEndpoint,
			a.saslAuthFlags,
		),
	)
	if err != nil {
		return errors.NewEf(err, stderr.String())
	}
	return nil
}

func (a *adminCli) CreateTopic(topicName string, partitionCount int) error {
	cmd := fmt.Sprintf(
		"rpk topic create %s -p %d --brokers %s %s",
		topicName,
		partitionCount,
		a.kafkaBrokers,
		a.withAuthnIfAvailable(),
	)
	err, stdout, stderr := fn.Exec(cmd)
	fmt.Println(stdout.String())
	if err != nil {
		return errors.NewEf(err, stderr.String())
	}
	return nil
}

func (a *adminCli) DeleteTopic(topicName string) error {
	err, _, stderr := fn.Exec(fmt.Sprintf("rpk topic delete %s --brokers %s %s", topicName, a.kafkaBrokers, a.withAuthnIfAvailable()))
	if err != nil {
		return errors.NewEf(err, stderr.String())
	}
	return nil
}

func (a *adminCli) AllowUserOnTopics(username string, allowedOperations string, topicNames ...string) error {
	topicFlags := ""
	for i := range topicNames {
		topicFlags += " --topic " + topicNames[i]
	}

	err, _, stderr := fn.Exec(
		fmt.Sprintf(
			"rpk acl create --allow-principal %s --operation %s %s --brokers %s %s", username, allowedOperations, topicFlags, a.kafkaBrokers, a.withAuthnIfAvailable(),
		),
	)
	if err != nil {
		if len(stderr.String()) == 0 {
			return nil
		}
		return errors.NewEf(err, stderr.String())
	}
	return nil
}

type AdminAuthOpts struct {
	Username string
	Password string
}

func NewAdminClient(adminEndpoint string, kafkaBrokers string, opts *AdminAuthOpts) AdminClient {
	return &adminCli{
		kafkaBrokers:  kafkaBrokers,
		adminEndpoint: adminEndpoint,
		adminAuthOpts: opts,
		//saslAuthFlags: fmt.Sprintf("--user %s --password %s --sasl-mechanism 'SCRAM-SHA-256'", username, password),
	}
}
