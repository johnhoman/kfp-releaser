package pipelines

import "time"


type GetOptions struct {
    ID string
}

type CreateOptions struct {
    Description string
    Name string
    Workflow map[string]interface{}
}

type CreateVersionOptions struct {
    Description string
    Name string
    Workflow map[string]interface{}
    PipelineID string
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
}

type PipelineVersion struct {
    ID string
    Name string
    CreatedAt time.Time
    PipelineID string
}