package handler

import (
	"encoding/json"
	"github.com/imulab/go-scim/core/errors"
	"github.com/imulab/go-scim/protocol/http"
)

// Handler function implemented by endpoint handlers in this package.
type Func func(request http.Request, response http.Response)

// Write an error to response.
func WriteError(response http.Response, err error) {
	var scimError *errors.Error
	{
		if _, ok := err.(*errors.Error); !ok {
			scimError = errors.Internal(err.Error()).(*errors.Error)
		} else {
			scimError = err.(*errors.Error)
		}
	}

	response.WriteStatus(scimError.Status)
	response.WriteSCIMContentType()
	raw, _ := json.Marshal(scimError)
	response.WriteBody(raw)
}

const (
	attributes         = "attributes"
	excludedAttributes = "excludedAttributes"
	filter             = "filter"
	sortBy             = "sortBy"
	sortOrder          = "sortOrder"
	startIndex         = "startIndex"
	count              = "count"
	space              = " "
)
