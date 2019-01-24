[![Build Status](https://travis-ci.org/kubernetes-csi/driver-registrar.svg?branch=master)](https://travis-ci.org/kubernetes-csi/driver-registrar)

# Cluster Driver Registrar

The cluster-driver-registrar is a sidecar container that creates a cluster-level
[CSIDriver object](https://github.com/kubernetes/kubernetes/blob/master/staging/src/k8s.io/csi-api/pkg/crd/manifests/csidriver.yaml)
for the CSI driver.

This sidecar container is only needed if you need one of the following Kubernetes
features:

* Skip attach: For drivers that don't support ControllerPublishVolume, this
  eliminates the need to deploy the external-attacher sidecar
* Pod info on mount: This passes Kubernetes metadata such as Pod name and
  namespace to the NodePublish call

For more details, please see the
[documentation](https://kubernetes-csi.github.io/docs/Setup.html#csidriver-custom-resource-alpha).

## Compatibility

| Latest stable release                                                                                       | Branch                                                                                  | Compatible with CSI Version                                                                | Container Image                                 | Min K8s Version | Max K8s Version |
| ----------------------------------------------------------------------------------------------------------- | --------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------ | ----------------------------------------------- | --------------- | --------------- |
| [cluster-driver-registrar v1.0.1](https://github.com/kubernetes-csi/cluster-driver-registrar/releases/tag/v1.0.1) | [release-1.0](https://github.com/kubernetes-csi/cluster-driver-registrar/tree/release-1.0) | [CSI Spec v1.0.0](https://github.com/container-storage-interface/spec/releases/tag/v1.0.0) | quay.io/k8scsi/csi-cluster-driver-registrar:v1.0.1 | 1.13            | -               |

## Usage

### Features

#### Skip attach

  * The cluster-driver-registrar populates the attachRequired field of
    CSIDriverInfo by checking for the PUBLISH_UNPUBLISH_VOLUME controller
    capability. No additional configuration is required.
  * Drivers that don't support this capability do not need to deploy the
    external-attacher sidecar.

#### Pod info on mount

  * The cluster-driver-registrar can indicate that the associated driver wants
    kubelet to send additional information about the driver Pod as
    VolumeAttributes on a volume mount.
  <!-- TODO: Upload pod info details to csi-docs, reference here -->
  * This is controlled by the `--pod-info-mount-version` argument. The
    supported values are:
    * `v1`

### Required arguments

Though not strictly required, the following parameter it typically customized:

* `--csi-address`: This is the path to the CSI driver UNIX domain socket inside
  the pod that the `cluster-driver-registrar` container will use to issue CSI
  operations (e.g. `/csi/csi.sock`).

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
            - "--pod-info-mount-version=v1"
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
