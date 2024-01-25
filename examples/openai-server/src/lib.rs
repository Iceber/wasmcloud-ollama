use core::time;

use serde::{Deserialize, Serialize};

wit_bindgen::generate!({
    world: "interfaces",
    exports: {
        "wasi:http/incoming-handler": HttpServer,
    },
});

use exports::wasi::http::incoming_handler;
use ollama::llm::llm;
use wasi::http::types::*;
use wasi::io::streams::StreamError;

use endpoints as openai_api;

struct HttpServer {}

impl incoming_handler::Guest for HttpServer {
    fn handle(request: IncomingRequest, response_out: ResponseOutparam) {
        let path = request.path_with_query().unwrap();
        let path = path.parse::<http::Uri>().unwrap();
        let (code, headers, body) = match path.path() {
            "/echo" => (200, Vec::new(), Vec::<u8>::from("echo test")),
            "/v1/models" => handle_list_models(),
            "/v1/chat/completions" => handle_openai_request(request),
            _ => (400, Vec::new(), Vec::<u8>::from("not found")),
        };

        let fields = Fields::new();
        headers.iter().for_each(|item| {
            let _ = fields.append(&item.0, &Vec::from(item.1.as_bytes()));
        });

        let response = OutgoingResponse::new(fields);
        // response.set_status_code(code).unwrap();
        let response_body = response.body().unwrap();
        response_body
            .write()
            .unwrap()
            .blocking_write_and_flush(&body)
            .unwrap();
        OutgoingBody::finish(response_body, None).expect("failed to finish response body");

        ResponseOutparam::set(response_out, Ok(response));
    }
}

fn handle_list_models() -> (u16, Vec<(String, String)>, Vec<u8>) {
    match llm::list() {
        Ok(models) => {
            let m = models
                .models
                .iter()
                .map(|m| openai_api::models::Model {
                    id: m.name.clone(),
                    created: time::Duration::from_millis(m.modified_at).as_secs(),
                    object: "model".to_string(),
                    owned_by: "Not specified".to_string(),
                })
                .collect();
            let response = openai_api::models::ListModelsResponse {
                object: "list".to_string(),
                data: m,
            };

            let data = serde_json::to_vec(&response).unwrap();
            (200, Vec::new(), data)
        }
        Err(status_error) => (
            status_error.status_code as u16,
            Vec::new(),
            Vec::from(status_error.error.as_bytes()),
        ),
    }
}

fn get_data_from_request(request: IncomingRequest) -> Result<Vec<u8>, StreamError> {
    let mut body_data = Vec::new();
    if let Ok(body) = request.consume() {
        if let Ok(stream) = body.stream() {
            loop {
                match stream.blocking_read(16 * 1024) {
                    Ok(buffer) => {
                        if buffer.is_empty() {
                            continue;
                        }
                        body_data.extend(buffer);
                        continue;
                    }
                    Err(StreamError::Closed) => {}
                    Err(StreamError::LastOperationFailed(error)) => {}
                }
                break;
            }
        } else {
        }
    }
    Ok(body_data)
}

fn handle_openai_request(request: IncomingRequest) -> (u16, Vec<(String, String)>, Vec<u8>) {
    let data = get_data_from_request(request).unwrap();
    let chat_request: openai_api::chat::ChatCompletionRequest =
        serde_json::from_slice(&data).unwrap();

    /*
    let format = match chat_request.response_format {
        None => String::from(""),
        Some(format) => match format.r#type {
            openai_api::ChatCompletionResponseFormatType::Text => String::from("text"),
            openai_api::ChatCompletionResponseFormatType::JsonObject => String::from("json"),
        },
    };
    */
    let format = "".to_string();

    let messages = chat_request
        .messages
        .iter()
        .filter_map(|m| {
            let role = match m.role {
                openai_api::chat::ChatCompletionRole::System => "system".to_string(),
                openai_api::chat::ChatCompletionRole::User => "user".to_string(),
                openai_api::chat::ChatCompletionRole::Assistant => "assistant".to_string(),
                _ => return None,
            };
            Some(llm::Message {
                role,
                content: m.content.clone(),
            })
        })
        .collect();

    let mut options = vec![];
    if let Some(max_tokens) = chat_request.max_tokens {
        options.push(("max_tokens".to_string(), format!("{max_tokens}")))
    }
    if let Some(temperature) = chat_request.temperature {
        options.push(("temperature".to_string(), format!("{temperature}")))
    }
    if let Some(top_p) = chat_request.top_p {
        options.push(("top_p".to_string(), format!("{top_p}")))
    }
    if let Some(frequency_penalty) = chat_request.frequency_penalty {
        options.push((
            "frequency_penalty".to_string(),
            format!("{frequency_penalty}"),
        ))
    }

    let request = llm::ChatRequest {
        messages,
        model: chat_request.model.unwrap_or_default(),
        with_stream: Some(false),
        format,
        options,
    };

    match llm::chat(&request) {
        Err(err) => (
            err.status_code as u16,
            Vec::new(),
            serde_json::to_vec(&ErrorResponse {
                error: err.error.clone(),
                status: err.status.clone(),
            })
            .unwrap(),
        ),
        Ok(response) => {
            let openai_response = openai_api::chat::ChatCompletionObject {
                id: uuid::Uuid::new_v4().to_string(),
                object: "chat.completion".to_string(),
                model: response.model.clone(),
                created: time::Duration::from_millis(response.created_at).as_secs(),
                choices: vec![openai_api::chat::ChatCompletionObjectChoice {
                    message: openai_api::chat::ChatCompletionObjectMessage {
                        role: openai_api::chat::ChatCompletionRole::Assistant,
                        content: response.message.content.clone(),
                        function_call: None,
                    },
                    index: 0,
                    finish_reason: openai_api::common::FinishReason::stop,
                }],

                // Usage is optional, default shoudle be None
                usage: openai_api::common::Usage {
                    prompt_tokens: 0,
                    completion_tokens: 0,
                    total_tokens: 0,
                },
            };

            (
                200,
                Vec::new(),
                serde_json::to_vec(&openai_response).unwrap(),
            )
        }
    }
}

#[derive(Debug, Deserialize, Serialize, Clone, PartialEq)]
struct ErrorResponse {
    error: String,
    status: String,
}
