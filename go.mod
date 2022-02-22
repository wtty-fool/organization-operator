module github.com/giantswarm/organization-operator

go 1.17

require (
	github.com/giantswarm/api-schema v0.7.1 // indirect
	github.com/giantswarm/companyd-client-go v0.6.1
	github.com/giantswarm/credentiald/v2 v2.17.0
	github.com/giantswarm/exporterkit v1.0.0
	github.com/giantswarm/k8sclient/v6 v6.1.0
	github.com/giantswarm/k8smetadata v0.9.2
	github.com/giantswarm/microendpoint v1.0.0
	github.com/giantswarm/microerror v0.4.0
	github.com/giantswarm/microkit v1.0.0
	github.com/giantswarm/micrologger v0.6.0
	github.com/giantswarm/operatorkit/v6 v6.1.0
	github.com/golang/mock v1.6.0
	github.com/prometheus/client_golang v1.12.1
	github.com/spf13/viper v1.10.1
	gopkg.in/resty.v1 v1.12.0
	k8s.io/api v0.20.15
	k8s.io/apiextensions-apiserver v0.20.15
	k8s.io/apimachinery v0.20.15
	k8s.io/client-go v0.20.15
	sigs.k8s.io/controller-runtime v0.8.3
)

replace (
	github.com/coreos/etcd v3.3.10+incompatible => github.com/coreos/etcd v3.3.25+incompatible
	github.com/coreos/etcd v3.3.13+incompatible => github.com/coreos/etcd v3.3.25+incompatible
	github.com/dgrijalva/jwt-go => github.com/dgrijalva/jwt-go/v4 v4.0.0-preview1
)
