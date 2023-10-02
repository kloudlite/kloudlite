package iac

import "embed"

//go:embed templates/*
var TemplatesDir embed.FS

func ClusterPlanAndApplyTemplate() ([]byte, error) {
	return TemplatesDir.ReadFile("templates/cluster-plan-and-apply-job.yml.tpl")
}
