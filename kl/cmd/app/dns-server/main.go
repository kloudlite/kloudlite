package dnsserver

import (
	"fmt"
	"strings"
	"time"

	"github.com/kloudlite/kl/domain/client"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/miekg/dns"
)

func GetDnsServers() []string {
	dns := []string{"1.1.1.1:53"}

	dc, err := client.GetDeviceContext()
	if err != nil {
		return dns
	}

	dns2 := []string{}
	for _, v := range dc.DeviceDns {
		dns2 = append(dns2, fmt.Sprintf("%s:%d", v, 53))
	}

	dns2 = append(dns2, dns...)

	return dns2
}

var healthyServers = make(map[string]bool)

func StartDnsServer(addr string) error {
	// Initial health check
	for _, server := range GetDnsServers() {
		healthyServers[server] = checkHealth(server)
	}

	// Start health check goroutine
	go healthCheckRoutine()

	// Set up DNS server
	dns.HandleFunc(".", handleDNSRequest)
	server := &dns.Server{Addr: addr, Net: "udp"}
	fn.Println("Starting DNS proxy server on", addr)
	err := server.ListenAndServe()
	if err != nil {
		return err
	}

	return nil
}

func healthCheckRoutine() {
	for {
		for _, server := range GetDnsServers() {
			healthyServers[server] = checkHealth(server)

			if !healthyServers[server] {
				delete(healthyServers, server)
			}
		}
		time.Sleep(5 * time.Second)
	}
}

func checkHealth(server string) bool {
	client := new(dns.Client)
	message := new(dns.Msg)
	message.SetQuestion("one.one.one.one.", dns.TypeA)
	response, _, err := client.Exchange(message, server)
	if err != nil || response.Rcode != dns.RcodeSuccess {
		fn.Logf("DNS server %s is unhealthy\n", server)
		return false
	}
	// fn.Printf("DNS server %s is healthy\n", server)
	return true
}

func handleDNSRequest(w dns.ResponseWriter, r *dns.Msg) {

	env, _ := client.CurrentEnv()

	// t := time.Now()
	// defer func() {
	// 	fn.Logf("DNS request took %s", time.Since(t))
	// }()

	if done := func() bool {
		if len(r.Question) > 0 &&
			(strings.Count(r.Question[0].Name, ".") == 1 ||
				(env != nil && (env.ClusterName != "" && env.TargetNs != "") && (strings.HasSuffix(r.Question[0].Name, fmt.Sprintf("%s.svc.%s.local.", env.TargetNs, env.ClusterName))))) {
			return false
		}

		client := new(dns.Client)
		// fn.Println("[#1] asking to server", server, r.Question[0].Name)
		response, _, err := client.Exchange(r, "1.1.1.1:53")
		if err == nil {
			if len(response.Answer) == 0 {
				return false
			}

			w.WriteMsg(response)
			return true
		}
		return false
	}(); done {
		return
	}

	if done := func() bool {
		searchDomain := ""
		if env != nil && (env.ClusterName != "" && env.TargetNs != "") {
			searchDomain = fmt.Sprintf("%s.svc.%s.local", env.TargetNs, env.ClusterName)
		}

		if searchDomain == "" {
			return false
		}

		questionMap := make(map[string]string)
		for i, question := range r.Question {
			if !(strings.HasSuffix(question.Name, ".local.") || strings.Count(question.Name, ".") == 1) {
				return false
			}

			originalName := question.Name
			if strings.Count(question.Name, ".") == 1 {
				r.Question[i].Name = strings.TrimSuffix(question.Name, ".") + "." + searchDomain + "."
			}
			questionMap[r.Question[i].Name] = originalName
		}

		for _, server := range GetDnsServers() {
			if healthyServers[server] {
				client := new(dns.Client)
				fn.Println("[#2] asking to server", server, r.Question[0].Name)
				response, _, err := client.Exchange(r, server)
				if err == nil {
					for i, r := range response.Answer {
						response.Answer[i].Header().Name = questionMap[r.Header().Name]
					}

					for i, q := range response.Question {
						if questionMap[q.Name] != "" {
							response.Question[i].Name = questionMap[q.Name]
							r.Question[i].Name = questionMap[q.Name]
						}
					}

					w.WriteMsg(response)
					return true
				}
				healthyServers[server] = false
			}
		}

		return true
	}(); done {
		return
	}

	m := new(dns.Msg)
	m.SetRcode(r, dns.RcodeServerFailure)
	w.WriteMsg(m)
}
