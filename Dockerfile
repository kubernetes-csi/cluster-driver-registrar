FROM gcr.io/distroless/static:latest
LABEL maintainers="Kubernetes Authors"
LABEL description="CSI Cluster Driver registrar"

COPY ./bin/csi-cluster-driver-registrar csi-cluster-driver-registrar
ENTRYPOINT ["/csi-cluster-driver-registrar"]
