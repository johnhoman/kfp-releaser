/*
Copyright 2022 John Homan.

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
	"github.com/johnhoman/go-kfp"
	"github.com/johnhoman/go-kfp/pipelines"
	kfpv1alpha1 "github.com/johnhoman/kfp-releaser/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"reflect"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	cu "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

const (
	Finalizer                    = "kfp.jackhoman.com/delete-pipeline"
	FieldOwner client.FieldOwner = "kfp-releaser"
)

// PipelineReconciler reconciles a Pipeline object
type PipelineReconciler struct {
	client.Client
	Scheme *runtime.Scheme

	Pipelines     kfp.Pipelines
	BlankWorkflow map[string]interface{}
}

//+kubebuilder:rbac:groups=kfp.jackhoman.com,resources=pipelineversions,verbs=get;list;watch
//+kubebuilder:rbac:groups=kfp.jackhoman.com,resources=pipelines,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=kfp.jackhoman.com,resources=pipelines/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=kfp.jackhoman.com,resources=pipelines/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Pipeline object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.10.0/pkg/reconcile
func (r *PipelineReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	k8s := client.NewNamespacedClient(r.Client, req.Namespace)

	instance := &kfpv1alpha1.Pipeline{}
	if err := k8s.Get(ctx, req.NamespacedName, instance); err != nil {
		logger.V(3).Info("instance not found", "error", err.Error())
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if !instance.GetDeletionTimestamp().IsZero() {
		logger.Info("Deleting pipeline resource")
		// Delete resources
		if cu.ContainsFinalizer(instance, Finalizer) {
			// Delete the resource
			if err := r.Pipelines.Delete(ctx, &kfp.DeleteOptions{ID: instance.Status.ID}); err != nil {
				if !kfp.IsNotFound(err) {
					return ctrl.Result{}, err
				}
			}

			patch := client.MergeFrom(instance.DeepCopy())
			cu.RemoveFinalizer(instance, Finalizer)
			if err := k8s.Patch(ctx, instance, patch, FieldOwner); err != nil {
				return ctrl.Result{}, err
			}
		}
		// Don't requeue
		return ctrl.Result{}, nil
	}
	if !cu.ContainsFinalizer(instance, Finalizer) {
		patch := &unstructured.Unstructured{Object: map[string]interface{}{
			"metadata": map[string]interface{}{
				"finalizers": []string{Finalizer},
			},
		}}
		patch.SetName(instance.GetName())
		patch.SetGroupVersionKind(instance.GroupVersionKind())
		logger.Info("Adding finalizer")
		if err := k8s.Patch(ctx, patch, client.Apply, FieldOwner, client.ForceOwnership); err != nil {
			return ctrl.Result{}, err
		}
		logger.Info("Added finalizer")
	}
	// Need to get the pipeline by the ID
	name := instance.GetNamespace() + "-" + instance.GetName()
	pipeline, err := r.Pipelines.Get(ctx, &kfp.GetOptions{Name: name})
	if err != nil {
		if !pipelines.IsNotFound(err) {
			return ctrl.Result{}, err
		}
		pipeline, err = r.Pipelines.Create(ctx, &kfp.CreateOptions{
			Name:        name,
			Description: instance.Spec.Description,
			Workflow:    r.BlankWorkflow,
		})
		if err != nil {
			return ctrl.Result{}, err
		}
		// Remove the blank after the pipeline stub is created
		err := r.Pipelines.DeleteVersion(ctx, &kfp.DeleteVersionOptions{ID: pipeline.ID})
		if err != nil {
			return ctrl.Result{}, err
		}
	}
	old := instance.DeepCopy()
	instance.Status.ID = pipeline.ID
	instance.Status.DefaultVersion = pipeline.DefaultVersionID
	instance.Status.CreatedAt = metav1.NewTime(pipeline.CreatedAt)
	if !reflect.DeepEqual(instance.Status, old.Status) {
		patch := client.MergeFrom(old)
		if err := k8s.Status().Patch(ctx, instance, patch); err != nil {
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *PipelineReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&kfpv1alpha1.Pipeline{}).
		Watches(
			&source.Kind{Type: &kfpv1alpha1.PipelineVersion{}},
			handler.EnqueueRequestsFromMapFunc(func(obj client.Object) []ctrl.Request {
				version, ok := obj.(*kfpv1alpha1.PipelineVersion)
				if !ok {
					return []ctrl.Request{}
				}
				ref := version.Spec.Pipeline
				return []ctrl.Request{{NamespacedName: types.NamespacedName{
					Name:      ref,
					Namespace: version.GetNamespace(),
				}}}
			}),
		).
		Complete(r)
}
