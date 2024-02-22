package router_controller

import "strings"

func FilterDomains(wildcardPatterns []string, domains []string) (wildcardDomains, nonWildcardDomains []string) {
	wildcardBases := map[string]struct{}{}
	for _, pattern := range wildcardPatterns {
		if strings.HasPrefix(pattern, "*.") {
			wildcardBases[pattern[2:]] = struct{}{}
			continue
		}
		wildcardBases[pattern] = struct{}{}
	}

	for _, domain := range domains {
		if _, ok := wildcardBases[domain]; ok {
			wildcardDomains = append(wildcardDomains, domain)
			continue
		}

		sp := strings.SplitN(domain, ".", 2)
		if len(sp) != 2 {
			nonWildcardDomains = append(nonWildcardDomains, domain)
			continue
		}

		if _, ok := wildcardBases[sp[1]]; ok {
			wildcardDomains = append(wildcardDomains, domain)
			continue
		}

		nonWildcardDomains = append(nonWildcardDomains, domain)
	}

	return
}
