FROM alpine:3.19.0

RUN apk add --no-cache ca-certificates

ADD ./organization-operator /organization-operator

ENTRYPOINT ["/organization-operator"]
