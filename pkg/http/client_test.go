package http

import (
	"io"
	configPkg "main/pkg/config"
	"main/pkg/constants"
	loggerPkg "main/pkg/logger"
	"main/pkg/metrics"
	"net/http"
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/require"
	"gopkg.in/guregu/null.v4"
)

func TestHttpClientMalformedURL(t *testing.T) {
	t.Parallel()

	logger := loggerPkg.GetNopLogger()
	metricsManager := metrics.NewManager(*logger, configPkg.MetricsConfig{Enabled: null.BoolFrom(false)})
	client := NewClient(*logger, metricsManager, "://", "chain")

	result := make(map[string]interface{})
	err := client.Get("", constants.QueryTypeBlock, &result)
	require.Error(t, err)
}

//nolint:paralleltest // disabled due to httpmock usage
func TestHttpClientMalformedJson(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	logger := loggerPkg.GetNopLogger()
	metricsManager := metrics.NewManager(*logger, configPkg.MetricsConfig{Enabled: null.BoolFrom(false)})
	client := NewClient(*logger, metricsManager, "http://example.com", "chain")

	httpmock.RegisterResponder(
		"GET",
		"http://example.com/",
		httpmock.NewBytesResponder(200, []byte("invalid\"")),
	)

	result := make(map[string]interface{})
	err := client.Get("/", constants.QueryTypeBlock, &result)
	require.Error(t, err)
}

type MockReader struct{}

func (m MockReader) Read(p []byte) (int, error) {
	return 0, io.ErrUnexpectedEOF
}

func (m MockReader) Close() error {
	return nil
}

//nolint:paralleltest // disabled due to httpmock usage
func TestHttpClientErrorReading(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	logger := loggerPkg.GetNopLogger()
	metricsManager := metrics.NewManager(*logger, configPkg.MetricsConfig{Enabled: null.BoolFrom(false)})
	client := NewClient(*logger, metricsManager, "http://example.com", "chain")

	httpmock.RegisterResponder(
		"GET",
		"http://example.com/",
		httpmock.ResponderFromResponse(&http.Response{
			ContentLength: 10,
			Body:          MockReader{},
			Header: http.Header{
				"Content-Type":   []string{"application/json"},
				"Content-Length": []string{"10"},
			},
		}),
	)

	_, err := client.GetPlain("/", constants.QueryTypeBlock, map[string]string{})
	require.Error(t, err)
}
