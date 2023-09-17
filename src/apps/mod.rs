use crate::{comm::MessageDigestor, config::Config, AppState};
use async_trait::async_trait;
use axum::Router;
use std::sync::Arc;
use tokio::sync::Mutex;

#[async_trait]
pub trait App {
    async fn initialize(
        self: Arc<Self>,
        _config: &'static Config,
        _app_state: Arc<Mutex<AppState>>,
    );
    fn api_routes(self: Arc<Self>) -> Router {
        Router::new()
    }
    fn message_digestors(self: Arc<Self>) -> Vec<Box<dyn MessageDigestor + Send + Sync>> {
        vec![]
    }
}

pub mod auth;
pub mod canvas_lms;
pub mod jpn_wordbook;
pub mod med;
pub mod server_info;
pub mod webcheck;
