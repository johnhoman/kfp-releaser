package controllers

import (
	kfpv1alpha1 "github.com/johnhoman/kfp-releaser/api/v1alpha1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/johnhoman/controller-tools/manager"
	"github.com/johnhoman/go-kfp"
	"github.com/johnhoman/go-kfp/fake"
)

func workflow() map[string]interface{} {
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

var _ = Describe("PipelineController", func() {
	var it manager.IntegrationTest
	var service kfp.Pipelines

	BeforeEach(func() {
		service = kfp.New(fake.NewPipelineService(), nil)
		it = manager.IntegrationTestBuilder().
			WithScheme(scheme.Scheme).
			Complete(cfg)
		err := (&PipelineReconciler{
			Client:        it.GetClient(),
			Scheme:        it.GetScheme(),
			Pipelines:     service,
			BlankWorkflow: workflow(),
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
