FROM alpine:3.14.3

RUN apk add --no-cache ca-certificates

ADD ./organization-operator /organization-operator

ENTRYPOINT ["/organization-operator"]
