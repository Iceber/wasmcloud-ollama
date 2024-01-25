use serde::{Deserialize, Serialize};

wit_bindgen::generate!({
    world: "interfaces",
    exports: {
        "ollama:llm/llm": OllamaLlm,
    },
});

use exports::ollama::llm::llm::{self};

struct OllamaLlm;

impl llm::Guest for OllamaLlm {
    fn chat(request: llm::ChatRequest) -> Result<llm::ChatResponse, llm::StatusError> {
        let r = ChatRequest {
            model: request.model,
            messages: request
                .messages
                .iter()
                .map(|m| Message {
                    role: m.role.clone(),
                    content: m.content.clone(),
                })
                .collect(),
            with_stream: request.with_stream,
            format: request.format,
            options: request.options,
        };

        let payload = match rmp_serde::to_vec_named(&r) {
            Ok(payload) => payload,
            Err(err) => {
                return Err(llm::StatusError {
                    error: err.to_string(),
                    status_code: 0,
                    status: "PreWasmCloudCall".to_string(),
                })
            }
        };
        match wasmcloud::bus::host::call_sync(None, "ollama:llm/Llm.Chat", &payload) {
            Err(err) => Err(llm::StatusError {
                error: err.to_string(),
                status_code: 0,
                status: "FailedWasmCloudCall".to_string(),
            }),
            Ok(result) => {
                match rmp_serde::from_slice::<Result<ChatResponse, StatusError>>(&result) {
                    Ok(result) => match result {
                        Ok(response) => Ok(llm::ChatResponse {
                            model: response.model,
                            created_at: response.created_at,
                            message: llm::Message {
                                role: response.message.role,
                                content: response.message.content,
                            },
                            done: response.done,
                            total_duration: response.total_duration,
                            load_duration: response.load_duration,
                            prompt_eval_count: response.prompt_eval_count,
                            prompt_eval_duration: response.prompt_eval_duration,
                            eval_count: response.eval_count,
                            eval_duration: response.eval_duration,
                        }),
                        Err(status_error) => Err(llm::StatusError {
                            status_code: status_error.status_code,
                            status: status_error.status,
                            error: status_error.error,
                        }),
                    },
                    Err(error) => Err(llm::StatusError {
                        error: error.to_string(),
                        status: "InvalidWasmCloudCallResponse".to_string(),
                        status_code: 0,
                    }),
                }
            }
        }
    }

    fn show(request: llm::ShowRequest) -> Result<llm::ShowResponse, llm::StatusError> {
        let r = ShowRequest {
            name: request.name,
            model: request.model,
            system: request.system,
            template: request.template,
            options: request.options,
        };

        let payload = match rmp_serde::to_vec_named(&r) {
            Ok(payload) => payload,
            Err(err) => {
                return Err(llm::StatusError {
                    error: err.to_string(),
                    status_code: 0,
                    status: "PreWasmCloudCall".to_string(),
                })
            }
        };
        match wasmcloud::bus::host::call_sync(None, "ollama:llm/Llm.Show", &payload) {
            Err(err) => Err(llm::StatusError {
                error: err.to_string(),
                status_code: 0,
                status: "FailedWasmCloudCall".to_string(),
            }),
            Ok(result) => {
                match rmp_serde::from_slice::<Result<ShowResponse, StatusError>>(&result) {
                    Ok(result) => match result {
                        Ok(response) => Ok(llm::ShowResponse {
                            license: response.license,
                            modelfile: response.modelfile,
                            parameters: response.parameters,
                            template: response.template,
                            system: response.system,
                            details: llm::ModelDetails {
                                format: response.details.format,
                                family: response.details.family,
                                families: response.details.families,
                                parameter_size: response.details.parameter_size,
                                quantization_level: response.details.quantization_level,
                            },
                        }),
                        Err(status_error) => Err(llm::StatusError {
                            status_code: status_error.status_code,
                            status: status_error.status,
                            error: status_error.error,
                        }),
                    },
                    Err(error) => Err(llm::StatusError {
                        error: error.to_string(),
                        status: "InvalidWasmCloudCallResponse".to_string(),
                        status_code: 0,
                    }),
                }
            }
        }
    }

