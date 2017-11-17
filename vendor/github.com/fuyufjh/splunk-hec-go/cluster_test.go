package hec

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	testSplunkURLs = []string{"http://127.0.0.1:8088", "http://localhost:8088"}
)

func TestCluster_WriteEvent(t *testing.T) {
	event := &Event{
		Index:      String("main"),
		Source:     String("test-hec-raw"),
		SourceType: String("manual"),
		Host:       String("localhost"),
		Time:       String("1485237827.123"),
		Event:      String("hello, world"),
	}

	c := NewCluster(testSplunkURLs, testSplunkToken)
	c.SetHTTPClient(testHttpClient)
	err := c.WriteEvent(event)
	assert.NoError(t, err)
}

func TestCluster_WriteEventBatch(t *testing.T) {
	eventBatches := [][]*Event{
		{
			{Event: "event one"},
			{Event: "event two"},
		},
		{
			{Event: "event foo"},
			{Event: "event bar"},
		},
	}

	c := NewCluster(testSplunkURLs, testSplunkToken)
	c.SetHTTPClient(testHttpClient)
	for _, batch := range eventBatches {
		err := c.WriteBatch(batch)
		assert.NoError(t, err)
	}
}

func TestCluster_WriteEventRaw(t *testing.T) {
	eventBlocks := []string{
		`2017-01-24T06:07:10.488Z Raw event one
2017-01-24T06:07:12.434Z Raw event two`,
		`2017-01-24T06:07:10.488Z Raw event foo
2017-01-24T06:07:12.434Z Raw event bar`,
	}
	metadata := EventMetadata{
		Source: String("test-hec-raw"),
	}
	c := NewCluster(testSplunkURLs, testSplunkToken)
	c.SetHTTPClient(testHttpClient)
	for _, block := range eventBlocks {
		err := c.WriteRaw(strings.NewReader(block), &metadata)
		assert.NoError(t, err)
	}
}

func TestCluster_Retrying(t *testing.T) {
	event := &Event{Event: "test retrying"}
	partlyBrokenUrls := []string{"http://127.0.0.1:8088", "http://example.com:8088", "http://example.com:88"}
	c := NewCluster(partlyBrokenUrls, testSplunkToken)
	c.SetHTTPClient(testHttpClient)
	for i := 0; i < 5; i++ {
		err := c.WriteEvent(event)
		assert.NoError(t, err)
	}
}
