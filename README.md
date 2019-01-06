[![Build Status](https://travis-ci.org/kubernetes-csi/driver-registrar.svg?branch=master)](https://travis-ci.org/kubernetes-csi/driver-registrar)

# Cluster Driver Registrar

A cluster-level sidecar container that

1. Creates a [CSIDriver
   object](https://github.com/kubernetes/kubernetes/blob/master/staging/src/k8s.io/csi-api/pkg/crd/manifests/csidriver.yaml)
   for the driver.

This sidecar container is only needed if you need one of the following Kubernetes
features:

* Skip attach: For drivers that don't support ControllerPublishVolume, this
  eliminates the need to deploy the external-attacher sidecar
* Pod info on mount: This passes Kubernetes metadata such as Pod name and
  namespace to the NodePublish call

For more details, please see the
[documentation](https://kubernetes-csi.github.io/docs/Setup.html#csidriver-custom-resource-alpha).

## Community, discussion, contribution, and support

Learn how to engage with the Kubernetes community on the [community page](http://kubernetes.io/community/).

You can reach the maintainers of this project at:

* Slack channels
  * [#wg-csi](https://kubernetes.slack.com/messages/wg-csi)
  * [#sig-storage](https://kubernetes.slack.com/messages/sig-storage)
  * [Mailing list](https://groups.google.com/forum/#!forum/kubernetes-sig-storage)

### Code of conduct

Participation in the Kubernetes community is governed by the [Kubernetes Code of Conduct](code-of-conduct.md).
