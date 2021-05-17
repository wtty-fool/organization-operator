# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- Also delete legacy `companyd` organizations when deleting `Organization` CRs.
- Prevent deleting legacy credentials when deleting `Organization` CRs.

## [0.5.0] - 2020-09-24

### Changed

- Updated Kubernetes dependencies to v1.18.9 and operatorkit to v2.0.1.

### Added

- Add monitoring labels.

## [0.4.0] - 2020-08-21

### Added

- Add NetworkPolicy.

## [0.3.0] - 2020-08-14

### Changed

- Updated backward incompatible Kubernetes dependencies to v1.18.5.

## [0.2.0] - 2020-08-13

### Added

- Add dependabot configuration.

### Changed

- Update operatorkit to v1.2.0 and k8sclient to v3.1.2.

## [0.1.0] - 2020-06-03

### Added

- First release.

[Unreleased]: https://github.com/giantswarm/organization-operator/compare/v0.5.0...HEAD
[0.5.0]: https://github.com/giantswarm/organization-operator/compare/v0.4.0...v0.5.0
[0.4.0]: https://github.com/giantswarm/organization-operator/compare/v0.3.0...v0.4.0
[0.3.0]: https://github.com/giantswarm/organization-operator/compare/v0.2.0...v0.3.0
[0.2.0]: https://github.com/giantswarm/organization-operator/compare/v0.1.0...v0.2.0
[0.1.0]: https://github.com/giantswarm/organization-operator/releases/tag/v0.1.0
