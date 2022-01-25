package controllers

import (
	kfpv1alpha1 "github.com/johnhoman/kfp-releaser/api/v1alpha1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/json"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/johnhoman/controller-tools/manager"
	"github.com/johnhoman/go-kfp"
	"github.com/johnhoman/go-kfp/fake"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("PipelineVersionController", func() {
	var it manager.IntegrationTest
	var service kfp.Pipelines
	var raw []byte

	BeforeEach(func() {
		service = kfp.New(fake.NewPipelineService(), nil)
		it = manager.IntegrationTestBuilder().
			WithScheme(scheme.Scheme).
			Complete(cfg)
		// I think I need this to get the pipeline ID from the status field
		err := (&PipelineReconciler{
			Client:        it.GetClient(),
			Scheme:        it.GetScheme(),
			Pipelines:     service,
			BlankWorkflow: workflow(),
		}).SetupWithManager(it)
		Expect(err).ToNot(HaveOccurred())
		err = (&PipelineVersionReconciler{
			Client:        it.GetClient(),
			Scheme:        it.GetScheme(),
			Pipelines:     service,
			EventRecorder: it.GetEventRecorderFor("kfp-releaser.controller-test"),
		}).SetupWithManager(it)
		Expect(err).ToNot(HaveOccurred())

		it.StartManager()

		raw, err = json.Marshal(workflow())
		Expect(err).ShouldNot(HaveOccurred())
	})
	AfterEach(func() { it.StopManager() })
	Context("Finalization", func() {
		When("The finalizer is missing from the PipelineVersion", func() {
			It("Should add a finalizer", func() {
				version := &kfpv1alpha1.PipelineVersion{}
				version.SetName("un-finalized")
				version.Spec.Workflow.Raw = raw
				version.Spec.Pipeline = "unknown"
				it.Eventually().Create(version).Should(Succeed())
				version = &kfpv1alpha1.PipelineVersion{}
				it.Eventually().GetWhen(types.NamespacedName{Name: "un-finalized"}, version, func(obj client.Object) bool {
					return len(obj.GetFinalizers()) == 1
				}).Should(Succeed())
				Expect(version.GetFinalizers()).To(ContainElement(VersionFinalizer))
				Expect(version.GetManagedFields()[0].Manager).To(Equal(string(FieldOwner)))
			})
		})
		When("No upstream resource exists", func() {
			It("Should remove the finalizer", func() {
				name := "no-resource-exists"
				versionName := "no-resource-exists-1"
				pipeline := &kfpv1alpha1.Pipeline{}
				pipeline.SetName(name)
				it.Eventually().Create(pipeline).Should(Succeed())
				pipeline = &kfpv1alpha1.Pipeline{}
				it.Eventually().GetWhen(types.NamespacedName{Name: name}, pipeline, func(obj client.Object) bool {
					return len(pipeline.Status.ID) > 0
				}).Should(Succeed())

				version := &kfpv1alpha1.PipelineVersion{}
				version.Spec.Workflow.Raw = raw
				version.Spec.Pipeline = name
				version.SetName(versionName)
				version.SetFinalizers([]string{"keep"})
				it.Eventually().Create(version).Should(Succeed())

				version = &kfpv1alpha1.PipelineVersion{}
				it.Eventually().GetWhen(types.NamespacedName{Name: versionName}, version, func(obj client.Object) bool {
					return len(obj.GetFinalizers()) == 2
				}).Should(Succeed())
				Expect(version.GetFinalizers()).To(ContainElement(VersionFinalizer))

				// Deleting the upstream resource should allow the controller
				// to safely remove the finalizer
				Expect(service.DeleteVersion(
					it.GetContext(), &kfp.DeleteVersionOptions{ID: version.Status.ID}),
				).Should(Succeed())

				version = &kfpv1alpha1.PipelineVersion{}
				version.SetName(versionName)
				it.Expect().Delete(version).Should(Succeed())
				it.Eventually().GetWhen(types.NamespacedName{Name: versionName}, version, func(obj client.Object) bool {
					return len(obj.GetFinalizers()) == 1
				}).Should(Succeed())
				Expect(version.GetFinalizers()).NotTo(ContainElement(VersionFinalizer))
			})
		})
		When("The upstream resource exists", func() {
			It("Should remove the resource", func() {
				pipeline := &kfpv1alpha1.Pipeline{}
				pipeline.SetName("create-version")
				it.Eventually().Create(pipeline).Should(Succeed())
				pipeline = &kfpv1alpha1.Pipeline{}
				it.Eventually().GetWhen(types.NamespacedName{Name: "create-version"}, pipeline, func(obj client.Object) bool {
					return len(pipeline.Status.ID) > 0
				}).Should(Succeed())

				version := &kfpv1alpha1.PipelineVersion{}
				version.SetName("create-version-v1")
				version.Spec.Pipeline = "create-version"
				version.Spec.Workflow.Raw = raw
				version.Spec.Description = "version 1.0.1"
				it.Eventually().Create(version).Should(Succeed())
				version = &kfpv1alpha1.PipelineVersion{}
				it.Eventually().GetWhen(types.NamespacedName{Name: "create-version-v1"}, version, func(obj client.Object) bool {
					return len(obj.(*kfpv1alpha1.PipelineVersion).Status.ID) > 0
				}).Should(Succeed())
				Expect(version.Status.Name).To(Equal("create-version-v1"))
			})
		})
	})
	Context("CreateVersion", func() {
		It("Should create a pipeline version", func() {
			pipeline := &kfpv1alpha1.Pipeline{}
			pipeline.SetName("create-version")
			it.Eventually().Create(pipeline).Should(Succeed())
			pipeline = &kfpv1alpha1.Pipeline{}
			it.Eventually().GetWhen(types.NamespacedName{Name: "create-version"}, pipeline, func(obj client.Object) bool {
				return len(pipeline.Status.ID) > 0
			}).Should(Succeed())

			version := &kfpv1alpha1.PipelineVersion{}
			version.SetName("create-version-v1")
			version.SetAnnotations(map[string]string{"kfp.jackhoman.com/pipeline-version": "1.0.1"})
			version.Spec.Pipeline = "create-version"
			version.Spec.Workflow.Raw = raw
			version.Spec.Description = "version 1.0.1"
			it.Eventually().Create(version).Should(Succeed())
			version = &kfpv1alpha1.PipelineVersion{}
			it.Eventually().GetWhen(types.NamespacedName{Name: "create-version-v1"}, version, func(obj client.Object) bool {
				return len(obj.(*kfpv1alpha1.PipelineVersion).Status.ID) > 0
			}).Should(Succeed())

			out, err := service.GetVersion(it.GetContext(), &kfp.GetVersionOptions{ID: version.Status.ID})
			Expect(err).ToNot(HaveOccurred())
			Expect(out).ToNot(BeNil())
			Expect(out.PipelineID).To(Equal(pipeline.Status.ID))
		})
		It("Should create a pipeline version with a specified name", func() {
			pipeline := &kfpv1alpha1.Pipeline{}
			pipeline.SetName("create-version")
			it.Eventually().Create(pipeline).Should(Succeed())
			pipeline = &kfpv1alpha1.Pipeline{}
			it.Eventually().GetWhen(types.NamespacedName{Name: "create-version"}, pipeline, func(obj client.Object) bool {
				return len(pipeline.Status.ID) > 0
			}).Should(Succeed())

			version := &kfpv1alpha1.PipelineVersion{}
			version.SetName("create-version-v1")
			version.SetAnnotations(map[string]string{"kfp.jackhoman.com/pipeline-version": "1.0.1"})
			version.Spec.Pipeline = "create-version"
			version.Spec.Workflow.Raw = raw
			version.Spec.Description = "version 1.0.1"
			it.Eventually().Create(version).Should(Succeed())
			version = &kfpv1alpha1.PipelineVersion{}
			it.Eventually().GetWhen(types.NamespacedName{Name: "create-version-v1"}, version, func(obj client.Object) bool {
				return len(obj.(*kfpv1alpha1.PipelineVersion).Status.ID) > 0
			}).Should(Succeed())

			out, err := service.GetVersion(it.GetContext(), &kfp.GetVersionOptions{
				Name:       "create-version-v1",
				PipelineID: pipeline.Status.ID,
			})
			Expect(err).ToNot(HaveOccurred())
			Expect(out).ToNot(BeNil())
			Expect(out.PipelineID).To(Equal(pipeline.Status.ID))
			Expect(out.ID).To(Equal(version.Status.ID))
		})
	})
})
