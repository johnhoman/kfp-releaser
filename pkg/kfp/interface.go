package kfp

import (
    "github.com/go-openapi/runtime"
    ps "github.com/johnhoman/kfp-releaser/pkg/kfp/http/client/pipeline_service"
)

type PipelineService interface {
    CreatePipeline(params *ps.CreatePipelineParams, authInfo runtime.ClientAuthInfoWriter, opts ...ps.ClientOption) (*ps.CreatePipelineOK, error)
    CreatePipelineVersion(params *ps.CreatePipelineVersionParams, authInfo runtime.ClientAuthInfoWriter, opts ...ps.ClientOption) (*ps.CreatePipelineVersionOK, error)
    DeletePipeline(params *ps.DeletePipelineParams, authInfo runtime.ClientAuthInfoWriter, opts ...ps.ClientOption) (*ps.DeletePipelineOK, error)
    DeletePipelineVersion(params *ps.DeletePipelineVersionParams, authInfo runtime.ClientAuthInfoWriter, opts ...ps.ClientOption) (*ps.DeletePipelineVersionOK, error)
    GetPipeline(params *ps.GetPipelineParams, authInfo runtime.ClientAuthInfoWriter, opts ...ps.ClientOption) (*ps.GetPipelineOK, error)
    GetPipelineVersion(params *ps.GetPipelineVersionParams, authInfo runtime.ClientAuthInfoWriter, opts ...ps.ClientOption) (*ps.GetPipelineVersionOK, error)
    GetPipelineVersionTemplate(params *ps.GetPipelineVersionTemplateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ps.ClientOption) (*ps.GetPipelineVersionTemplateOK, error)
    GetTemplate(params *ps.GetTemplateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ps.ClientOption) (*ps.GetTemplateOK, error)
    ListPipelineVersions(params *ps.ListPipelineVersionsParams, authInfo runtime.ClientAuthInfoWriter, opts ...ps.ClientOption) (*ps.ListPipelineVersionsOK, error)
    ListPipelines(params *ps.ListPipelinesParams, authInfo runtime.ClientAuthInfoWriter, opts ...ps.ClientOption) (*ps.ListPipelinesOK, error)
    UpdatePipelineDefaultVersion(params *ps.UpdatePipelineDefaultVersionParams, authInfo runtime.ClientAuthInfoWriter, opts ...ps.ClientOption) (*ps.UpdatePipelineDefaultVersionOK, error)
}