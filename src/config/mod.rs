use std::collections::HashMap;

use serde::Deserialize;

pub mod comm;

#[derive(Debug, Deserialize)]
pub struct Config {
    pub listen: ListenConfig,
    pub db: DbConfig,

    pub canvas_lms: CanvasLmsConfig,

    pub auth: AuthConfig,

    pub session: SessionConfig,

    pub webcheck: Option<HashMap<String, HashMap<String, String>>>,

    pub comm: comm::Config,
}

impl Config {
    pub fn load_yaml<R: std::io::Read>(reader: R) -> Self {
        let config = serde_yaml::from_reader(reader).expect("Failed to parse config");
        config
    }
    pub fn load_yaml_file<P: AsRef<std::path::Path>>(path: P) -> Self {
        let file = std::fs::File::open(path).expect("Failed to open config file");
        Self::load_yaml(file)
    }
}

#[derive(Debug, Deserialize)]
pub struct ListenConfig {
    pub addr: String,
    pub cert: Option<String>,
    pub key: Option<String>,
}

#[derive(Debug, Deserialize)]
pub struct DbConfig {
    pub data_dir: String,
    pub url: String,
}

#[derive(Debug, Deserialize)]
pub struct CanvasLmsConfig {
    pub token: String,
    pub endpoint: String,
    pub refresh_interval: u64,
}

#[derive(Debug, Deserialize)]
pub struct AuthConfig {
    pub users: HashMap<String, AuthUser>,
}

#[derive(Debug, Deserialize)]
pub struct SessionConfig {
    pub secret: String,
}

#[derive(Debug, Deserialize)]
pub struct AuthUser {
    pub password: String,
    pub roles: Vec<String>,
}

static mut CURRENT_CONFIG: Option<Config> = None;

pub unsafe fn set_config(config: Config) {
    unsafe {
        CURRENT_CONFIG = Some(config);
    }
}

pub fn get_config() -> &'static Config {
    unsafe { CURRENT_CONFIG.as_ref().expect("Config not initialized") }
}
