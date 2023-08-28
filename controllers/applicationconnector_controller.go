/*
Copyright 2022.

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

package controllers

import (
	"context"
	"reflect"

	"github.com/kyma-project/application-connector-manager/api/v1alpha1"
	"github.com/kyma-project/application-connector-manager/pkg/reconciler"
	"go.uber.org/zap"
	v2 "k8s.io/api/autoscaling/v2"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

const (
	chartNs = "kyma-system"
)

type State string

// Valid CustomObject States.
const (
	// StateReady signifies application-connector is ready and has been installed successfully.
	StateReady State = "Ready"

	// StateProcessing signifies application-connector is reconciling and is in the process of installation.
	// Processing can also signal that the Installation previously encountered an error and is now recovering.
	StateProcessing State = "Processing"

	// StateError signifies an error for application-connector. This signifies that the Installation
	// process encountered an error.
	// Contrary to Processing, it can be expected that this state should change on the next retry.
	StateError State = "Error"

	// StateDeleting signifies application-connector is being deleted. This is the state that is used
	// when a deletionTimestamp was detected and Finalizers are picked up.
	StateDeleting State = "Deleting"
)

type ApplicationConnetorReconciler interface {
	reconcile.Reconciler
	SetupWithManager(mgr ctrl.Manager) error
}

type applicationConnectorReconciler struct {
	log *zap.SugaredLogger
	reconciler.Cfg
	reconciler.K8s
}

func NewApplicationConnetorReconciler(c client.Client, r record.EventRecorder, log *zap.SugaredLogger, o []unstructured.Unstructured) ApplicationConnetorReconciler {
	return &applicationConnectorReconciler{
		log: log,
		Cfg: reconciler.Cfg{
			Finalizer: v1alpha1.Finalizer,
			Objs:      o,
		},
		K8s: reconciler.K8s{
			Client:        c,
			EventRecorder: r,
		},
	}
}

func (r *applicationConnectorReconciler) mapFunction(object client.Object) []reconcile.Request {
	var applicationConnectors v1alpha1.ApplicationConnectorList
	err := r.List(context.Background(), &applicationConnectors)

	if apierrors.IsNotFound(err) {
		return nil
	}

	if err != nil {
		r.log.Error(err)
		return nil
	}

	if len(applicationConnectors.Items) < 1 {
		return nil
	}

	// instance is being deleted, do not notify it about changes
	instanceIsBeingDeleted := !applicationConnectors.Items[0].GetDeletionTimestamp().IsZero()
	if instanceIsBeingDeleted {
		return nil
	}

	r.log.
		With("name", object.GetName()).
		With("ns", object.GetNamespace()).
		With("gvk", object.GetObjectKind().GroupVersionKind()).
		With("rscVer", object.GetResourceVersion()).
		With("appConRscVer", applicationConnectors.Items[0].ResourceVersion).
		Debug("redirecting")

	// make sure only 1 controller will handle change
	return []ctrl.Request{
		{
			NamespacedName: types.NamespacedName{
				Namespace: applicationConnectors.Items[0].Namespace,
				Name:      applicationConnectors.Items[0].Name,
			},
		},
	}
}

var ommitStatusChanged = predicate.Or(
	predicate.LabelChangedPredicate{},
	predicate.AnnotationChangedPredicate{},
	predicate.GenerationChangedPredicate{},
)

type hpaResourceVersionChangedPredicate struct {
	predicate.ResourceVersionChangedPredicate
	log *zap.SugaredLogger
}

func (h hpaResourceVersionChangedPredicate) Update(e event.UpdateEvent) bool {
	if update := h.ResourceVersionChangedPredicate.Update(e); !update {
		return false
	}

	var newObj v2.HorizontalPodAutoscaler
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(e.ObjectNew.(*unstructured.Unstructured).Object, &newObj); err != nil {
		return true
	}

	var oldObj v2.HorizontalPodAutoscaler
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(e.ObjectOld.(*unstructured.Unstructured).Object, &oldObj); err != nil {
		return true
	}

	conditionsEqual := reflect.DeepEqual(oldObj.Status.Conditions, newObj.Status.Conditions)
	replicasEqual := oldObj.Status.CurrentReplicas == newObj.Status.CurrentReplicas

	result := !conditionsEqual || !replicasEqual
	if result {
		h.log.With("conditionsEqual", conditionsEqual, "replicasEqual", replicasEqual).Debugf("reconciliation triggered by HPA: %s/%s", oldObj.Namespace, oldObj.Name)
	}
	return result
}

var hpaGroupVersionKind = schema.GroupVersionKind{
	Group:   v2.GroupName,
	Version: "v2",
	Kind:    "HorizontalPodAutoscaler",
}

// SetupWithManager sets up the controller with the Manager.
func (r *applicationConnectorReconciler) SetupWithManager(mgr ctrl.Manager) error {
	labelSelectorPredicate, err := predicate.LabelSelectorPredicate(
		metav1.LabelSelector{
			MatchLabels: map[string]string{
				"app.kubernetes.io/part-of": "application-connector-manager",
			},
		},
	)
	if err != nil {
		return err
	}

	b := ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.ApplicationConnector{}, builder.WithPredicates(ommitStatusChanged))

	// create functtion to register wached objects
	watchFn := func(u unstructured.Unstructured) {
		var objPredicate predicate.Predicate = &predicate.ResourceVersionChangedPredicate{}
		if u.GroupVersionKind() == hpaGroupVersionKind {
			objPredicate = hpaResourceVersionChangedPredicate{
				log: r.log,
			}
		}

		r.log.With("gvk", u.GroupVersionKind().String()).Infoln("adding watcher")
		b = b.Watches(
			&source.Kind{Type: &u},
			handler.EnqueueRequestsFromMapFunc(r.mapFunction),
			builder.WithPredicates(
				predicate.And(
					labelSelectorPredicate,
					objPredicate,
				),
			),
		)
	}
	// register watch for each managed type of object
	if err := registerWatchDistinct(r.Objs, watchFn); err != nil {
		return err
	}

	return b.Complete(r)
}

// ManifestResolver represents the chart information for the passed Sample resource.
type ManifestResolver struct {
	chartPath string
}

func (r *applicationConnectorReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	var instance v1alpha1.ApplicationConnector
	if err := r.Get(ctx, req.NamespacedName, &instance); err != nil {
		return ctrl.Result{
			Requeue: true,
		}, client.IgnoreNotFound(err)
	}

	stateFSM := reconciler.NewFsm(r.log, r.Cfg, reconciler.K8s{
		Client:        r.Client,
		EventRecorder: r.EventRecorder,
	})
	return stateFSM.Run(ctx, instance)
}