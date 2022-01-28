package controllers

import (
	"encoding/json"
	"os"
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

		it.StartManager()

		raw, err = json.Marshal(workflow())
		Expect(err).ShouldNot(HaveOccurred())
	})
	AfterEach(func() { it.StopManager() })
})
