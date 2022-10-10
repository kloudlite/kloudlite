package redpanda

import (
	"fmt"
	"os/exec"
	"strings"

	exec2 "k8s.io/utils/exec"
	"operators.kloudlite.io/lib/errors"
	fn "operators.kloudlite.io/lib/functions"
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
	username      string
	password      string
}

func exitCode(err error) int {
	if exitErr, ok := err.(exec2.ExitError); ok {
		return exitErr.ExitStatus()
	}
	return 17
}

func (admin adminCli) UserExists(username string) (bool, error) {
	err, stdout, stderr := fn.Exec(
		fmt.Sprintf(
			"rpk acl user list --user %s --password '%s' --api-urls %s | grep -i %s",
			admin.username,
			admin.password,
			admin.adminEndpoint,
			username,
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

func (admin adminCli) TopicExists(topicName string) (bool, error) {
	err, _, stderr := fn.Exec(
		fmt.Sprintf(
			"rpk topic list --brokers %s %s | grep -i %s", admin.kafkaBrokers, admin.saslAuthFlags, topicName,
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

func (admin adminCli) CreateUser(username, password string) error {
	err, _, stderr := fn.Exec(fmt.Sprintf("rpk acl user create %s -p %s --api-urls %s", username, password, admin.adminEndpoint))
	if err != nil {
		return errors.NewEf(err, stderr.String())
	}
	return nil
}

func (admin adminCli) DeleteUser(username string) error {
	err, _, stderr := fn.Exec(
		fmt.Sprintf(
			"rpk acl user delete %s  --api-urls %s %s", username, admin.adminEndpoint,
			admin.saslAuthFlags,
		),
	)
	if err != nil {
		return errors.NewEf(err, stderr.String())
	}
	return nil
}

func (admin adminCli) CreateTopic(topicName string, partitionCount int) error {
	cmd := fmt.Sprintf(
		"rpk topic create %s -p %d --brokers %s %s",
		topicName,
		partitionCount,
		admin.kafkaBrokers,
		admin.saslAuthFlags,
	)
	err, stdout, stderr := fn.Exec(cmd)
	fmt.Println(stdout.String())
	if err != nil {
		return errors.NewEf(err, stderr.String())
	}
	return nil
}

func (admin adminCli) DeleteTopic(topicName string) error {
	err, _, stderr := fn.Exec(fmt.Sprintf("rpk topic delete %s --brokers %s %s", topicName, admin.kafkaBrokers, admin.saslAuthFlags))
	if err != nil {
		return errors.NewEf(err, stderr.String())
	}
	return nil
}

func (admin adminCli) AllowUserOnTopics(username string, allowedOperations string, topicNames ...string) error {
	topicFlags := ""
	for i := range topicNames {
		topicFlags += " --topic " + topicNames[i]
	}

	err, _, stderr := fn.Exec(
		fmt.Sprintf(
			"rpk acl create --allow-principal %s --operation %s %s --brokers %s %s", username, allowedOperations, topicFlags, admin.kafkaBrokers,
			admin.saslAuthFlags,
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

func NewAdminClient(username, password, kafkaBrokers, adminEndpoint string) AdminClient {
	return &adminCli{
		username:      username,
		password:      password,
		kafkaBrokers:  kafkaBrokers,
		adminEndpoint: adminEndpoint,
		saslAuthFlags: fmt.Sprintf("--user %s --password %s --sasl-mechanism 'SCRAM-SHA-256'", username, password),
	}
}
