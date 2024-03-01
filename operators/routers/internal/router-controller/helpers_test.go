package router_controller

import (
	"reflect"
	"testing"
)

func TestFilterDomains(t *testing.T) {
	type args struct {
		wildcardPatterns []string
		domains          []string
	}
	tests := []struct {
		name                   string
		args                   args
		wantWildcardDomains    []string
		wantNonWildcardDomains []string
	}{
		{
			name: "1. wildcard patterns only with *.",
			args: args{
				wildcardPatterns: []string{"*.example.com"},
				domains: []string{
					"sample.example.com",
					"example.com",
					"abc.xyz.example.com",
				},
			},
			wantWildcardDomains:    []string{"sample.example.com", "example.com"},
			wantNonWildcardDomains: []string{"abc.xyz.example.com"},
		},
		{
			name: "2. empty wildcard patterns",
			args: args{
				wildcardPatterns: nil,
				domains: []string{
					"sample.example.com",
					"example.com",
					"abc.xyz.example.com",
				},
			},
			wantWildcardDomains:    nil,
			wantNonWildcardDomains: []string{"sample.example.com", "example.com", "abc.xyz.example.com"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotWildcardDomains, gotNonWildcardDomains := FilterDomains(tt.args.wildcardPatterns, tt.args.domains)
			if !reflect.DeepEqual(gotWildcardDomains, tt.wantWildcardDomains) {
				t.Errorf("FilterDomains() gotWildcardDomains = %v, want %v", gotWildcardDomains, tt.wantWildcardDomains)
			}
			if !reflect.DeepEqual(gotNonWildcardDomains, tt.wantNonWildcardDomains) {
				t.Errorf("FilterDomains() gotNonWildcardDomains = %v, want %v", gotNonWildcardDomains, tt.wantNonWildcardDomains)
			}
		})
	}
}
