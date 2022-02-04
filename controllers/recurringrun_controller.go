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
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	cu "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/johnhoman/go-kfp"
	kfpv1alpha1 "github.com/johnhoman/kfp-releaser/api/v1alpha1"
)

// RecurringRunReconciler reconciles a RecurringRun object
type RecurringRunReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	record.EventRecorder

	Pipelines kfp.Interface
}

const (
	RecurringRunFinalizer = "kfp.jackhoman.com/delete-recurring-run"
)

//+kubebuilder:rbac:groups=core,resources=events,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=kfp.jackhoman.com,resources=recurringruns,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=kfp.jackhoman.com,resources=recurringruns/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=kfp.jackhoman.com,resources=recurringruns/finalizers,verbs=update
//+kubebuilder:rbac:groups=kfp.jackhoman.com,resources=pipelineversions,verbs=get;list;watch;

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *RecurringRunReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	k8s := client.NewNamespacedClient(r.Client, req.Namespace)
	api := kfp.NewNamespaced(r.Pipelines, req.Namespace)

	instance := &kfpv1alpha1.RecurringRun{}
	if err := k8s.Get(ctx, req.NamespacedName, instance); err != nil {
		logger.Info("instance not found", "error", err.Error())
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if !instance.GetDeletionTimestamp().IsZero() {
		// RecurringRun is being deleted
		if cu.ContainsFinalizer(instance, RecurringRunFinalizer) {
			job, err := api.GetJob(ctx, &kfp.GetOptions{Name: instance.GetName()})
			if err != nil {
				if !kfp.IsNotFound(err) {
					logger.Error(err, "unable to get job")
					return ctrl.Result{}, err
				}
			} else {
				// err == nil
				if err := api.DeleteJob(ctx, &kfp.DeleteOptions{ID: job.ID}); err != nil {
					logger.Error(err, "unable to delete recurring run")
					return ctrl.Result{}, err
				}
				r.Eventf(instance, corev1.EventTypeNormal, "Deleted", fmt.Sprintf(
					"Removed recurring run %s", instance.Status.ID,
				))
			}
			patch := client.MergeFrom(instance.DeepCopy())
			cu.RemoveFinalizer(instance, RecurringRunFinalizer)
			if err := k8s.Patch(ctx, instance, patch); err != nil {
				logger.Error(err, "unable to remove finalizer")
				return ctrl.Result{}, err
			}
			logger.Info("removed finalizer")
		}

		return ctrl.Result{}, nil
	}
	if !cu.ContainsFinalizer(instance, RecurringRunFinalizer) {
		patch := &unstructured.Unstructured{Object: map[string]interface{}{
			"metadata": map[string]interface{}{
				"finalizers": []string{RecurringRunFinalizer},
			},
		}}
		patch.SetGroupVersionKind(instance.GroupVersionKind())
		patch.SetName(instance.GetName())
		if err := k8s.Patch(ctx, patch, client.Apply, FieldOwner); err != nil {
			logger.Error(err, "unable to add finalizer")
			return ctrl.Result{}, err
		}
		logger.Info("added finalizer")
	}

	version := &kfpv1alpha1.PipelineVersion{}
	if err := k8s.Get(ctx, types.NamespacedName{Name: instance.Spec.VersionRef}, version); err != nil {
		r.Eventf(instance, corev1.EventTypeWarning, "VersionNotFound", fmt.Sprintf(
			"Could not find version %s", instance.Spec.VersionRef,
		))
		return ctrl.Result{}, err
	}

	var err error
	job := &kfp.Job{}
	job, err = api.GetJob(ctx, &kfp.GetOptions{Name: instance.GetName()})
	if err != nil {
		if !kfp.IsNotFound(err) {
			return ctrl.Result{}, err
		}
		var parameters []kfp.Parameter = nil
		if instance.Spec.Parameters != nil && len(instance.Spec.Parameters) > 0 {
			parameters = make([]kfp.Parameter, 0, len(instance.Spec.Parameters))
			for _, param := range instance.Spec.Parameters {
				parameters = append(parameters, kfp.Parameter{
					Name:  param.Name,
					Value: param.Value,
				})
			}
		}
		job, err = api.CreateJob(ctx, &kfp.CreateJobOptions{
			Name:           instance.GetName(),
			PipelineID:     version.Status.PipelineID,
			VersionID:      version.Status.ID,
			CronSchedule:   instance.Spec.Schedule.Cron,
			Enabled:        true,
			MaxConcurrency: 1,
			Parameters:     parameters,
		})
		if err != nil {
			logger.Error(err, "an error occurred creating the recurring run")
			return ctrl.Result{}, err
		}
		r.Eventf(instance, corev1.EventTypeNormal, "Created", fmt.Sprintf(
			"Created recurring run %s:%s", job.Name, job.ID,
		))
	}
	original := instance.DeepCopy()

	instance.Status.ID = job.ID
	instance.Status.VersionID = job.VersionID
	instance.Status.PipelineID = job.PipelineID
	if !reflect.DeepEqual(instance.Status, original.Status) {
		if err := k8s.Status().Update(ctx, instance); err != nil {
			logger.Error(err, "unable to update status")
			return ctrl.Result{}, err
		}
		logger.Info("updated status")
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *RecurringRunReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&kfpv1alpha1.RecurringRun{}).
		Complete(r)
}
