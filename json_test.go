package herodot

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	exampleError = &richError{
		CodeField:   http.StatusNotFound,
		ErrorField:  errors.New("foo").Error(),
		ReasonField: "some-reason",
		StatusField: "some-status",
		DetailsField: []map[string]interface{}{
			map[string]interface{}{"foo": "bar"},
		},
	}
)

func TestWriteError(t *testing.T) {
	for k, tc := range []error{
		exampleError,
		errors.WithStack(exampleError),
	} {
		t.Run(fmt.Sprintf("case=%d", k), func(t *testing.T) {
			j := &jsonError{
				Error: &richError{},
			}

			h := NewJSONWriter(nil)
			r := mux.NewRouter()
			r.HandleFunc("/do", func(w http.ResponseWriter, r *http.Request) {
				r.Header.Set("X-Request-ID", "foo")
				h.WriteError(w, r, tc)
			})
			ts := httptest.NewServer(r)

			resp, err := http.Get(ts.URL + "/do")
			require.Nil(t, err)
			defer resp.Body.Close()

			require.Nil(t, json.NewDecoder(resp.Body).Decode(j))
			assert.Equal(t, http.StatusNotFound, resp.StatusCode)
			assert.Equal(t, "foo", j.Error.RequestID())
			assert.Equal(t, "some-status", j.Error.Status())
			assert.Equal(t, "some-reason", j.Error.Reason())
		})
	}
}

func TestWriteErrorCode(t *testing.T) {
	var j jsonError

	h := NewJSONWriter(nil)
	r := mux.NewRouter()
	r.HandleFunc("/do", func(w http.ResponseWriter, r *http.Request) {
		r.Header.Set("X-Request-ID", "foo")
		h.WriteErrorCode(w, r, http.StatusBadRequest, errors.Wrap(exampleError, ""))
	})
	ts := httptest.NewServer(r)

	resp, err := http.Get(ts.URL + "/do")
	require.Nil(t, err)

	require.Nil(t, json.NewDecoder(resp.Body).Decode(&j))
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	assert.Equal(t, "foo", j.Error.RequestID())
}

func TestWriteJSON(t *testing.T) {
	foo := map[string]string{"foo": "bar"}

	h := NewJSONWriter(nil)
	r := mux.NewRouter()
	r.HandleFunc("/do", func(w http.ResponseWriter, r *http.Request) {
		h.Write(w, r, &foo)
	})
	ts := httptest.NewServer(r)

	resp, err := http.Get(ts.URL + "/do")
	require.Nil(t, err)

	result := map[string]string{}
	require.Nil(t, json.NewDecoder(resp.Body).Decode(&result))
	assert.Equal(t, foo["foo"], result["foo"])
}

func TestWriteCreatedJSON(t *testing.T) {
	foo := map[string]string{"foo": "bar"}

	h := NewJSONWriter(nil)
	r := mux.NewRouter()
	r.HandleFunc("/do", func(w http.ResponseWriter, r *http.Request) {
		h.WriteCreated(w, r, "/new", &foo)
	})
	ts := httptest.NewServer(r)

	resp, err := http.Get(ts.URL + "/do")
	require.Nil(t, err)

	result := map[string]string{}
	require.Nil(t, json.NewDecoder(resp.Body).Decode(&result))
	assert.Equal(t, foo["foo"], result["foo"])
	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	assert.Equal(t, "/new", resp.Header.Get("Location"))
}

func TestWriteCodeJSON(t *testing.T) {
	foo := map[string]string{"foo": "bar"}

	h := NewJSONWriter(nil)
	r := mux.NewRouter()
	r.HandleFunc("/do", func(w http.ResponseWriter, r *http.Request) {
		h.WriteCode(w, r, 400, &foo)
	})
	ts := httptest.NewServer(r)

	resp, err := http.Get(ts.URL + "/do")
	require.Nil(t, err)

	result := map[string]string{}
	require.Nil(t, json.NewDecoder(resp.Body).Decode(&result))
	assert.Equal(t, foo["foo"], result["foo"])
	assert.Equal(t, 400, resp.StatusCode)
}
