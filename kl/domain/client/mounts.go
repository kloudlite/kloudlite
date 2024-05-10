package client

import (
	"fmt"
	"strings"

	"github.com/kloudlite/kl/pkg/functions"
)

type CSType string

const (
	ConfigType CSType = "config"
	SecretType CSType = "secret"
)

type Mount struct {
	Path      string  `json:"path"`
	ConfigRef *string `json:"configRef,omitempty" yaml:"configRef,omitempty"`
	SecretRef *string `json:"secretRef,omitempty" yaml:"secretRef,omitempty"`
}

type FileEntry struct {
	Path string `json:"path"`
	Type CSType `json:"type"`
	Name string `json:"Name"`
	Key  string `json:"key"`
}

type Mounts []Mount

func (m *Mounts) GetMounts() []FileEntry {
	resp := make([]FileEntry, 0)
	if m == nil {
		return resp
	}

	hist := map[string]bool{}

	for _, mt := range *m {
		if hist[mt.Path] {
			continue
		}

		var ref *string

		if mt.ConfigRef != nil {
			ref = mt.ConfigRef
		} else if mt.SecretRef != nil {
			ref = mt.SecretRef
		}

		if ref == nil {
			continue
		}

		s := strings.Split(*ref, "/")
		if len(s) != 2 {
			continue
		}

		mName, mKey := s[0], s[1]
		resp = append(resp, FileEntry{
			Path: mt.Path,
			Type: func() CSType {
				if mt.ConfigRef != nil {
					return ConfigType
				}

				return SecretType
			}(),
			Name: mName,
			Key:  mKey,
		})
	}

	return resp
}

func (m *Mounts) AddMounts(fes []FileEntry) {
	if m == nil {
		m = &Mounts{}
	}

	hist := map[string]bool{}

	for _, mt := range *m {
		hist[mt.Path] = true
	}

	for _, fe := range fes {
		if hist[fe.Path] {
			continue
		}

		*m = append(*m, Mount{
			Path: fe.Path,
			ConfigRef: func() *string {
				if fe.Type == ConfigType {
					return functions.Ptr(fmt.Sprint(fe.Name, "/", fe.Key))
				}

				return nil
			}(),
			SecretRef: func() *string {
				if fe.Type == SecretType {
					return functions.Ptr(fmt.Sprint(fe.Name, "/", fe.Key))
				}

				return nil
			}(),
		})
	}
}
