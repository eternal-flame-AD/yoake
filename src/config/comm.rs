use serde::Deserialize;

#[derive(Debug, Deserialize)]
pub struct Config {
    pub email: EmailConfig,
    pub gotify: GotifyConfig,
}

#[derive(Debug, Deserialize)]
pub struct EmailConfig {
    pub from: String,
    pub to: String,
    pub host: String,
    pub port: u16,
    pub username: String,
    pub password: String,
    pub default_subject: String,
}

#[derive(Debug, Deserialize)]
pub struct GotifyConfig {
    pub url: String,
    pub token: String,
}
