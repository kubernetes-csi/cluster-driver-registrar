/*
Copyright 2017 The Kubernetes Authors.

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

package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"google.golang.org/grpc"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/kubernetes-csi/csi-lib-utils/connection"
	csirpc "github.com/kubernetes-csi/csi-lib-utils/rpc"

	k8scsi "k8s.io/api/storage/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog"
)

const (
	// Default timeout of short CSI calls like GetPluginInfo
	csiTimeout = time.Second

	// Verify (and update, if needed) the node ID at this freqeuency.
	sleepDuration = 2 * time.Minute
)

// Command line flags
var (
	kubeconfig        = flag.String("kubeconfig", "", "Absolute path to the kubeconfig file. Required only when running out of cluster.")
	k8sPodInfoOnMount = flag.Bool("pod-info-mount", false,
		"This indicates that the associated CSI volume driver"+
			"requires additional pod information (like podName, podUID, etc.) during mount."+
			"When set to true, Kubelet will send the followings pod information "+
			"during NodePublishVolume() calls to the driver as VolumeAttributes:"+
			"- csi.storage.k8s.io/pod.name: pod.Name\n"+
			"- csi.storage.k8s.io/pod.namespace: pod.Namespace\n"+
			"- csi.storage.k8s.io/pod.uid: string(pod.UID)",
	)
	connectionTimeout = flag.Duration("connection-timeout", 0, "The --connection-timeout flag is deprecated")
	csiAddress        = flag.String("csi-address", "/run/csi/socket", "Address of the CSI driver socket.")
	showVersion       = flag.Bool("version", false, "Show version.")
	version           = "unknown"
	// List of supported versions
	supportedVersions = []string{"1.0.0"}
)

func main() {
	klog.InitFlags(nil)
	flag.Set("logtostderr", "true")
	flag.Parse()

	if *showVersion {
		fmt.Println(os.Args[0], version)
		return
	}
	klog.Infof("Version: %s", version)

	if *connectionTimeout != 0 {
		klog.Warning("--connection-timeout is deprecated and will have no effect")
	}

	// Connect to CSI.
	klog.V(1).Infof("Attempting to open a gRPC connection with: %q", *csiAddress)
	csiConn, err := connection.Connect(*csiAddress)
	if err != nil {
		klog.Errorf("error connecting to CSI driver: %v", err)
		os.Exit(1)
	}

	// Get connection context
	ctx, cancel := context.WithTimeout(context.Background(), csiTimeout)
	defer cancel()

	// Get CSI driver name.
	klog.V(4).Infof("Calling CSI driver to discover driver name.")
	csiDriverName, err := csirpc.GetDriverName(ctx, csiConn)
	if err != nil {
		klog.Error(err.Error())
		os.Exit(1)
	}
	klog.V(2).Infof("CSI driver name: %q", csiDriverName)

	// Check if volume attach is required
	klog.V(4).Infof("Checking if CSI driver implements ControllerPublishVolume().")
	k8sAttachmentRequired, err := isAttachRequired(ctx, csiConn)
	if err != nil {
		klog.Errorf("error checking if attach is required: %v", err)
		os.Exit(1)
	}

	// Create CSIDriver object
	csiDriver := &k8scsi.CSIDriver{
		ObjectMeta: metav1.ObjectMeta{
			Name: csiDriverName,
		},
		Spec: k8scsi.CSIDriverSpec{
			AttachRequired: &k8sAttachmentRequired,
			PodInfoOnMount: k8sPodInfoOnMount,
		},
	}

	klog.V(2).Infof("CSIDriver object: %+v", *csiDriver)

	// Create the client config. Use kubeconfig if given, otherwise assume
	// in-cluster.
	klog.V(1).Infof("Loading kubeconfig.")
	config, err := buildConfig(*kubeconfig)
	if err != nil {
		klog.Error(err.Error())
		os.Exit(1)
	}

	// Run forever
	kubernetesRegister(config, csiDriver)
}

func buildConfig(kubeconfig string) (*rest.Config, error) {
	if kubeconfig != "" {
		return clientcmd.BuildConfigFromFlags("", kubeconfig)
	}

	// Return config object which uses the service account kubernetes gives to
	// pods. It's intended for clients that are running inside a pod running on
	// kubernetes.
	return rest.InClusterConfig()
}

func isAttachRequired(ctx context.Context, conn *grpc.ClientConn) (bool, error) {
	capabilities, err := csirpc.GetControllerCapabilities(ctx, conn)
	if err != nil {
		return false, err
	}

	return capabilities[csi.ControllerServiceCapability_RPC_PUBLISH_UNPUBLISH_VOLUME], nil
}
