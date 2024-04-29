package entities

//type ProjectManagedService struct {
//	repos.BaseEntity             `json:",inline" graphql:"noinput"`
//	crdsv1.ProjectManagedService `json:",inline"`
//
//	AccountName string `json:"accountName" graphql:"noinput"`
//	ProjectName string `json:"projectName" graphql:"noinput"`
//
//	SyncedOutputSecretRef *corev1.Secret `json:"syncedOutputSecretRef" graphql:"ignore" struct-json-path:",ignore-nesting"`
//
//	common.ResourceMetadata `json:",inline"`
//	SyncStatus              t.SyncStatus `json:"syncStatus" graphql:"noinput"`
//}
//
//func (s *ProjectManagedService) GetDisplayName() string {
//	return s.ResourceMetadata.DisplayName
//}
//
//func (s *ProjectManagedService) GetStatus() operator.Status {
//	return s.ProjectManagedService.Status
//}
//
//// GetResourceType implements domain.resource.
//func (*ProjectManagedService) GetResourceType() ResourceType {
//	return ResourceTypeProjectManagedService
//}
//
//var ProjectManagedServiceIndices = []repos.IndexField{
//	{
//		Field: []repos.IndexKey{
//			{Key: fields.Id, Value: repos.IndexAsc},
//		},
//		Unique: true,
//	},
//	{
//		Field: []repos.IndexKey{
//			{Key: fields.MetadataName, Value: repos.IndexAsc},
//			{Key: fields.AccountName, Value: repos.IndexAsc},
//			{Key: fields.ProjectName, Value: repos.IndexAsc},
//		},
//		Unique: true,
//	},
//}
