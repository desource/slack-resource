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

	"github.com/nlopes/slack"
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
		fmt.Fprintf(os.Stdout, `{"version":[]}`)

	case "in":
		if len(args) != 1 {
			r.Fail("usage: in <destination>")
		}
		destination := args[0]

		var req InRequest
		err := json.NewDecoder(os.Stdin).Decode(&req)
		if err != nil {
			r.Fail("invalid JSON request: %s", err)
		}

		resp := execIn(&r, destination, req)

		err = json.NewEncoder(os.Stdout).Encode(&resp)
		if err != nil {
			r.Fail("invalid JSON response: %s", err)
		}

	case "out":
		if len(args) != 1 {
			r.Fail("usage: out <source>")
		}
		source := args[0]

		var req OutRequest
		err := json.NewDecoder(os.Stdin).Decode(&req)
		if err != nil {
			r.Fail("invalid JSON request: %s", err)
		}

		resp := execOut(&r, source, req)

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

type InRequest struct {
	Source  Source           `json:"source"`
	Version TimestampVersion `json:"version"`
	Params  OutParams        `json:"params"`
}

type InResponse struct {
	Version TimestampVersion `json:"version"`
}

type OutRequest struct {
	Source  Source           `json:"source"`
	Version TimestampVersion `json:"version"`
	Params  OutParams        `json:"params"`
}

type Source struct {
	State   string `json:"state"`
	Token   string `json:"token"`
	Channel string `json:"channel"`
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

var now = func() int {
	return int(time.Now().Unix())
}

func execIn(r *Runner, destination string, req InRequest) (resp InResponse) {
	r.Log("IN %#v", req.Version)

	if req.Version.Timestamp != "" {
		resp.Version = req.Version
	} else {
		resp.Version.Timestamp = "none"
	}
	return
}

func execOut(r *Runner, source string, req OutRequest) (resp OutResponse) {
	resp.Version.Timestamp = strconv.Itoa(now())

	client := slack.New(req.Source.Token)

	externalURL := os.Getenv("ATC_EXTERNAL_URL")
	team := os.Getenv("BUILD_TEAM_NAME")
	pipeline := os.Getenv("BUILD_PIPELINE_NAME")
	job := os.Getenv("BUILD_JOB_NAME")
	build := os.Getenv("BUILD_NAME")

	shortLink := fmt.Sprintf("%s/teams/%s/pipelines/%s/jobs/%s", externalURL, team, pipeline, job)
	fullLink := fmt.Sprintf("%s/builds/%s", shortLink, build)

	postfix := ""
	color := ""
	switch req.Source.State {
	case "success":
		postfix = " :bowtie:"
		color = "#2ECC71"
	case "failure":
		postfix = " :sob:"
		color = "#E74C3C"
	}

	fallback := fmt.Sprintf("Build #%s %s/%s was a %s : %s", build, pipeline, job, req.Source.State, fullLink)

	text := fmt.Sprintf("Build *<%s|#%s>* in *<%s|%s/%s>* was a *%s* %s", fullLink, build, shortLink, pipeline, job, req.Source.State, postfix)

	params := slack.PostMessageParameters{
		Attachments: []slack.Attachment{
			{
				Fallback:   fallback,
				Color:      color,
				Text:       text,
				MarkdownIn: []string{"text"},
			},
		},
	}

	// verry cude incremental backoff
	backoffs := []int64{1, 3, 9}
	var err error
	for attempt := 0; ; attempt += 1 {
		if _, _, err = client.PostMessage(req.Source.Channel, "", params); err == nil {
			break
		} else if attempt >= len(backoffs) {
			break
		}
		time.Sleep(time.Duration(backoffs[attempt]) * time.Second)
	}
	if err != nil {
		r.Fail("Failed to post message: %s", err)
		return
	}

	return
}
