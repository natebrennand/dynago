package aws

import (
	"bytes"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
)

var MaxResponseSize int64 = 5 * 1024 * 1024 // 5MB maximum response

const DynamoTargetPrefix = "DynamoDB_20120810." // This is the Dynamo API version we support

type Signer interface {
	SignRequest(*http.Request, []byte)
}

/*
RequestMaker is the default AwsRequester used by Dynago.

The RequestMaker has its properties exposed as public to allow easier
construction. Directly modifying properties on the RequestMaker after
construction is not goroutine-safe so it should be avoided except for in
special cases (testing, mocking).
*/
type RequestMaker struct {
	// These are required to be set
	Endpoint   string
	Signer     Signer
	BuildError func(*http.Request, []byte, *http.Response) error

	// These can be optionally set
	Caller         http.Client
	DebugRequests  bool
	DebugResponses bool
	DebugFunc      func(string, ...interface{})
}

func (r *RequestMaker) MakeRequest(target string, body []byte) (io.Reader, error) {
	req, err := http.NewRequest("POST", r.Endpoint, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	if !strings.Contains(target, ".") {
		target = DynamoTargetPrefix + target
	}
	req.Header.Add("x-amz-target", target)
	req.Header.Add("content-type", "application/x-amz-json-1.0")
	req.Header.Set("Host", req.URL.Host)
	r.Signer.SignRequest(req, body)
	if r.DebugRequests {
		r.DebugFunc("Request:%#v\n\nRequest Body: %s\n\n", req, body)
	}
	response, err := r.Caller.Do(req)
	if err != nil {
		return nil, err
	}

	if r.DebugResponses {
		buf, _ := ioutil.ReadAll(response.Body)
		r.DebugFunc("Response: %#v\nBody:%s\n", response, buf)
		response.Body = ioutil.NopCloser(bytes.NewBuffer(buf))
	}

	if response.StatusCode != http.StatusOK {
		return response.Body, r.BuildError(req, body, response)
	}
	return io.LimitReader(response.Body, MaxResponseSize), err
}
