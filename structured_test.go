package main

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_structuredWriter_JSON_singleObject(t *testing.T) {
	var out bytes.Buffer
	sw := &structuredWriter{format: "json", contextName: "prod-us", w: &out}

	_, err := sw.Write([]byte(`{"kind":"Pod","metadata":{"name":"nginx"}}`))
	require.NoError(t, err)
	require.NoError(t, sw.Close())

	var got map[string]any
	require.NoError(t, json.Unmarshal(bytes.TrimSpace(out.Bytes()), &got))
	assert.Equal(t, "prod-us", got["context"])
	assert.Equal(t, "Pod", got["kind"])
}

func Test_structuredWriter_JSON_list(t *testing.T) {
	var out bytes.Buffer
	sw := &structuredWriter{format: "json", contextName: "staging", w: &out}

	input := `{"apiVersion":"v1","kind":"List","items":[{"kind":"Pod","metadata":{"name":"a"}},{"kind":"Pod","metadata":{"name":"b"}}]}`
	_, err := sw.Write([]byte(input))
	require.NoError(t, err)
	require.NoError(t, sw.Close())

	lines := strings.Split(strings.TrimSpace(out.String()), "\n")
	assert.Len(t, lines, 2)

	for _, line := range lines {
		var obj map[string]any
		require.NoError(t, json.Unmarshal([]byte(line), &obj))
		assert.Equal(t, "staging", obj["context"])
		assert.Equal(t, "Pod", obj["kind"])
	}
}

func Test_structuredWriter_JSON_array(t *testing.T) {
	var out bytes.Buffer
	sw := &structuredWriter{format: "json", contextName: "dev", w: &out}

	input := `[{"name":"a"},{"name":"b"}]`
	_, err := sw.Write([]byte(input))
	require.NoError(t, err)
	require.NoError(t, sw.Close())

	lines := strings.Split(strings.TrimSpace(out.String()), "\n")
	assert.Len(t, lines, 2)

	for _, line := range lines {
		var obj map[string]any
		require.NoError(t, json.Unmarshal([]byte(line), &obj))
		assert.Equal(t, "dev", obj["context"])
	}
}

func Test_structuredWriter_JSON_emptyInput(t *testing.T) {
	var out bytes.Buffer
	sw := &structuredWriter{format: "json", contextName: "ctx", w: &out}

	require.NoError(t, sw.Close())
	assert.Empty(t, out.String())
}

func Test_structuredWriter_JSON_invalidFallback(t *testing.T) {
	var out bytes.Buffer
	sw := &structuredWriter{format: "json", contextName: "ctx", w: &out}

	_, err := sw.Write([]byte("not json at all"))
	require.NoError(t, err)
	require.NoError(t, sw.Close())
	assert.Equal(t, "not json at all", out.String())
}

func Test_structuredWriter_YAML_singleDoc(t *testing.T) {
	var out bytes.Buffer
	sw := &structuredWriter{format: "yaml", contextName: "prod-eu", w: &out}

	input := "kind: Pod\nmetadata:\n  name: nginx\n"
	_, err := sw.Write([]byte(input))
	require.NoError(t, err)
	require.NoError(t, sw.Close())

	assert.Contains(t, out.String(), "context: prod-eu")
	assert.Contains(t, out.String(), "kind: Pod")
}

func Test_structuredWriter_YAML_multiDoc(t *testing.T) {
	var out bytes.Buffer
	sw := &structuredWriter{format: "yaml", contextName: "multi", w: &out}

	input := "kind: Pod\nmetadata:\n  name: a\n---\nkind: Service\nmetadata:\n  name: b\n"
	_, err := sw.Write([]byte(input))
	require.NoError(t, err)
	require.NoError(t, sw.Close())

	assert.Contains(t, out.String(), "---")
	assert.Equal(t, 2, strings.Count(out.String(), "context: multi"))
}

func Test_structuredWriter_YAML_emptyInput(t *testing.T) {
	var out bytes.Buffer
	sw := &structuredWriter{format: "yaml", contextName: "ctx", w: &out}

	require.NoError(t, sw.Close())
	assert.Empty(t, out.String())
}

func Test_structuredWriter_YAML_invalidFallback(t *testing.T) {
	var out bytes.Buffer
	sw := &structuredWriter{format: "yaml", contextName: "ctx", w: &out}

	_, err := sw.Write([]byte(":::invalid yaml[[["))
	require.NoError(t, err)
	require.NoError(t, sw.Close())
	assert.Equal(t, ":::invalid yaml[[[", out.String())
}
