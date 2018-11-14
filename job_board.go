package main

import (
	"context"
	"encoding/json"
	"io"
	"io/ioutil"
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

// JobBoardImage is a registered macOS-based image in job board.
type JobBoardImage struct {
	ID   int64
	Tag  string
	Name string
}

type imageListPayload struct {
	Data []imagePayload `json:"data"`
}

type imagePayload struct {
	ID        int64
	Infra     string
	Name      string
	Tags      map[string]string
	IsDefault bool   `json:"is_default"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

// ListImages lists images by their osx_image tag.
func (jb *JobBoard) ListImages(ctx context.Context) ([]JobBoardImage, error) {
	req, err := jb.newRequest("GET", "/images?infra=jupiterbrain", nil)
	if err != nil {
		return nil, err
	}

	resp, err := jb.client.Do(req)
	if err != nil {
		return nil, err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var imageList imageListPayload
	if err = json.Unmarshal(body, &imageList); err != nil {
		return nil, err
	}

	images := make([]JobBoardImage, len(imageList.Data))
	for i, payload := range imageList.Data {
		images[i] = JobBoardImage{
			ID:   payload.ID,
			Tag:  payload.Tags["osx_image"],
			Name: payload.Name,
		}
	}

	return images, nil
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
