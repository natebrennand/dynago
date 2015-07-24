package dynago

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestErrorStringRepresentation(t *testing.T) {
	e := Error{
		Type:          ErrorThrottling,
		Exception:     "ThrottlingException",
		AmazonRawType: "dynamodb#ThrottlingBlah",
		Message:       "FooBar",
	}
	assert.Equal(t, "dynago.Error(ErrorThrottling): ThrottlingException: FooBar", e.Error())
	e.Exception = ""
	assert.Equal(t, "dynago.Error(ErrorThrottling): dynamodb#ThrottlingBlah: FooBar", e.Error())
}

func TestErrorBuildError(t *testing.T) {
	assert := assert.New(t)
	input := `{"__type": "dynamodb#MissingAction", "message": "The FooBar happened."}`
	req, _ := http.NewRequest("POST", "http://fake/fake", nil)
	resp := &http.Response{
		Body: ioutil.NopCloser(bytes.NewBuffer([]byte(input))),
	}
	err := buildError(req, nil, resp)
	e := err.(*Error)
	assert.Equal(ErrorInvalidParameter, e.Type)
	assert.Equal([]byte(input), e.ResponseBody)
	assert.Equal("MissingAction", e.Exception)

	resp = &http.Response{
		Body: ioutil.NopCloser(bytes.NewBuffer([]byte("\n"))),
	}
	e = buildError(req, nil, resp).(*Error)
	assert.Equal(ErrorUnknown, e.Type)
	assert.Equal("unexpected end of JSON input", e.Message)
}
