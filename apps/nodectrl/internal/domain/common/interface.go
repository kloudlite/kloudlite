package common

import "context"

/*
Tasks needs to be performed by this job
	- create node
	- attach node
	- delete node
	- craete cluster
	- delete cluster
*/

type ProviderClient interface {
	/*
		ssh generation
		create node
		AttachNode
	*/
	NewNode(ctx context.Context) error
	DeleteNode(ctx context.Context) error
	SaveToDbGuranteed(ctx context.Context)

	CreateCluster(ctx context.Context) error

	AddWorker(ctx context.Context) error
	AddMaster(ctx context.Context) error

	/*
		It will perform generation of ssh
		create node
		install master
		fetch agent token and Master URL and save it to db
	*/
}
