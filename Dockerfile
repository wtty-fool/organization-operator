FROM quay.io/giantswarm/alpine:3.15.4

RUN apk add --no-cache ca-certificates

ADD ./organization-operator /organization-operator

ENTRYPOINT ["/organization-operator"]
