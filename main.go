package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

var cmd string

func init() {
	flag.StringVar(&cmd, "cmd", filepath.Base(os.Args[0]), "check, in, out")
}

func main() {
	flag.Parse()
	Runner{
		Stdin:  os.Stdin,
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	}.Exec(cmd, flag.Args()...)
}

type Runner struct {
	Stdin  io.Reader
	Stdout io.Writer
	Stderr io.Writer
}

func (r Runner) Exec(cmd string, args ...string) {
	switch cmd {
	case "check":
		fmt.Fprintf(os.Stdout, `{"version":{"ref":"none"}}`)

	case "in":
		if len(args) != 1 {
			r.Fail("usage: in <destination>")
		}
		fmt.Fprintf(os.Stdout, `{"version":{"ref":"none"}}`)

	case "out":
		if len(args) != 1 {
			r.Fail("usage: out <destination>")
		}

		var req OutRequest
		err := json.NewDecoder(os.Stdin).Decode(&req)
		if err != nil {
			r.Fail("invalid JSON request: %s", err)
		}

		resp := execOut(&r, req)

		err = json.NewEncoder(os.Stdout).Encode(&resp)
		if err != nil {
			r.Fail("invalid JSON response: %s", err)
		}

	default:
		r.Fail("unexpected command %s; must be check, in, out", cmd)
	}
}

func (r *Runner) Log(msg string, args ...interface{}) {
	fmt.Fprintf(r.Stderr, msg, args...)
	fmt.Fprintln(r.Stderr)
}

func (r *Runner) Fail(msg string, args ...interface{}) {
	r.Log(msg, args...)
	os.Exit(1)
}

var now = func() int {
	return int(time.Now().Unix())
}

func execOut(r *Runner, req OutRequest) (resp OutResponse) {
	resp.Version.Timestamp = strconv.Itoa(now())
	return
}

type OutRequest struct {
	Source  Source    `json:"source"`
	Version Version   `json:"version"`
	Params  OutParams `json:"params"`
}

type Source struct {
	URL string `json:"url"`
}

type Version struct {
	Ref string `json:"ref"`
}

type OutParams struct {
	Status string `json:"status"`
}

type OutResponse struct {
	Version  TimestampVersion `json:"version"`
	Metadata Metadata         `json:"metadata,omitempty"`
}

type TimestampVersion struct {
	Timestamp string `json:"timestamp"`
}

type Metadata []MetadataField

type MetadataField struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}
