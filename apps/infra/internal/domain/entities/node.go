package entities

import (
	"encoding/json"
	cmgrV1 "github.com/kloudlite/cluster-operator/apis/cmgr/v1"
	infraV1 "github.com/kloudlite/cluster-operator/apis/infra/v1"
	"io"
	"kloudlite.io/pkg/repos"
)

type MasterNode struct {
	repos.BaseEntity  `json:",inline"`
	cmgrV1.MasterNode `json:",inline"`
}

func (m *MasterNode) UnmarshalGQL(v interface{}) error {
	switch t := v.(type) {
	case map[string]any:
		b, err := json.Marshal(t)
		if err != nil {
			return err
		}

		if err := json.Unmarshal(b, m); err != nil {
			return err
		}

	case string:
		if err := json.Unmarshal([]byte(t), m); err != nil {
			return err
		}
	}

	return nil
}

func (m MasterNode) MarshalGQL(w io.Writer) {
	b, err := json.Marshal(m)
	if err != nil {
		w.Write([]byte("{}"))
	}
	w.Write(b)
}

var MasterNodeIndices = []repos.IndexField{
	{
		Field: []repos.IndexKey{
			{Key: "id", Value: repos.IndexAsc},
		},
		Unique: true,
	},
	{
		Field: []repos.IndexKey{
			{Key: "metadata.name", Value: repos.IndexAsc},
			{Key: "spec.clusterName", Value: repos.IndexAsc},
		},
		Unique: true,
	},
}

type WorkerNode struct {
	repos.BaseEntity   `json:",inline"`
	infraV1.WorkerNode `json:",inline"`
}

func (wn *WorkerNode) UnmarshalGQL(v interface{}) error {
	switch t := v.(type) {
	case map[string]any:
		b, err := json.Marshal(t)
		if err != nil {
			return err
		}

		if err := json.Unmarshal(b, wn); err != nil {
			return err
		}

	case string:
		if err := json.Unmarshal([]byte(t), wn); err != nil {
			return err
		}
	}

	return nil
}

func (wn WorkerNode) MarshalGQL(w io.Writer) {
	b, err := json.Marshal(wn)
	if err != nil {
		w.Write([]byte("{}"))
	}
	w.Write(b)
}

var WorkerNodeIndices = []repos.IndexField{
	{
		Field: []repos.IndexKey{
			{Key: "id", Value: repos.IndexAsc},
		},
		Unique: true,
	},
	{
		Field: []repos.IndexKey{
			{Key: "metadata.name", Value: repos.IndexAsc},
			{Key: "spec.clusterName", Value: repos.IndexAsc},
		},
		Unique: true,
	},
}

type NodePool struct {
	repos.BaseEntity `json:",inline"`
	infraV1.NodePool `json:",inline"`
}

func (np *NodePool) UnmarshalGQL(v interface{}) error {
	switch t := v.(type) {
	case map[string]any:
		b, err := json.Marshal(t)
		if err != nil {
			return err
		}

		if err := json.Unmarshal(b, np); err != nil {
			return err
		}

	case string:
		if err := json.Unmarshal([]byte(t), np); err != nil {
			return err
		}
	}

	return nil
}

func (np NodePool) MarshalGQL(w io.Writer) {
	b, err := json.Marshal(np)
	if err != nil {
		w.Write([]byte("{}"))
	}
	w.Write(b)
}

var NodePoolIndices = []repos.IndexField{
	{
		Field: []repos.IndexKey{
			{Key: "id", Value: repos.IndexAsc},
		},
		Unique: true,
	},
	{
		Field: []repos.IndexKey{
			{Key: "metadata.name", Value: repos.IndexAsc},
			{Key: "spec.edgeName", Value: repos.IndexAsc},
			{Key: "spec.clusterName", Value: repos.IndexAsc},
		},
		Unique: true,
	},
}
