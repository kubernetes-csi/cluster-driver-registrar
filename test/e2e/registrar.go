/*
Copyright 2019 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package e2e

import (
	"context"
	"flag"

	appsv1 "k8s.io/api/apps/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/kubernetes/test/e2e/framework"
	"k8s.io/kubernetes/test/e2e/framework/podlogs"

	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega"
)

var (
	image = flag.String("cluster-driver-registrar-image", "", "Docker image to use during testing instead of the one from the deploy .yaml file")

	csiDriverGVR = schema.GroupVersionResource{Group: "storage.k8s.io", Version: "v1beta1", Resource: "csidrivers"}
)

var _ = ginkgo.Describe("CSIDriverInfo", func() {
	f := framework.NewDefaultFramework("csidriverinfo")

	ginkgo.It("should be created and removed", func() {
		cs := f.ClientSet
		dc := f.DynamicClient
		var err error

		// Ensure that we run on a cluster which has csidrivers.storage.k8s.io.
		_, err = dc.Resource(csiDriverGVR).Namespace("").List(metav1.ListOptions{})
		if err != nil {
			if apierrors.IsNotFound(err) {
				framework.Skipf("cluster does not have csidrivers.storage.k8s.io")
			}
			framework.ExpectNoError(err, "checking for csidrivers.storage.k8s.io")
		}

		ginkgo.By("deploying cluster-driver-registrar and csi mock driver")
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		to := podlogs.LogOutput{
			StatusWriter: ginkgo.GinkgoWriter,
			LogWriter:    ginkgo.GinkgoWriter,
		}
		podlogs.CopyAllLogs(ctx, cs, f.Namespace.Name, to)
		podlogs.WatchPods(ctx, cs, f.Namespace.Name, ginkgo.GinkgoWriter)

		uniqueName := "csi-mock-" + f.UniqueName
		cleanup, err := f.CreateFromManifests(func(item interface{}) error {
			switch item := item.(type) {
			case *appsv1.StatefulSet:
				containers := &item.Spec.Template.Spec.Containers
				for i := range *containers {
					container := &(*containers)[i]
					switch container.Name {
					case "mock-driver":
						// Rename driver.
						container.Args = append(container.Args, "--name="+uniqueName)
					case "cluster-driver-registrar":
						// Replace image.
						if *image != "" {
							container.Image = *image
						}
					}
				}
			}
			return nil
		},
			"deploy/kubernetes/rbac.yaml",
			"deploy/kubernetes/statefulset.yaml",
		)
		framework.ExpectNoError(err, "deploying mock driver")
		defer cleanup()

		haveInfo := func() bool {
			_, err := dc.Resource(csiDriverGVR).Namespace("").Get(uniqueName, metav1.GetOptions{})
			if err == nil {
				return true
			}
			if apierrors.IsNotFound(err) {
				return false
			}
			framework.ExpectNoError(err, "getting csidrivers.storage.k8s.io")
			return false
		}

		// Ensure that CSIDriverInfo appears.
		ginkgo.By("waiting for creation of csidrivers.storage.k8s.io/" + uniqueName)
		gomega.Eventually(haveInfo, "120s").Should(gomega.Equal(true), "have csidriverinfos.storage.k8s.io")

		// Uninstall by scaling down the stateful set to 0.
		// This is how a pod managed by a stateful set can be
		// terminated reliably
		// (https://kubernetes.io/docs/concepts/workloads/controllers/statefulset/#limitations).
		//
		// In practice, removing a stateful set does remove the pod. But if we do that with
		// cleanup(), then we remove both RBAC rules and the stateful set at the same time,
		// which causes permission issues when cluster-driver-registrar removes the CSIDriver
		// when the necessary ClusterRole is already gone.
		ginkgo.By("scaling down to zero replicas")
		gomega.Eventually(func() int32 {
			scale, err := cs.AppsV1().StatefulSets(f.Namespace.Name).GetScale("csi-cluster-driver-registrar", metav1.GetOptions{})
			framework.ExpectNoError(err, "get cluster-driver-registrar replicas")
			if scale.Status.Replicas == 0 {
				return 0
			}

			scale.Spec.Replicas = 0
			scale2, err := cs.AppsV1().StatefulSets(f.Namespace.Name).UpdateScale("csi-cluster-driver-registrar", scale)
			if err != nil {
				// Sometimes this fails because the
				// stateful set has been modified in
				// the meantime.  Ignore errors and
				// just retry.
				return scale.Status.Replicas
			}
			return scale2.Status.Replicas
		}, "120s").Should(gomega.Equal(int32(0)), "zero driver replicas")

		// Ensure that CSIDriverInfo disappeared.
		gomega.Expect(haveInfo()).Should(gomega.Equal(false), "not have csidriverinfos.storage.k8s.io")
	})
})
