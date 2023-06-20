use std::sync::Arc;

use crate::{
    apps::App,
    comm::{discord::MIME_DISCORD, Message, MessageDigestor},
    config::Config,
    http::ApiResponse,
    AppState,
};
use anyhow::Result;
use async_trait::async_trait;
use axum::{
    body::HttpBody,
    http::Request,
    response::{IntoResponse, Response},
    routing::get,
    Router,
};
use lazy_static::lazy_static;
use serde::{Deserialize, Serialize};
use tokio::sync::Mutex;

pub struct ServerInfoApp {}

#[cfg(debug_assertions)]
const PROFILE: &str = "debug";
#[cfg(not(debug_assertions))]
const PROFILE: &str = "release";

#[derive(Debug, Serialize, Deserialize)]
pub struct ServerInfo {
    version: String,
    profile: String,
}

async fn get_server_info<B: HttpBody>(_req: Request<B>) -> Response
where
    <B as HttpBody>::Error: std::fmt::Debug,
{
    let server_info = ServerInfo {
        version: env!("CARGO_PKG_VERSION").to_string(),
        profile: PROFILE.to_string(),
    };
    ApiResponse::ok(
        "Server info retrieved successfully".to_string(),
        Some(server_info),
    )
    .into_response()
}

impl ServerInfoApp {
    pub fn new() -> Self {
        Self {}
    }
}

#[async_trait]
impl MessageDigestor for Arc<ServerInfoApp> {
    async fn digest(&self, message: &Message) -> Result<Option<Message>> {
        lazy_static! {
            static ref REGEXP_ASK_SERVER_INFO: regex::Regex =
                regex::Regex::new(r"server info$").unwrap();
        }
        if REGEXP_ASK_SERVER_INFO.is_match(&message.body) {
            let server_info = ServerInfo {
                version: env!("CARGO_PKG_VERSION").to_string(),
                profile: PROFILE.to_string(),
            };
            let ret = format!(
                "Server info:\nVersion: {}\nProfile: {}",
                server_info.version, server_info.profile
            );
            return Ok(Some(Message {
                body: ret,
                subject: "".to_string(),
                priority: 0,
                mime: MIME_DISCORD,
            }));
        }

        Ok(None)
    }
}

#[async_trait]
impl App for ServerInfoApp {
    async fn initialize(
        self: Arc<Self>,
        _config: &'static Config,
        _app_state: Arc<Mutex<AppState>>,
    ) {
    }
    fn api_routes(self: Arc<Self>) -> Router {
        Router::new().route("/server_info", get(get_server_info))
    }
    fn message_digestors(self: Arc<Self>) -> Vec<Box<dyn MessageDigestor + Send + Sync>> {
        vec![Box::new(self)]
    }
}
