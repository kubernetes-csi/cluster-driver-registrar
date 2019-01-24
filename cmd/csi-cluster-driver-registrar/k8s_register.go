/*
Copyright 2018 The Kubernetes Authors.

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
	"os"
	"os/signal"
	"time"

	"github.com/golang/glog"
	crdclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/util/retry"
	k8scsi "k8s.io/csi-api/pkg/apis/csi/v1alpha1"
	k8scsiclient "k8s.io/csi-api/pkg/client/clientset/versioned"
	k8scsicrd "k8s.io/csi-api/pkg/crd"
)

func kubernetesRegister(
	config *rest.Config,
	csiDriver *k8scsi.CSIDriver,
) {
	// Get client info to CSIDriver
	clientset, err := k8scsiclient.NewForConfig(config)
	if err != nil {
		glog.Error(err.Error())
		os.Exit(1)
	}

	// Register CRD
	glog.V(1).Info("Registering " + k8scsi.CsiDriverResourcePlural)
	crdclient, err := crdclient.NewForConfig(config)
	if err != nil {
		glog.Error(err.Error())
		os.Exit(1)
	}
	crdv1beta1client := crdclient.ApiextensionsV1beta1().CustomResourceDefinitions()
	_, err = crdv1beta1client.Create(k8scsicrd.CSIDriverCRD())
	if apierrors.IsAlreadyExists(err) {
		glog.V(1).Info("CSIDriver CRD already had been registered")
	} else if err != nil {
		glog.Error(err.Error())
		os.Exit(1)
	}
	glog.V(1).Info("CSIDriver CRD registered")
	// Set up goroutine to cleanup (aka deregister) on termination.
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go cleanup(c, clientset, csiDriver)

	// Run forever
	for {
		verifyAndAddCSIDriverInfo(clientset, csiDriver)
		time.Sleep(sleepDuration)
	}
}

func cleanup(c <-chan os.Signal, clientSet *k8scsiclient.Clientset, csiDriver *k8scsi.CSIDriver) {
	<-c
	verifyAndDeleteCSIDriverInfo(clientSet, csiDriver)
	os.Exit(1)
}

// Registers CSI driver by creating a CSIDriver object
func verifyAndAddCSIDriverInfo(
	csiClientset *k8scsiclient.Clientset,
	csiDriver *k8scsi.CSIDriver,
) error {
	retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		csidrivers := csiClientset.CsiV1alpha1().CSIDrivers()

		_, err := csidrivers.Create(csiDriver)
		if err == nil {
			glog.V(1).Infof("CSIDriver object created for driver %s", csiDriver.Name)
			return nil
		} else if apierrors.IsAlreadyExists(err) {
			glog.V(1).Info("CSIDriver CRD already had been registered")
			return nil
		}
		glog.Errorf("Failed to create CSIDriver object: %v", err)
		return err
	})
	return retryErr
}

// Deregister CSI Driver by deleting CSIDriver object
func verifyAndDeleteCSIDriverInfo(
	csiClientset *k8scsiclient.Clientset,
	csiDriver *k8scsi.CSIDriver,
) error {
	retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		csidrivers := csiClientset.CsiV1alpha1().CSIDrivers()
		err := csidrivers.Delete(csiDriver.Name, &metav1.DeleteOptions{})
		if err == nil {
			glog.V(1).Infof("CSIDriver object deleted for driver %s", csiDriver.Name)
			return nil
		} else if apierrors.IsNotFound(err) {
			glog.V(1).Info("No need to clean up CSIDriver since it does not exist")
			return nil
		}
		glog.Errorf("Failed to delete CSIDriver object: %v", err)
		return err
	})
	return retryErr
}
