package pipelines_test

import (
	"context"
	"github.com/google/uuid"
	"github.com/johnhoman/kfp-releaser/pkg/kfp/pipelines"
	"strings"
	"time"

	"github.com/johnhoman/kfp-releaser/pkg/kfp/fake"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func newWhaleSay() map[string]interface{} {
	// Not sure if the name actually matters -- might be able to swap it for a uuid
	content := map[string]interface{}{
		"apiVersion": "argoproj.io/v1alpha1",
		"kind": "Workflow",
		"metadata": map[string]interface{}{
			"name": "whalesay",
		},
		"spec": map[string]interface{}{
			"entrypoint": "whalesay",
			"arguments": map[string]interface{}{
				"parameters": []interface{}{
					map[string]interface{}{
						"name": "name",
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
						"image": "docker/whalesay",
						"command": []string{"cowsay"},
						"args": []string{"Hello", "{{inputs.parameters.name}}"},
					},
				},
			},
		},
	}
	return content
}
var _ = Describe("PipelinesApi", func() {
	var api pipelines.Interface
	var pipeline *pipelines.Pipeline
	var ctx context.Context
	var cancelFunc context.CancelFunc
	var name string
	var description string
	BeforeEach(func() {
		name = "testcase-" + uuid.New().String()[:8]
		description = strings.ToTitle(strings.Join(strings.Split(name, "-"), " "))
		api = pipelines.New(fake.NewPipelineService(), nil)
		ctx, cancelFunc = context.WithCancel(context.Background())
	})
	AfterEach(func() {
		Expect(api.Delete(ctx, &pipelines.DeleteOptions{ID: pipeline.ID})).To(Or(
			Succeed(),
			Equal(pipelines.NewNotFound()),
		))
		cancelFunc()
	})
	Context("CreatePipeline", func() {
		It("Can create a pipeline", func() {
			var err error
			pipeline, err = api.Create(ctx, &pipelines.CreateOptions{
				Name:        name,
				Workflow:    newWhaleSay(),
				Description: description,
			})
			Expect(err).ToNot(HaveOccurred())
			Expect(pipeline).ToNot(BeNil())
			Expect(pipeline.Description).To(Equal(description))
			Expect(pipeline.Name).To(Equal(name))
			Expect(pipeline.ID).ToNot(Equal(""))
			Expect(pipeline.DefaultVersionID).To(Equal(pipeline.ID))
		})
		It("Should return 409 conflict when a pipeline doesn't exist", func() {
			var err error
			pipeline, err = api.Create(ctx, &pipelines.CreateOptions{
				Name:        name,
				Workflow:    newWhaleSay(),
				Description: description,
			})
			Expect(err).ToNot(HaveOccurred())
			out, err := api.Create(ctx, &pipelines.CreateOptions{
				Name:        name,
				Workflow:    newWhaleSay(),
				Description: description,
			})
			Expect(err).Should(HaveOccurred())
			Expect(pipelines.IsConflict(err)).To(BeTrue())
			Expect(out).To(Equal(&pipelines.Pipeline{}))
		})
	})
	Context("DeletePipeline", func() {
		It("Should remove a pipeline", func() {
			var err error
			pipeline, err = api.Create(ctx, &pipelines.CreateOptions{
				Name:        name,
				Workflow:    newWhaleSay(),
				Description: description,
			})
			Expect(err).To(Succeed())
			Expect(api.Delete(ctx, &pipelines.DeleteOptions{ID: pipeline.ID})).To(Succeed())
		})
		It("Should return 404 when the pipeline doesn't exist", func() {
			err := api.Delete(ctx, &pipelines.DeleteOptions{ID: uuid.New().String()})
			Expect(err).To(HaveOccurred())
			Expect(err).To(Equal(pipelines.NewNotFound()))
		})
	})
	Context("UpdatePipeline", func() {
		It("Should not change the default version of a pipeline that doesn't exist", func(){
			// TODO:
		})
		It("Should not change the default version of a pipeline to a version that doesn't exist", func() {
			var err error
			pipeline, err = api.Create(ctx, &pipelines.CreateOptions{
				Name:        name,
				Workflow:    newWhaleSay(),
				Description: description,
			})
			Expect(err).To(Succeed())
			Expect(pipeline).ToNot(BeNil())

			pipeline, err = api.Update(ctx, &pipelines.UpdateOptions{
				ID: pipeline.ID,
				DefaultVersionID: uuid.New().String(),
			})
			Expect(pipeline).To(Equal(&pipelines.Pipeline{}))
			Expect(err).To(HaveOccurred())
			Expect(err).To(Equal(pipelines.NewNotFound()))
		})
	})
	Context("CreateVersion", func() {
		It("Should create a new version", func(){
			var err error
			pipeline, err = api.Create(ctx, &pipelines.CreateOptions{
				Name:        name,
				Workflow:    newWhaleSay(),
				Description: description,
			})
			Expect(err).To(Succeed())
			Expect(pipeline).ToNot(BeNil())
			version, err := api.CreateVersion(ctx, &pipelines.CreateVersionOptions{
				PipelineID: pipeline.ID,
				Name: name + "-1",
				Description: description,
				Workflow: newWhaleSay(),
			})
			Expect(err).To(Succeed())
			Expect(version.Name).To(Equal(name + "-1"))
		})
		It("Should return not found if the pipeline doesn't exist", func(){
			_, err := api.CreateVersion(ctx, &pipelines.CreateVersionOptions{
				PipelineID: uuid.New().String(),
				Name: name + "-1",
				Description: description,
				Workflow: newWhaleSay(),
			})
			Expect(err).Should(HaveOccurred())
			Expect(pipelines.IsNotFound(err)).To(BeTrue())
		})
		It("Should return 409 if version name exists", func(){
			var err error
			pipeline, err = api.Create(ctx, &pipelines.CreateOptions{
				Name:        name,
				Workflow:    newWhaleSay(),
				Description: description,
			})
			Expect(err).To(Succeed())
			Expect(pipeline).ToNot(BeNil())
			_, err = api.CreateVersion(ctx, &pipelines.CreateVersionOptions{
				PipelineID: pipeline.ID,
				Name: name,
				Description: description,
				Workflow: newWhaleSay(),
			})
			Expect(err).Should(HaveOccurred())
			Expect(pipelines.IsConflict(err)).To(BeTrue())
		})
	})
	FContext("GetVersion", func() {
		It("Should get the version info", func() {
			var err error
			pipeline, err = api.Create(ctx, &pipelines.CreateOptions{
				Name:        name,
				Workflow:    newWhaleSay(),
				Description: description,
			})
			Expect(err).To(Succeed())
			Expect(pipeline).ToNot(BeNil())

			version, err := api.GetVersion(ctx, &pipelines.GetOptions{ID: pipeline.ID})
			Expect(err).To(Succeed())
			Expect(version.PipelineID).To(Equal(pipeline.ID))
			Expect(version.Name).To(Equal(pipeline.Name))
			Expect(version.ID).To(Equal(pipeline.ID))
			Expect(time.Now().UTC().Sub(version.CreatedAt)).To(BeNumerically("~", 0, time.Second))
		})
	})
	Context("DeleteVersion", func() {
		It("Should delete a version", func() {
			var err error
			pipeline, err = api.Create(ctx, &pipelines.CreateOptions{
				Name:        name,
				Workflow:    newWhaleSay(),
				Description: description,
			})
			Expect(err).To(Succeed())
			Expect(pipeline).ToNot(BeNil())
			version, err := api.CreateVersion(ctx, &pipelines.CreateVersionOptions{
				PipelineID: pipeline.ID,
				Name: name,
				Description: description,
				Workflow: newWhaleSay(),
			})
			Expect(err).Should(HaveOccurred())
			Expect(pipelines.IsConflict(err)).To(BeTrue())

			Expect(api.DeleteVersion(ctx, &pipelines.DeleteOptions{ID: version.ID})).Should(Succeed())
		})
		It("Should return a 404 for a version that doesn't exist", func() {
			err := api.DeleteVersion(ctx, &pipelines.DeleteOptions{ID: uuid.New().String()})
			Expect(err).Should(HaveOccurred())
			Expect(pipelines.IsNotFound(err)).To(BeTrue())
		})
	})
})
