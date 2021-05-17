package key

import (
	securityv1alpha1 "github.com/giantswarm/apiextensions/v2/pkg/apis/security/v1alpha1"
	"github.com/giantswarm/microerror"
)

const (
	// uiName is used to identify organizations
	// with names that don't adhere to the DNS standard.
	//
	// While the org CR name would be converted to the
	// DNS standard, this annotation is used to determine
	// the companyd counterpart of the org CR.
	uiName = "ui.giantswarm.io/original-organization-name"
)

func ToOrganization(v interface{}) (securityv1alpha1.Organization, error) {
	if v == nil {
		return securityv1alpha1.Organization{}, microerror.Maskf(wrongTypeError, "expected non-nil, got %#v", v)
	}

	p, ok := v.(*securityv1alpha1.Organization)
	if !ok {
		return securityv1alpha1.Organization{}, microerror.Maskf(wrongTypeError, "expected %T, got %T", p, v)
	}

	c := p.DeepCopy()

	return *c, nil
}

func LegacyOrganizationName(cr *securityv1alpha1.Organization) string {
	if cr.GetAnnotations() != nil {
		name := cr.GetAnnotations()[uiName]
		if len(name) > 0 {
			return name
		}
	}

	return cr.GetName()
}
