<!--

    TODO:

    - Add the project to the CircleCI:
      https://circleci.com/setup-project/gh/giantswarm/REPOSITORY_NAME

    - Import RELEASE_TOKEN variable from template repository for the builds:
      https://circleci.com/gh/giantswarm/REPOSITORY_NAME/edit#env-vars

    - Change the badge (with style=shield):
      https://circleci.com/gh/giantswarm/REPOSITORY_NAME/edit#badges
      If this is a private repository token with scope `status` will be needed.

    - Run `devctl replace -i "REPOSITORY_NAME" "$(basename $(git rev-parse --show-toplevel))" *.md`
      and commit your changes.

-->
[![CircleCI](https://circleci.com/gh/giantswarm/template-operator.svg?&style=shield)](https://circleci.com/gh/giantswarm/template-operator) [![Docker Repository on Quay](https://quay.io/repository/giantswarm/template-operator/status "Docker Repository on Quay")](https://quay.io/repository/giantswarm/template-operator)

# organization-operator

Organization operator manages namespaces based on Organization CR.
