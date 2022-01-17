// Code generated by go-swagger; DO NOT EDIT.

package models

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"context"

	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/swag"
)

// APIGetTemplateResponse api get template response
//
// swagger:model apiGetTemplateResponse
type APIGetTemplateResponse struct {

	// The template of the pipeline specified in a GetTemplate request, or of a
	// pipeline version specified in a GetPipelinesVersionTemplate request.
	Template string `json:"template,omitempty"`
}

// Validate validates this api get template response
func (m *APIGetTemplateResponse) Validate(formats strfmt.Registry) error {
	return nil
}

// ContextValidate validates this api get template response based on context it is used
func (m *APIGetTemplateResponse) ContextValidate(ctx context.Context, formats strfmt.Registry) error {
	return nil
}

// MarshalBinary interface implementation
func (m *APIGetTemplateResponse) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *APIGetTemplateResponse) UnmarshalBinary(b []byte) error {
	var res APIGetTemplateResponse
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}
