package stats

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/tufin/oasdiff/diff"
)

type SpecInfo struct {
	Extension string
	Proto     string
}

func getPathProto(path string) string {
	frags := strings.Split(path, "://")
	if len(frags) < 2 {
		return ""
	}
	return frags[0]
}

func getPathExtension(path string) string {
	frags := strings.Split(path, ".")
	if len(frags) < 2 {
		return ""
	}
	return frags[len(frags)-1]
}

func getSpecInfo(path string) *SpecInfo {

	return &SpecInfo{
		Extension: getPathExtension(path),
		Proto:     getPathProto(path),
	}

}

type Durations struct {
	Load    time.Duration
	Diff    time.Duration
	Summary time.Duration
	Output  time.Duration
}

type Info struct {
	Config     *diff.Config
	Base       *SpecInfo
	Revision   *SpecInfo
	StatusCode int
	Diff       bool
	Durations  Durations
	Summary    *diff.Summary
	Err        string
}

func getErrStr(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

func diffExists(d *diff.Diff) bool {
	if d == nil {
		return false
	}
	return !d.Empty()
}

func GetInfo(statusCode int, config *diff.Config, base, revision string, durations Durations, d *diff.Diff, err error) *Info {
	return &Info{
		Config:     config,
		Base:       getSpecInfo(base),
		Revision:   getSpecInfo(revision),
		StatusCode: statusCode,
		Diff:       diffExists(d),
		Durations:  durations,
		Summary:    d.GetSummary(),
		Err:        getErrStr(err),
	}
}

func Send(info *Info, url string) {
	var buf bytes.Buffer
	err := json.NewEncoder(&buf).Encode(info)
	if err != nil {

	}
	http.Post(url, "application/json", &buf)
}
