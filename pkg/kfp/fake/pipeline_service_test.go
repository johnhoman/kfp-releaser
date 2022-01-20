package fake_test

import (
	"context"
	"github.com/johnhoman/kfp-releaser/pkg/kfp/fake"
	"net/http"
	"strings"

	"github.com/go-openapi/runtime"
	"github.com/google/uuid"
	up "github.com/johnhoman/kfp-releaser/pkg/kfp/pipeline_upload/client/pipeline_upload_service"

	"github.com/johnhoman/kfp-releaser/pkg/kfp"
	ps "github.com/johnhoman/kfp-releaser/pkg/kfp/pipeline/client/pipeline_service"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func stringPtr(s string) *string {
	return &s
}

func int32Ptr(i int32) *int32 {
	return &i
}

func newCowSay(name string) runtime.NamedReadCloser {
	// Not sure if the name actually matters -- might be able to swap it for a uuid
	reader := runtime.NamedReader(name +".yaml", strings.NewReader(`
    apiVersion: argoproj.io/v1alpha1
    kind: Workflow
    metadata:
      name: whalesay
    spec:
      entrypoint: whalesay
      arguments:
        parameters:
        - {name: name, value: Jack}
      templates:
      - name: whalesay
        inputs:
          parameters:
          - {name: name}
        container:
          image: docker/whalesay
          command: [cowsay]
          args: ["Hello, {{inputs.parameters.name}}!"]
`))
	return reader
}

type UploadService = up.ClientService
type Service = ps.ClientService

type PipelineService struct {
	UploadService
	Service
}

var _ = Describe("PipelineService", func() {
	var service kfp.PipelineService
	var ctx context.Context
	var cancelFunc context.CancelFunc
	var pipelineIds []string
	var reader runtime.NamedReadCloser
	var name string
	var description string
	BeforeEach(func() {
		name = "testpipeline-" + uuid.New().String()[:8]
		description = strings.Join(strings.Split(name, "-"), " ")

		pipelineIds = make([]string, 0)

		service = fake.NewPipelineService()
		// transport := httptransport.New("localhost:8888", "", []string{"http"})
		// service = PipelineService{
		// 	UploadService: up.New(transport, strfmt.Default),
		// 	Service: ps.New(transport, strfmt.Default),
		// }
		reader = newCowSay(name)

		ctx, cancelFunc = context.WithCancel(context.Background())
	})
	AfterEach(func() {
		if pipelineIds != nil {
			for _, id := range pipelineIds {
				out, err := service.DeletePipeline(&ps.DeletePipelineParams{
					ID: id,
					Context: ctx,
				}, nil)
				if err != nil {
					Expect(err.(*ps.DeletePipelineDefault).Code()).To(Or(Equal(http.StatusOK), Equal(http.StatusNotFound)))
				}
				Expect(out).ShouldNot(BeNil())
			}
		}
		pipelineIds = nil
		cancelFunc()
	})
	It("should upload a pipeline", func() {
		out, err := service.UploadPipeline(&up.UploadPipelineParams{
			Description: stringPtr(description),
			Name:        stringPtr(name),
			Uploadfile:  reader,
			Context:     ctx,
		}, nil)
		Expect(err).ShouldNot(HaveOccurred())
		Expect(out).ShouldNot(BeNil())
		pipelineIds = append(pipelineIds, out.Payload.ID)
		// TODO: check the default version

		version, err := service.GetPipelineVersion(&ps.GetPipelineVersionParams{
			VersionID: out.GetPayload().ID,
			Context: ctx,
		}, nil)
		Expect(version.GetPayload().ID).To(Equal(out.GetPayload().ID))
		Expect(version.GetPayload().Name).To(Equal(out.GetPayload().Name))
		Expect(version.GetPayload().Description).To(Equal(""))
		Expect(version.GetPayload().CreatedAt).To(Equal(out.GetPayload().CreatedAt))
	})
	It("should return 404 for a pipeline that doesn't exist", func() {
		out, err := service.GetPipeline(&ps.GetPipelineParams{
			ID: uuid.New().String(),
			Context: ctx,
		}, nil)
		Expect(err).Should(HaveOccurred())
		Expect(out).To(BeNil())
		Expect(err.(*ps.GetPipelineDefault).Code()).To(Equal(http.StatusNotFound))
	})
	Context("PipelineVersion", func() {
		It("should upload a pipeline version", func() {
			out, err := service.UploadPipeline(&up.UploadPipelineParams{
				Description: stringPtr(description),
				Name:        stringPtr(name),
				Uploadfile:  reader,
				Context:     ctx,
			}, nil)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(out).ShouldNot(BeNil())
			// Expect(out.Payload.Parameters[0].Name).To(Equal("name"))
			// Expect(out.Payload.Parameters[0].Value).To(Equal("Jack"))
			pipelineIds = append(pipelineIds, out.Payload.ID)

			reader = newCowSay(name)
			vsOut, err := service.UploadPipelineVersion(&up.UploadPipelineVersionParams{
				Description: stringPtr(description),
				Name: stringPtr(name + "-v1"),
				Uploadfile: reader,
				Context: ctx,
				Pipelineid: stringPtr(out.Payload.ID),
			}, nil)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(vsOut).ShouldNot(BeNil())
			// I don't think I need this functionality
			// Expect(vsOut.Payload.Parameters[0].Name).To(Equal("name"))
			// Expect(vsOut.Payload.Parameters[0].Value).To(Equal("Jack"))
		})
		It("Should delete a pipeline version", func() {
			out, err := service.UploadPipeline(&up.UploadPipelineParams{
				Description: stringPtr(description),
				Name:        stringPtr(name),
				Uploadfile:  reader,
				Context:     ctx,
			}, nil)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(out).ShouldNot(BeNil())
			// Expect(out.Payload.Parameters[0].Name).To(Equal("name"))
			// Expect(out.Payload.Parameters[0].Value).To(Equal("Jack"))
			pipelineIds = append(pipelineIds, out.Payload.ID)

			reader = newCowSay(name)
			vsOut, err := service.UploadPipelineVersion(&up.UploadPipelineVersionParams{
				Description: stringPtr(description),
				Name: stringPtr(name + "-v1"),
				Uploadfile: reader,
				Context: ctx,
				Pipelineid: stringPtr(out.Payload.ID),
			}, nil)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(vsOut).ShouldNot(BeNil())
			Expect(vsOut.GetPayload().ID).ToNot(Equal(out.GetPayload().ID))

			delOut, err := service.DeletePipelineVersion(&ps.DeletePipelineVersionParams{
				VersionID: vsOut.GetPayload().ID,
				Context: ctx,
			}, nil)
			Expect(err).ToNot(HaveOccurred())
			Expect(delOut).ToNot(BeNil())
			Expect(*delOut).To(Equal(ps.DeletePipelineVersionOK{Payload: map[string]interface{}{}}))
		})
		It("Cannot delete a pipeline version that doesn't exist", func() {
			delOut, err := service.DeletePipelineVersion(&ps.DeletePipelineVersionParams{
				VersionID: uuid.New().String(),
				Context: ctx,
			}, nil)
			Expect(err).To(HaveOccurred())
			Expect(delOut).To(BeNil())
			out, ok := err.(*ps.DeletePipelineVersionDefault)
			Expect(ok).To(BeTrue())
			Expect(out.Code()).To(Equal(http.StatusNotFound))
		})
		It("Should get a pipeline version that does exist", func() {
			out, err := service.UploadPipeline(&up.UploadPipelineParams{
				Description: stringPtr(description),
				Name:        stringPtr(name),
				Uploadfile:  reader,
				Context:     ctx,
			}, nil)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(out).ShouldNot(BeNil())
			// Expect(out.Payload.Parameters[0].Name).To(Equal("name"))
			// Expect(out.Payload.Parameters[0].Value).To(Equal("Jack"))
			pipelineIds = append(pipelineIds, out.Payload.ID)

			reader = newCowSay(name)
			vsOut, err := service.UploadPipelineVersion(&up.UploadPipelineVersionParams{
				Description: stringPtr(description),
				Name: stringPtr(name + "-v1"),
				Uploadfile: reader,
				Context: ctx,
				Pipelineid: stringPtr(out.Payload.ID),
			}, nil)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(vsOut).ShouldNot(BeNil())
			Expect(vsOut.GetPayload().ID).ToNot(Equal(out.GetPayload().ID))

			getVersion, err := service.GetPipelineVersion(&ps.GetPipelineVersionParams{
				VersionID: vsOut.GetPayload().ID,
				Context: ctx,
			}, nil)
			Expect(err).ToNot(HaveOccurred())
			Expect(getVersion).ToNot(BeNil())
			Expect(getVersion.GetPayload().Name).To(Equal(vsOut.GetPayload().Name))
		})
		It("Cannot get a pipeline version that does not exist", func() {
			out, err := service.UploadPipeline(&up.UploadPipelineParams{
				Description: stringPtr(description),
				Name:        stringPtr(name),
				Uploadfile:  reader,
				Context:     ctx,
			}, nil)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(out).ShouldNot(BeNil())
			pipelineIds = append(pipelineIds, out.Payload.ID)

			getVersion, err := service.GetPipelineVersion(&ps.GetPipelineVersionParams{
				VersionID: uuid.New().String(),
				Context: ctx,
			}, nil)
			Expect(err).To(HaveOccurred())
			Expect(getVersion).To(BeNil())
			_, ok := err.(*ps.GetPipelineVersionDefault)
			Expect(ok).To(BeTrue())
			Expect(err.(*ps.GetPipelineVersionDefault).Code()).To(Equal(http.StatusNotFound))
		})
	})
	Context("UpdatePipelineDefaultVersion", func() {
		It("Should update the pipeline default version", func() {
			out, err := service.UploadPipeline(&up.UploadPipelineParams{
				Description: stringPtr(description),
				Name:        stringPtr(name),
				Uploadfile:  reader,
				Context:     ctx,
			}, nil)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(out).ShouldNot(BeNil())
			pipelineIds = append(pipelineIds, out.Payload.ID)

			reader = newCowSay(name)
			v1, err := service.UploadPipelineVersion(&up.UploadPipelineVersionParams{
				Description: stringPtr(description),
				Name: stringPtr(name + "-v1"),
				Uploadfile: reader,
				Context: ctx,
				Pipelineid: stringPtr(out.Payload.ID),
			}, nil)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(v1).ShouldNot(BeNil())
			Expect(v1.GetPayload().ID).ToNot(Equal(out.GetPayload().ID))

			reader = newCowSay(name)
			v2, err := service.UploadPipelineVersion(&up.UploadPipelineVersionParams{
				Description: stringPtr(description),
				Name: stringPtr(name + "-v2"),
				Uploadfile: reader,
				Context: ctx,
				Pipelineid: stringPtr(out.Payload.ID),
			}, nil)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(v2).ShouldNot(BeNil())
			Expect(v2.GetPayload().ID).ToNot(Equal(out.GetPayload().ID))

			getPipeline, err := service.GetPipeline(&ps.GetPipelineParams{
				ID: out.GetPayload().ID,
				Context: ctx,
			}, nil)
			// This is not guaranteed - this has to do with default server behaviour
			Expect(getPipeline.GetPayload().DefaultVersion.ID).To(Equal(v2.GetPayload().ID))

			updateOut, err := service.UpdatePipelineDefaultVersion(&ps.UpdatePipelineDefaultVersionParams{
				PipelineID: out.GetPayload().ID,
				VersionID: v1.GetPayload().ID,
				Context: ctx,
			}, nil)
			Expect(err).ToNot(HaveOccurred())
			Expect(updateOut).ToNot(BeNil())
			Expect(updateOut).To(Equal(&ps.UpdatePipelineDefaultVersionOK{Payload: map[string]interface{}{}}))

			getPipeline, err = service.GetPipeline(&ps.GetPipelineParams{
				ID: out.GetPayload().ID,
				Context: ctx,
			}, nil)
			Expect(err).ToNot(HaveOccurred())
			Expect(getPipeline).ToNot(BeNil())
			Expect(getPipeline.GetPayload().DefaultVersion.ID).To(Equal(v1.GetPayload().ID))
		})
		// Kubeflow bug
		It("Should fail silently for a version that does not exist", func() {
			out, err := service.UploadPipeline(&up.UploadPipelineParams{
				Description: stringPtr(description),
				Name:        stringPtr(name),
				Uploadfile:  reader,
				Context:     ctx,
			}, nil)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(out).ShouldNot(BeNil())
			pipelineIds = append(pipelineIds, out.Payload.ID)

			getOut, err := service.GetPipeline(&ps.GetPipelineParams{
				ID: out.GetPayload().ID,
				Context: ctx,
			}, nil)
			// Didn't update
			Expect(getOut.GetPayload().DefaultVersion.ID).To(Equal(out.GetPayload().ID))

			id := uuid.New().String()
			updateOut, err := service.UpdatePipelineDefaultVersion(&ps.UpdatePipelineDefaultVersionParams{
				PipelineID: out.GetPayload().ID,
				VersionID: id,
				Context: ctx,
			}, nil)
			Expect(err).ToNot(HaveOccurred())
			Expect(updateOut).ToNot(BeNil())
			Expect(updateOut).To(Equal(&ps.UpdatePipelineDefaultVersionOK{Payload: map[string]interface{}{}}))

			getOut, err = service.GetPipeline(&ps.GetPipelineParams{
				ID: out.GetPayload().ID,
				Context: ctx,
			}, nil)
			// Didn't update -- Kubeflow Bug
			Expect(getOut.GetPayload().DefaultVersion.ID).To(Equal(""))
		})
		// Kubeflow bug
		It("Silently fails to update a default version of a pipeline that does not exist", func() {
			out, err := service.UploadPipeline(&up.UploadPipelineParams{
				Description: stringPtr(description),
				Name:        stringPtr(name),
				Uploadfile:  reader,
				Context:     ctx,
			}, nil)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(out).ShouldNot(BeNil())
			pipelineIds = append(pipelineIds, out.Payload.ID)

			reader = newCowSay(name)
			v1, err := service.UploadPipelineVersion(&up.UploadPipelineVersionParams{
				Description: stringPtr(description),
				Name: stringPtr(name + "-v1"),
				Uploadfile: reader,
				Context: ctx,
				Pipelineid: stringPtr(out.Payload.ID),
			}, nil)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(v1).ShouldNot(BeNil())
			Expect(v1.GetPayload().ID).ToNot(Equal(out.GetPayload().ID))

			updateOut, err := service.UpdatePipelineDefaultVersion(&ps.UpdatePipelineDefaultVersionParams{
				PipelineID: uuid.New().String(),
				VersionID: v1.GetPayload().ID,
				Context: ctx,
			}, nil)
			Expect(err).ToNot(HaveOccurred())
			Expect(updateOut).ToNot(BeNil())

			getOut, err := service.GetPipeline(&ps.GetPipelineParams{
				ID: out.GetPayload().ID,
				Context: ctx,
			}, nil)
			// Didn't update
			Expect(getOut.GetPayload().DefaultVersion.ID).To(Equal(v1.GetPayload().ID))
		})
	})
	Context("CreatePipeline", func() {
		It("Should return useless error when pipeline exists", func() {
			out, err := service.UploadPipeline(&up.UploadPipelineParams{
				Description: stringPtr(description),
				Name:        stringPtr(name),
				Uploadfile:  reader,
				Context:     ctx,
			}, nil)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(out).ShouldNot(BeNil())
			pipelineIds = append(pipelineIds, out.Payload.ID)

			reader = newCowSay(name)
			out, err = service.UploadPipeline(&up.UploadPipelineParams{
				Description: stringPtr(description),
				Name:        stringPtr(name),
				Uploadfile:  reader,
				Context:     ctx,
			}, nil)
			Expect(err).Should(HaveOccurred())
			Expect(out).Should(BeNil())
			Expect(err.Error()).To(ContainSubstring("is not supported by the TextConsumer"))
		})
	})
	// Internal server errors are common errors returned by Kubeflow
	Context("DeletePipeline", func() {
		It("should delete a pipeline", func() {
			out, err := service.UploadPipeline(&up.UploadPipelineParams{
				Description: stringPtr(description),
				Name:        stringPtr(name),
				Uploadfile:  reader,
				Context:     ctx,
			}, nil)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(out).ShouldNot(BeNil())

			delOut, err := service.DeletePipeline(&ps.DeletePipelineParams{
				Context: ctx,
				ID: out.Payload.ID,
			}, nil)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(*delOut).To(Equal(ps.DeletePipelineOK{Payload: map[string]interface{}{}}))
		})
		It("Should return 404 when deleting a pipeline that doesn't exist", func() {
			_, err := service.DeletePipeline(&ps.DeletePipelineParams{
				Context: ctx,
				ID: uuid.New().String(),
			}, nil)
			Expect(err).Should(HaveOccurred())
			Expect(err.(*ps.DeletePipelineDefault).Code()).To(Equal(http.StatusNotFound))
		})

	})
	// This list implementation will be somewhat complicated, and I don't think I need it
	/*
	It("Should list pipelines", func() {
		Expect(reader.Close()).To(Succeed())
		for k := 0; k < 5; k++ {
			reader = newCowSay(name)
			out, err := service.UploadPipeline(&up.UploadPipelineParams{
				Description: stringPtr(description),
				Name:        stringPtr(fmt.Sprintf("%s-%d", name, k)),
				Uploadfile:  reader,
				Context:     ctx,
			}, nil)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(out).ShouldNot(BeNil())
			pipelineIds = append(pipelineIds, out.Payload.ID)
		}
		filter := map[string]interface{}{
			"predicates": []interface{}{
				map[string]interface{}{"op": "IS_SUBSTRING", "key": "name", "string_value": name},
			},
		}
		raw, err := json.Marshal(filter)
		Expect(err).To(Succeed())

		pipelines := make([]*models.APIPipeline, 0, 5)

		listOut, err := pipelineService.ListPipelines(&ps.ListPipelinesParams{
			Filter: stringPtr(string(raw)),
			PageSize: int32Ptr(3),
			PageToken: nil,
			Context: ctx,
			ResourceReferenceKeyType: stringPtr(string(models.APIResourceTypeNAMESPACE)),
		}, nil)
		// eyJTb3J0QnlGaWVsZE5hbWUiOiJDcmVhdGVkQXRJblNlYyIsIlNvcnRCeUZpZWxkVmFsdWUiOjE2NDI0NzU2MzQsIlNvcnRCeUZpZWxkUHJlZml4IjoicGlwZWxpbmVzLiIsIktleUZpZWxkTmFtZSI6IlVVSUQiLCJLZXlGaWVsZFZhbHVlIjoiYjUxMDU0YjctYzc4OS00YTQ5LWJiNGQtODZkOTEwMzQwMDZhIiwiS2V5RmllbGRQcmVmaXgiOiJwaXBlbGluZXMuIiwiSXNEZXNjIjpmYWxzZSwiTW9kZWxOYW1lIjoicGlwZWxpbmVzIiwiRmlsdGVyIjp7IkZpbHRlclByb3RvIjoie1wicHJlZGljYXRlc1wiOlt7XCJvcFwiOlwiSVNfU1VCU1RSSU5HXCIsXCJrZXlcIjpcInBpcGVsaW5lcy5OYW1lXCIsXCJzdHJpbmdWYWx1ZVwiOlwidGVzdHBpcGVsaW5lLTA2MmYzY2FjXCJ9XX0iLCJFUSI6e30sIk5FUSI6e30sIkdUIjp7fSwiR1RFIjp7fSwiTFQiOnt9LCJMVEUiOnt9LCJJTiI6e30sIlNVQlNUUklORyI6eyJwaXBlbGluZXMuTmFtZSI6WyJ0ZXN0cGlwZWxpbmUtMDYyZjNjYWMiXX19fQ==
		Expect(err).ToNot(HaveOccurred())
		Expect(listOut.Payload.Pipelines).To(HaveLen(3))
		Expect(listOut.Payload.TotalSize).To(Equal(int32(5)))
		pipelines = append(pipelines, listOut.Payload.Pipelines...)
		listOut, err = pipelineService.ListPipelines(&ps.ListPipelinesParams{
			Filter: stringPtr(string(raw)),
			PageSize: int32Ptr(2),
			PageToken: &listOut.Payload.NextPageToken,
			ResourceReferenceKeyType: stringPtr(string(models.APIResourceTypeNAMESPACE)),
			Context: ctx,
		}, nil)
		Expect(err).ToNot(HaveOccurred())
		Expect(listOut.Payload.Pipelines).To(HaveLen(2))
		Expect(listOut.Payload.TotalSize).To(Equal(int32(5)))
		pipelines = append(pipelines, listOut.Payload.Pipelines...)

		for _, k := range []string{"-0", "-1", "-2", "-3", "-4"} {
			found := false
			for _, pl := range pipelines {
				if strings.HasSuffix(pl.Name, k) {
					found = true
				}
			}
			Expect(found).To(BeTrue())
		}
	})
	table.DescribeTable("", func(filter map[string]interface{}) {
		raw, err := json.Marshal(filter)
		Expect(err).ToNot(HaveOccurred())
		_, err = pipelineService.ListPipelines(&ps.ListPipelinesParams{
			Filter: stringPtr(string(raw)),
			ResourceReferenceKeyType: stringPtr(string(models.APIResourceTypeNAMESPACE)),
			Context: ctx,
		}, nil)
		Expect(err).To(HaveOccurred())
		Expect(err.(*ps.ListPipelinesDefault).Code()).To(Equal(http.StatusBadRequest))
	}, table.Entry("", map[string]interface{}{
		"predicate": map[string]interface{}{"op": "IS_SUBSTRING", "key": "name", "string_value": name},
	}), table.Entry("", map[string]interface{}{
		"predicates": map[string]interface{}{"op": "IS_SUBSTRING", "key": "name", "string_value": name},
	}))
	 */
	/*
		DescribeTable("create should fail", func(body *models.APIPipeline, messagePrefix string, code int) {
			params := &ps.CreatePipelineParams{Body: body, Context: ctx}
			_, err := service.CreatePipeline(params, nil)
			Expect(err).Should(HaveOccurred())
			Expect(err.(*ps.CreatePipelineDefault).Code()).To(Equal(code))
			Expect(err.(*ps.CreatePipelineDefault).Payload.Error).To(HavePrefix(messagePrefix))
		},
			Entry("no pipeline url", &models.APIPipeline{
				Description:    "[Tutorial] DSL - Control structures",
				Name:           "[Tutorial] DSL - Control structures",
				DefaultVersion: &models.APIPipelineVersion{},
			}, "Invalid input error: Pipeline URL is empty", http.StatusBadRequest),
		)
	*/
})