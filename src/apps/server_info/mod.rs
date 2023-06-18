use std::sync::Arc;

use crate::{apps::App, http::ApiResponse};
use axum::{
    body::HttpBody,
    http::Request,
    response::{IntoResponse, Response},
    routing::get,
    Router,
};
use serde::{Deserialize, Serialize};

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

impl App for ServerInfoApp {
    fn api_routes(self: Arc<Self>) -> Router {
        Router::new().route("/server_info", get(get_server_info))
    }
}
