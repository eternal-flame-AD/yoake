use std::sync::Arc;

use crate::{
    comm::{Message, MessageDigestor},
    config::Config,
    AppState,
};

use super::App;

use anyhow::Result;
use async_trait::async_trait;
use axum::{
    routing::{delete, get, post},
    Extension, Router,
};
use chrono::DateTime;
use lazy_static::lazy_static;
use serenity::utils::MessageBuilder;
use tokio::sync::Mutex;

mod directive;
mod log;

pub struct MedManagementApp {
    state: Mutex<Option<MedManagementAppState>>,
}

struct MedManagementAppState {
    global_app_state: Arc<Mutex<AppState>>,
}

impl MedManagementApp {
    pub fn new() -> Self {
        Self {
            state: Mutex::new(None),
        }
    }
}

pub fn format_relative_time<Tz: chrono::TimeZone>(
    time: chrono::DateTime<Tz>,
    relative_to: chrono::DateTime<Tz>,
) -> String {
    let duration = time.signed_duration_since(relative_to);
    let duration = chrono_humanize::HumanTime::from(duration);
    duration.to_string()
}

#[async_trait]
impl MessageDigestor for Arc<MedManagementApp> {
    async fn digest(&self, message: &Message) -> Result<Option<Message>> {
        lazy_static! {
            static ref REGEX_GET_DIRECTIVE: regex::Regex =
                regex::Regex::new(r"get med info$").unwrap();
        }
        if REGEX_GET_DIRECTIVE.is_match(&message.body) {
            let next_doses = log::project_next_doses(self).await?;

            let mut msg = MessageBuilder::new();
            msg.push_line("");

            for next_dose in next_doses.iter() {
                msg.push_bold(next_dose.0.name.to_string());
                msg.push_line(":");
                msg.push_line(format!("Offset: {:.2}", next_dose.1.dose_offset,));
                msg.push_line(format!(
                    "Next dose: {}",
                    format_relative_time(
                        DateTime::<chrono::Utc>::from_utc(next_dose.1.time_expected, chrono::Utc),
                        chrono::Utc::now()
                    )
                ));
                msg.push_line("");
            }

            return Ok(Some(Message {
                subject: "".to_string(),
                priority: 0,
                body: msg.build(),
                mime: message.mime.clone(),
            }));
        }

        Ok(None)
    }
}

#[async_trait]
impl App for MedManagementApp {
    async fn initialize(
        self: Arc<Self>,
        _config: &'static Config,
        app_state: Arc<Mutex<AppState>>,
    ) {
        let mut state = self.state.lock().await;
        *state = Some(MedManagementAppState {
            global_app_state: app_state,
        });
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
    fn message_digestors(self: Arc<Self>) -> Vec<Box<dyn MessageDigestor + Send + Sync>> {
        vec![Box::new(self)]
    }
}
