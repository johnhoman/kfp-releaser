// Code generated by go-swagger; DO NOT EDIT.

package pipeline_upload_service

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"fmt"
	"io"

	"github.com/go-openapi/runtime"
	"github.com/go-openapi/strfmt"

	"github.com/johnhoman/kfp-releaser/pkg/kfp/pipeline_upload/models"
)

// UploadPipelineReader is a Reader for the UploadPipeline structure.
type UploadPipelineReader struct {
	formats strfmt.Registry
}

// ReadResponse reads a server response into the received o.
func (o *UploadPipelineReader) ReadResponse(response runtime.ClientResponse, consumer runtime.Consumer) (interface{}, error) {
	switch response.Code() {
	case 200:
		result := NewUploadPipelineOK()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return result, nil
	default:
		result := NewUploadPipelineDefault(response.Code())
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		if response.Code()/100 == 2 {
			return result, nil
		}
		return nil, result
	}
}

// NewUploadPipelineOK creates a UploadPipelineOK with default headers values
func NewUploadPipelineOK() *UploadPipelineOK {
	return &UploadPipelineOK{}
}

/* UploadPipelineOK describes a response with status code 200, with default header values.

UploadPipelineOK upload pipeline o k
*/
type UploadPipelineOK struct {
	Payload *models.APIPipeline
}

func (o *UploadPipelineOK) Error() string {
	return fmt.Sprintf("[POST /apis/v1beta1/pipelines/upload][%d] uploadPipelineOK  %+v", 200, o.Payload)
}
func (o *UploadPipelineOK) GetPayload() *models.APIPipeline {
	return o.Payload
}

func (o *UploadPipelineOK) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.APIPipeline)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewUploadPipelineDefault creates a UploadPipelineDefault with default headers values
func NewUploadPipelineDefault(code int) *UploadPipelineDefault {
	return &UploadPipelineDefault{
		_statusCode: code,
	}
}

/* UploadPipelineDefault describes a response with status code -1, with default header values.

UploadPipelineDefault upload pipeline default
*/
type UploadPipelineDefault struct {
	_statusCode int

	Payload *models.APIStatus
}

// Code gets the status code for the upload pipeline default response
func (o *UploadPipelineDefault) Code() int {
	return o._statusCode
}

func (o *UploadPipelineDefault) Error() string {
	return fmt.Sprintf("[POST /apis/v1beta1/pipelines/upload][%d] UploadPipeline default  %+v", o._statusCode, o.Payload)
}
func (o *UploadPipelineDefault) GetPayload() *models.APIStatus {
	return o.Payload
}

func (o *UploadPipelineDefault) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.APIStatus)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}
