package stats_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tufin/oasdiff/diff"
	"github.com/tufin/oasdiff/stats"
)

func TestModel_PathInfoHTTPAndYaml(t *testing.T) {
	info := stats.GetInfo(1, nil, "https://raw.githubusercontent.com/Tufin/oasdiff/main/data/openapi-test1.yaml", "", nil, nil)
	require.Equal(t, "yaml", info.Base.Extension)
	require.Equal(t, "https", info.Base.Proto)
}

func TestModel_PathInfoYaml(t *testing.T) {
	info := stats.GetInfo(1, nil, "openapi-test1.yaml", "", nil, nil)
	require.Equal(t, "yaml", info.Base.Extension)
	require.Equal(t, "", info.Base.Proto)
}

func TestModel_PathInfoJson(t *testing.T) {
	info := stats.GetInfo(1, nil, "openapi-test1.json", "", nil, nil)
	require.Equal(t, "json", info.Base.Extension)
	require.Equal(t, "", info.Base.Proto)
}

func TestModel_PathInfoMultiDot(t *testing.T) {
	info := stats.GetInfo(1, nil, "a.b.c", "", nil, nil)
	require.Equal(t, "c", info.Base.Extension)
	require.Equal(t, "", info.Base.Proto)
}

func TestModel_PathInfoEmpty(t *testing.T) {
	info := stats.GetInfo(1, nil, "", "", nil, nil)
	require.Equal(t, "", info.Base.Extension)
	require.Equal(t, "", info.Base.Proto)
}

func TestModel_ErrNil(t *testing.T) {
	info := stats.GetInfo(1, nil, "", "", nil, nil)
	require.Equal(t, "", info.Err)
}

func TestModel_ErrEmpty(t *testing.T) {
	info := stats.GetInfo(1, nil, "", "", nil, errors.New(""))
	require.Equal(t, "", info.Err)
}

func TestModel_ErrText(t *testing.T) {
	info := stats.GetInfo(1, nil, "", "", nil, errors.New("reuven"))
	require.Equal(t, "reuven", info.Err)
}

func TestModel_DiffNil(t *testing.T) {
	info := stats.GetInfo(1, nil, "", "", nil, errors.New("reuven"))
	require.False(t, info.Diff)
}

func TestModel_DiffEmpty(t *testing.T) {
	info := stats.GetInfo(1, nil, "", "", &diff.Diff{}, errors.New("reuven"))
	require.False(t, info.Diff)
}

func TestModel_DiffNotEmpty(t *testing.T) {
	info := stats.GetInfo(1, nil, "", "", &diff.Diff{ExtensionsDiff: &diff.ExtensionsDiff{}}, errors.New("reuven"))
	require.True(t, info.Diff)
}
