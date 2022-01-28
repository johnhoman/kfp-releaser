package controllers

import (
	"encoding/json"
	kfpv1alpha1 "github.com/johnhoman/kfp-releaser/api/v1alpha1"
	"k8s.io/apimachinery/pkg/types"
	"os"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strings"

	httptransport "github.com/go-openapi/runtime/client"
	"k8s.io/client-go/kubernetes/scheme"

	"github.com/johnhoman/controller-tools/manager"
	"github.com/johnhoman/go-kfp"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("RecurringRunController", func() {

	var it manager.IntegrationTest
	var service kfp.Interface
	var raw []byte

	BeforeEach(func() {
		address, ok := os.LookupEnv("GO_KFP_API_SERVER_ADDRESS")
		if !ok {
			Fail("could not run tests without kubeflow api service address (export GO_KFP_API_SERVER_ADDRESS=)")
		}
		if strings.HasPrefix(address, "http://") {
			address = strings.TrimPrefix(address, "http://")
		}
		transport := httptransport.New(address, "", []string{"http"})
		service = kfp.New(kfp.NewPipelineService(transport), nil)
		it = manager.IntegrationTestBuilder().
			WithScheme(scheme.Scheme).
			Complete(cfg)
		// I think I need this to get the pipeline ID from the status field
		err := (&PipelineReconciler{
			Client:        it.GetClient(),
			Scheme:        it.GetScheme(),
			Pipelines:     service,
			BlankWorkflow: workflow(),
			EventRecorder: it.GetEventRecorderFor("kfp-releaser.controller-test"),
		}).SetupWithManager(it)
		Expect(err).ToNot(HaveOccurred())
		err = (&PipelineVersionReconciler{
			Client:        it.GetClient(),
			Scheme:        it.GetScheme(),
			Pipelines:     service,
			EventRecorder: it.GetEventRecorderFor("kfp-releaser.controller-test"),
		}).SetupWithManager(it)
		Expect(err).ToNot(HaveOccurred())
		err = (&RecurringRunReconciler{
			Client:        it.GetClient(),
			Scheme:        it.GetScheme(),
			EventRecorder: it.GetEventRecorderFor("kfp-releaser.controller-test"),
			api:           service,
		}).SetupWithManager(it)
		Expect(err).ToNot(HaveOccurred())

		it.StartManager()

		raw, err = json.Marshal(workflow())
		Expect(err).ShouldNot(HaveOccurred())
	})
	AfterEach(func() { it.StopManager() })
	When("a pipeline exists", func() {
		var pipeline *kfpv1alpha1.Pipeline
		var version *kfpv1alpha1.PipelineVersion
		var name string
		var key types.NamespacedName
		var versionKey types.NamespacedName
		BeforeEach(func() {
			name = "whalesay"
			key = types.NamespacedName{Name: name}
			versionKey = types.NamespacedName{Name: name + "-v1"}

			pipeline = &kfpv1alpha1.Pipeline{}
			pipeline.SetName(name)
			pipeline.Spec.Description = "Whalesay Pipeline"
			it.Eventually().Create(pipeline).Should(Succeed())
			it.Eventually().GetWhen(key, pipeline, func(obj client.Object) bool {
				return len(obj.(*kfpv1alpha1.Pipeline).Status.ID) > 0
			}).Should(Succeed())

			version = &kfpv1alpha1.PipelineVersion{}
			version.SetName(versionKey.Name)
			version.Spec.Workflow.Raw = raw
			version.Spec.Pipeline = pipeline.GetName()
			version.Spec.Description = pipeline.Spec.Description

			it.Eventually().Create(version).Should(Succeed())
			it.Eventually().GetWhen(versionKey, version, func(obj client.Object) bool {
				return len(obj.(*kfpv1alpha1.PipelineVersion).Status.ID) > 0
			}).Should(Succeed())
		})
		AfterEach(func() {
			// Just call delete, pass or fail doesn't matter all that much
			it.Expect().Delete(version).Should(Or(Succeed(), Not(Succeed())))
			it.Expect().Delete(pipeline).Should(Or(Succeed(), Not(Succeed())))
		})
		When("the recurring run exists", func() {
			var recurringRun *kfpv1alpha1.RecurringRun
			BeforeEach(func() {
				recurringRun = &kfpv1alpha1.RecurringRun{}
				recurringRun.SetName(versionKey.Name)
				recurringRun.SetFinalizers([]string{"keep"})
				recurringRun.Spec.Schedule = kfpv1alpha1.RecurringRunSchedule{Cron: "* * * * *"}
				recurringRun.Spec.VersionRef = version.GetName()
				it.Eventually().Create(recurringRun).Should(Succeed())
			})
			AfterEach(func() {
				it.Expect().Delete(recurringRun).Should(Or(Succeed(), Not(Succeed())))
			})
			When("the recurring run is not being deleted", func() {
				When("it has a finalizer", func() {
					BeforeEach(func() {
						instance := &kfpv1alpha1.RecurringRun{}
						it.Eventually().GetWhen(versionKey, instance, func(obj client.Object) bool {
							return len(obj.GetFinalizers()) > 1
						}).Should(Succeed())
					})
					It("should contain the correct finalizer", func() {
						instance := &kfpv1alpha1.RecurringRun{}
						it.Eventually().Get(versionKey, instance).Should(Succeed())
						Expect(instance.GetFinalizers()).Should(ContainElement(RecurringRunFinalizer))
					})
				})
				When("it does not have a finalizer", func() {

				})
			})
			When("the recurring run is being deleted", func() {
				BeforeEach(func() {
					it.Expect().Delete(recurringRun).Should(Succeed())
					it.Eventually().GetWhen(versionKey, &kfpv1alpha1.RecurringRun{}, func(obj client.Object) bool {
						return len(obj.GetFinalizers()) == 1
					}).Should(Succeed())
				})
				It("should remove the finalizer", func() {
					instance := &kfpv1alpha1.RecurringRun{}
					it.Eventually().Get(versionKey, instance).Should(Succeed())
					Expect(instance.GetFinalizers()).ShouldNot(ContainElement(RecurringRunFinalizer))
				})
			})
		})
	})

})
