package organization

import (
	"context"

	companyclient "github.com/giantswarm/companyd-client-go"
	"github.com/giantswarm/credentiald/v2/service/lister"
)

type CompanydClient interface {
	CreateCompany(companyID string, fields companyclient.CompanyFields) error
	DeleteCompany(companyID string) error
}

type CredentialdClient interface {
	List(ctx context.Context, request lister.Request) ([]lister.Response, error)
}
