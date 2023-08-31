package pinecone

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

var (
	ErrEmptyAPIKey      = errors.New("error: api key is empty")
	ErrEmptyEnvironment = errors.New("error: environment is empty")
)

var (
	_ PineconeClientInterface = &PineconeClient{}
)

type PineconeClientInterface interface {
	GetAPIKey() string
	GetEnvironment() string
	GetBaseURL() string
	ListIndexes(ctx context.Context) ([]string, error)
	CreateIndex(ctx context.Context, req CreateIndexRequest) error
	DescribeIndex(ctx context.Context, indexName string) (*DescribeIndexResponse, error)
	DeleteIndex(ctx context.Context, indexName string) error
	ConfigureIndex(ctx context.Context, indexName string, req ConfigureIndexRequest) error
}

type PineconeClient struct {
	APIKey      string
	Environment string
}

func (c *PineconeClient) GetAPIKey() string {
	return c.APIKey
}

func (c *PineconeClient) GetEnvironment() string {
	return c.Environment
}

type Options struct {
	APIKey      string
	Environment string
}

type Option func(*Options)

func WithAPIKey(apiKey string) Option {
	return func(o *Options) {
		o.APIKey = apiKey
	}
}

func WithEnvironment(environment string) Option {
	return func(o *Options) {
		o.Environment = environment
	}
}

// Finally, update the NewClient function to call NewClientWithInterfaces, passing in instances of jsonMarshaler and jsonUnmarshaler:
func NewClient(apiKey string, environment string) (*PineconeClient, error) {
	return &PineconeClient{
		APIKey:      apiKey,
		Environment: environment,
	}, nil
}

type ListIndexesResponse []string

// GetBaseURL get base url
func (c *PineconeClient) GetBaseURL() string {
	if c.Environment == "" {
		return ""
	}
	return fmt.Sprintf("https://controller.%s.pinecone.io", c.Environment)
}

// ListIndexes lists all indexes
func (c *PineconeClient) ListIndexes(ctx context.Context) ([]string, error) {
	baseURL := c.GetBaseURL()
	if baseURL == "" {
		return nil, fmt.Errorf("error: environment is empty")
	}
	url := fmt.Sprintf("%s/databases", baseURL)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("accept", "application/json; charset=utf-8")
	req.Header.Add("Api-Key", c.APIKey)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	if res.StatusCode >= http.StatusBadRequest {
		return nil, fmt.Errorf("error: error: %s", res.Status)
	}
	defer res.Body.Close()

	// parse json response
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	var resp ListIndexesResponse
	// unmarshal the response body into resp
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, err
	}
	return resp, nil
}

type Metric int

const (
	MetricEuclidean Metric = iota
	MetricCosine
	MetricDotProduct
)

func (m Metric) String() string {
	metricTypes := [...]string{"euclidean", "cosine", "dotproduct"}
	if m < 0 || int(m) >= len(metricTypes) {
		return "cosine" // default value
	}
	return metricTypes[m]
}

func NewMetric(metricStr string) (Metric, error) {
	switch metricStr {
	case "euclidean":
		return MetricEuclidean, nil
	case "cosine":
		return MetricCosine, nil
	case "dotproduct":
		return MetricDotProduct, nil
	default:
		return MetricEuclidean, fmt.Errorf("error: invalid metric value: %s", metricStr)
	}
}

func (m Metric) MarshalJSON() ([]byte, error) {
	return []byte(`"` + m.String() + `"`), nil
}

func (m *Metric) UnmarshalJSON(data []byte) error {
	var metricStr string
	if err := json.Unmarshal(data, &metricStr); err != nil {
		return err
	}

	switch metricStr {
	case "euclidean":
		*m = MetricEuclidean
	case "cosine":
		*m = MetricCosine
	case "dotproduct":
		*m = MetricDotProduct
	default:
		return fmt.Errorf("error: invalid metric value: %s", metricStr)
	}

	return nil
}

type PodType struct {
	Class string
	Size  string
}

func (p PodType) String() string {
	return p.Class + "." + p.Size
}

// NewPodType creates a new PodType
func NewPodType(podTypeStr string) (PodType, error) {
	splitPodType := strings.Split(podTypeStr, ".")
	if len(splitPodType) != 2 {
		return PodType{}, fmt.Errorf("error: invalid pod type: %s", podTypeStr)
	}

	class := splitPodType[0]
	size := splitPodType[1]

	if class != "s1" && class != "p1" && class != "p2" {
		return PodType{}, fmt.Errorf("error: invalid pod class: %s", class)
	}

	if size != "x1" && size != "x2" && size != "x4" && size != "x8" {
		return PodType{}, fmt.Errorf("error: invalid pod size: %s", size)
	}

	return PodType{Class: class, Size: size}, nil
}

func (p PodType) MarshalJSON() ([]byte, error) {
	return []byte(`"` + p.String() + `"`), nil
}

func (p *PodType) UnmarshalJSON(data []byte) error {
	var podTypeStr string
	if err := json.Unmarshal(data, &podTypeStr); err != nil {
		return err
	}

	podType, err := NewPodType(podTypeStr)
	if err != nil {
		return err
	}

	*p = podType
	return nil
}

// Nested struct that supports the format {"indexed": ["example_metadata_field"]}
type MetadataConfig struct {
	Indexed []string `json:"indexed"`
}

