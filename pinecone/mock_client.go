package pinecone

import (
	"context"
	"fmt"
	"os"
	"sync"
)

var (
	_ PineconeClientInterface = &MockPineconeClient{}
)

type MockPineconeClient struct {
	APIKey      string
	Environment string
	indexes     map[string]*DescribeIndexResponse
	mutex       sync.Mutex
}

func NewMockClient(options ...Option) (*MockPineconeClient, error) {
	opts := &Options{
		APIKey:      os.Getenv("PINECONE_API_KEY"),
		Environment: os.Getenv("PINECONE_ENVIRONMENT"),
	}

	for _, option := range options {
		option(opts)
	}

	return &MockPineconeClient{
		APIKey:      opts.APIKey,
		Environment: opts.Environment,
		indexes:     make(map[string]*DescribeIndexResponse),
	}, nil
}

func (c *MockPineconeClient) GetBaseURL() string {
	if c.Environment == "" {
		return ""
	}
	return fmt.Sprintf("https://controller.%s.pinecone.io", c.Environment)
}

func (c *MockPineconeClient) GetAPIKey() string {
	return c.APIKey
}

func (c *MockPineconeClient) GetEnvironment() string {
	return c.Environment
}

func (c *MockPineconeClient) ListIndexes(ctx context.Context) ([]string, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	indexNames := make([]string, 0, len(c.indexes))
	for indexName := range c.indexes {
		indexNames = append(indexNames, indexName)
	}

	return indexNames, nil
}

func (c *MockPineconeClient) CreateIndex(ctx context.Context, req CreateIndexRequest) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// save the index
	c.indexes[req.Name] = &DescribeIndexResponse{
		Database: DescribeDatabaseResponse{
			Name:      req.Name,
			Metric:    req.Metric,
			Dimension: req.Dimension,
			Replicas:  req.Replicas,
			Shards:    1,
			Pods:      req.Pods,
			PodType:   req.PodType,
		},
		Status: DescribeStatusResponse{
			Ready: true,
		},
	}
	return nil
}

func (c *MockPineconeClient) DescribeIndex(ctx context.Context, indexName string) (*DescribeIndexResponse, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	index, exists := c.indexes[indexName]
	if !exists {
		return nil, nil
	}
	index.Database.Name = indexName
	return index, nil
}

func (c *MockPineconeClient) DeleteIndex(ctx context.Context, indexName string) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	delete(c.indexes, indexName)
	return nil
}

func (c *MockPineconeClient) ConfigureIndex(ctx context.Context, indexName string, req ConfigureIndexRequest) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	index, exists := c.indexes[indexName]
	if !exists {
		return fmt.Errorf("index not found: %s", indexName)
	}

	index.Database.Replicas = req.Replicas
	index.Database.PodType = req.PodType

	// save the index
	c.indexes[indexName] = index

	return nil
}
