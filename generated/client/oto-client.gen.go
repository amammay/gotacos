// Code generated by oto; DO NOT EDIT.

package client

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/pkg/errors"
)

// Client is used to access Pace services.
type Client struct {
	// RemoteHost is the URL of the remote server that this Client should
	// access.
	RemoteHost string
	// HTTPClient is the http.Client to use when making HTTP requests.
	HTTPClient *http.Client
	// Debug writes a line of debug log output.
	Debug func(s string)
}

// New makes a new Client.
func New(remoteHost string) *Client {
	c := &Client{
		RemoteHost: remoteHost,
		Debug:      func(s string) {},
		HTTPClient: &http.Client{Timeout: 10 * time.Second},
	}
	return c
}

// TacoService contains all knowledge around consumption of tacos
type TacoService struct {
	client *Client
}

// NewTacoService makes a new client for accessing TacoService services.
func NewTacoService(client *Client) *TacoService {
	return &TacoService{
		client: client,
	}
}

// EatTaco handles keeping track of eating tacos
func (s *TacoService) EatTaco(ctx context.Context, r EatTacoRequest) (*EatTacoResponse, error) {
	requestBodyBytes, err := json.Marshal(r)
	if err != nil {
		return nil, errors.Wrap(err, "TacoService.EatTaco: marshal EatTacoRequest")
	}
	url := s.client.RemoteHost + "TacoService.EatTaco"
	s.client.Debug(fmt.Sprintf("POST %s", url))
	s.client.Debug(fmt.Sprintf(">> %s", string(requestBodyBytes)))
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(requestBodyBytes))
	if err != nil {
		return nil, errors.Wrap(err, "TacoService.EatTaco: NewRequest")
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept-Encoding", "gzip")
	req = req.WithContext(ctx)
	resp, err := s.client.HTTPClient.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "TacoService.EatTaco")
	}
	defer resp.Body.Close()
	var response struct {
		EatTacoResponse
		Error string
	}
	var bodyReader io.Reader = resp.Body
	if strings.Contains(resp.Header.Get("Content-Encoding"), "gzip") {
		decodedBody, err := gzip.NewReader(resp.Body)
		if err != nil {
			return nil, errors.Wrap(err, "TacoService.EatTaco: new gzip reader")
		}
		defer decodedBody.Close()
		bodyReader = decodedBody
	}
	respBodyBytes, err := ioutil.ReadAll(bodyReader)
	if err != nil {
		return nil, errors.Wrap(err, "TacoService.EatTaco: read response body")
	}
	if err := json.Unmarshal(respBodyBytes, &response); err != nil {
		if resp.StatusCode != http.StatusOK {
			return nil, errors.Errorf("TacoService.EatTaco: (%d) %v", resp.StatusCode, string(respBodyBytes))
		}
		return nil, err
	}
	if response.Error != "" {
		return nil, errors.New(response.Error)
	}
	return &response.EatTacoResponse, nil
}

// EatTacoRequest is the request for TacoService.EatTaco.
type EatTacoRequest struct {

	// Name is your name
	Name string `json:"name"`

	// All of the Tacos you have consumed 🌮
	Tacos []string `json:"tacos"`
}

// EatTacoResponse is the response for TacoService.EatTaco.
type EatTacoResponse struct {

	// TacoConsumptionStatus is your current taco consumption status
	TacoConsumptionStatus string `json:"tacoConsumptionStatus"`
}
