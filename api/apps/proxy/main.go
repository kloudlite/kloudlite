package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"os"
	"sync"
)

func main() {

	type Service struct {
		Name   string `json:"name"`
		Port   int    `json:"proxyPort"`
		Target int    `json:"servicePort"`
	}

	var data struct {
		Services []Service `json:"services"`
	}

	confFile := os.Getenv("CONFIG_FILE")
	configData, err := ioutil.ReadFile(confFile)
	if err != nil {
		fmt.Println("Error reading config file:", confFile, err)
	}

	err = json.Unmarshal(configData, &data)
	if err != nil {
		fmt.Println("Error unmarshalling config file:", err)
	}

	for _, service := range data.Services {
		go func(service Service) {

			fmt.Println(fmt.Sprint(service.Name, ":", service.Target, "->", service.Port))

			listener, err := net.Listen("tcp", fmt.Sprintf(":%d", service.Port))
			if err != nil {
				fmt.Println("Error listening:", err)
				return
			}
			defer listener.Close()

			for {
				conn, err := listener.Accept()
				fmt.Println("Accepted connection from: R", conn.RemoteAddr())
				fmt.Println("Accepted connection from: L", conn.LocalAddr())
				if err != nil {
					fmt.Println("Error accepting connection: ", err)
					continue
				}
				upconn, err := net.Dial("tcp", fmt.Sprint(service.Name, ":", service.Target))
				if err != nil {
					fmt.Println("Error dialing target: ", err)
					continue
				}
				go func() {
					io.Copy(conn, upconn)
					conn.Close()
					upconn.Close()
				}()
				go func() {
					io.Copy(upconn, conn)
					upconn.Close()
					conn.Close()
				}()
			}
		}(service)
	}
	fmt.Println("running")

	var wg sync.WaitGroup
	wg.Add(1)
	wg.Wait()
}
