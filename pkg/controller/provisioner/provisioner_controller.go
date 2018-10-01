/*
Copyright 2018 The Knative Authors.

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

package provisioner

import (
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/test-infra/boskos/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

type ProvisionerController struct {
	client.Client
	scheme *runtime.Scheme
	log    logr.Logger
}

func Add(mgr manager.Manager) error {
	// Create a new Controller
	c, err := controller.New("k8sevents-provisioner-controller", mgr,
		controller.Options{Reconciler: &ProvisionerController{
			Client: mgr.GetClient(),
			scheme: mgr.GetScheme(),
		}})
	if err != nil {
		return err
	}

	// Watch for changes to ClusterProvisioner
	//TODO add a filtered source
	err = c.Watch(
		&source.Kind{Type: &eventingv1alpha1.ClusterProvisioner{}},
		&handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	return nil
}

func (pc *ProvisionerController) Reconcile(req reconcile.Request)
