use crate::{config::Config, AppState};
use axum::Router;

use std::{
    future::Future,
    pin::Pin,
    sync::{Arc, Mutex},
};

pub trait App {
    fn initialize(
        self: Arc<Self>,
        _config: &'static Config,
        _app_state: Arc<Mutex<AppState>>,
    ) -> Pin<Box<dyn Future<Output = ()>>> {
        Box::pin(async {})
    }
    fn api_routes(self: Arc<Self>) -> Router {
        Router::new()
    }
}

pub mod auth;
pub mod canvas_lms;
pub mod med;
pub mod server_info;
pub mod webcheck;
