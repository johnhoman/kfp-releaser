package pipelines

import (
    "context"
    "github.com/go-openapi/runtime"
    ps "github.com/johnhoman/kfp-releaser/pkg/kfp/pipeline/client/pipeline_service"
    up "github.com/johnhoman/kfp-releaser/pkg/kfp/pipeline_upload/client/pipeline_upload_service"
)

type PipelineService interface {
    DeletePipeline(params *ps.DeletePipelineParams, authInfo runtime.ClientAuthInfoWriter, opts ...ps.ClientOption) (*ps.DeletePipelineOK, error)
    DeletePipelineVersion(params *ps.DeletePipelineVersionParams, authInfo runtime.ClientAuthInfoWriter, opts ...ps.ClientOption) (*ps.DeletePipelineVersionOK, error)
    GetPipeline(params *ps.GetPipelineParams, authInfo runtime.ClientAuthInfoWriter, opts ...ps.ClientOption) (*ps.GetPipelineOK, error)
    GetPipelineVersion(params *ps.GetPipelineVersionParams, authInfo runtime.ClientAuthInfoWriter, opts ...ps.ClientOption) (*ps.GetPipelineVersionOK, error)
    UpdatePipelineDefaultVersion(params *ps.UpdatePipelineDefaultVersionParams, authInfo runtime.ClientAuthInfoWriter, opts ...ps.ClientOption) (*ps.UpdatePipelineDefaultVersionOK, error)
    ListPipelines(params *ps.ListPipelinesParams, authInfo runtime.ClientAuthInfoWriter, opts ...ps.ClientOption) (*ps.ListPipelinesOK, error)
    ListPipelineVersions(params *ps.ListPipelineVersionsParams, authInfo runtime.ClientAuthInfoWriter, opts ...ps.ClientOption) (*ps.ListPipelineVersionsOK, error)

    UploadPipeline(params *up.UploadPipelineParams, authInfo runtime.ClientAuthInfoWriter, opts ...up.ClientOption) (*up.UploadPipelineOK, error)
    UploadPipelineVersion(params *up.UploadPipelineVersionParams, authInfo runtime.ClientAuthInfoWriter, opts ...up.ClientOption) (*up.UploadPipelineVersionOK, error)
}

type Interface interface {
    Create(ctx context.Context, options *CreateOptions) (*Pipeline, error)
    Get(ctx context.Context, options *GetOptions) (*Pipeline, error)
    Update(ctx context.Context, options *UpdateOptions) (*Pipeline, error)
    Delete(ctx context.Context, options *DeleteOptions) error

    GetVersion(ctx context.Context, options *GetOptions) (*PipelineVersion, error)
    CreateVersion(ctx context.Context, options *CreateVersionOptions) (*PipelineVersion, error)
    DeleteVersion(ctx context.Context, options *DeleteOptions) error
}