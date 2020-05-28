package project

var (
	bundleVersion = "0.0.1"
	description   = "The organization-operator does something."
	gitSHA        = "n/a"
	name          = "organization-operator"
	source        = "https://github.com/giantswarm/organization-operator"
	version       = "1.0.0-dev"
)

func BundleVersion() string {
	return bundleVersion
}

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
