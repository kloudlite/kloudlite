package entities

import (
	fc "github.com/kloudlite/api/apps/message-office/internal/entities/field-constants"
	"github.com/kloudlite/api/pkg/repos"
)

type ClusterAllocationCluster struct {
	Name           string `json:"name"`
	Region         string `json:"region"`
	OwnedByAccount string `json:"owned_by_account"`
}

type ClusterAllocation struct {
	repos.BaseEntity `json:",inline"`
	To               string                   `json:"to"`
	Cluster          ClusterAllocationCluster `json:"cluster"`
}

var ClusterAllocationIndexes = []repos.IndexField{
	{
		Field: []repos.IndexKey{
			{Key: fc.Id, Value: repos.IndexAsc},
		},
		Unique: true,
	},
	{
		Field: []repos.IndexKey{
			{Key: fc.ClusterAllocationTo, Value: repos.IndexAsc},
			{Key: fc.ClusterAllocationClusterName, Value: repos.IndexAsc},
		},
		Unique: true,
	},
	{
		Field: []repos.IndexKey{
			{Key: fc.ClusterAllocationClusterRegion, Value: repos.IndexAsc},
		},
	},
}

/*
```javascript
records = []
for (let i = 1000; i < 10000; i++) {
  records.push({
    id: i,
    to: `A-${i}`,
    cluster_name: `cluster-${Math.ceil(Math.random() * 10)}`
  });
}
db.cluster_allocations.insertMany(records);

//   [
//   { id: "1", to: "A", cluster_name: "cluster-1" },
//   { id: "2", to: "B", cluster_name: "cluster-1" },
//   { id: "3", to: "C", cluster_name: "cluster-2" },
//   { id: "4", to: "D", cluster_name: "cluster-2" },
//   { id: "5", to: "E", cluster_name: "cluster-3" },
//   { id: "6", to: "F", cluster_name: "cluster-3" },
//   { id: "7", to: "G", cluster_name: "cluster-4" },
//   ]
// );

db.cluster_allocations.aggregate([
    {
        $group: {
            _id: "$cluster_name", // Replace with the field you want to group by
            count: { $sum: 1 } // Count the number of occurrences
        }
    },
    {
        $sort: { count: 1 } // Sort by count in descending order
    },
    {
        $limit: 1 // Limit the results to the top 10
    }
])

```
*/
