package fake

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/go-openapi/runtime"
	"github.com/go-openapi/strfmt"
	"github.com/google/uuid"
	"github.com/johnhoman/kfp-releaser/pkg/kfp/pipelines"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	ps "github.com/johnhoman/kfp-releaser/pkg/kfp/pipeline/client/pipeline_service"
	"github.com/johnhoman/kfp-releaser/pkg/kfp/pipeline/models"
	up "github.com/johnhoman/kfp-releaser/pkg/kfp/pipeline_upload/client/pipeline_upload_service"
	upmodels "github.com/johnhoman/kfp-releaser/pkg/kfp/pipeline_upload/models"
)


type PagedRequest struct {
	Items []*models.APIPipeline
	LastSent int
	Total int
}

type PagedVersionRequest struct {
	Items []*models.APIPipelineVersion
	LastSent int
	Total int
}

type PipelineService struct {
	sync.Mutex
	// map[models.APIPipeline.ID]models.APIPipeline
	Pipelines        map[string]models.APIPipeline
	PipelineVersions map[string]map[string]models.APIPipelineVersion
	PagedRequests    map[string]PagedRequest
	PagedVersionRequests map[string]PagedVersionRequest
}

var internalServerError = fmt.Errorf("\"&{0 [] }\" (*models.APIStatus) is not supported by the TextConsumer, %s",
	"can be resolved by supporting TextUnmarshaler interface")

