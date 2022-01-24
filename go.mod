module github.com/giantswarm/organization-operator

go 1.14

require (
	github.com/evanphx/json-patch v4.11.0+incompatible // indirect
	github.com/getsentry/sentry-go v0.11.0 // indirect
	github.com/giantswarm/api-schema v0.7.1 // indirect
	github.com/giantswarm/apiextensions/v3 v3.39.0
	github.com/giantswarm/companyd-client-go v0.6.1
	github.com/giantswarm/credentiald/v2 v2.17.0
	github.com/giantswarm/exporterkit v1.0.0
	github.com/giantswarm/k8sclient/v5 v5.12.0
	github.com/giantswarm/k8smetadata v0.8.0
	github.com/giantswarm/microendpoint v1.0.0
	github.com/giantswarm/microerror v0.4.0
	github.com/giantswarm/microkit v1.0.0
	github.com/giantswarm/micrologger v0.6.0
	github.com/giantswarm/operatorkit/v5 v5.0.0
	github.com/golang/mock v1.6.0
	github.com/googleapis/gnostic v0.5.5 // indirect
	github.com/imdario/mergo v0.3.12 // indirect
	github.com/prometheus/client_golang v1.12.0
	github.com/spf13/viper v1.10.1
	golang.org/x/term v0.0.0-20210220032956-6a3ed077a48d // indirect
	gopkg.in/resty.v1 v1.12.0
	k8s.io/api v0.20.12
	k8s.io/apiextensions-apiserver v0.20.12
	k8s.io/apimachinery v0.20.12
	k8s.io/client-go v0.20.12
	k8s.io/klog/v2 v2.9.0 // indirect
	k8s.io/kube-openapi v0.0.0-20210421082810-95288971da7e // indirect
	k8s.io/utils v0.0.0-20210819203725-bdf08cb9a70a // indirect
	sigs.k8s.io/controller-runtime v0.6.5
	sigs.k8s.io/yaml v1.3.0
)

replace (
	github.com/coreos/etcd v3.3.10+incompatible => github.com/coreos/etcd v3.3.25+incompatible
	github.com/coreos/etcd v3.3.13+incompatible => github.com/coreos/etcd v3.3.25+incompatible
	github.com/dgrijalva/jwt-go => github.com/dgrijalva/jwt-go/v4 v4.0.0-preview1
	github.com/gogo/protobuf v1.3.1 => github.com/gogo/protobuf v1.3.2
	sigs.k8s.io/cluster-api => github.com/giantswarm/cluster-api v0.3.13-gs
)
