use std::{
    collections::HashMap,
    future::Future,
    pin::Pin,
    sync::{Arc, Mutex},
};

use async_trait::async_trait;
use axum::{routing::get, Extension, Router};
use chrono::DateTime;
use log::info;
use serde::{Deserialize, Serialize};
use thirtyfour::{DesiredCapabilities, WebDriver};
use tokio::sync::Mutex as AsyncMutex;

use crate::{
    comm::{Communicator, Message},
    config::Config,
    http::{ApiResponse, ApiResult},
    AppState,
};

use super::{
    auth::{middleware::AuthInfo, Role},
    App,
};

mod driver;
mod utd_app;

pub struct WebcheckApp {
    state: AsyncMutex<WebcheckAppState>,
}

struct WebcheckAppState {
    config: Option<&'static Config>,
    global_app_state: Option<Arc<Mutex<AppState>>>,
    last_response: HashMap<String, LastResponse>,
    checkers: HashMap<String, Box<dyn WebDriverChecker + Send + Sync>>,
}

#[derive(Serialize, Deserialize, Clone, Debug)]
struct LastResponse {
    pub response: String,
    pub timestamp: DateTime<chrono::Utc>,
}

async fn route_get_results(
    auth: AuthInfo,
    app: Extension<Arc<WebcheckApp>>,
) -> ApiResult<HashMap<String, LastResponse>> {
    auth.check_for_any_role(&[Role::Admin])?;

    let state = app.state.lock().await;

    Ok(ApiResponse::ok(
        "Results retrieved successfully".to_string(),
        Some(state.last_response.to_owned()),
    ))
}

#[async_trait]
pub trait WebDriverChecker {
    fn init(&mut self, config: &HashMap<String, String>) -> anyhow::Result<()>;
    fn interval(&self) -> u64;
    async fn check(&self, driver: &WebDriver) -> anyhow::Result<String>;
}

impl WebcheckApp {
    pub fn new() -> Self {
        Self {
            state: AsyncMutex::new(WebcheckAppState {
                config: None,
                global_app_state: None,
                last_response: HashMap::new(),
                checkers: HashMap::new(),
            }),
        }
    }

    pub async fn run_single_check(self: &Self, key: &str) -> anyhow::Result<()> {
        let mut state = self.state.lock().await;

        let checker = state.checkers.get_mut(key).unwrap();

        let mut driver = driver::chrome::ChromeDriver::new();
        driver.spawn(&["--enable-chrome-logs"])?;
        tokio::time::sleep(std::time::Duration::from_secs(5)).await;

        let mut caps = DesiredCapabilities::chrome();
        caps.set_headless().unwrap();
        caps.set_disable_gpu().unwrap();

        let driver = driver.connect(caps).await?;
        let response = match checker.check(&driver).await {
            Ok(response) => response,
            Err(e) => {
                driver.quit().await?;
                return Err(e);
            }
        };
        driver.quit().await?;

        let new_response = LastResponse {
            response: response.clone(),
            timestamp: chrono::Utc::now(),
        };

        let last_response = state.last_response.get(key);

        match last_response {
            Some(last_response) => {
                if last_response.response != response {
                    state
                        .global_app_state
                        .as_ref()
                        .unwrap()
                        .lock()
                        .unwrap()
                        .comm
                        .send_message(&Message {
                            subject: format!("webcheck {} changed", key),
                            body: format!("{} changed to {}", key, response),
                            mime: "text/plain",
                            priority: 0,
                        })?;
                }
            }
            None => {}
        }

        state.last_response.insert(key.to_string(), new_response);

        Ok(())
    }

    pub async fn run_check_loops(self: Arc<Self>) {
        let self_clone = self.clone();

        let state = self.state.lock().await;
        for key in state.checkers.keys() {
            let key = key.clone();
            let self_clone = self_clone.clone();
            tokio::spawn(async move {
                let interval = {
                    let state = self_clone.state.lock().await;
                    let checker = state.checkers.get(key.as_str()).unwrap();
                    checker.interval()
                };

                let mut ticker = tokio::time::interval(std::time::Duration::from_secs(interval));

                loop {
                    ticker.tick().await;

                    info!("Running webcheck for {}", key);
                    self_clone
                        .run_single_check(key.as_str())
                        .await
                        .map_err(|e| {
                            log::error!("Error running webcheck for {}: {}", key, e);
                        })
                        .ok();
                }
            });
        }
    }
}

impl App for WebcheckApp {
    fn initialize(
        self: Arc<Self>,
        config: &'static Config,
        app_state: Arc<Mutex<AppState>>,
    ) -> Pin<Box<dyn Future<Output = ()>>> {
        Box::pin(async move {
            let mut state = self.state.lock().await;
            state.config = Some(config);
            state.global_app_state = Some(app_state);

            let Some(ref config) = config.webcheck else {
                return;
            };

            config.keys().for_each(|key| match key.as_str() {
                "utd_app" => {
                    let mut checker = utd_app::UTDAppChecker::new();
                    checker
                        .init(config.get(key).unwrap())
                        .expect("Failed to initialize UTDAppChecker");
                    state.checkers.insert(key.clone(), Box::new(checker));
                }
                _ => panic!("Invalid key in webcheck config: {}", key),
            });

            let self_clone = self.clone();
            tokio::spawn(self_clone.run_check_loops());
        })
    }

    fn api_routes(self: Arc<Self>) -> Router {
        Router::new()
            .route("/webcheck/results", get(route_get_results))
            .layer(Extension(self.clone()))
    }
}
