package controllers

import (
	"os"
	"strings"

	httptransport "github.com/go-openapi/runtime/client"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/johnhoman/controller-tools/manager"
	"github.com/johnhoman/go-kfp"
	kfpv1alpha1 "github.com/johnhoman/kfp-releaser/api/v1alpha1"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("PipelineController", func() {
	var it manager.IntegrationTest
	var service kfp.Interface

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
		err := (&PipelineReconciler{
			Client:        it.GetClient(),
			Scheme:        it.GetScheme(),
			Pipelines:     service,
			BlankWorkflow: workflow(),
			EventRecorder: it.GetEventRecorderFor("kfp-releaser.controller-test"),
		}).SetupWithManager(it)
		Expect(err).ToNot(HaveOccurred())

		it.StartManager()
	})
	AfterEach(func() { it.StopManager() })
	Context("Finalization", func() {
		It("Should add a finalizer", func() {
			instance := &kfpv1alpha1.Pipeline{}
			instance.SetName("un-finalized")
			it.Eventually().Create(instance).Should(Succeed())
			instance = &kfpv1alpha1.Pipeline{}
			it.Eventually().GetWhen(types.NamespacedName{Name: "un-finalized"}, instance, func(obj client.Object) bool {
				return len(obj.GetFinalizers()) == 1
			}).Should(Succeed())
			Expect(instance.Finalizers).Should(ContainElement(Finalizer))
			Expect(instance.ManagedFields[0].Manager).To(Equal(string(FieldOwner)))
		})
		It("Should remove the finalizer", func() {
			instance := &kfpv1alpha1.Pipeline{}
			instance.SetName("finalized")
			instance.SetFinalizers([]string{"keep"}) // keeps the object from being deleted
			it.Eventually().Create(instance).Should(Succeed())
			instance = &kfpv1alpha1.Pipeline{}
			it.Eventually().GetWhen(types.NamespacedName{Name: "finalized"}, instance, func(obj client.Object) bool {
				return len(obj.GetFinalizers()) == 2
			}).Should(Succeed())
			Expect(instance.Finalizers).Should(ContainElement(Finalizer))
			Expect(instance.ManagedFields[0].Manager).To(Equal(string(FieldOwner)))

			it.Expect().Delete(instance).Should(Succeed())
			instance = &kfpv1alpha1.Pipeline{}
			it.Eventually().GetWhen(types.NamespacedName{Name: "finalized"}, instance, func(obj client.Object) bool {
				return len(obj.GetFinalizers()) == 1
			}).Should(Succeed())
		})
		It("Should remove the upstream resource on deletion", func() {
			instance := &kfpv1alpha1.Pipeline{}
			instance.SetName("finalized")
			instance.SetFinalizers([]string{"keep"}) // keeps the object from being deleted
			it.Eventually().Create(instance).Should(Succeed())
			instance = &kfpv1alpha1.Pipeline{}
			it.Eventually().GetWhen(types.NamespacedName{Name: "finalized"}, instance, func(obj client.Object) bool {
				return len(obj.GetFinalizers()) == 2
			}).Should(Succeed())
			Expect(instance.Finalizers).Should(ContainElement(Finalizer))
			Expect(instance.ManagedFields[0].Manager).To(Equal(string(FieldOwner)))
			it.Eventually().GetWhen(types.NamespacedName{Name: "finalized"}, instance, func(obj client.Object) bool {
				return len(obj.(*kfpv1alpha1.Pipeline).Status.ID) > 0
			})

			it.Expect().Delete(instance).Should(Succeed())
			instance = &kfpv1alpha1.Pipeline{}
			it.Eventually().GetWhen(types.NamespacedName{Name: "finalized"}, instance, func(obj client.Object) bool {
				return len(obj.GetFinalizers()) == 1
			}).Should(Succeed())
			Eventually(func() error {
				_, err := service.Get(it.GetContext(), &kfp.GetOptions{ID: instance.Status.ID})
				return err
			}).Should(Equal(kfp.NewNotFound()))
		})
	})
	Context("CreatePipeline", func() {
		It("Should fill in the status fields", func() {
			instance := &kfpv1alpha1.Pipeline{}
			instance.SetName("create-a-pipeline")
			it.Eventually().Create(instance).Should(Succeed())
			instance = &kfpv1alpha1.Pipeline{}
			it.Eventually().GetWhen(types.NamespacedName{Name: "create-a-pipeline"}, instance, func(obj client.Object) bool {
				return len(obj.(*kfpv1alpha1.Pipeline).Status.ID) > 0
			}).Should(Succeed())
			_, err := service.Get(it.GetContext(), &kfp.GetOptions{ID: instance.Status.ID})
			Expect(err).ToNot(HaveOccurred())
		})
	})
})