func NewMetadataConfig(receivedMetadataConfig types.Object) (*MetadataConfig, error) {
	if receivedMetadataConfig.IsNull() || len(receivedMetadataConfig.Attributes()) == 0 {
		return nil, nil
	}

	var indexed []string
	values := receivedMetadataConfig.Attributes()["indexed"]

	listValues := values.(basetypes.ListValue)
	for _, val := range listValues.Elements() {
		str, ok := val.(basetypes.StringValue)
		if !ok {
			return nil, fmt.Errorf("error: invalid type for indexed element")
		}

		indexed = append(indexed, str.ValueString())
	}

	return &MetadataConfig{
		Indexed: indexed,
	}, nil
}

type CreateIndexRequest struct {
	Name           string          `json:"name"` // The name of the index to be created. The maximum length is 45 characters.
	Dimension      int             `json:"dimension"`
	Metric         Metric          `json:"metric"` // You can use 'euclidean', 'cosine', or 'dotproduct'.
	Pods           int             `json:"pods"`
	Replicas       int             `json:"replicas"`
	PodType        PodType         `json:"pod_type"` // The type of pod to use. One of s1, p1, or p2 appended with . and one of x1, x2, x4, or x8.
	MetadataConfig *MetadataConfig `json:"metadata_config,omitempty"`
}

// CreateIndex creates an index
func (c *PineconeClient) CreateIndex(ctx context.Context, req CreateIndexRequest) error {
	baseURL := c.GetBaseURL()
	if baseURL == "" {
		return fmt.Errorf("error: environment is empty")
	}
	url := fmt.Sprintf("%s/databases", baseURL)

	// marshal request
	body, err := json.Marshal(req)
	if err != nil {
		return err
	}
	payload := strings.NewReader(string(body))

	httpReq, err := http.NewRequest("POST", url, payload)
	if err != nil {
		return err
	}
	httpReq.Header.Add("accept", "text/plain; charset=utf-8")
	httpReq.Header.Add("content-type", "application/json")
	httpReq.Header.Add("Api-Key", c.APIKey)
	res, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return err
	}
	if res.StatusCode >= http.StatusBadRequest {
		return fmt.Errorf("error: status code: %d", res.StatusCode)
	}

	defer res.Body.Close()
	return nil
}

// DescribeIndexResponse is the response of DescribeIndex
type DescribeIndexResponse struct {
	Database DescribeDatabaseResponse `json:"database"`
	Status   DescribeStatusResponse   `json:"status"`
}

type DescribeDatabaseResponse struct {
	Name           string          `json:"name"`
	Metric         Metric          `json:"metric"`
	Dimension      int             `json:"dimension"`
	Replicas       int             `json:"replicas"`
	Shards         int             `json:"shards"`
	Pods           int             `json:"pods"`
	PodType        PodType         `json:"pod_type"`
	MetadataConfig *MetadataConfig `json:"metadata_config,omitempty"`
}

type DescribeStatusResponse struct {
	Waiting []string `json:"waiting"`
	Crashed []string `json:"crashed"`
	Host    string   `json:"host"`
	Port    int      `json:"port"`
	State   string   `json:"state"`
	Ready   bool     `json:"ready"`
}

// DescribeIndex describes an index
func (c *PineconeClient) DescribeIndex(ctx context.Context, indexName string) (*DescribeIndexResponse, error) {
	baseURL := c.GetBaseURL()
	if baseURL == "" {
		return nil, fmt.Errorf("error: environment is empty")
	}
	url := fmt.Sprintf("%s/databases/%s", baseURL, indexName)
	httpReq, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	httpReq.Header.Add("accept", "application/json")
	httpReq.Header.Add("Api-Key", c.APIKey)
	res, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusNotFound {
		return nil, nil
	}

	if res.StatusCode >= http.StatusBadRequest {
		return nil, fmt.Errorf("error: index %s status code: %d", indexName, res.StatusCode)
	}
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	var resp DescribeIndexResponse
	// unmarshal response
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// DeleteIndex deletes an index
func (c *PineconeClient) DeleteIndex(ctx context.Context, indexName string) error {
	baseURL := c.GetBaseURL()
	if baseURL == "" {
		return fmt.Errorf("error: environment is empty")
	}
	url := fmt.Sprintf("%s/databases/%s", baseURL, indexName)
	httpReq, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return err
	}
	httpReq.Header.Add("accept", "text/plain")
	httpReq.Header.Add("Api-Key", c.APIKey)
	res, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return err
	}
	if res.StatusCode >= http.StatusBadRequest {
		return fmt.Errorf("error: status code is greater than or equal to %d: %d", http.StatusBadRequest, res.StatusCode)
	}
	defer res.Body.Close()
	return nil
}

type ConfigureIndexRequest struct {
	Replicas int     `json:"replicas"`
	PodType  PodType `json:"pod_type"`
}

// ConfigureIndex configures an index
func (c *PineconeClient) ConfigureIndex(ctx context.Context, indexName string, req ConfigureIndexRequest) error {
	baseURL := c.GetBaseURL()
	if baseURL == "" {
		return fmt.Errorf("error: environment is empty")
	}
	url := fmt.Sprintf("%s/databases/%s", baseURL, indexName)
	// marshal request
	body, err := json.Marshal(req)
	if err != nil {
		return err
	}
	payload := strings.NewReader(string(body))

	httpReq, err := http.NewRequest("PATCH", url, payload)
	if err != nil {
		return err
	}
	httpReq.Header.Add("accept", "text/plain")
	httpReq.Header.Add("content-type", "application/json")
	httpReq.Header.Add("Api-Key", c.APIKey)
	res, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return err
	}
	if res.StatusCode >= http.StatusBadRequest {
		return fmt.Errorf("error: status code is not 200: %d", res.StatusCode)
	}
	defer res.Body.Close()
	return nil
}
