####################################################################################################
# base
####################################################################################################
FROM alpine:3.12.3 as base
RUN apk update && apk upgrade && \
    apk add ca-certificates && \
    apk --no-cache add tzdata

COPY dist/kafka-source /bin/kafka-source
RUN chmod +x /bin/kafka-source

####################################################################################################
# kafka-source
####################################################################################################
FROM scratch as kafka-source
ARG ARCH
COPY --from=base /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=base /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=base /bin/kafka-source /bin/kafka-source
ENTRYPOINT [ "/bin/kafka-source" ]