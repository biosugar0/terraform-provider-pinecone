package pinecone

import (
	"fmt"
	"reflect"
	"testing"
)

func TestNewPodType(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected PodType
		err      error
	}{
		{
			name:     "Valid pod type",
			input:    "s1.x1",
			expected: PodType{Class: "s1", Size: "x1"},
			err:      nil,
		},
		{
			name:     "Invalid pod type",
			input:    "invalidpodtype",
			expected: PodType{},
			err:      fmt.Errorf("error: invalid pod type: invalidpodtype"),
		},
		{
			name:     "Invalid pod class",
			input:    "s2.x2",
			expected: PodType{},
			err:      fmt.Errorf("error: invalid pod class: s2"),
		},
		{
			name:     "Invalid pod size",
			input:    "s1.x3",
			expected: PodType{},
			err:      fmt.Errorf("error: invalid pod size: x3"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := NewPodType(tc.input)

			if err != nil {
				if tc.err == nil {
					t.Fatalf("test '%s' failed: expected no error, but received error: %v", tc.name, err)
				}
				if err.Error() != tc.err.Error() {
					t.Fatalf("test '%s' failed: expected error %v, but received error %v", tc.name, tc.err, err)
				}
			} else {
				if tc.err != nil {
					t.Fatalf("test '%s' failed: expected error %v, but received no error", tc.name, tc.err)
				}
				if result != tc.expected {
					t.Fatalf("test '%s' failed: expected %v, but received %v", tc.name, tc.expected, result)
				}
			}
		})
	}
}

func TestMetadataConfig(t *testing.T) {
	testCases := []struct {
		name     string
		indexed  []string
		expected MetadataConfig
	}{
		{
			name:     "MetadataConfig with one indexed field",
			indexed:  []string{"field1"},
			expected: MetadataConfig{Indexed: []string{"field1"}},
		},
		{
			name:     "MetadataConfig with multiple indexed fields",
			indexed:  []string{"field1", "field2", "field3"},
			expected: MetadataConfig{Indexed: []string{"field1", "field2", "field3"}},
		},
		{
			name:     "MetadataConfig with no indexed fields",
			indexed:  []string{},
			expected: MetadataConfig{Indexed: []string{}},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create new MetadataConfig instance
			mc := MetadataConfig{Indexed: tc.indexed}

			// Compare the created MetadataConfig instance with the expected one
			if !reflect.DeepEqual(mc, tc.expected) {
				t.Fatalf("test '%s' failed: expected %v, but received %v", tc.name, tc.expected, mc)
			}
		})
	}
}

func TestNewMetric(t *testing.T) {
	testCases := []struct {
		metricStr string
		expected  Metric
		err       error
	}{
		{metricStr: "euclidean", expected: MetricEuclidean, err: nil},
		{metricStr: "cosine", expected: MetricCosine, err: nil},
		{metricStr: "dotproduct", expected: MetricDotProduct, err: nil},
		{metricStr: "manhattan", expected: MetricEuclidean, err: fmt.Errorf("error: invalid metric value: manhattan")},
	}

	for _, tC := range testCases {
		actual, err := NewMetric(tC.metricStr)
		if actual != tC.expected || (err == nil && tC.err != nil) || (err != nil && tC.err == nil) {
			t.Errorf("TestNewMetric(%s) = (%v, %v), expected (%v, %v)", tC.metricStr, actual, err, tC.expected, tC.err)
		}
	}
}

func TestMarshalJSON(t *testing.T) {
	testCases := []struct {
		metric   Metric
		expected []byte
		err      error
	}{
		{metric: MetricEuclidean, expected: []byte(`"euclidean"`)},
		{metric: MetricCosine, expected: []byte(`"cosine"`)},
		{metric: MetricDotProduct, expected: []byte(`"dotproduct"`)},
	}

	for _, tC := range testCases {
		actual, err := tC.metric.MarshalJSON()
		if string(actual) != string(tC.expected) || (err == nil && tC.err != nil) || (err != nil && tC.err == nil) {
			t.Errorf("TestMarshalJSON(%v) = (%v, %v), expected (%v, %v)", tC.metric, actual, err, tC.expected, tC.err)
		}
	}
}

func TestUnmarshalJSON(t *testing.T) {
	testCases := []struct {
		jsonStr  string
		expected Metric
		err      error
	}{
		{jsonStr: `"euclidean"`, expected: MetricEuclidean, err: nil},
		{jsonStr: `"cosine"`, expected: MetricCosine, err: nil},
		{jsonStr: `"dotproduct"`, expected: MetricDotProduct, err: nil},
		{jsonStr: `"manhattan"`, expected: MetricEuclidean, err: fmt.Errorf("error: invalid metric value: manhattan")},
	}

	for _, tC := range testCases {
		var actual Metric
		err := actual.UnmarshalJSON([]byte(tC.jsonStr))
		if actual != tC.expected || (err == nil && tC.err != nil) || (err != nil && tC.err == nil) {
			t.Errorf("TestUnmarshalJSON(%s) = (%v, %v), expected (%v, %v)", tC.jsonStr, actual, err, tC.expected, tC.err)
		}
	}
}
