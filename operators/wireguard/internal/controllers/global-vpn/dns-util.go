package globalvpn

import (
	"fmt"
	"strings"

	wgv1 "github.com/kloudlite/operator/apis/wireguard/v1"
	// "github.com/kloudlite/operator/operators/wireguard/apps/multi-cluster/mpkg/wg"
	rApi "github.com/kloudlite/operator/pkg/operator"
)

func (r *Reconciler) getCorednsConfig(req *rApi.Request[*wgv1.GlobalVPN], current string, corednsSvcIP string) (string, error) {
	obj, _ := req.Object, req.Context()

	updatedContent := current

	for _, p := range obj.Spec.Peers {
		if p.ClusterName == "" || p.ClusterName == r.Env.ClusterName {
			continue
		}

		// 		updatedContent = addOrUpdateSectionInString(updatedContent, fmt.Sprintf("%s.local:53", p.ClusterName), fmt.Sprintf(`
		//       errors
		//
		//       rewrite name regex (.*)\.svc\.%s\.local {1}.svc.cluster.local
		//
		//       forward . %s
		//
		//       cache 30
		//       loop
		//       reload
		//       loadbalance
		// `, p.ClusterName, p.IP))

		// as we are just modifying only one file, we can just upsert the section
		updatedContent += fmt.Sprintf(`
      %s.local:53 {
        errors

        rewrite name regex (.*)\.svc\.%s\.local {1}.svc.cluster.local

        forward . %s

        cache 30
        loop
        reload
        loadbalance
      }
`, p.ClusterName, p.ClusterName, p.IP)
	}

	return strings.TrimSpace(updatedContent), nil
}

// func addOrUpdateSectionInString(content, sectionName, newSectionContent string) string {
// 	// Split the content by lines
// 	lines := strings.Split(content, "\n")
//
// 	// Look for the section, and if found, remember its position
// 	var startIndex int = -1
// 	var endIndex int = -1
// 	for i, line := range lines {
// 		trimmedLine := strings.TrimSpace(line)
// 		if trimmedLine == sectionName+" {" {
// 			startIndex = i
// 		}
// 		if startIndex != -1 && trimmedLine == "}" {
// 			endIndex = i
// 			break
// 		}
// 	}
//
// 	// If the section exists, remove it
// 	if startIndex != -1 && endIndex != -1 {
// 		lines = append(lines[:startIndex], lines[endIndex+1:]...)
// 	}
//
// 	// Add or re-add the section at the end
// 	oldLines := lines
// 	lines = []string{}
// 	lines = append(lines, sectionName+" {")
// 	for _, line := range strings.Split(newSectionContent, "\n") {
// 		lines = append(lines, "    "+line) // Indent the content to match the style
// 	}
// 	lines = append(lines, "}")
//
// 	lines = append(lines, oldLines...)
//
// 	return strings.Join(lines, "\n")
// }
