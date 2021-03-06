package apigroups

import (
	"net/http"

	"github.com/rancher/steve/pkg/schemaserver/store/empty"
	"github.com/rancher/steve/pkg/schemaserver/types"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/discovery"
)

func Register(schemas *types.APISchemas, discovery discovery.DiscoveryInterface) {
	schemas.MustImportAndCustomize(v1.APIGroup{}, func(schema *types.APISchema) {
		schema.CollectionMethods = []string{http.MethodGet}
		schema.ResourceMethods = []string{http.MethodGet}
		schema.Store = NewStore(discovery)
		schema.Formatter = func(request *types.APIRequest, resource *types.RawResource) {
			resource.ID = resource.APIObject.Data().String("name")
		}
	})
}

type Store struct {
	empty.Store

	discovery discovery.DiscoveryInterface
}

func NewStore(discovery discovery.DiscoveryInterface) types.Store {
	return &Store{
		Store:     empty.Store{},
		discovery: discovery,
	}
}

func (e *Store) ByID(apiOp *types.APIRequest, schema *types.APISchema, id string) (types.APIObject, error) {
	return types.DefaultByID(e, apiOp, schema, id)
}

func toAPIObject(schema *types.APISchema, group v1.APIGroup) types.APIObject {
	if group.Name == "" {
		group.Name = "core"
	}
	return types.APIObject{
		Type:   schema.ID,
		ID:     group.Name,
		Object: group,
	}

}

func (e *Store) List(apiOp *types.APIRequest, schema *types.APISchema) (types.APIObjectList, error) {
	groupList, err := e.discovery.ServerGroups()
	if err != nil {
		return types.APIObjectList{}, err
	}

	var result types.APIObjectList
	for _, item := range groupList.Groups {
		result.Objects = append(result.Objects, toAPIObject(schema, item))
	}

	return result, nil
}
