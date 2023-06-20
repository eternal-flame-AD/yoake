use serde::Deserialize;

#[derive(Debug, Deserialize)]
pub struct Config {
    pub email: Option<EmailConfig>,
    pub gotify: Option<GotifyConfig>,
    pub discord: Option<DiscordConfig>,
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

#[derive(Debug, Deserialize)]
pub struct DiscordConfig {
    pub token: String,
    pub channel_id: u64,
}
