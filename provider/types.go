package ollamaprovider

type ProviderActionResponse struct {
	Ok  interface{}  `msgpack:"Ok,omitempty"`
	Err *StatusError `msgpack:"Err,omitempty"`
}

type StatusError struct {
	StatusCode int    `msgpack:"statusCode"`
	Status     string `msgpack:"status"`
	Error      string `msgpack:"error"`
}

type ChatRequest struct {
	Model    string    `msgpack:"model"`
	Messages []Message `msgpack:"messages"`
	Stream   *bool     `msgpack:"withStream"`
	Format   string    `msgpack:"format"`

	Options []OptionKV `msgpack:"options"`
}

type OptionKV struct {
	_msgpack struct{} `msgpack:",as_array"`

	Key   string
	Value string
}

type Message struct {
	Role    string `msgpack:"role"` // one of ["system", "user", "assistant"]
	Content string `msgpack:"content"`
}

type ChatResponse struct {
	Model     string  `msgpack:"model"`
	CreatedAt uint64  `msgpack:"createdAt"`
	Message   Message `msgpack:"message"`

	Done bool `msgpack:"done"`

	Metrics `msgpack:",inline"`
}

type Metrics struct {
	TotalDuration      uint64 `msgpack:"totalDuration"`
	LoadDuration       uint64 `msgpack:"loadDuration"`
	PromptEvalCount    uint32 `msgpack:"promptEvalCount"`
	PromptEvalDuration uint64 `msgpack:"promptEvalDuration"`
	EvalCount          uint32 `msgpack:"evalCount"`
	EvalDuration       uint64 `msgpack:"evalDuration"`
}

type ShowRequest struct {
	Name     string `msgpack:"name"`
	Model    string `msgpack:"model"`
	System   string `msgpack:"system"`
	Template string `msgpack:"template"`

	Options []OptionKV `msgpack:"options"`
}

type ShowResponse struct {
	License    string       `msgpack:"license"`
	Modelfile  string       `msgpack:"modelfile"`
	Parameters string       `msgpack:"parameters"`
	Template   string       `msgpack:"template"`
	System     string       `msgpack:"system"`
	Details    ModelDetails `msgpack:"details"`
}

type ListResponse struct {
	Models []ModelResponse `msgpack:"models"`
}

type ModelResponse struct {
	Name       string       `msgpack:"name"`
	ModifiedAt uint64       `msgpack:"modifiedAt"`
	Size       uint64       `msgpack:"size"`
	Digest     string       `msgpack:"digest"`
	Details    ModelDetails `msgpack:"details"`
}

type ModelDetails struct {
	Format            string   `msgpack:"format"`
	Family            string   `msgpack:"family"`
	Families          []string `msgpack:"families"`
	ParameterSize     string   `msgpack:"parameterSize"`
	QuantizationLevel string   `msgpack:"quantizationLevel"`
}
