package wg

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/kloudlite/operator/apps/multi-cluster/constants"
	"github.com/kloudlite/operator/pkg/logging"
	"golang.zx2c4.com/wireguard/wgctrl"
)

func (c *client) execCmd(cmdString string, env map[string]string) ([]byte, error) {
	out, errOut := &bytes.Buffer{}, &bytes.Buffer{}

	r := csv.NewReader(strings.NewReader(cmdString))
	r.Comma = ' '
	cmdArr, err := r.Read()
	if err != nil {
		return nil, err
	}
	cmd := exec.Command(cmdArr[0], cmdArr[1:]...)
	if c.verbose {
		c.logger.Infof(strings.Join(cmdArr, " "))

		cmd.Stdout = out
	}

	if env == nil {
		env = map[string]string{}
	}

	for k, v := range env {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
	}

	cmd.Stderr = errOut
	if err := cmd.Run(); err != nil {
		return errOut.Bytes(), err
	}
	return out.Bytes(), nil
}

func (c *client) Stop() error {
	b, err := c.execCmd(fmt.Sprintf("ip link delete %s", c.ifName), nil)
	if err != nil {
		c.logger.Infof(string(b))
		return err
	}

	return nil
}

func (c *client) startWg(conf []byte) error {
	if err := os.WriteFile(fmt.Sprintf("/etc/wireguard/%s.conf", c.ifName), conf, 0644); err != nil {
		return err
	}

	b, err := c.execCmd(fmt.Sprintf("wg-quick up %s", c.ifName), nil)
	if err != nil {
		if c.verbose {
			c.logger.Error(err)
			c.logger.Infof(string(b))
		}
		return err
	}

	c.logger.Infof("started")
	return nil
}

var prevConf []byte

func (c *client) updateWg(conf []byte) error {

	fName := fmt.Sprintf("/etc/wireguard/%s.conf", c.ifName)

	if prevConf != nil && string(prevConf) == string(conf) {
		c.logger.Infof("configuration is the same, skipping update")
		return nil
	}

	prevConf = conf

	if err := os.WriteFile(fName, conf, 0644); err != nil {
		return err
	}

	b, err := c.execCmd(fmt.Sprintf("wg-quick down %s", c.ifName), nil)
	if err != nil {
		c.logger.Error(err)
		fmt.Println(string(b))
		return err
	}

	if err := os.WriteFile(fName, conf, 0644); err != nil {
		return err
	}

	b, err = c.execCmd(fmt.Sprintf("wg-quick up %s", c.ifName), nil)

	if err != nil {
		c.logger.Error(err)
		fmt.Println(string(b))
		return err
	}

	c.logger.Infof("configuration updated")

	return nil
}

func (c *client) Sync(conf []byte) error {
	if !c.IsConfigured() {
		c.logger.Infof("not running, starting it")
		if err := c.startWg(conf); err != nil {
			c.logger.Error(err)
		}
		return nil
	}

	c.logger.Infof("already running, updating configuration")
	if err := c.updateWg(conf); err != nil {
		c.logger.Error(err)
	}
	return nil
}

type client struct {
	logger  logging.Logger
	ifName  string
	verbose bool
}

type Client interface {
	Sync(conf []byte) error
	Stop() error
}

func (c *client) IsConfigured() bool {
	cl, err := wgctrl.New()
	if err != nil {
		if c.verbose {
			c.logger.Error(err)
		}
		return false
	}

	dev, err := cl.Device(c.ifName)
	if err != nil {
		if c.verbose {
			c.logger.Error(err)
		}
		return false
	}

	return len(dev.Peers) > 0
}

func NewClient() (Client, error) {
	l, err := logging.New(&logging.Options{})
	if err != nil {
		return nil, err
	}

	return &client{
		logger:  l.WithName("wirec"),
		ifName:  constants.IfaceName,
		verbose: false,
	}, nil
}
