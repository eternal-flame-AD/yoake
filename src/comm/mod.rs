use std::{collections::HashMap, sync::Arc};

use anyhow::Result;
use log::{error, warn};

pub mod email;
pub mod gotify;

#[derive(Clone, Debug)]
pub struct Message {
    pub subject: String,
    pub body: String,
    pub mime: &'static str,
    pub priority: i8,
}

impl Default for Message {
    fn default() -> Self {
        Self {
            subject: String::new(),
            body: String::new(),
            mime: MIME_PLAIN,
            priority: 0,
        }
    }
}

pub const MIME_PLAIN: &'static str = "text/plain";
pub const MIME_HTML: &'static str = "text/html";

pub trait Communicator {
    fn name(&self) -> &'static str;
    fn supported_mimes(&self) -> Vec<&'static str>;
    fn send_message(&self, message: &Message) -> Result<()>;
}

pub struct GlobalCommunicator {
    communicators: HashMap<&'static str, Vec<Arc<dyn Communicator + Sync + Send>>>,
}

impl GlobalCommunicator {
    pub fn new() -> Self {
        Self {
            communicators: HashMap::new(),
        }
    }
    pub fn add_communicator(&mut self, communicator: Arc<dyn Communicator + Sync + Send>) {
        for mime in communicator.supported_mimes() {
            if !self.communicators.contains_key(mime) {
                self.communicators.insert(mime, Vec::new());
            }
            self.communicators
                .get_mut(mime)
                .unwrap()
                .push(communicator.clone());
        }
    }
    pub fn by_mime(&self, mime: &'static str) -> Option<&Vec<Arc<dyn Communicator + Sync + Send>>> {
        self.communicators.get(mime)
    }
    pub fn by_name(&self, name: &'static str) -> Option<&Arc<dyn Communicator + Sync + Send>> {
        for communicators in self.communicators.values() {
            for communicator in communicators {
                if communicator.name() == name {
                    return Some(communicator);
                }
            }
        }
        None
    }
}

impl Communicator for GlobalCommunicator {
    fn name(&self) -> &'static str {
        "global"
    }

    fn supported_mimes(&self) -> Vec<&'static str> {
        self.communicators.keys().map(|k| *k).collect()
    }

    fn send_message(&self, message: &Message) -> Result<()> {
        let mime = message.mime;
        if let Some(communicators) = self.communicators.get(mime) {
            for communicator in communicators {
                if let Err(e) = communicator.send_message(message) {
                    warn!("Failed to send message with {}: {}", communicator.name(), e);
                    continue;
                }
                return Ok(());
            }
        }
        error!("No communicators available for mime {}", mime);
        Err(anyhow::anyhow!(
            "No communicators available for mime {}",
            mime
        ))
    }
}
