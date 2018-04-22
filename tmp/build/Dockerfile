FROM alpine:3.6

ADD tmp/_output/bin/wildfly-operator /usr/local/bin/wildfly-operator

RUN adduser -D wildfly-operator
USER wildfly-operator
