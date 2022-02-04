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
	"fmt"
	"reflect"
	"sort"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	cu "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/source"

	"github.com/johnhoman/go-kfp"
	kfpv1alpha1 "github.com/johnhoman/kfp-releaser/api/v1alpha1"
)

const (
	Finalizer                    = "kfp.jackhoman.com/delete-pipeline"
	FieldOwner client.FieldOwner = "kfp-releaser"
)

// PipelineReconciler reconciles a Pipeline object
type PipelineReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	record.EventRecorder

	Pipelines     kfp.Interface
	BlankWorkflow map[string]interface{}
}

//+kubebuilder:rbac:groups=core,resources=events,verbs=get;list;watch;create;update;patch;delete
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
			// TODO: For some reason it's going here with an empty status
			//       which doesn't make a whole lot of sense since only
			//       a single object is supposed to be reconciled at a time
			if err := r.Pipelines.Delete(ctx, &kfp.DeleteOptions{ID: instance.Status.ID}); err != nil {
				logger.Info("upstream resource not found")
				return ctrl.Result{}, err
			} else {
				logger.Info("Deleted upstream resource")
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
		if !kfp.IsNotFound(err) {
			r.Event(instance, corev1.EventTypeWarning, "GetPipelineError", err.Error())
			return ctrl.Result{}, err
		}
		pipeline, err = r.Pipelines.Create(ctx, &kfp.CreateOptions{
			Name:        name,
			Description: instance.Spec.Description,
			Workflow:    r.BlankWorkflow,
		})
		if err != nil {
			if kfp.IsConflict(err) {
				// This shouldn't come down this path but for some reason it has been
				// probably because kubeflow doesn't really work
				r.Eventf(instance, corev1.EventTypeWarning, "Conflict", "Pipeline %s already exists", name)
				return ctrl.Result{}, nil
			}
			return ctrl.Result{}, err
		}
		// Remove the blank after the pipeline stub is created
		err := r.Pipelines.DeleteVersion(ctx, &kfp.DeleteOptions{ID: pipeline.ID})
		if err != nil {
			r.Event(instance, corev1.EventTypeWarning, "DeleteVersionError", err.Error())
			return ctrl.Result{}, err
		}
		r.Eventf(instance, corev1.EventTypeNormal, "Created", fmt.Sprintf(
			"created pipeline with id %s",
			pipeline.ID,
		))
	}
	// Own all versions
	versionList := &kfpv1alpha1.PipelineVersionList{}
	if err := k8s.List(ctx, versionList, client.MatchingFields{"spec.pipeline": instance.GetName()}); err != nil {
		r.Eventf(instance, corev1.EventTypeWarning, "ListVersionsFailed", "could not list versions %s", err.Error())
		return ctrl.Result{}, err
	}
	versions := make([]string, 0, len(versionList.Items))
	for _, version := range versionList.Items {
		versions = append(versions, version.GetName())
	}
	sort.Sort(sort.StringSlice(versions))
	versionRefs := make([]corev1.LocalObjectReference, 0, len(versions))
	for _, name := range versions {
		versionRefs = append(versionRefs, corev1.LocalObjectReference{Name: name})
	}
	old := instance.DeepCopy()
	instance.Status.ID = pipeline.ID
	instance.Status.DefaultVersion = pipeline.DefaultVersionID
	instance.Status.CreatedAt = metav1.NewTime(pipeline.CreatedAt)
	instance.Status.Versions = versionRefs
	if !reflect.DeepEqual(instance.Status, old.Status) {
		patch := client.MergeFrom(old)
		if err := k8s.Status().Patch(ctx, instance, patch); err != nil {
			logger.Error(err, "failed to update status")
			return ctrl.Result{}, err
		}
		logger.Info("Updated status")
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *PipelineReconciler) SetupWithManager(mgr ctrl.Manager) error {
	if err := mgr.GetFieldIndexer().IndexField(context.Background(), &kfpv1alpha1.PipelineVersion{}, "spec.pipeline", func(obj client.Object) []string {
		version, ok := obj.(*kfpv1alpha1.PipelineVersion)
		if !ok {
			return []string{}
		}
		return []string{version.Spec.Pipeline}
	}); err != nil {
		return err
	}
	return ctrl.NewControllerManagedBy(mgr).
		For(&kfpv1alpha1.Pipeline{}).
		Watches(
			&source.Kind{Type: &kfpv1alpha1.PipelineVersion{}},
			handler.EnqueueRequestsFromMapFunc(func(obj client.Object) []ctrl.Request {
				version, ok := obj.(*kfpv1alpha1.PipelineVersion)
				if ok {
					return []ctrl.Request{{NamespacedName: types.NamespacedName{
						Name:      version.Spec.Pipeline,
						Namespace: version.GetNamespace()},
					}}
				}

				return []ctrl.Request{}
			}),
		).
		Complete(r)
}
