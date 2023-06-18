use argon2::{
    password_hash::{rand_core::OsRng, SaltString},
    Argon2, PasswordHasher, PasswordVerifier,
};
use serde::{Deserialize, Serialize};

use crate::http::{ApiResponse, ApiResult, JsonApiForm};

pub fn hash_password(password: &str) -> String {
    let salt = SaltString::generate(&mut OsRng);

    let argon2 = Argon2::default();
    argon2
        .hash_password(password.as_bytes(), &salt)
        .unwrap()
        .to_string()
}

pub fn verify_password(hash: &str, password: &str) -> bool {
    let argon2 = Argon2::default();
    let hash = argon2::PasswordHash::new(hash).unwrap();
    argon2.verify_password(password.as_bytes(), &hash).is_ok()
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct HashPasswordForm {
    password: String,
}

pub async fn route_hash_password(
    JsonApiForm(form): JsonApiForm<HashPasswordForm>,
) -> ApiResult<String> {
    let hash = hash_password(&form.password);

    Ok(ApiResponse::ok("".to_string(), Some(hash)))
}
#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_hash_password() {
        let password = "password";
        let hash = hash_password(password);
        assert!(verify_password(&hash, password));
    }

    #[test]
    fn test_hash_invalid_password() {
        let password = "password";
        let hash = hash_password(password);
        assert!(!verify_password(&hash, "invalid"));
    }
}
