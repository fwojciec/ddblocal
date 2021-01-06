package ddblocal_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"

	"github.com/fwojciec/ddblocal"
)

func TestReportsPresenceCorrectly(t *testing.T) {
	t.Parallel()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/x-amz-json-1.0")
		w.WriteHeader(http.StatusBadRequest)
		msg := struct {
			Type string `json:"__type"`
		}{
			Type: "com.amazonaws.dynamodb.v20120810#MissingAuthenticationToken",
		}
		_ = json.NewEncoder(w).Encode(msg)
	}))
	defer ts.Close()

	pc := ddblocal.NewPresenceChecker()

	port, err := testServerPort(ts.URL)
	ok(t, err)

	res := pc.IsPresent(port)
	equals(t, true, res)
}

func TestFailsPresenceTestIfStatusIsNotBadRequest(t *testing.T) {
	t.Parallel()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/x-amz-json-1.0")
		w.WriteHeader(http.StatusOK)
		msg := struct {
			Type string `json:"__type"`
		}{
			Type: "com.amazonaws.dynamodb.v20120810#MissingAuthenticationToken",
		}
		_ = json.NewEncoder(w).Encode(msg)
	}))
	defer ts.Close()

	pc := ddblocal.NewPresenceChecker()

	port, err := testServerPort(ts.URL)
	ok(t, err)

	res := pc.IsPresent(port)
	equals(t, false, res)
}

func TestFailsPresenceTestIfStatusDoesntMatch(t *testing.T) {
	t.Parallel()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/x-amz-json-1.0")
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer ts.Close()

	pc := ddblocal.NewPresenceChecker()

	port, err := testServerPort(ts.URL)
	ok(t, err)

	res := pc.IsPresent(port)
	equals(t, false, res)
}

func TestFailsPresenceTestIfServerDoesntRespond(t *testing.T) {
	t.Parallel()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	}))
	ts.Close()

	pc := ddblocal.NewPresenceChecker()

	port, err := testServerPort(ts.URL)
	ok(t, err)

	res := pc.IsPresent(port)
	equals(t, false, res)
}

func testServerPort(addr string) (int, error) {
	u, err := url.Parse(addr)
	if err != nil {
		return 0, err
	}
	ui, err := strconv.Atoi(u.Port())
	if err != nil {
		return 0, err
	}
	return ui, nil
}
