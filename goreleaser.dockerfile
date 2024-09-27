FROM cgr.dev/chainguard/wolfi-base

RUN apk add --no-cache tini

COPY ochami /bin/ochami

# nobody 65534:65534
USER 65534:65534

CMD [ "/bin/ochami" ]
ENTRYPOINT [ "/sbin/tini", "--" ]
