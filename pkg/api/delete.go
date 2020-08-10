package api

import (
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/crossplane/terraform-provider-runtime/pkg/client"
	"github.com/crossplane/terraform-provider-runtime/pkg/registry"
	"github.com/hashicorp/terraform/providers"
	"github.com/zclconf/go-cty/cty"
)

// Delete deletes the given resource from the provider
// In terraform slang this is expressed as asking the provider
// to act on a Nil planned state.
func Delete(p *client.Provider, r *registry.Registry, res resource.Managed) error {
	gvk := res.GetObjectKind().GroupVersionKind()
	s, err := SchemaForGVK(gvk, p, r)
	if err != nil {
		return err
	}
	ctyEncoder, err := r.GetCtyEncoder(gvk)
	if err != nil {
		return err
	}
	encoded, err := ctyEncoder(res, s)
	if err != nil {
		return err
	}
	tfName, err := r.GetTerraformNameForGVK(gvk)
	if err != nil {
		return err
	}

	req := providers.ApplyResourceChangeRequest{
		TypeName:   tfName,
		PriorState: encoded,
		// TODO: For the purposes of Delete, I am assuming that it's fine for
		// Config and PlannedState to be the same
		Config:       cty.NullVal(s.Block.ImpliedType()),
		PlannedState: cty.NullVal(s.Block.ImpliedType()),
	}
	resp := p.GRPCProvider.ApplyResourceChange(req)
	if resp.Diagnostics.HasErrors() {
		return resp.Diagnostics.NonFatalErr()
	}
	return nil
}
