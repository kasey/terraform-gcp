package api

import (
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/crossplane/terraform-provider-runtime/pkg/client"
	"github.com/crossplane/terraform-provider-runtime/pkg/registry"
	"github.com/hashicorp/terraform/providers"
)

// Update syncs with an existing resource and modifies mutable values
func Update(p *client.Provider, r *registry.Registry, res resource.Managed) (resource.Managed, error) {
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

	prior, err := Read(p, r, res)
	if err != nil {
		return nil, err
	}
	priorEncoded, err := ctyEncoder(prior, s)
	if err != nil {
		return nil, err
	}

	// TODO: research how/if the major providers are using Config
	// same goes for the private state blobs that are shuffled around
	req := providers.ApplyResourceChangeRequest{
		TypeName:   tfName,
		PriorState: priorEncoded,
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
