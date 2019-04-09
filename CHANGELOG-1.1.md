# Changelog since v1.0.1

## Actions needed

* Cluster admins must update RBAC rules. Previously, the cluster driver registrar created CSIDriver in `csi.storage.k8s.io/v1alpha1` API group, now it creates it in `csi.k8s.io/v1beta1`. See [deploy/kubernetes/rbac.yaml](deploy/kubernetes/rbac.yaml) for example.

* Command line flag `-driver-requires-attachment` has been removed. It is autodetected from the CSI driver.

* Command line flag `-pod-info-mount-version <string>` has been reworked into `-pod-info-mount <bool>`.

Please update your cluster-driver-registrar deployment files with these changes.

## Deprecations

* Command line flag `-connection-timeout` is deprecated and has no effect.


## Notable Features

* The cluster driver registrar now tries to connect to CSI driver indefinitely. ([#32](https://github.com/kubernetes-csi/cluster-driver-registrar/pull/32))

* `CSIDriver` object is now beta and has been moved from `csi.storage.k8s.io/v1alpha1` to `csi.k8s.io/v1beta1`. New RBAC rules are required!

## Other notable changes

* Use distroless as base image ([#39](https://github.com/kubernetes-csi/cluster-driver-registrar/pull/39))
* Use GetDriverName from csi-lib-utils ([#35](https://github.com/kubernetes-csi/cluster-driver-registrar/pull/35))
* Minor RBAC update ([#34](https://github.com/kubernetes-csi/cluster-driver-registrar/pull/34))
* Update compatibility matrix to only reflect branch head ([#27](https://github.com/kubernetes-csi/cluster-driver-registrar/pull/27))
* Minor README updates ([#25](https://github.com/kubernetes-csi/cluster-driver-registrar/pull/25))
* Migrate to k8s.io/klog from glog. ([#24](https://github.com/kubernetes-csi/cluster-driver-registrar/pull/24))
* release tools: fix "canary" + enhance testing ([#23](https://github.com/kubernetes-csi/cluster-driver-registrar/pull/23))
* Update usage and deployment documentation ([#20](https://github.com/kubernetes-csi/cluster-driver-registrar/pull/20))
* Add more details about cluster-driver-registrar ([#10](https://github.com/kubernetes-csi/cluster-driver-registrar/pull/10))
* Auto-detect AttachRequired ([#4](https://github.com/kubernetes-csi/cluster-driver-registrar/pull/4))
* Use protosanitizer when logging ([#5](https://github.com/kubernetes-csi/cluster-driver-registrar/pull/5))
