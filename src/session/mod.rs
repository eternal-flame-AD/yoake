use axum::{http::Request, middleware::Next, response::Response, Extension};
use base64::{engine::general_purpose, Engine};
use serde::{Deserialize, Serialize};

use crate::{apps::auth::SessionAuth, config::Config};

mod wrap;

#[derive(Debug, Clone, Serialize, Deserialize, Default)]
pub struct SessionStore {
    pub auth: SessionAuth,
}

const SESSION_COOKIE_NAME: &str = "session_state";

pub async fn middleware<B>(
    Extension(config): Extension<&Config>,
    mut req: Request<B>,
    next: Next<B>,
) -> Response {
    let mut session_state = SessionStore::default();

    let mut key = [0u8; 32];
    general_purpose::STANDARD
        .decode_slice(config.session.secret.as_bytes(), &mut key)
        .expect("Failed to decode session secret");

    {
        let cookies = req
            .headers()
            .get("Cookie")
            .map(|c| c.to_str().unwrap_or(""))
            .unwrap_or("")
            .split("; ")
            .map(|c| {
                let mut parts = c.splitn(2, '=');
                (
                    parts.next().unwrap_or("").to_string(),
                    parts.next().unwrap_or("").to_string(),
                )
            });

        for (name, value) in cookies {
            if name == SESSION_COOKIE_NAME {
                if let Some(store) = wrap::unwrap_json::<SessionStore>(&value, key.as_ref()) {
                    session_state = store;
                }
            }
        }
    }

    req.extensions_mut().insert(session_state);

    let mut resp = next.run(req).await;

    if let Some(new_store) = resp.extensions().get::<SessionStore>() {
        let wrapped = wrap::wrap_json(new_store, key.as_ref());

        resp.headers_mut().insert(
            "Set-Cookie",
            format!(
                "{}={}; Path=/; HttpOnly; Max-Age=31536000",
                SESSION_COOKIE_NAME, wrapped
            )
            .parse()
            .unwrap(),
        );
    }

    resp
}
