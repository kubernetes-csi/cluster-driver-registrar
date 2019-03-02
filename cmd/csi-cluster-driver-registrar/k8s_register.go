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

	k8scsi "k8s.io/api/storage/v1beta1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/util/retry"
	"k8s.io/klog"
)

func kubernetesRegister(
	config *rest.Config,
	csiDriver *k8scsi.CSIDriver,
) {
	// Get client info to CSIDriver
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		klog.Error(err.Error())
		os.Exit(1)
	}

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

func cleanup(c <-chan os.Signal, clientSet *kubernetes.Clientset, csiDriver *k8scsi.CSIDriver) {
	<-c
	verifyAndDeleteCSIDriverInfo(clientSet, csiDriver)
	os.Exit(1)
}

// Registers CSI driver by creating a CSIDriver object
func verifyAndAddCSIDriverInfo(
	csiClientset *kubernetes.Clientset,
	csiDriver *k8scsi.CSIDriver,
) error {
	retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		csidrivers := csiClientset.StorageV1beta1().CSIDrivers()

		_, err := csidrivers.Create(csiDriver)
		if err == nil {
			klog.V(1).Infof("CSIDriver object created for driver %s", csiDriver.Name)
			return nil
		} else if apierrors.IsAlreadyExists(err) {
			klog.V(1).Info("CSIDriver CRD already had been registered")
			return nil
		}
		klog.Errorf("Failed to create CSIDriver object: %v", err)
		return err
	})
	return retryErr
}

// Deregister CSI Driver by deleting CSIDriver object
func verifyAndDeleteCSIDriverInfo(
	csiClientset *kubernetes.Clientset,
	csiDriver *k8scsi.CSIDriver,
) error {
	retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		csidrivers := csiClientset.StorageV1beta1().CSIDrivers()
		err := csidrivers.Delete(csiDriver.Name, &metav1.DeleteOptions{})
		if err == nil {
			klog.V(1).Infof("CSIDriver object deleted for driver %s", csiDriver.Name)
			return nil
		} else if apierrors.IsNotFound(err) {
			klog.V(1).Info("No need to clean up CSIDriver since it does not exist")
			return nil
		}
		klog.Errorf("Failed to delete CSIDriver object: %v", err)
		return err
	})
	return retryErr
}
