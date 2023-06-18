use std::{
    future::Future,
    pin::Pin,
    sync::{Arc, Mutex},
};

use axum::{
    body::HttpBody,
    error_handling::HandleErrorLayer,
    http::Request,
    response::{IntoResponse, Response},
    routing::{get, post},
    BoxError, Extension, Router,
};
use chrono::{DateTime, Utc};
use serde::{Deserialize, Serialize};
use tower::ServiceBuilder;

use crate::{
    apps::App,
    config::Config,
    http::{ApiResponse, JsonApiForm},
    session::SessionStore,
    AppState,
};

use self::middleware::AuthInfo;

pub mod middleware;
mod password;

#[derive(Debug, Clone, Serialize, Deserialize, Default)]
pub struct SessionAuth {
    pub user: String,
    pub expire: DateTime<Utc>,
    pub roles: Vec<Role>,
}

#[derive(Debug, Serialize, Deserialize, Clone, PartialEq)]
pub enum Role {
    Admin,
    User,
    Unknown,
}

impl From<&str> for Role {
    fn from(s: &str) -> Self {
        match s {
            "Admin" => Self::Admin,
            "User" => Self::User,
            _ => Self::Unknown,
        }
    }
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct LoginForm {
    pub username: String,
    pub password: String,
}

pub async fn route_login(
    app: Extension<Arc<AuthApp>>,
    Extension(mut session_store): Extension<SessionStore>,
    JsonApiForm(form): JsonApiForm<LoginForm>,
) -> Result<(Extension<SessionStore>, ApiResponse<()>), ApiResponse<()>> {
    let failed_response = ApiResponse::<()>::error("Invalid credentials".to_string(), 401, None);

    let state = app.state.lock().unwrap();
    let state = state.as_ref().unwrap();

    let Some(user) = state.config.auth.users.get(&form.username) else {
        const DUMMY_HASH: &str = "$argon2id$v=19$m=19456,t=2,p=1$U7zg/pa1Wf9Hi9NM+ns9aA$tivXyIMw+wo9ZZoz0I+6yLm7+1SfkW9fF5hONy/qq1Y";
        password::verify_password(DUMMY_HASH, form.password.as_str());
        return Err(failed_response);
    };

    let hash = user.password.as_str();
    if !password::verify_password(hash, form.password.as_str()) {
        return Err(failed_response);
    }

    session_store.auth.user = form.username.to_string();
    session_store.auth.expire = Utc::now() + chrono::Duration::days(7);
    session_store.auth.roles = user.roles.iter().map(|r| r.as_str().into()).collect();

    Ok((
        Extension(session_store),
        ApiResponse::ok("Login successful".to_string(), None),
    ))
}

pub async fn route_logout(
    Extension(mut session_store): Extension<SessionStore>,
) -> Result<(Extension<SessionStore>, ApiResponse<()>), ApiResponse<()>> {
    session_store.auth.user = String::new();

    Ok((
        Extension(session_store),
        ApiResponse::ok("Logout successful".to_string(), None),
    ))
}

pub async fn route_self<B: HttpBody>(auth: AuthInfo, _req: Request<B>) -> Response
where
    <B as HttpBody>::Error: std::fmt::Debug,
{
    ApiResponse::ok("".to_string(), Some(auth)).into_response()
}

pub struct AuthApp {
    state: Mutex<Option<AuthAppState>>,
}

pub struct AuthAppState {
    config: &'static Config,
    _app_state: Arc<Mutex<AppState>>,
}

impl AuthApp {
    pub fn new() -> Self {
        Self {
            state: Mutex::new(None),
        }
    }
}

impl App for AuthApp {
    fn initialize(
        self: Arc<Self>,
        config: &'static Config,
        app_state: Arc<Mutex<AppState>>,
    ) -> Pin<Box<dyn Future<Output = ()>>> {
        let mut state = self.state.lock().unwrap();
        *state = Some(AuthAppState {
            config,
            _app_state: app_state,
        });

        Box::pin(async {})
    }

    fn api_routes(self: Arc<Self>) -> Router {
        let rate_limiter = tower::limit::RateLimitLayer::new(1, std::time::Duration::from_secs(1));
        Router::new()
            .route(
                "/auth/hash_password",
                post(password::route_hash_password)
                    // rate limit
                    .layer(
                        ServiceBuilder::new()
                            .layer(HandleErrorLayer::new(|err: BoxError| async move {
                                log::error!("Error: {:?}", err);
                                (
                                    axum::http::StatusCode::INTERNAL_SERVER_ERROR,
                                    "Internal Server Error".to_string(),
                                )
                            }))
                            .buffer(64)
                            .concurrency_limit(1)
                            .rate_limit(1, std::time::Duration::from_secs(1))
                            .layer(rate_limiter),
                    ),
            )
            .route("/auth/logout", post(route_logout))
            .route("/auth/login", post(route_login))
            .route("/auth/self", get(route_self))
            .layer(Extension(self.clone()))
    }
}
