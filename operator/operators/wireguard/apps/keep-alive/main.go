package main

import (
	"fmt"
	"net"
	"os"
	"sync"
	"time"

	"github.com/kloudlite/operator/operators/wireguard/apps/keep-alive/env"
	"github.com/kloudlite/operator/operators/wireguard/apps/keep-alive/types"
	"github.com/kloudlite/operator/operators/wireguard/apps/keep-alive/utils"
	"github.com/kloudlite/operator/pkg/logging"
)

func main() {
	ev := env.GetEnvOrDie()

	l := logging.NewOrDie(&logging.Options{
		Name: "keep-alive",
		Dev:  true,
	})

	r := run{
		ev:     ev,
		logger: l,
		id:     0,
	}

	if err := r.Run(); err != nil {
		panic(err)
	}

}

func (c *run) hc(conn *net.UDPConn) error {
	// defer func() {
	// 	time.Sleep(5 * time.Second)
	// }()

	message := GetMsg()

	if err := conn.SetDeadline(time.Now().Add(100 * time.Millisecond)); err != nil {
		return err
	}

	if _, err := conn.Write(message); err != nil {
		return err
	}

	return nil
}

var msg []byte

func GetMsg() []byte {
	if msg == nil {
		msg = make([]byte, 8)
	}
	// Create ICMP Echo request

	msg[0] = 8 // ICMP type for Echo request
	msg[1] = 0 // ICMP code for Echo request
	msg[2] = 0 // ICMP checksum (auto-calculated by the system)
	msg[3] = 0 // ICMP checksum (auto-calculated by the system)
	msg[4] = 0 // Identifier (arbitrary)
	msg[5] = 0 // Identifier (arbitrary)
	msg[6] = 0 // Sequence number (arbitrary)
	msg[7] = 0 // Sequence number (arbitrary)

	return msg
}

type run struct {
	logger logging.Logger
	ev     *env.Env
	id     int
}

func (r *run) Run() error {
	b, err := os.ReadFile(r.ev.ConfigPath)
	if err != nil {
		return err
	}

	var conf types.Conf

	if err := conf.FromYaml(b); err != nil {
		return err
	}

	for {
		for _, deviceCidr := range conf.Cidrs {
			if err := r.pingCidr(deviceCidr); err != nil {
				fmt.Println(err)
			}
		}

		time.Sleep(time.Duration(conf.Interval) * time.Second)
	}
}

func (r *run) pingCidr(cidr string) error {
	count := 0

	r.id++
	log := r.logger.WithName(fmt.Sprint(r.id))

	log.Infof("pinging CIDR: %s", cidr)

	defer func() {
		log.Infof("finished pinging CIDR: %s", cidr)
	}()

	wg := sync.WaitGroup{}
	for {
		s, err := utils.GenIPAddr(cidr, count)
		if err != nil {
			if err == utils.ErrIPsMaxedOut {
				break
			}

			log.Errorf(err, "Error getting Ip")
			time.After(1 * time.Second)
			continue
		}
		wg.Add(1)

		go func(ip string) {
			defer wg.Done()

			if err := pingIP(ip); err != nil {
				// log.Errorf(err, "Error pinging IP: %s", ip)
			} else {
				log.Infof("success ip: %s", ip)
			}

		}(s)

		count++
		// time.Sleep(10 * time.Millisecond)
	}

	log.Infof("added %d IPs to ping", count)

	wg.Wait()
	return nil
}

// Function to ping an IP using ICMP Echo
func pingIP(ip string) error {
	conn, err := net.Dial("ip4:icmp", ip)
	if err != nil {
		return err
	}
	defer conn.Close()

	// Send ICMP Echo request
	_, err = conn.Write(GetMsg())
	if err != nil {
		return err
	}

	// Wait for ICMP Echo reply (if any)
	reply := make([]byte, 8)
	conn.SetReadDeadline(time.Now().Add(100 * time.Millisecond))
	_, err = conn.Read(reply)
	if err != nil {
		return err
	}

	return nil
}
