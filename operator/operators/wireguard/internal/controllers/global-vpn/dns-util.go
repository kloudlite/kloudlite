package globalvpn

import (
	"fmt"
	"strings"

	wgv1 "github.com/kloudlite/operator/apis/wireguard/v1"
	// "github.com/kloudlite/operator/operators/wireguard/apps/multi-cluster/mpkg/wg"
	rApi "github.com/kloudlite/operator/pkg/operator"
)

func (r *Reconciler) getCorednsConfig(req *rApi.Request[*wgv1.GlobalVPN], current []byte) ([]byte, error) {
	obj, _ := req.Object, req.Context()

	updatedContent := string(current)

	for _, p := range obj.Spec.Peers {
		// ip, err := wg.GetRemoteDeviceIp(int64(p.Id), r.Env.WgIpBase)
		// if err != nil {
		// 	return nil, fmt.Errorf("failed to get remote device ip: %w", err)
		// }

		updatedContent = addOrUpdateSectionInString(updatedContent, fmt.Sprintf("cluster%d.local:53", 2), fmt.Sprintf(`
      errors
      forward . %s

      cache 30
      loop
      reload
      loadbalance
`, p.IP))
		// if err != nil {
		// 	return nil, fmt.Errorf("failed to add or update section in coredns config: %w", err)
		// }
	}

	return []byte(updatedContent), nil
}

func addOrUpdateSectionInString(content, sectionName, newSectionContent string) string {
	// Split the content by lines
	lines := strings.Split(content, "\n")

	// Look for the section, and if found, remember its position
	var startIndex int = -1
	var endIndex int = -1
	for i, line := range lines {
		trimmedLine := strings.TrimSpace(line)
		if trimmedLine == sectionName+" {" {
			startIndex = i
		}
		if startIndex != -1 && trimmedLine == "}" {
			endIndex = i
			break
		}
	}

	// If the section exists, remove it
	if startIndex != -1 && endIndex != -1 {
		lines = append(lines[:startIndex], lines[endIndex+1:]...)
	}

	// Add or re-add the section at the end
	oldLines := lines
	lines = []string{}
	lines = append(lines, sectionName+" {")
	for _, line := range strings.Split(newSectionContent, "\n") {
		lines = append(lines, "    "+line) // Indent the content to match the style
	}
	lines = append(lines, "}")

	lines = append(lines, oldLines...)

	return strings.Join(lines, "\n")
}
