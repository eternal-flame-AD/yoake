use std::{path::PathBuf, sync::Arc};

use async_trait::async_trait;
use axum::{routing::get, Extension, Router};
use lazy_static::lazy_static;
use reqwest::redirect::Policy;
use tokio::sync::Mutex;

use crate::{
    comm::{Message, MessageDigestor},
    config::Config,
    AppState,
};

use self::{
    routes_sources::{
        route_combo_search, route_combo_search_top, route_goo_search, route_jisho_search,
        route_jisho_search_top, route_tatoeba_search,
    },
    routes_wordbook::{route_get_wordbook, route_get_wordbook_csv, route_store_wordbook},
    sources::Lookup,
};

use super::App;

mod routes_sources;
mod routes_wordbook;
pub mod sources;

pub struct JpnWordBookApp {
    state: Mutex<Option<JpnWordBookAppState>>,
}

struct JpnWordBookAppState {
    global_app_state: Arc<Mutex<AppState>>,
    jisho: sources::jisho::Client,
    goo: sources::goo::Client,
    tatoeba: sources::tatoeba::Client,
}

impl JpnWordBookApp {
    pub fn new() -> Self {
        Self {
            state: Mutex::new(None),
        }
    }
}

#[async_trait]
impl MessageDigestor for Arc<JpnWordBookApp> {
    async fn digest(&self, message: &Message) -> anyhow::Result<Option<Message>> {
        lazy_static! {
            static ref REGEX_WORDBOOK_QUERY: regex::Regex =
                regex::Regex::new(r"^jisho (.*)$").unwrap();
        }
        if REGEX_WORDBOOK_QUERY.is_match(message.body.as_str()) {
            let captures = REGEX_WORDBOOK_QUERY
                .captures(message.body.as_str())
                .unwrap();
            let query = captures.get(1).unwrap().as_str();
            let state = self.state.lock().await;
            let state = state.as_ref().unwrap();
            let results = state
                .jisho
                .lookup(query)
                .await
                .map_err(|e| anyhow::anyhow!("Failed to lookup word: {}", e))?;
            let mut msg = Message::default();
            msg.subject = "Jisho search successful".to_string();
            for (i, result) in results.iter().enumerate() {
                msg.body
                    .push_str(format!("\n{}. {}\n", i + 1, result.ja).as_str());
                msg.body.push_str(
                    result
                        .fu
                        .clone()
                        .unwrap_or("(no furigana)".to_string())
                        .as_str(),
                );
                msg.body.push_str("\n");
                msg.body.push_str(
                    result
                        .en
                        .clone()
                        .map(|s| s.join("\n"))
                        .unwrap_or("(no english)".to_string())
                        .as_str(),
                );
                msg.body.push_str("\n");
                msg.body.push_str(
                    result
                        .jm
                        .clone()
                        .map(|s| s.join("\n"))
                        .unwrap_or("(解説を見つからない)".to_string())
                        .as_str(),
                );
                msg.body.push_str("\n");
            }
            return Ok(Some(msg));
        }
        Ok(None)
    }
}

#[async_trait]
impl App for JpnWordBookApp {
    async fn initialize(self: Arc<Self>, config: &'static Config, app_state: Arc<Mutex<AppState>>) {
        let mut state = self.state.lock().await;
        *state = Some(JpnWordBookAppState {
            global_app_state: app_state,
            jisho: sources::jisho::Client::new(reqwest::Client::new()),
            goo: sources::goo::Client::new(
                reqwest::ClientBuilder::new()
                    .redirect(Policy::none())
                    .user_agent("Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:109.0) Gecko/20100101 Firefox/116.0")
                    .cookie_store(true)
                    .build()
                    .unwrap(),
            ),
            tatoeba: sources::tatoeba::Client::new(reqwest::Client::new(), PathBuf::from(config.db.data_dir.clone())).await.expect("Failed to initialize tatoeba client"),
        });
    }
    fn api_routes(self: Arc<Self>) -> Router {
        Router::new()
            .route(
                "/jpn_wordbook/sources/combo/search",
                get(route_combo_search),
            )
            .route(
                "/jpn_wordbook/sources/combo/search_top",
                get(route_combo_search_top),
            )
            .route(
                "/jpn_wordbook/sources/tatoeba/search",
                get(route_tatoeba_search),
            )
            .route("/jpn_wordbook/sources/goo/search", get(route_goo_search))
            .route(
                "/jpn_wordbook/sources/jisho/search",
                get(route_jisho_search),
            )
            .route(
                "/jpn_wordbook/sources/jisho/search_top",
                get(route_jisho_search_top),
            )
            .route(
                "/jpn_wordbook/wordbook",
                get(route_get_wordbook).post(route_store_wordbook),
            )
            .route(
                "/jpn_wordbook/wordbook/csv_export",
                get(route_get_wordbook_csv),
            )
            .layer(Extension(self.clone()))
    }

    fn message_digestors(self: Arc<Self>) -> Vec<Box<dyn MessageDigestor + Send + Sync>> {
        vec![Box::new(self)]
    }
}
