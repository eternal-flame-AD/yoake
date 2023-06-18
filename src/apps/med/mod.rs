use std::{
    future::Future,
    pin::Pin,
    sync::{Arc, Mutex},
};

use crate::{config::Config, AppState};

use super::App;

use axum::{
    routing::{delete, get, post},
    Extension, Router,
};
use tokio::sync::Mutex as AsyncMutex;

mod directive;
mod log;

pub struct MedManagementApp {
    state: AsyncMutex<Option<MedManagementAppState>>,
}

struct MedManagementAppState {
    global_app_state: Arc<Mutex<AppState>>,
}

impl MedManagementApp {
    pub fn new() -> Self {
        Self {
            state: AsyncMutex::new(None),
        }
    }
}

impl App for MedManagementApp {
    fn initialize(
        self: Arc<Self>,
        _config: &'static Config,
        app_state: Arc<Mutex<AppState>>,
    ) -> Pin<Box<dyn Future<Output = ()>>> {
        Box::pin(async move {
            let mut state = self.state.lock().await;
            *state = Some(MedManagementAppState {
                global_app_state: app_state,
            });
        })
    }

    fn api_routes(self: Arc<Self>) -> Router {
        Router::new()
            .route(
                "/med/parse_shorthand",
                post(directive::route_parse_shorthand),
            )
            .route(
                "/med/format_shorthand",
                post(directive::route_format_shorthand),
            )
            .route(
                "/med/directive",
                get(directive::route_get_directive)
                    .post(directive::route_post_directive)
                    .patch(directive::route_patch_directive),
            )
            .route(
                "/med/directive/:med_uuid",
                delete(directive::route_delete_directive),
            )
            .route(
                "/med/directive/:med_uuid/project_next_dose",
                get(log::route_project_next_dose),
            )
            .route(
                "/med/directive/:med_uuid/log",
                get(log::route_get_log).post(log::route_post_log),
            )
            .route(
                "/med/directive/:med_uuid/log/:log_uuid",
                delete(log::route_delete_log),
            )
            .layer(Extension(self.clone()))
    }
}
