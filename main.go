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

package main

import (
	"flag"
	"io/ioutil"
	"os"

	httptransport "github.com/go-openapi/runtime/client"
	"go.uber.org/zap/zapcore"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	"github.com/johnhoman/go-kfp"
	kfpv1alpha1 "github.com/johnhoman/kfp-releaser/api/v1alpha1"
	"github.com/johnhoman/kfp-releaser/controllers"
	//+kubebuilder:scaffold:imports
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))

	utilruntime.Must(kfpv1alpha1.AddToScheme(scheme))
	//+kubebuilder:scaffold:scheme
}

func newWhaleSay() map[string]interface{} {
	// Not sure if the name actually matters -- might be able to swap it for a uuid
	content := map[string]interface{}{
		"apiVersion": "argoproj.io/v1alpha1",
		"kind":       "Workflow",
		"metadata": map[string]interface{}{
			"name": "whalesay",
		},
		"spec": map[string]interface{}{
			"entrypoint": "whalesay",
			"arguments": map[string]interface{}{
				"parameters": []interface{}{
					map[string]interface{}{
						"name":  "name",
						"value": "Jack",
					},
				},
			},
			"templates": []interface{}{
				map[string]interface{}{
					"name": "whalesay",
					"inputs": map[string]interface{}{
						"parameters": []interface{}{
							map[string]interface{}{"name": "name"},
						},
					},
					"container": map[string]interface{}{
						"image":   "docker/whalesay",
						"command": []string{"cowsay"},
						"args":    []string{"Hello", "{{inputs.parameters.name}}"},
					},
				},
			},
		},
	}
	return content
}

func main() {
	var metricsAddr string
	var enableLeaderElection bool
	var probeAddr string
	var apiServer string
	flag.StringVar(&metricsAddr, "metrics-bind-address", ":8080", "The address the metric endpoint binds to.")
	flag.StringVar(&probeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
	flag.StringVar(&apiServer, "ml-pipeline-api-server", "ml-pipeline.kubeflow.svc.cluster.local:8888", "The ml-pipeline api server address")
	flag.BoolVar(&enableLeaderElection, "leader-elect", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")
	opts := zap.Options{
		Development: true,
		TimeEncoder: zapcore.ISO8601TimeEncoder,
	}
	opts.BindFlags(flag.CommandLine)
	flag.Parse()

	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:                 scheme,
		MetricsBindAddress:     metricsAddr,
		Port:                   9443,
		HealthProbeBindAddress: probeAddr,
		LeaderElection:         enableLeaderElection,
		LeaderElectionID:       "e1ae2438.jackhoman.com",
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}
	transport := httptransport.New(apiServer, "", []string{"http"})
	pipelines := kfp.NewPipelineService(transport)
	api := kfp.New(pipelines, nil)
	// This will need to be reloaded regularly
	token, err := ioutil.ReadFile("/var/run/secrets/kubeflow/pipelines")
	if os.IsNotExist(err) {
		setupLog.Info(
			"kubeflow pipelines service account token not found",
			"warning",
			"file /var/run/secrets/kubeflow/pipelines not found",
		)
	} else if err != nil {
		setupLog.Error(err, "an error occurred reading the service account token")
	} else {
		api = kfp.New(pipelines, httptransport.BearerToken(string(token)))
	}

	if err = (&controllers.PipelineReconciler{
		Client:        mgr.GetClient(),
		Scheme:        mgr.GetScheme(),
		Pipelines:     api,
		BlankWorkflow: newWhaleSay(),
		EventRecorder: mgr.GetEventRecorderFor("controller.pipeline"),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Pipeline")
		os.Exit(1)
	}
	if err = (&controllers.PipelineVersionReconciler{
		Client:        mgr.GetClient(),
		Scheme:        mgr.GetScheme(),
		Pipelines:     api,
		EventRecorder: mgr.GetEventRecorderFor("controller.pipelineversion"),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "PipelineVersion")
		os.Exit(1)
	}
	if err = (&controllers.RecurringRunReconciler{
		Client:        mgr.GetClient(),
		Scheme:        mgr.GetScheme(),
		EventRecorder: mgr.GetEventRecorderFor("controller.recurringrun"),
		Pipelines:     api,
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "RecurringRun")
		os.Exit(1)
	}
	//+kubebuilder:scaffold:builder

	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up health check")
		os.Exit(1)
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up ready check")
		os.Exit(1)
	}

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}