    fn list() -> Result<llm::ListResponse, llm::StatusError> {
        match wasmcloud::bus::host::call_sync(None, "ollama:llm/Llm.List", &vec![]) {
            Err(err) => Err(llm::StatusError {
                error: err.to_string(),
                status_code: 0,
                status: "FailedWasmCloudCall".to_string(),
            }),
            Ok(result) => {
                match rmp_serde::from_slice::<Result<ListResponse, StatusError>>(&result) {
                    Ok(result) => match result {
                        Ok(response) => Ok(llm::ListResponse {
                            models: response
                                .models
                                .iter()
                                .map(|m| llm::ModelResponse {
                                    name: m.name.clone(),
                                    modified_at: m.modified_at,
                                    size: m.size,
                                    digest: m.digest.clone(),
                                    details: llm::ModelDetails {
                                        format: m.details.format.clone(),
                                        family: m.details.family.clone(),
                                        families: m.details.families.clone(),
                                        parameter_size: m.details.parameter_size.clone(),
                                        quantization_level: m.details.quantization_level.clone(),
                                    },
                                })
                                .collect(),
                        }),
                        Err(status_error) => Err(llm::StatusError {
                            status_code: status_error.status_code,
                            status: status_error.status,
                            error: status_error.error,
                        }),
                    },
                    Err(error) => Err(llm::StatusError {
                        error: error.to_string(),
                        status: "InvalidWasmCloudCallResponse".to_string(),
                        status_code: 0,
                    }),
                }
            }
        }
    }
}

#[derive(Debug, Serialize, Deserialize)]
#[serde(rename_all = "camelCase")]
struct ChatRequest {
    model: String,
    messages: Vec<Message>,
    with_stream: Option<bool>,
    format: String,
    options: Vec<(String, String)>,
}

#[derive(Debug, Serialize, Deserialize)]
#[serde(rename_all = "camelCase")]
struct Message {
    role: String,
    content: String,
}

#[derive(Debug, Serialize, Deserialize)]
#[serde(rename_all = "camelCase")]
struct ChatResponse {
    model: String,
    created_at: u64,
    message: Message,
    done: bool,
    total_duration: u64,
    load_duration: u64,
    prompt_eval_count: u32,
    prompt_eval_duration: u32,
    eval_count: u32,
    eval_duration: u32,
}

#[derive(Debug, Serialize, Deserialize)]
#[serde(rename_all = "camelCase")]
struct StatusError {
    status_code: u32,
    status: String,
    error: String,
}

#[derive(Debug, Serialize, Deserialize)]
#[serde(rename_all = "camelCase")]
struct ShowRequest {
    name: String,
    model: String,
    system: String,
    template: String,
    options: Vec<(String, String)>,
}

#[derive(Debug, Serialize, Deserialize)]
#[serde(rename_all = "camelCase")]
struct ShowResponse {
    license: String,
    modelfile: String,
    parameters: String,
    template: String,
    system: String,
    details: ModelDetails,
}

#[derive(Debug, Serialize, Deserialize)]
#[serde(rename_all = "camelCase")]
struct ModelDetails {
    format: String,
    family: String,
    families: Vec<String>,
    parameter_size: String,
    quantization_level: String,
}

#[derive(Debug, Serialize, Deserialize)]
#[serde(rename_all = "camelCase")]
struct ModelResponse {
    name: String,
    modified_at: u64,
    size: u64,
    digest: String,
    details: ModelDetails,
}

#[derive(Debug, Serialize, Deserialize)]
#[serde(rename_all = "camelCase")]
struct ListResponse {
    models: Vec<ModelResponse>,
}
