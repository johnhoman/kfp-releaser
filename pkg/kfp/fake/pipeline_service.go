package fake

import (
	"fmt"
	"github.com/go-openapi/runtime"
	"github.com/go-openapi/strfmt"
	"github.com/google/uuid"
	"net/http"
	"sync"
	"time"

	"github.com/johnhoman/kfp-releaser/pkg/kfp"
	ps "github.com/johnhoman/kfp-releaser/pkg/kfp/pipeline/client/pipeline_service"
	"github.com/johnhoman/kfp-releaser/pkg/kfp/pipeline/models"
	up "github.com/johnhoman/kfp-releaser/pkg/kfp/pipeline_upload/client/pipeline_upload_service"
	upmodels "github.com/johnhoman/kfp-releaser/pkg/kfp/pipeline_upload/models"
)


type PipelineService struct {
	sync.Mutex
	// map[models.APIPipeline.Name]models.APIPipeline
	Pipelines        map[string]models.APIPipeline
	PipelineVersions map[string][]models.APIPipelineVersion
}

var internalServerError = fmt.Errorf("\"&{0 [] }\" (*models.APIStatus) is not supported by the TextConsumer, %s",
	"can be resolved by supporting TextUnmarshaler interface")

func (p *PipelineService) UploadPipeline(params *up.UploadPipelineParams, authInfo runtime.ClientAuthInfoWriter, opts ...up.ClientOption) (*up.UploadPipelineOK, error) {
	if _, found := p.Pipelines[*params.Name]; found {
		// Kubeflow doesn't return a response body for this, so the swagger generated client
		// isn't able to create the APIStatus response, so for now just mimic that behaviour
		// E0117 16:21:25.254145       7 pipeline_upload_server.go:241] Failed to upload pipelines. Error: Error creating pipeline: Create pipeline failed: Already exist error: Failed to create a new pipeline. The name production-cowsay already exist. Please specify a new name.
		//
		// Unexpected error:
		//   <*errors.errorString | 0xc0004180c0>: {
		//   s: "&{0 [] } (*models.APIStatus) is not supported by the TextConsumer, can be resolved by supporting TextUnmarshaler interface",
		//   }
		// &{0 [] } (*models.APIStatus) is not supported by the TextConsumer, can be resolved by supporting TextUnmarshaler interface

		// Always return internal server error. Many errors aren't handled by api server
		return nil, internalServerError
	}
	uid := uuid.New().String()
	m := models.APIPipeline{
		CreatedAt: strfmt.DateTime(time.Now().UTC()),
		DefaultVersion: &models.APIPipelineVersion{
			CreatedAt: strfmt.DateTime(time.Now().UTC()),
			Description: "",
		},
		Description: *params.Description,
		Error: "",
		ID: uid,
		Name: *params.Name,
		Parameters: nil,
		ResourceReferences: nil,
		URL: nil,
	}
	p.Mutex.Lock()
	defer p.Mutex.Unlock()
	p.Pipelines[m.Name] = m
	// Kubeflow api tracks upload api models and pipeline api models as separate objects in the swagger specs,
	// so we have to duplicate it for the response here
	payload := &upmodels.APIPipeline{}
	payload.CreatedAt = m.CreatedAt
	payload.Description = m.Description
	payload.Error = m.Error
	payload.ID = m.ID
	payload.Name = m.Name
	payload.Parameters = nil
	if m.Parameters != nil {
		parameters := make([]*upmodels.APIParameter, len(m.Parameters))
		for _, p := range m.Parameters {
			parameters = append(parameters, &upmodels.APIParameter{
				Name: p.Name,
				Value: p.Value,
			})
		}
		payload.Parameters = parameters
	}
	return &up.UploadPipelineOK{Payload: payload}, nil
}

