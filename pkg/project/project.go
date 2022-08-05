package project

var (
	description = "Organization operator manages namespaces based on Organization CR."
	gitSHA      = "n/a"
	name        = "organization-operator"
	source      = "https://github.com/giantswarm/organization-operator"
	version     = "1.0.3"
)

func Description() string {
	return description
}

func GitSHA() string {
	return gitSHA
}

func Name() string {
	return name
}

func Source() string {
	return source
}

func Version() string {
	return version
}
