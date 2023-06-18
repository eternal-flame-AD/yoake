use super::Communicator;
use crate::config::{comm::GotifyConfig, Config};
use serde::Serialize;

pub struct GotifyCommunicator {
    config: &'static GotifyConfig,
}

impl GotifyCommunicator {
    pub fn new(config: &'static Config) -> Self {
        Self {
            config: &config.comm.gotify,
        }
    }
}

#[derive(Serialize)]
struct GotifyMessage {
    title: String,
    message: String,
    priority: i8,
    extras: GotifyMessageExtras,
}

impl Into<GotifyMessage> for &super::Message {
    fn into(self) -> GotifyMessage {
        GotifyMessage {
            title: self.subject.clone(),
            message: self.body.clone(),
            priority: self.priority,
            extras: GotifyMessageExtras {
                client_display: GotifyMessageExtrasClientDisplay {
                    content_type: self.mime.clone().to_string(),
                },
            },
        }
    }
}

#[derive(Serialize)]
struct GotifyMessageExtras {
    #[serde(rename = "client::display")]
    client_display: GotifyMessageExtrasClientDisplay,
}

#[derive(Serialize)]
struct GotifyMessageExtrasClientDisplay {
    #[serde(rename = "contentType")]
    content_type: String,
}

impl Communicator for GotifyCommunicator {
    fn name(&self) -> &'static str {
        "gotify"
    }
    fn supported_mimes(&self) -> Vec<&'static str> {
        vec!["text/plain", "text/markdown"]
    }
    fn send_message(&self, message: &super::Message) -> anyhow::Result<()> {
        let client = reqwest::blocking::Client::new();
        let response = client
            .post(&format!("{}/message", self.config.url))
            .header("X-Gotify-Key", &self.config.token)
            .json::<GotifyMessage>(&message.into())
            .send()?;

        if !response.status().is_success() {
            anyhow::bail!("Gotify returned an error: {:?}", response);
        }

        Ok(())
    }
}
