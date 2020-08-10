package registry

import xpresource "github.com/crossplane/crossplane-runtime/pkg/resource"

type ResourceMerger func(xpresource.Managed, xpresource.Managed) MergeDescription

type MergeDescription struct {
	LateInitializedSpec bool
	StatusUpdated       bool
	AnnotationsUpdated  bool
	NeedsProviderUpdate bool
}
