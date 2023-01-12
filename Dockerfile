FROM alpine:3.17.1

RUN apk add --no-cache ca-certificates

ADD ./organization-operator /organization-operator

ENTRYPOINT ["/organization-operator"]