func (p *PipelineService) UploadPipeline(params *up.UploadPipelineParams, authInfo runtime.ClientAuthInfoWriter, opts ...up.ClientOption) (*up.UploadPipelineOK, error) {
	for _, pipeline := range p.Pipelines {
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
		if pipeline.Name == *params.Name {
			return nil, internalServerError
		}
	}
	uid := uuid.New().String()
	now := strfmt.DateTime(time.Now().UTC())
	m := models.APIPipeline{
		CreatedAt: now,
		DefaultVersion: &models.APIPipelineVersion{
			CreatedAt: now,
			Description: "",
			ID: uid,
			Name: *params.Name,
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
	p.Pipelines[m.ID] = m
	p.PipelineVersions[m.ID] = map[string]models.APIPipelineVersion{m.ID: *m.DefaultVersion}
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
	p.PipelineVersions[id][m.ID] = m
	pipeline := p.Pipelines[id]
	pipeline.DefaultVersion = &m
	p.Pipelines[id] = pipeline

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

func (p *PipelineService) DeletePipeline(params *ps.DeletePipelineParams, authInfo runtime.ClientAuthInfoWriter, opts ...ps.ClientOption) (*ps.DeletePipelineOK, error) {

	p.Mutex.Lock()
	defer p.Mutex.Unlock()

	_, ok := p.Pipelines[params.ID]
	if ok {
		delete(p.Pipelines, params.ID)
		delete(p.PipelineVersions, params.ID)
		return &ps.DeletePipelineOK{Payload: map[string]interface{}{}}, nil
	}

	return nil, ps.NewDeletePipelineDefault(http.StatusNotFound)
}

func (p *PipelineService) DeletePipelineVersion(params *ps.DeletePipelineVersionParams, authInfo runtime.ClientAuthInfoWriter, opts ...ps.ClientOption) (*ps.DeletePipelineVersionOK, error) {
	for id, versions := range p.PipelineVersions {
		_, ok := versions[params.VersionID]
		if ok {
			delete(p.PipelineVersions[id], params.VersionID)
			return &ps.DeletePipelineVersionOK{Payload: map[string]interface{}{}}, nil
		}
	}
	return nil, ps.NewDeletePipelineVersionDefault(http.StatusNotFound)
}

func (p *PipelineService) GetPipeline(params *ps.GetPipelineParams, authInfo runtime.ClientAuthInfoWriter, opts ...ps.ClientOption) (*ps.GetPipelineOK, error) {
	pipeline, ok := p.Pipelines[params.ID]
	if !ok {
		return nil, ps.NewGetPipelineDefault(http.StatusNotFound)
	}
	return &ps.GetPipelineOK{Payload: &pipeline}, nil
}

func (p *PipelineService) GetPipelineVersion(params *ps.GetPipelineVersionParams, authInfo runtime.ClientAuthInfoWriter, opts ...ps.ClientOption) (*ps.GetPipelineVersionOK, error) {
	for _, versions := range p.PipelineVersions {
		version, ok := versions[params.VersionID]
		if ok {
			return &ps.GetPipelineVersionOK{Payload: &version}, nil
		}
	}
	return nil, ps.NewGetPipelineVersionDefault(http.StatusNotFound)
}

func (p *PipelineService) UpdatePipelineDefaultVersion(params *ps.UpdatePipelineDefaultVersionParams, authInfo runtime.ClientAuthInfoWriter, opts ...ps.ClientOption) (*ps.UpdatePipelineDefaultVersionOK, error) {
	// There's a few silent failures here
	p.Mutex.Lock()
	defer p.Mutex.Unlock()

	pipeline, ok := p.Pipelines[params.PipelineID]
	if ok {
		versions := p.PipelineVersions[pipeline.ID]
		version, ok := versions[params.VersionID]
		if !ok {
			// Bug in kubeflow I think
			// If the version doesn't exist it just deletes the default version from the pipeline
			pipeline.DefaultVersion = &models.APIPipelineVersion{}
		} else {
			pipeline.DefaultVersion = &version
		}
		p.Pipelines[params.PipelineID] = pipeline
	}
	return &ps.UpdatePipelineDefaultVersionOK{Payload: map[string]interface{}{}}, nil
}

func (p *PipelineService) ListPipelines(params *ps.ListPipelinesParams, authInfo runtime.ClientAuthInfoWriter, opts ...ps.ClientOption) (*ps.ListPipelinesOK, error) {
	p.Mutex.Lock()
	defer p.Mutex.Unlock()
	if params.PageToken != nil {
		req := p.PagedRequests[*params.PageToken]
		pageSize := req.Total - req.LastSent - 1
		if params.PageSize != nil && *params.PageSize < int32(pageSize) {
			pageSize = int(*params.PageSize)
		}
		apiPipelines := make([]*models.APIPipeline, 0, pageSize)
		for pageSize > 0 {
			req.LastSent++
			apiPipelines = append(apiPipelines, req.Items[req.LastSent])
			pageSize--
		}
		token := *params.PageToken
		if req.LastSent == req.Total - 1 {
			// Finished sending everything
			delete(p.PagedRequests, token)
			token = ""
		} else {
			p.PagedRequests[token] = req
		}
		return &ps.ListPipelinesOK{Payload: &models.APIListPipelinesResponse{
			NextPageToken: token,
			Pipelines: apiPipelines,
			TotalSize: int32(req.Total),
		}}, nil
	}

	apiPipelines := make([]*models.APIPipeline, 0, len(p.Pipelines))
	for _, pipeline := range p.Pipelines {
		pl := pipeline
		apiPipelines = append(apiPipelines, &pl)
	}

	decoded, err := url.QueryUnescape(*params.Filter)
	if err != nil {
		return &ps.ListPipelinesOK{}, err
	}
	var filter map[string]interface{}
	if err := json.Unmarshal([]byte(decoded), &filter); err != nil {
		return &ps.ListPipelinesOK{}, err
	}

	for _, predicate := range filter["predicates"].([]interface{}) {
		m := predicate.(map[string]interface{})

		var pred func(string, string) bool

		if op, ok := m["op"]; !ok {
			return nil, ps.NewListPipelinesDefault(http.StatusBadRequest)
		} else {
			switch op {
			case "EQUALS": {
				pred = func(a, b string) bool {
					return a == b
				}
			}
			case "IS_SUBSTRING": {
				pred = strings.Contains
			}
			default:
				return nil, ps.NewListPipelinesDefault(http.StatusBadRequest)
			}
		}
		if key, ok := m["key"]; !ok || key != "name" {
			return nil, ps.NewListPipelinesDefault(http.StatusBadRequest)
		}
		if _, ok := m["string_value"]; !ok {
			return nil, ps.NewListPipelinesDefault(http.StatusBadRequest)
		}
		for k, item := range apiPipelines {
			if !pred(item.Name, m["string_value"].(string)) {
				apiPipelines[k] = nil
			}
		}
	}

	validPipelines := make([]*models.APIPipeline, 0, len(apiPipelines))
	for _, pl := range apiPipelines {
		if pl != nil {
			validPipelines = append(validPipelines, pl)
		}
	}

	token := ""
	page := validPipelines
	if int(*params.PageSize) < len(validPipelines) {
		token = base64.StdEncoding.EncodeToString([]byte(uuid.New().String()))
		p.PagedRequests[token] = PagedRequest{
			Total: len(validPipelines),
			Items: validPipelines,
			LastSent: int(*params.PageSize) - 1,
		}
		page = validPipelines[:*params.PageSize]
	}

	out := &ps.ListPipelinesOK{Payload: &models.APIListPipelinesResponse{
		Pipelines: page,
		TotalSize: int32(len(validPipelines)),
		NextPageToken: token,
	}}
	return out, nil
}

func (p *PipelineService) ListPipelineVersions(params *ps.ListPipelineVersionsParams, authInfo runtime.ClientAuthInfoWriter, opts ...ps.ClientOption) (*ps.ListPipelineVersionsOK, error) {
	p.Mutex.Lock()
	defer p.Mutex.Unlock()
	if params.PageToken != nil {
		req := p.PagedVersionRequests[*params.PageToken]
		pageSize := req.Total - req.LastSent - 1
		if params.PageSize != nil && *params.PageSize < int32(pageSize) {
			pageSize = int(*params.PageSize)
		}
		versions := make([]*models.APIPipelineVersion, 0, pageSize)
		for pageSize > 0 {
			req.LastSent++
			versions = append(versions, req.Items[req.LastSent])
			pageSize--
		}
		token := *params.PageToken
		if req.LastSent == req.Total - 1 {
			// Finished sending everything
			delete(p.PagedVersionRequests, token)
			token = ""
		} else {
			p.PagedVersionRequests[token] = req
		}
		return &ps.ListPipelineVersionsOK{Payload: &models.APIListPipelineVersionsResponse{
			NextPageToken: token,
			Versions: versions,
			TotalSize: int32(req.Total),
		}}, nil
	}

	versions := make([]*models.APIPipelineVersion, 0, len(p.Pipelines))
	for _, version := range p.PipelineVersions[*params.ResourceKeyID] {
		vs := version
		versions = append(versions, &vs)
	}

	decoded, err := url.QueryUnescape(*params.Filter)
	if err != nil {
		return &ps.ListPipelineVersionsOK{}, err
	}
	var filter map[string]interface{}
	if err := json.Unmarshal([]byte(decoded), &filter); err != nil {
		return &ps.ListPipelineVersionsOK{}, err
	}

	for _, predicate := range filter["predicates"].([]interface{}) {
		m := predicate.(map[string]interface{})

		var pred func(string, string) bool

		if op, ok := m["op"]; !ok {
			return nil, ps.NewListPipelineVersionsDefault(http.StatusBadRequest)
		} else {
			switch op {
			case "EQUALS": {
				pred = func(a, b string) bool {
					return a == b
				}
			}
			case "IS_SUBSTRING": {
				pred = strings.Contains
			}
			default:
				return nil, ps.NewListPipelineVersionsDefault(http.StatusBadRequest)
			}
		}
		if key, ok := m["key"]; !ok || key != "name" {
			return nil, ps.NewListPipelinesDefault(http.StatusBadRequest)
		}
		if _, ok := m["string_value"]; !ok {
			return nil, ps.NewListPipelinesDefault(http.StatusBadRequest)
		}
		for k, item := range versions {
			if !pred(item.Name, m["string_value"].(string)) {
				versions[k] = nil
			}
		}
	}

	validVersions := make([]*models.APIPipelineVersion, 0, len(versions))
	for _, pl := range versions {
		if pl != nil {
			validVersions = append(validVersions, pl)
		}
	}

	token := ""
	page := validVersions
	if int(*params.PageSize) < len(validVersions) {
		token = base64.StdEncoding.EncodeToString([]byte(uuid.New().String()))
		p.PagedVersionRequests[token] = PagedVersionRequest{
			Total: len(validVersions),
			Items: validVersions,
			LastSent: int(*params.PageSize) - 1,
		}
		page = validVersions[:*params.PageSize]
	}

	out := &ps.ListPipelineVersionsOK{Payload: &models.APIListPipelineVersionsResponse{
		Versions: page,
		TotalSize: int32(len(validVersions)),
		NextPageToken: token,
	}}
	return out, nil
}

var _ pipelines.PipelineService = &PipelineService{}

func NewPipelineService() *PipelineService {
	return &PipelineService{
		Pipelines:        make(map[string]models.APIPipeline),
		PipelineVersions: map[string]map[string]models.APIPipelineVersion{},
		PagedRequests: make(map[string]PagedRequest),
		PagedVersionRequests: make(map[string]PagedVersionRequest),
	}
}