func (p *PipelineService) UploadPipelineVersion(params *up.UploadPipelineVersionParams, authInfo runtime.ClientAuthInfoWriter, opts ...up.ClientOption) (*up.UploadPipelineVersionOK, error) {
	id := *params.Pipelineid
	name := *params.Name

	versions, ok := p.PipelineVersions[id]
	if !ok {
		return nil, internalServerError
	}
	found := false
	for _, v := range versions {
		if v.Name == name {
			found = true
		}
	}
	if found {
		return nil, internalServerError
	}
	uid := uuid.New().String()
	description := *params.Description

	m := models.APIPipelineVersion{}
	m.Name = name
	m.ID = uid
	m.CreatedAt = strfmt.DateTime(time.Now().UTC())
	m.Parameters = nil
	pipelineType := models.APIResourceTypePIPELINE
	owner := models.APIRelationshipOWNER
	m.ResourceReferences = []*models.APIResourceReference{{
		Key: &models.APIResourceKey{ID: id, Type: &pipelineType},
		Name: "",
		Relationship: &owner,
	}}
	m.Description = description

	p.Mutex.Lock()
	defer p.Mutex.Unlock()
	p.PipelineVersions[id] = append(p.PipelineVersions[id], m)

	out := &upmodels.APIPipelineVersion{}
	out.CreatedAt = m.CreatedAt
	out.ID = m.ID
	out.Name = m.Name
	out.PackageURL = nil
	out.Parameters = nil
	out.ResourceReferences = nil
	if m.ResourceReferences != nil {
		out.ResourceReferences = make([]*upmodels.APIResourceReference, 0, len(m.ResourceReferences))
		for _, ref := range m.ResourceReferences {
			out.ResourceReferences = append(out.ResourceReferences, &upmodels.APIResourceReference{
				Key: &upmodels.APIResourceKey{
					ID: ref.Key.ID,
					Type: upmodels.NewAPIResourceType(upmodels.APIResourceType(*ref.Key.Type)),
				},
				Name: ref.Name,
				Relationship: upmodels.NewAPIRelationship(upmodels.APIRelationship(*ref.Relationship)),
			})
		}
	}
	return &up.UploadPipelineVersionOK{Payload: out}, nil
}

func (p *PipelineService) CreatePipeline(params *ps.CreatePipelineParams, authInfo runtime.ClientAuthInfoWriter, opts ...ps.ClientOption) (*ps.CreatePipelineOK, error) {
	panic("implement me")
}

func (p *PipelineService) CreatePipelineVersion(params *ps.CreatePipelineVersionParams, authInfo runtime.ClientAuthInfoWriter, opts ...ps.ClientOption) (*ps.CreatePipelineVersionOK, error) {
	panic("implement me")
}

func (p *PipelineService) DeletePipeline(params *ps.DeletePipelineParams, authInfo runtime.ClientAuthInfoWriter, opts ...ps.ClientOption) (*ps.DeletePipelineOK, error) {

	p.Mutex.Lock()
	defer p.Mutex.Unlock()
	pipelines := make(map[string]models.APIPipeline, len(p.Pipelines))
	found := false
	for key, value := range p.Pipelines {
		if value.ID == params.ID {
			found = true
		} else {
			pipelines[key] = value
		}
	}
	if found {
		p.Pipelines = pipelines
		return &ps.DeletePipelineOK{}, nil
	}
	return nil, ps.NewDeletePipelineDefault(http.StatusNotFound)
}

func (p *PipelineService) DeletePipelineVersion(params *ps.DeletePipelineVersionParams, authInfo runtime.ClientAuthInfoWriter, opts ...ps.ClientOption) (*ps.DeletePipelineVersionOK, error) {
	panic("implement me")
}

func (p *PipelineService) GetPipeline(params *ps.GetPipelineParams, authInfo runtime.ClientAuthInfoWriter, opts ...ps.ClientOption) (*ps.GetPipelineOK, error) {
	panic("implement me")
}

func (p *PipelineService) GetPipelineVersion(params *ps.GetPipelineVersionParams, authInfo runtime.ClientAuthInfoWriter, opts ...ps.ClientOption) (*ps.GetPipelineVersionOK, error) {
	panic("implement me")
}

func (p *PipelineService) GetPipelineVersionTemplate(params *ps.GetPipelineVersionTemplateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ps.ClientOption) (*ps.GetPipelineVersionTemplateOK, error) {
	panic("implement me")
}

func (p *PipelineService) GetTemplate(params *ps.GetTemplateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ps.ClientOption) (*ps.GetTemplateOK, error) {
	panic("implement me")
}

func (p *PipelineService) ListPipelineVersions(params *ps.ListPipelineVersionsParams, authInfo runtime.ClientAuthInfoWriter, opts ...ps.ClientOption) (*ps.ListPipelineVersionsOK, error) {
	panic("implement me")
}

func (p *PipelineService) ListPipelines(params *ps.ListPipelinesParams, authInfo runtime.ClientAuthInfoWriter, opts ...ps.ClientOption) (*ps.ListPipelinesOK, error) {
	panic("implement me")
}

func (p *PipelineService) UpdatePipelineDefaultVersion(params *ps.UpdatePipelineDefaultVersionParams, authInfo runtime.ClientAuthInfoWriter, opts ...ps.ClientOption) (*ps.UpdatePipelineDefaultVersionOK, error) {
	panic("implement me")
}

var _ kfp.PipelineService = &PipelineService{}
var _ kfp.PipelineUploadService = &PipelineService{}

func NewPipelineService() *PipelineService {
	return &PipelineService{
		Pipelines:        make(map[string]models.APIPipeline),
		PipelineVersions: make(map[string][]models.APIPipelineVersion),
	}
}
