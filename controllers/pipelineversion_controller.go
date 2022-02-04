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

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/json"
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

// PipelineVersionReconciler reconciles a PipelineVersion object
type PipelineVersionReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	record.EventRecorder

	Pipelines kfp.Interface
}

const (
	VersionFinalizer              = "kfp.jackhoman.com/delete-pipeline-version"
	ReasonPipelineNotFound        = "PipelineNotFound"
	ReasonPipelineVersionCreated  = "Created"
	ReasonPipelineVersionDeleted  = "Deleted"
	ReasonPipelineVersionConflict = "Conflict"
	ReasonAPIError                = "APIError"
)

//+kubebuilder:rbac:groups=core,resources=events,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=kfp.jackhoman.com,resources=pipelines,verbs=get;list;watch
//+kubebuilder:rbac:groups=kfp.jackhoman.com,resources=pipelines/status,verbs=get
//+kubebuilder:rbac:groups=kfp.jackhoman.com,resources=pipelineversions,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=kfp.jackhoman.com,resources=pipelineversions/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=kfp.jackhoman.com,resources=pipelineversions/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *PipelineVersionReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	// This is going to be kind of a weird controller -- this should
	// be owned by the pipeline that controls it

	logger := log.FromContext(ctx)

	k8s := client.NewNamespacedClient(r.Client, req.Namespace)

	instance := &kfpv1alpha1.PipelineVersion{}
	if err := k8s.Get(ctx, req.NamespacedName, instance); err != nil {
		logger.Info("instance not found", "error", err.Error())
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if !instance.GetDeletionTimestamp().IsZero() {
		logger.Info("Deleting pipeline resource")
		// Under deletion
		if cu.ContainsFinalizer(instance, VersionFinalizer) {
			err := r.Pipelines.DeleteVersion(ctx, &kfp.DeleteOptions{ID: instance.Status.ID})
			if err != nil && !kfp.IsNotFound(err) {
				return ctrl.Result{}, err
			}
			logger.Info("Deleted pipeline resource")
			patch := client.MergeFrom(instance.DeepCopy())
			cu.RemoveFinalizer(instance, VersionFinalizer)
			if err := k8s.Patch(ctx, instance, patch, FieldOwner); err != nil {
				return ctrl.Result{}, err
			}
			logger.Info("Removed finalizer")
		}
		return ctrl.Result{}, nil
	}
	if !cu.ContainsFinalizer(instance, Finalizer) {
		patch := &unstructured.Unstructured{Object: map[string]interface{}{
			"metadata": map[string]interface{}{
				"finalizers": []string{VersionFinalizer},
			},
		}}
		patch.SetGroupVersionKind(instance.GroupVersionKind())
		patch.SetName(instance.GetName())
		if err := k8s.Patch(ctx, patch, client.Apply, FieldOwner, client.ForceOwnership); err != nil {
			return ctrl.Result{}, err
		}
	}

	pipeline := &kfpv1alpha1.Pipeline{}
	key := types.NamespacedName{Name: instance.Spec.Pipeline, Namespace: instance.GetNamespace()}
	if err := k8s.Get(ctx, key, pipeline); err != nil {
		r.Event(instance, corev1.EventTypeWarning, ReasonPipelineNotFound, fmt.Sprintf(
			"Could not find pipeline %s", instance.Spec.Pipeline,
		))
		return ctrl.Result{}, err
	}
	name := instance.GetName()
	version, err := r.Pipelines.GetVersion(ctx, &kfp.GetVersionOptions{
		Name:       name,
		PipelineID: pipeline.Status.ID,
	})
	if err != nil {
		if !kfp.IsNotFound(err) {
			return ctrl.Result{}, err
		}
		var m map[string]interface{}
		if err := json.Unmarshal(instance.Spec.Workflow.Raw, &m); err != nil {
			return ctrl.Result{}, err
		}
		out, err := r.Pipelines.CreateVersion(ctx, &kfp.CreateVersionOptions{
			PipelineID:  pipeline.Status.ID,
			Name:        name,
			Description: instance.Spec.Description,
			Workflow:    m,
		})
		if err != nil {
			r.Event(instance, corev1.EventTypeWarning, ReasonAPIError, fmt.Sprintf(
				"Unknown error occured %s", err.Error(),
			))
			return ctrl.Result{}, err
		}
		r.Event(instance, corev1.EventTypeNormal, ReasonPipelineVersionCreated, fmt.Sprintf(
			"Created pipeline version %s for pipeline %s", out.ID, pipeline.Status.ID,
		))
		*version = *out
	}

	old := instance.DeepCopy()
	instance.Status.ID = version.ID
	instance.Status.Name = version.Name
	instance.Status.PipelineID = version.PipelineID
	if version.Parameters != nil && len(version.Parameters) > 0 {
		params := make([]kfpv1alpha1.Parameter, 0, len(version.Parameters))
		for _, param := range version.Parameters {
			params = append(params, kfpv1alpha1.Parameter{Name: param.Name, Value: param.Value})
		}
		instance.Status.Parameters = params
	}
	if !reflect.DeepEqual(old.Status, instance.Status) {
		logger.Info("Updating status")
		patch := client.MergeFrom(old)
		if err := k8s.Status().Patch(ctx, instance, patch); err != nil {
			return ctrl.Result{}, err
		}
		logger.Info("Status updated")
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *PipelineVersionReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&kfpv1alpha1.PipelineVersion{}).
		Watches(
			&source.Kind{Type: &kfpv1alpha1.Pipeline{}},
			handler.EnqueueRequestsFromMapFunc(func(obj client.Object) []ctrl.Request {
				pipeline, ok := obj.(*kfpv1alpha1.Pipeline)
				if ok {
					versions := make([]ctrl.Request, 0, len(pipeline.Status.Versions))
					for _, version := range pipeline.Status.Versions {
						versions = append(versions, ctrl.Request{
							NamespacedName: types.NamespacedName{Name: version.Name, Namespace: pipeline.GetNamespace()},
						})
					}
					return versions
				}

				return []ctrl.Request{}
			}),
		).
		Complete(r)
}
