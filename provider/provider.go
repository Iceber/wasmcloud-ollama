package ollamaprovider

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/jmorganca/ollama/api"
	"github.com/vmihailenco/msgpack/v5"
	provider "github.com/wasmCloud/provider-sdk-go"
	core "github.com/wasmcloud/interfaces/core/tinygo"
)

var ProviderContract = "ollama:llm"

type OllamaAdaptor interface {
	Chat(context.Context, *api.ChatRequest, api.ChatResponseFunc) error
	List(context.Context) (*api.ListResponse, error)
	Show(context.Context, *api.ShowRequest) (*api.ShowResponse, error)
}

type ollamaProvider struct {
	ollama OllamaAdaptor

	provider provider.WasmcloudProvider
	Logger   logr.Logger
}

func New(adaptor OllamaAdaptor) (*provider.WasmcloudProvider, error) {
	p := ollamaProvider{ollama: adaptor}

	provider, err := provider.New(ProviderContract,
		provider.WithProviderActionFunc(p.HandleAction),
		provider.WithNewLinkFunc(func(core.LinkDefinition) error { return nil }),
		provider.WithDelLinkFunc(func(core.LinkDefinition) error { return nil }),
		provider.WithShutdownFunc(func() error { return nil }),
	)
	if err != nil {
		return nil, err
	}
	p.Logger = provider.Logger
	return provider, nil
}

func (h *ollamaProvider) HostData() core.HostData {
	return h.provider.HostData()
}

func (h *ollamaProvider) HandleAction(action provider.ProviderAction) (*provider.ProviderResponse, error) {
	h.Logger.Info("handle action", "action", action.Operation)

	var err error
	var actionResp ProviderActionResponse

	var ctx = context.Background()
	switch action.Operation {
	case "Llm.Chat":
		var arequest ChatRequest
		if err = msgpack.Unmarshal(action.Msg, &arequest); err == nil {
			actionResp = h.handleChatRequest(ctx, arequest)
		}

	case "Llm.Show":
		var arequest ShowRequest
		if err = msgpack.Unmarshal(action.Msg, &arequest); err == nil {
			actionResp = h.handleShowRequest(ctx, arequest)
		}

	case "Llm.List":
		actionResp = h.handleListRequest(ctx)

	default:
		err = fmt.Errorf("Invalid method name: %s", action.Operation)
	}

	response := &provider.ProviderResponse{}
	if err != nil {
		response.Error = err.Error()
	} else {
		if response.Msg, err = msgpack.Marshal(actionResp); err != nil {
			response.Error = err.Error()
		}
	}
	return response, nil
}

func (h *ollamaProvider) handleChatRequest(ctx context.Context, arequest ChatRequest) ProviderActionResponse {
	var stream bool
	request := api.ChatRequest{
		Model:  arequest.Model,
		Stream: &stream,
		Format: arequest.Format,
	}
	for _, msg := range arequest.Messages {
		request.Messages = append(request.Messages, api.Message{
			Role:    msg.Role,
			Content: msg.Content,
		})
	}

	var aresponse ChatResponse
	fn := func(oresponse api.ChatResponse) error {
		aresponse = ChatResponse{
			Model:     oresponse.Model,
			CreatedAt: uint64(oresponse.CreatedAt.UnixMilli()),
			Message: Message{
				Role:    oresponse.Message.Role,
				Content: oresponse.Message.Content,
			},
			Done: oresponse.Done,
			Metrics: Metrics{
				TotalDuration:      uint64(oresponse.TotalDuration.Milliseconds()),
				LoadDuration:       uint64(oresponse.LoadDuration.Milliseconds()),
				PromptEvalCount:    uint32(oresponse.PromptEvalCount),
				PromptEvalDuration: uint64(oresponse.PromptEvalDuration.Milliseconds()),
				EvalCount:          uint32(oresponse.EvalCount),
				EvalDuration:       uint64(oresponse.EvalDuration.Milliseconds()),
			},
		}
		return nil
	}

	err := h.ollama.Chat(ctx, &request, fn)
	if statusErr := checkResponseError(err); statusErr != nil {
		return ProviderActionResponse{Err: statusErr}
	}
	return ProviderActionResponse{Ok: aresponse}
}

func (h *ollamaProvider) handleShowRequest(ctx context.Context, arequest ShowRequest) ProviderActionResponse {
	request := api.ShowRequest{
		Name:     arequest.Name,
		Model:    arequest.Model,
		System:   arequest.System,
		Template: arequest.Template,
	}
	oresponse, err := h.ollama.Show(ctx, &request)
	if statusErr := checkResponseError(err); statusErr != nil {
		return ProviderActionResponse{Err: statusErr}
	}

	aresponse := ShowResponse{
		License:    oresponse.License,
		Modelfile:  oresponse.Modelfile,
		Parameters: oresponse.Parameters,
		Template:   oresponse.Parameters,
		System:     oresponse.System,
		Details: ModelDetails{
			Format:            oresponse.Details.Format,
			Family:            oresponse.Details.Family,
			Families:          oresponse.Details.Families,
			ParameterSize:     oresponse.Details.ParameterSize,
			QuantizationLevel: oresponse.Details.QuantizationLevel,
		},
	}
	return ProviderActionResponse{Ok: aresponse}
}

func (h *ollamaProvider) handleListRequest(ctx context.Context) ProviderActionResponse {
	oresponse, err := h.ollama.List(ctx)
	if statusErr := checkResponseError(err); statusErr != nil {
		return ProviderActionResponse{Err: statusErr}
	}

	aresponse := ListResponse{
		Models: make([]ModelResponse, 0),
	}
	for _, model := range oresponse.Models {
		// TODO(Iceber): Modify families field in ollama.wit to Opional
		families := make([]string, 0)
		if model.Details.Families != nil {
			families = model.Details.Families
		}
		aresponse.Models = append(aresponse.Models, ModelResponse{
			Name:       model.Name,
			ModifiedAt: uint64(model.ModifiedAt.UnixMilli()),
			Size:       uint64(model.Size),
			Digest:     model.Digest,
			Details: ModelDetails{
				Format:            model.Details.Format,
				Family:            model.Details.Family,
				Families:          families,
				ParameterSize:     model.Details.ParameterSize,
				QuantizationLevel: model.Details.QuantizationLevel,
			},
		})
	}

	return ProviderActionResponse{Ok: aresponse}
}

func checkResponseError(err error) *StatusError {
	if err == nil {
		return nil
	}

	statusErr := &StatusError{}
	if e, ok := err.(api.StatusError); ok {
		statusErr.Error = e.ErrorMessage
		statusErr.Status = e.Status
		statusErr.StatusCode = e.StatusCode
	} else {
		statusErr.Error = err.Error()
		statusErr.Status = "FailedOllamaRequest"
		statusErr.StatusCode = 0
	}
	return statusErr
}
