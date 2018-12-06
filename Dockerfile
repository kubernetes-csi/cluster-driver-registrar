FROM alpine
LABEL maintainers="Kubernetes Authors"
LABEL description="CSI Cluster Driver registrar"

COPY ./bin/cluster-driver-registrar cluster-driver-registrar
ENTRYPOINT ["/cluster-driver-registrar"]
