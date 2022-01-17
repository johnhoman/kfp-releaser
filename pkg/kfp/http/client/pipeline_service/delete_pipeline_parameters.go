// Code generated by go-swagger; DO NOT EDIT.

package pipeline_service

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"context"
	"net/http"
	"time"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/runtime"
	cr "github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"
)

// NewDeletePipelineParams creates a new DeletePipelineParams object,
// with the default timeout for this client.
//
// Default values are not hydrated, since defaults are normally applied by the API server side.
//
// To enforce default values in parameter, use SetDefaults or WithDefaults.
func NewDeletePipelineParams() *DeletePipelineParams {
	return &DeletePipelineParams{
		timeout: cr.DefaultTimeout,
	}
}

// NewDeletePipelineParamsWithTimeout creates a new DeletePipelineParams object
// with the ability to set a timeout on a request.
func NewDeletePipelineParamsWithTimeout(timeout time.Duration) *DeletePipelineParams {
	return &DeletePipelineParams{
		timeout: timeout,
	}
}

// NewDeletePipelineParamsWithContext creates a new DeletePipelineParams object
// with the ability to set a context for a request.
func NewDeletePipelineParamsWithContext(ctx context.Context) *DeletePipelineParams {
	return &DeletePipelineParams{
		Context: ctx,
	}
}

// NewDeletePipelineParamsWithHTTPClient creates a new DeletePipelineParams object
// with the ability to set a custom HTTPClient for a request.
func NewDeletePipelineParamsWithHTTPClient(client *http.Client) *DeletePipelineParams {
	return &DeletePipelineParams{
		HTTPClient: client,
	}
}

/* DeletePipelineParams contains all the parameters to send to the API endpoint
   for the delete pipeline operation.

   Typically these are written to a http.Request.
*/
type DeletePipelineParams struct {

	/* ID.

	   The ID of the pipeline to be deleted.
	*/
	ID string

	timeout    time.Duration
	Context    context.Context
	HTTPClient *http.Client
}

// WithDefaults hydrates default values in the delete pipeline params (not the query body).
//
// All values with no default are reset to their zero value.
func (o *DeletePipelineParams) WithDefaults() *DeletePipelineParams {
	o.SetDefaults()
	return o
}

// SetDefaults hydrates default values in the delete pipeline params (not the query body).
//
// All values with no default are reset to their zero value.
func (o *DeletePipelineParams) SetDefaults() {
	// no default values defined for this parameter
}

// WithTimeout adds the timeout to the delete pipeline params
func (o *DeletePipelineParams) WithTimeout(timeout time.Duration) *DeletePipelineParams {
	o.SetTimeout(timeout)
	return o
}

// SetTimeout adds the timeout to the delete pipeline params
func (o *DeletePipelineParams) SetTimeout(timeout time.Duration) {
	o.timeout = timeout
}

// WithContext adds the context to the delete pipeline params
func (o *DeletePipelineParams) WithContext(ctx context.Context) *DeletePipelineParams {
	o.SetContext(ctx)
	return o
}

// SetContext adds the context to the delete pipeline params
func (o *DeletePipelineParams) SetContext(ctx context.Context) {
	o.Context = ctx
}

// WithHTTPClient adds the HTTPClient to the delete pipeline params
func (o *DeletePipelineParams) WithHTTPClient(client *http.Client) *DeletePipelineParams {
	o.SetHTTPClient(client)
	return o
}

// SetHTTPClient adds the HTTPClient to the delete pipeline params
func (o *DeletePipelineParams) SetHTTPClient(client *http.Client) {
	o.HTTPClient = client
}

// WithID adds the id to the delete pipeline params
func (o *DeletePipelineParams) WithID(id string) *DeletePipelineParams {
	o.SetID(id)
	return o
}

// SetID adds the id to the delete pipeline params
func (o *DeletePipelineParams) SetID(id string) {
	o.ID = id
}

// WriteToRequest writes these params to a swagger request
func (o *DeletePipelineParams) WriteToRequest(r runtime.ClientRequest, reg strfmt.Registry) error {

	if err := r.SetTimeout(o.timeout); err != nil {
		return err
	}
	var res []error

	// path param id
	if err := r.SetPathParam("id", o.ID); err != nil {
		return err
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}
