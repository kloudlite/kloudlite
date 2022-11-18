package controllers

import (
	"fmt"
)

func GetClusterIssuerName(region string) string {
	if region == "" {
		return "kl-cert-issuer"
	}
	return fmt.Sprintf("kl-cert-issuer-%s", region)
}

func GetCertIssuerSecret(region string) string {
	return GetClusterIssuerName(region) + "-tls"
}

func GetIngressClassName(region string) string {
	if region == "" {
		return "ingress-nginx"
	}
	return fmt.Sprintf("ingress-nginx-%s", region)
}
