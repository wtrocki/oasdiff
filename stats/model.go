package stats

import (
	"bytes"
	"encoding/json"
	"net/http"
	"time"

	"github.com/tufin/oasdiff/diff"
)

type Data struct {
	Conf    *diff.Config
	Diff    bool
	Error   error
	Time    time.Duration
	Summary *diff.Summary
}

func Send(data Data) {
	var buf bytes.Buffer
	err := json.NewEncoder(&buf).Encode(data)
	if err != nil {

	}
	http.Post("https://oasdiff-stats-xiixymmvca-ew.a.run.app", "application/json", &buf)
}
