package stats_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tufin/oasdiff/diff"
	"github.com/tufin/oasdiff/stats"
)

func TestModel_PathInfoHTTPAndYaml(t *testing.T) {
	info := stats.GetInfo(1, nil, "https://raw.githubusercontent.com/Tufin/oasdiff/main/data/openapi-test1.yaml", "", stats.Durations{}, nil, nil)
	require.Equal(t, "yaml", info.Base.Extension)
	require.Equal(t, "https", info.Base.Proto)
}

func TestModel_PathInfoYaml(t *testing.T) {
	info := stats.GetInfo(1, nil, "openapi-test1.yaml", "", stats.Durations{}, nil, nil)
	require.Equal(t, "yaml", info.Base.Extension)
	require.Equal(t, "", info.Base.Proto)
}

func TestModel_PathInfoJson(t *testing.T) {
	info := stats.GetInfo(1, nil, "openapi-test1.json", "", stats.Durations{}, nil, nil)
	require.Equal(t, "json", info.Base.Extension)
	require.Equal(t, "", info.Base.Proto)
}

func TestModel_PathInfoMultiDot(t *testing.T) {
	info := stats.GetInfo(1, nil, "a.b.c", "", stats.Durations{}, nil, nil)
	require.Equal(t, "c", info.Base.Extension)
	require.Equal(t, "", info.Base.Proto)
}

func TestModel_PathInfoEmpty(t *testing.T) {
	info := stats.GetInfo(1, nil, "", "", stats.Durations{}, nil, nil)
	require.Equal(t, "", info.Base.Extension)
	require.Equal(t, "", info.Base.Proto)
}

func TestModel_ErrNil(t *testing.T) {
	info := stats.GetInfo(1, nil, "", "", stats.Durations{}, nil, nil)
	require.Equal(t, "", info.Err)
}

func TestModel_ErrEmpty(t *testing.T) {
	info := stats.GetInfo(1, nil, "", "", stats.Durations{}, nil, errors.New(""))
	require.Equal(t, "", info.Err)
}

func TestModel_ErrText(t *testing.T) {
	info := stats.GetInfo(1, nil, "", "", stats.Durations{}, nil, errors.New("reuven"))
	require.Equal(t, "reuven", info.Err)
}

func TestModel_DiffNil(t *testing.T) {
	info := stats.GetInfo(1, nil, "", "", stats.Durations{}, nil, errors.New("reuven"))
	require.False(t, info.Diff)
}

func TestModel_DiffEmpty(t *testing.T) {
	info := stats.GetInfo(1, nil, "", "", stats.Durations{}, &diff.Diff{}, errors.New("reuven"))
	require.False(t, info.Diff)
}

func TestModel_DiffNotEmpty(t *testing.T) {
	info := stats.GetInfo(1, nil, "", "", stats.Durations{}, &diff.Diff{ExtensionsDiff: &diff.ExtensionsDiff{}}, errors.New("reuven"))
	require.True(t, info.Diff)
}

func TestModel_DiffSummary(t *testing.T) {
	info := stats.GetInfo(1, nil, "", "", stats.Durations{}, &diff.Diff{ExtensionsDiff: &diff.ExtensionsDiff{}}, nil)
	require.True(t, info.Diff)
}

func TestSend(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		info := stats.Info{}
		err := json.NewDecoder(r.Body).Decode(&info)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		require.Equal(t, "http", info.Base.Proto)
		require.Equal(t, "", info.Revision.Proto)
		require.True(t, info.Diff)
		require.True(t, info.Summary.Diff)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	info := stats.GetInfo(1, diff.NewConfig(), "http://1.json", "/tmp/2.yaml", stats.Durations{}, &diff.Diff{PathsDiff: &diff.PathsDiff{Added: diff.StringList{"1"}}}, errors.New("reuven"))
	stats.Send(info, server.URL)
}
