package fake

import (
    "github.com/go-openapi/runtime"
    "github.com/johnhoman/kfp-releaser/pkg/kfp/http/models"
    "sync"

    "github.com/johnhoman/kfp-releaser/pkg/kfp"
    ps "github.com/johnhoman/kfp-releaser/pkg/kfp/http/client/pipeline_service"
)

type PipelineService struct {
    sync.Mutex
    // map[models.APIPipeline.Name]models.APIPipeline
    Pipelines map[string]models.APIPipeline
    PipelineVersions map[string]models.APIPipelineVersion
}

func (p* PipelineService) CreatePipeline(params *ps.CreatePipelineParams, authInfo runtime.ClientAuthInfoWriter, opts ...ps.ClientOption) (*ps.CreatePipelineOK, error) {
    _, ok := p.Pipelines[params.Body.Name]
    if ok {
        // not sure what to do here
        resp := ps.CreatePipelineDefault{}
        return nil, runtime.NewAPIError(
            "unexpected success response: content available as default response in error",
            resp,
            resp.Code())
    }
    return &ps.CreatePipelineOK{Payload: params.Body}, nil
}

func (p* PipelineService) CreatePipelineVersion(params *ps.CreatePipelineVersionParams, authInfo runtime.ClientAuthInfoWriter, opts ...ps.ClientOption) (*ps.CreatePipelineVersionOK, error) {
    panic("implement me")
}

func (p* PipelineService) DeletePipeline(params *ps.DeletePipelineParams, authInfo runtime.ClientAuthInfoWriter, opts ...ps.ClientOption) (*ps.DeletePipelineOK, error) {
    panic("implement me")
}

func (p* PipelineService) DeletePipelineVersion(params *ps.DeletePipelineVersionParams, authInfo runtime.ClientAuthInfoWriter, opts ...ps.ClientOption) (*ps.DeletePipelineVersionOK, error) {
    panic("implement me")
}

func (p* PipelineService) GetPipeline(params *ps.GetPipelineParams, authInfo runtime.ClientAuthInfoWriter, opts ...ps.ClientOption) (*ps.GetPipelineOK, error) {
    panic("implement me")
}

func (p* PipelineService) GetPipelineVersion(params *ps.GetPipelineVersionParams, authInfo runtime.ClientAuthInfoWriter, opts ...ps.ClientOption) (*ps.GetPipelineVersionOK, error) {
    panic("implement me")
}

func (p* PipelineService) GetPipelineVersionTemplate(params *ps.GetPipelineVersionTemplateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ps.ClientOption) (*ps.GetPipelineVersionTemplateOK, error) {
    panic("implement me")
}

func (p* PipelineService) GetTemplate(params *ps.GetTemplateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ps.ClientOption) (*ps.GetTemplateOK, error) {
    panic("implement me")
}

func (p* PipelineService) ListPipelineVersions(params *ps.ListPipelineVersionsParams, authInfo runtime.ClientAuthInfoWriter, opts ...ps.ClientOption) (*ps.ListPipelineVersionsOK, error) {
    panic("implement me")
}

func (p* PipelineService) ListPipelines(params *ps.ListPipelinesParams, authInfo runtime.ClientAuthInfoWriter, opts ...ps.ClientOption) (*ps.ListPipelinesOK, error) {
    panic("implement me")
}

func (p* PipelineService) UpdatePipelineDefaultVersion(params *ps.UpdatePipelineDefaultVersionParams, authInfo runtime.ClientAuthInfoWriter, opts ...ps.ClientOption) (*ps.UpdatePipelineDefaultVersionOK, error) {
    panic("implement me")
}

var _ kfp.PipelineService = &PipelineService{}