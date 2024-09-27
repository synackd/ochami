FROM cgr.dev/chainguard/wolfi-base

RUN apk add --no-cache tini

COPY ochami /ochami

# nobody 65534:65534
USER 65534:65534

CMD [ "/ochami" ]
ENTRYPOINT [ "/sbin/tini", "--" ]
