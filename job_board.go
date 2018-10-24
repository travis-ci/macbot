package main

import (
	"context"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// JobBoard is a client for interacting with an instance of job-board.
type JobBoard struct {
	Host     string
	Password string
	client   *http.Client
}

// NewJobBoard creates a new job board client.
func NewJobBoard(host, password string) *JobBoard {
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	return &JobBoard{
		Host:     host,
		Password: password,
		client:   client,
	}
}

// RegisterImage adds an image to job board with the given osx_image tag.
func (jb *JobBoard) RegisterImage(ctx context.Context, image, tag string) error {
	v := url.Values{}
	v.Set("infra", "jupiterbrain")
	v.Set("name", image)
	v.Set("tags", "os:osx,osx_image:"+tag)

	body := strings.NewReader(v.Encode())
	req, err := jb.newRequest("POST", "/images", body)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	_, err = jb.client.Do(req)
	return err
}

func (jb *JobBoard) newRequest(method, path string, body io.Reader) (*http.Request, error) {
	url := jb.Host + path
	r, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}

	r.SetBasicAuth("macbot", jb.Password)
	return r, nil
}
