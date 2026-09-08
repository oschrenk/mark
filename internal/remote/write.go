// Package remote builds and sends a Prometheus remote-write request.
//
// The payload is hand-encoded rather than imported from
// github.com/prometheus/prometheus, whose module pulls in client-go and the
// rest of the Prometheus server for four message types:
//
//	WriteRequest { repeated TimeSeries timeseries = 1; }
//	TimeSeries   { repeated Label labels = 1; repeated Sample samples = 2; }
//	Label        { string name = 1; string value = 2; }
//	Sample       { double value = 1; int64 timestamp = 2; }
package remote

import (
	"bytes"
	"fmt"
	"io"
	"math"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/golang/snappy"
	"google.golang.org/protobuf/encoding/protowire"
)

// Event is one thing that happened, at one instant.
type Event struct {
	Metric      string
	Description string
	Tags        []string
	At          time.Time
}

// Label is a single name/value pair on the series.
type Label struct {
	Name  string
	Value string
}

// Labels returns the series labels, sorted by name. Remote write requires
// lexicographic order and Prometheus rejects a series that is not sorted.
func (e Event) Labels() []Label {
	ls := []Label{
		{Name: "__name__", Value: e.Metric},
		{Name: "description", Value: e.Description},
	}
	if len(e.Tags) > 0 {
		ls = append(ls, Label{Name: "tags", Value: strings.Join(e.Tags, ",")})
	}
	sort.Slice(ls, func(i, j int) bool { return ls[i].Name < ls[j].Name })
	return ls
}

// TimestampMS is the sample time in milliseconds, which is what the wire
// format carries.
func (e Event) TimestampMS() int64 { return e.At.UnixMilli() }

// Value is the sample value. A single sample at 1 draws one vertical mark.
const Value = 1.0

// Encode returns the uncompressed protobuf WriteRequest.
func (e Event) Encode() []byte {
	var labels []byte
	for _, l := range e.Labels() {
		var f []byte
		f = protowire.AppendTag(f, 1, protowire.BytesType)
		f = protowire.AppendString(f, l.Name)
		f = protowire.AppendTag(f, 2, protowire.BytesType)
		f = protowire.AppendString(f, l.Value)

		labels = protowire.AppendTag(labels, 1, protowire.BytesType)
		labels = protowire.AppendBytes(labels, f)
	}

	var sample []byte
	sample = protowire.AppendTag(sample, 1, protowire.Fixed64Type)
	sample = protowire.AppendFixed64(sample, math.Float64bits(Value))
	sample = protowire.AppendTag(sample, 2, protowire.VarintType)
	sample = protowire.AppendVarint(sample, uint64(e.TimestampMS()))

	series := labels
	series = protowire.AppendTag(series, 2, protowire.BytesType)
	series = protowire.AppendBytes(series, sample)

	var req []byte
	req = protowire.AppendTag(req, 1, protowire.BytesType)
	req = protowire.AppendBytes(req, series)
	return req
}

// Send posts the snappy-compressed payload to a Prometheus remote-write
// endpoint. The endpoint only exists when Prometheus runs with
// --web.enable-remote-write-receiver.
func (e Event) Send(url string) error {
	body := snappy.Encode(nil, e.Encode())

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("building request: %w", err)
	}
	req.Header.Set("Content-Encoding", "snappy")
	req.Header.Set("Content-Type", "application/x-protobuf")
	req.Header.Set("X-Prometheus-Remote-Write-Version", "0.1.0")
	req.Header.Set("User-Agent", "mark")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("posting to %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode/100 != 2 {
		msg, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return fmt.Errorf("%s returned %s: %s", url, resp.Status, strings.TrimSpace(string(msg)))
	}
	return nil
}
