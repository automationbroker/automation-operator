FROM alpine:3.6

RUN adduser -D automation-operator
USER automation-operator

ADD tmp/_output/bin/automation-operator /usr/local/bin/automation-operator
