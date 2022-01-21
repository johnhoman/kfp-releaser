package pipelines

import (
    "fmt"
    "net/http"
)

type ApiError struct {
    status string
    code int
}

func (a *ApiError) Error() string {
    return fmt.Sprintf("status=%s,code=%d", a.status, a.code)
}

func (a *ApiError) Code() int {
    return a.code
}

func NewConflict() *ApiError {
    return &ApiError{code: http.StatusConflict, status: "Conflict"}
}

func IsConflict(err error) bool {
    _, ok := err.(*ApiError)
    if !ok {
        return false
    }
    return err.(*ApiError).Code() == http.StatusConflict
}

func NewNotFound() *ApiError {
    return &ApiError{code: http.StatusNotFound, status: "NotFound"}
}

func IsNotFound(err error) bool {
    _, ok := err.(*ApiError)
    if !ok {
        return false
    }
    return err.(*ApiError).Code() == http.StatusNotFound
}