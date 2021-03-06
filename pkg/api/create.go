package api

import (
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/crossplane/terraform-provider-runtime/pkg/client"
	"github.com/crossplane/terraform-provider-runtime/pkg/registry"
	"github.com/hashicorp/terraform/providers"
	"github.com/zclconf/go-cty/cty"
)

// Create returns an up-to-date version of the resource
// TODO: If `id` is unset for a new resource, how do we figure out
// what value needs to be used as the id?
func Create(p *client.Provider, r *registry.Registry, res resource.Managed) (resource.Managed, error) {
	// read resource yaml
	// lookup terraform schema from type name
	// TODO: registry of types from indexing provider getschema
	// - we should get schemas from provider, but use codgen to map schema back to go types
	// serialization function should traverse provider schema and then use codegen'd method to look up values for each field
	//
	// traverse provider schema and
	gvk := res.GetObjectKind().GroupVersionKind()
	s, err := SchemaForGVK(gvk, p, r)
	if err != nil {
		return nil, err
	}
	ctyEncoder, err := r.GetCtyEncoder(gvk)
	if err != nil {
		return nil, err
	}
	encoded, err := ctyEncoder(res, s)
	if err != nil {
		return nil, err
	}
	tfName, err := r.GetTerraformNameForGVK(gvk)
	if err != nil {
		return nil, err
	}

	// TODO: research how/if the major providers are using Config
	// same goes for the private state blobs that are shuffled around
	req := providers.ApplyResourceChangeRequest{
		TypeName:   tfName,
		PriorState: cty.NullVal(s.Block.ImpliedType()),
		// TODO: For the purposes of Create, I am assuming that it's fine for
		// Config and PlannedState to be the same
		Config:       encoded,
		PlannedState: encoded,
	}
	resp := p.GRPCProvider.ApplyResourceChange(req)
	if resp.Diagnostics.HasErrors() {
		return res, resp.Diagnostics.NonFatalErr()
	}
	ctyDecoder, err := r.GetCtyDecoder(gvk)
	if err != nil {
		return nil, err
	}
	return ctyDecoder(res, resp.NewState, s)
}
