package main

import (
	"encoding/json"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"io"
	"net"
	"os"
	"sync"
)

type Service struct {
	Name     string `json:"name"`
	Target   int    `json:"servicePort"`
	Port     int    `json:"proxyPort"`
	Listener net.Listener
	Closed   bool
}

var ServiceMap map[string]*Service

func reloadConfig(configData []byte) error {
	var data struct {
		Services []Service `json:"services"`
	}
	if configData == nil {
		confFile := os.Getenv("CONFIG_FILE")
		configData, err := os.ReadFile(confFile)
		if err != nil {
			return err
		}
		err = json.Unmarshal(configData, &data)
	} else {
		err := json.Unmarshal(configData, &data)
		if err != nil {
			return err
		}
	}
	oldServiceMap := make(map[string]*Service)
	for _, service := range ServiceMap {
		oldServiceMap[getKey(service)] = service
	}
	ServiceMap = make(map[string]*Service)
	for key, _ := range data.Services {
		s := data.Services[key]
		if _, ok := oldServiceMap[getKey(&s)]; !ok {
			ServiceMap[getKey(&s)] = &s
		} else {
			ServiceMap[getKey(&s)] = oldServiceMap[getKey(&s)]
		}
	}
	for key, _ := range oldServiceMap {
		s := oldServiceMap[key]
		if _, ok := ServiceMap[key]; !ok {
			err := stopService(s)
			if err != nil {
				return err
			}
		}
	}
	for key, _ := range ServiceMap {
		s := ServiceMap[key]
		if _, ok := oldServiceMap[getKey(s)]; !ok {
			err := startService(s)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
func getKey(service *Service) string {
	return fmt.Sprint(service.Name, ":", service.Port, ":", service.Target)
}
func stopService(service *Service) error {
	if service.Listener != nil {
		err := service.Listener.Close()
		service.Closed = true
		fmt.Println("- stopping :: ", getKey(service))
		return err
	}
	return nil
}
func startService(service *Service) error {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", service.Port))
	if err != nil {
		return err
	}
	service.Listener = listener
	service.Closed = false
	go runLoop(service)
	return nil
}
func runLoop(service *Service) error {
	fmt.Println("+ starting :: ", getKey(service))
	for {
		if service.Closed || service.Listener == nil {
			return nil
		}
		conn, err := service.Listener.Accept()
		if err != nil {
			return err
		}
		go func() {
			upconn, err := net.Dial("tcp", fmt.Sprint(service.Name, ":", service.Target))
			if err != nil {
				return
			}
			defer upconn.Close()
			defer conn.Close()
			go io.Copy(upconn, conn)
			io.Copy(conn, upconn)
		}()
	}
}
func startApi() {
	app := fiber.New(fiber.Config{
		DisableStartupMessage: true,
	})
	app.Post("/post", func(c *fiber.Ctx) error {
		err := reloadConfig(c.Body())
		if err != nil {
			return err
		}
		c.Send([]byte("done"))
		return nil
	})
	app.Listen(":2999")
}
func main() {
	go startApi()
	err := reloadConfig(nil)
	if err != nil {
		panic(err)
	}
	var wg sync.WaitGroup
	wg.Add(1)
	wg.Wait()
}
