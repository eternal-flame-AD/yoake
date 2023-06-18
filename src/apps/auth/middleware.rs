use std::convert::Infallible;

use axum::{async_trait, extract::FromRequestParts};
use hyper::http::request::Parts;
use serde::{Deserialize, Serialize};

use crate::{http::ApiResponse, session::SessionStore};

use super::Role;

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct AuthInfo {
    valid: bool,
    user: String,
    display_name: String,
    roles: Vec<Role>,
}

impl AuthInfo {
    pub fn is_valid(&self) -> bool {
        self.valid
    }
    pub fn user(&self) -> &str {
        &self.user
    }
    pub fn display_name(&self) -> &str {
        &self.display_name
    }
    pub fn roles(&self) -> &[Role] {
        &self.roles
    }
    pub fn has_any_role(&self, roles: &[Role]) -> bool {
        for role in roles {
            if self.roles.contains(role) {
                return true;
            }
        }
        false
    }
    pub fn check_for_any_role(&self, roles: &[Role]) -> Result<(), ApiResponse<()>> {
        if self.has_any_role(roles) {
            Ok(())
        } else {
            Err(ApiResponse::unauthorized("Unauthorized".to_string(), None))
        }
    }
}

impl Default for AuthInfo {
    fn default() -> Self {
        Self {
            valid: false,
            user: String::new(),
            display_name: "anonymous".to_string(),
            roles: Vec::new(),
        }
    }
}
#[async_trait]
impl<S> FromRequestParts<S> for AuthInfo
where
    S: Send + Sync,
{
    type Rejection = Infallible;

    async fn from_request_parts(parts: &mut Parts, _state: &S) -> Result<Self, Self::Rejection> {
        let reject = AuthInfo::default();

        let Some(session) = parts.extensions.get::<SessionStore>() else {
            return Ok(reject);
        };

        if session.auth.user.is_empty() {
            return Ok(reject);
        }
        let now = chrono::Utc::now();
        if session.auth.expire < now {
            return Ok(reject);
        }

        let mut res = AuthInfo::default();
        res.valid = true;
        res.user = session.auth.user.clone();
        res.display_name = session.auth.user.clone();
        res.roles = session.auth.roles.clone();

        Ok(res)
    }
}

/*
#[macro_export]
macro_rules! require_role {
    ($auth:ident, $roles:expr) => {
        if !$auth.has_any_role(&[$roles]) {
            return ApiResponse::<()>::error(
                format!("You do not have permission to access this resource. Acceptable role: {:?}, you have: {:?}", $roles, $auth.roles()),
                403,
                None,
            )
            .into_response();
        }
    };
    ($auth:ident, [$($roles:expr),*]) => {
        if !$auth.has_any_role(&[$($roles),*]) {
            return ApiResponse::<()>::error(
                format!("You do not have permission to access this resource. Acceptable roles: {:?}, you have: {:?}", [$($roles),*], $auth.roles()),
                403,
                None,
            )
            .into_response();
        }
    };
}
*/
