package kfp

import (
    "context"
    "github.com/go-openapi/runtime"
    ps "github.com/johnhoman/kfp-releaser/pkg/kfp/pipeline/client/pipeline_service"
    up "github.com/johnhoman/kfp-releaser/pkg/kfp/pipeline_upload/client/pipeline_upload_service"
    "time"
)

type PipelineService interface {
    CreatePipeline(params *ps.CreatePipelineParams, authInfo runtime.ClientAuthInfoWriter, opts ...ps.ClientOption) (*ps.CreatePipelineOK, error)
    CreatePipelineVersion(params *ps.CreatePipelineVersionParams, authInfo runtime.ClientAuthInfoWriter, opts ...ps.ClientOption) (*ps.CreatePipelineVersionOK, error)
    DeletePipeline(params *ps.DeletePipelineParams, authInfo runtime.ClientAuthInfoWriter, opts ...ps.ClientOption) (*ps.DeletePipelineOK, error)
    DeletePipelineVersion(params *ps.DeletePipelineVersionParams, authInfo runtime.ClientAuthInfoWriter, opts ...ps.ClientOption) (*ps.DeletePipelineVersionOK, error)
    GetPipeline(params *ps.GetPipelineParams, authInfo runtime.ClientAuthInfoWriter, opts ...ps.ClientOption) (*ps.GetPipelineOK, error)
    GetPipelineVersion(params *ps.GetPipelineVersionParams, authInfo runtime.ClientAuthInfoWriter, opts ...ps.ClientOption) (*ps.GetPipelineVersionOK, error)
    GetPipelineVersionTemplate(params *ps.GetPipelineVersionTemplateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ps.ClientOption) (*ps.GetPipelineVersionTemplateOK, error)

    UpdatePipelineDefaultVersion(params *ps.UpdatePipelineDefaultVersionParams, authInfo runtime.ClientAuthInfoWriter, opts ...ps.ClientOption) (*ps.UpdatePipelineDefaultVersionOK, error)


    // I don't think I actually need these
    // GetTemplate(params *ps.GetTemplateParams, authInfo runtime.ClientAuthInfoWriter, opts ...ps.ClientOption) (*ps.GetTemplateOK, error)
    // ListPipelineVersions(params *ps.ListPipelineVersionsParams, authInfo runtime.ClientAuthInfoWriter, opts ...ps.ClientOption) (*ps.ListPipelineVersionsOK, error)
    // ListPipelines(params *ps.ListPipelinesParams, authInfo runtime.ClientAuthInfoWriter, opts ...ps.ClientOption) (*ps.ListPipelinesOK, error)

}

type PipelineUploadService interface {
    UploadPipeline(params *up.UploadPipelineParams, authInfo runtime.ClientAuthInfoWriter, opts ...up.ClientOption) (*up.UploadPipelineOK, error)
    UploadPipelineVersion(params *up.UploadPipelineVersionParams, authInfo runtime.ClientAuthInfoWriter, opts ...up.ClientOption) (*up.UploadPipelineVersionOK, error)
}

type GetOptions struct {
    ID string
}
type CreateOptions struct {
    Name string
    Content string
    Description string
}
type UpdateOptions struct {
    ID string
    DefaultVersionID string
}
type DeleteOptions struct {
    ID string
}
type Pipeline struct {
    ID string
    Name string
    Description string
    CreatedAt time.Time
    DefaultVersionID string
    Parameters []struct{
        Name string
        Value string
    }
}


type Kubeflow interface {
    CreatePipeline(ctx context.Context, options *CreateOptions) (*Pipeline, error)
    GetPipeline(ctx context.Context, options *GetOptions) (*Pipeline, error)
    UpdatePipeline(ctx context.Context, options *UpdateOptions) (*Pipeline, error)
    DeletePipeline(ctx context.Context, options *DeleteOptions) error
}