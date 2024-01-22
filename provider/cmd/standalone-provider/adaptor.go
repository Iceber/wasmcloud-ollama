package main

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jmorganca/ollama/api"
	"github.com/jmorganca/ollama/cmd"
	"github.com/jmorganca/ollama/server"

	ollamaprovider "github.com/iceber/wasmcloud-ollama-provider"
)

type OllamaServer struct {
	r *gin.Engine
}

func (s *OllamaServer) Run() error {
	return cmd.RunServer(nil, nil)
}

var _ ollamaprovider.OllamaAdaptor = &OllamaServer{}

func (s *OllamaServer) Chat(ctx context.Context, request *api.ChatRequest, fn api.ChatResponseFunc) error {
	panic("no implemented")
}

func (s *OllamaServer) Show(ctx context.Context, request *api.ShowRequest) (*api.ShowResponse, error) {
	var buf bytes.Buffer
	writer := &responseWriter{body: &buf}
	context := gin.CreateTestContextOnly(writer, s.r)

	data, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}

	context.Request, _ = http.NewRequestWithContext(ctx, "", "", bytes.NewReader(data))

	server.ShowModelHandler(context)
	if writer.code >= 400 {
		apiError := api.StatusError{StatusCode: writer.code, Status: http.StatusText(writer.code)}
		if err := json.Unmarshal(buf.Bytes(), &apiError); err != nil {
			apiError.ErrorMessage = err.Error()
			return nil, apiError
		}

		if apiError.ErrorMessage == "" {
			apiError.ErrorMessage = context.Errors.String()
		}
		return nil, apiError
	}

	var response api.ShowResponse
	if err := json.Unmarshal(buf.Bytes(), &response); err != nil {
		return nil, err
	}
	return &response, nil
}

func (s *OllamaServer) List(ctx context.Context) (*api.ListResponse, error) {
	var buf bytes.Buffer
	writer := &responseWriter{body: &buf}
	context := gin.CreateTestContextOnly(writer, s.r)

	context.Request, _ = http.NewRequestWithContext(ctx, "", "", nil)
	server.ListModelsHandler(context)
	if writer.code >= 400 {
		apiError := api.StatusError{StatusCode: writer.code, Status: http.StatusText(writer.code)}
		if err := json.Unmarshal(buf.Bytes(), &apiError); err != nil {
			apiError.ErrorMessage = err.Error()
			return nil, apiError
		}

		if apiError.ErrorMessage == "" {
			apiError.ErrorMessage = context.Errors.String()
		}
		return nil, apiError
	}

	var response api.ListResponse
	if err := json.Unmarshal(buf.Bytes(), &response); err != nil {
		return nil, err
	}
	return &response, nil
}

type responseWriter struct {
	code        int
	body        io.Writer
	emptyHeader http.Header
}

func (w *responseWriter) WriteHeader(code int) {
	w.code = code
}

func (w *responseWriter) Header() http.Header {
	return w.emptyHeader
}

func (w *responseWriter) Write(data []byte) (int, error) {
	return w.body.Write(data)
}
