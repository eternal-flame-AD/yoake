use async_trait::async_trait;
use lettre::{
    message::header::ContentType, transport::smtp::authentication::Credentials, Transport,
};

use crate::config::{comm::EmailConfig, Config};

use super::Communicator;

pub struct EmailCommunicator {
    config: &'static EmailConfig,
}
impl EmailCommunicator {
    pub fn new(config: &'static Config) -> Self {
        Self {
            config: &config
                .comm
                .email
                .as_ref()
                .expect("Email communicator not configured"),
        }
    }
}

#[async_trait]
impl Communicator for EmailCommunicator {
    fn name(&self) -> &'static str {
        "email"
    }
    fn supported_mimes(&self) -> Vec<&'static str> {
        vec!["text/plain", "text/html"]
    }
    async fn send_message(&self, message: &super::Message) -> anyhow::Result<()> {
        let mailer = lettre::SmtpTransport::relay(&self.config.host)?
            .credentials(Credentials::new(
                self.config.username.clone(),
                self.config.password.clone(),
            ))
            .build();

        let email = lettre::Message::builder()
            .from(self.config.from.parse()?)
            .to(self.config.to.parse()?)
            .subject(message.subject.clone())
            .header(ContentType::parse(message.mime)?)
            .body(message.body.clone())?;

        mailer.send(&email)?;

        Ok(())
    }
}
