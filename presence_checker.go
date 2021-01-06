package ddblocal

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type presenceChecker func(port int) bool

func (p presenceChecker) IsPresent(port int) bool {
	return p(port)
}

// Checks if DynamoDB local instance is running by checking the sponse to a GET
// request to the instance endpoint.
//
// The "correct" response looks like this:
//
// HTTP/1.1 400 Bad Request
// Content-Type: application/x-amz-json-1.0
//
// {
//    "__type": "com.amazonaws.dynamodb.v20120810#MissingAuthenticationToken",
//    "message": "Request must contain either a valid (registered) AWS access key ID or X.509 certificate."
// }
func isPresent(port int) bool {
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*500)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("http://localhost:%d", port), nil)
	if err != nil {
		return false
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return false
	}

	if resp.StatusCode != 400 {
		return false
	}

	var respBody struct {
		Type string `json:"__type"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&respBody); err != nil {
		return false
	}
	if !strings.HasPrefix(respBody.Type, "com.amazonaws.dynamodb") {
		return false
	}

	return true
}

// NewPresenceChecker returns a new instance of PresenceChecker with default
// configuration.
func NewPresenceChecker() PresenceChecker {
	return presenceChecker(isPresent)
}
