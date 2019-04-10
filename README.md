[![Build Status](https://travis-ci.org/kubernetes-csi/driver-registrar.svg?branch=master)](https://travis-ci.org/kubernetes-csi/driver-registrar)

# !NOTE!
:warning: :warning: Under Construction :warning: :warning:

Due to Issue
[#44](https://github.com/kubernetes-csi/cluster-driver-registrar/issues/44) this
side car container is being redesigned and has not been released for Kubernetes
1.14. Please read
[docs/cluster-driver-registrar](https://kubernetes-csi.github.io/docs/cluster-driver-registrar.html)
for more information.

# Cluster Driver Registrar

The cluster-driver-registrar is a sidecar container that creates a cluster-level
[CSIDriver object](https://github.com/kubernetes/kubernetes/blob/master/staging/src/k8s.io/csi-api/pkg/crd/manifests/csidriver.yaml)
for the CSI driver.

This sidecar container is only needed if you need one of the following Kubernetes
features:

<!-- TODO: Reference skip attach docs here -->
* Skip attach: For drivers that don't support ControllerPublishVolume, this
  eliminates the need to deploy the external-attacher sidecar
<!-- TODO: Reference pod info docs here -->
* Pod info on mount: This passes Kubernetes metadata such as Pod name and
  namespace to the NodePublish call

If you are not using one of these features, this sidecar container (and the
creation of the CSIDriver Object) is not required. However, it is still
recommended, because the CSIDriver Object makes it easier for users to easily
discover the CSI drivers installed on their clusters.

## Compatibility

This information reflects the head of this branch.

| Compatible with CSI Version                                                                | Container Image                                 | Min K8s Version |
| ------------------------------------------------------------------------------------------ | ----------------------------------------------- | --------------- |
| [CSI Spec v1.0.0](https://github.com/container-storage-interface/spec/releases/tag/v1.0.0) | quay.io/k8scsi/csi-cluster-driver-registrar     | 1.14            |

## Usage

### Common arguments

Though not strictly required, the following parameters are typically
customized:

* `--csi-address`: This is the path to the CSI driver UNIX domain socket inside
  the pod that the `cluster-driver-registrar` container will use to issue CSI
  operations (e.g. `/csi/csi.sock`).
* `--pod-info-mount`: This allows Pod information to be passed to
  the NodePublish call. This should only be set if the CSI driver requires Pod
  information for mounting.

### Required permissions

The cluster-driver-registrar needs to be able to create and delete CSIDriver
objects. A sample RBAC configuration can be found at
[deploy/kubernetes/rbac.yaml](deploy/kubernetes/rbac.yaml).

### Example

Here is an example sidecar spec in the driver's controller StatefulSet.
`<drivername.example.com>` should be replaced by the actual driver's name.

```bash
      containers:
        - name: cluster-driver-registrar
          image: quay.io/k8scsi/csi-cluster-driver-registrar:v1.0.2
          args:
            - "--csi-address=/csi/csi.sock"
            - "--pod-info-mount=true"
          volumeMounts:
            - name: plugin-dir
              mountPath: /csi
      volumes:
        - name: plugin-dir
          emptyDir: {}
```

## Community, discussion, contribution, and support

Learn how to engage with the Kubernetes community on the [community page](http://kubernetes.io/community/).

You can reach the maintainers of this project at:

* Slack channels
  * [#wg-csi](https://kubernetes.slack.com/messages/wg-csi)
  * [#sig-storage](https://kubernetes.slack.com/messages/sig-storage)
* [Mailing list](https://groups.google.com/forum/#!forum/kubernetes-sig-storage)

### Code of conduct

Participation in the Kubernetes community is governed by the [Kubernetes Code of Conduct](code-of-conduct.md).
