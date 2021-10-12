# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.10.0] - 2021-10-12

### Added

- Ensure `Organization` CR in Azure MCs have the `subscriptionid` annotation set.

### Changed

- Use `Patch` to save `Namespace` in `Status` to avoid write conflicts.

## [0.9.0] - 2021-07-05

### Added

- Introduce local interfaces to make easier to test the organization handler.
- Save created namespace in Organization status field.

### Changed

- Updated `apiextensions` dependency to v3.
- Updated `operatorkit` dependency to v5.
- Updated `k8sclient` dependency to v5.
- Always try to create organization in `companyd` even when organization namespace already exists.
- Pin `jwt-go` dependency to avoid version with vulnerabilities.

## [0.8.0] - 2021-05-24

### Changed

- Update `architect-orb` to v3.0.0.

## [0.7.0] - 2021-05-21

### Changed

- Set config version in `Chart.yaml`.

## [0.6.0] - 2021-05-17

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

[Unreleased]: https://github.com/giantswarm/organization-operator/compare/v0.10.0...HEAD
[0.10.0]: https://github.com/giantswarm/organization-operator/compare/v0.9.0...v0.10.0
[0.9.0]: https://github.com/giantswarm/organization-operator/compare/v0.8.0...v0.9.0
[0.8.0]: https://github.com/giantswarm/organization-operator/compare/v0.7.0...v0.8.0
[0.7.0]: https://github.com/giantswarm/organization-operator/compare/v0.6.0...v0.7.0
[0.6.0]: https://github.com/giantswarm/organization-operator/compare/v0.5.0...v0.6.0
[0.5.0]: https://github.com/giantswarm/organization-operator/compare/v0.4.0...v0.5.0
[0.4.0]: https://github.com/giantswarm/organization-operator/compare/v0.3.0...v0.4.0
[0.3.0]: https://github.com/giantswarm/organization-operator/compare/v0.2.0...v0.3.0
[0.2.0]: https://github.com/giantswarm/organization-operator/compare/v0.1.0...v0.2.0
[0.1.0]: https://github.com/giantswarm/organization-operator/releases/tag/v0.1.0
