use std::sync::Arc;

use anyhow::Result;
use async_trait::async_trait;

use log::info;
use serenity::model::channel::Message;
use serenity::model::gateway::Ready;
use serenity::prelude::*;

use serenity::model::prelude::{ChannelId, UserId};

use crate::config::comm::DiscordConfig;

use super::{Communicator, MessageDigestor};

struct Handler {
    state: Mutex<HandlerState>,
    channel_id: ChannelId,
    digestor: Arc<Vec<Box<dyn MessageDigestor + Send + Sync>>>,
}

impl Handler {
    fn new(
        channel_id: ChannelId,
        digestor: Arc<Vec<Box<dyn MessageDigestor + Send + Sync>>>,
    ) -> Self {
        Self {
            state: Mutex::new(HandlerState { bot_id: UserId(0) }),
            channel_id,
            digestor,
        }
    }
}

struct HandlerState {
    bot_id: UserId,
}

#[async_trait]
impl EventHandler for Handler {
    async fn message(&self, ctx: Context, msg: Message) {
        if msg.channel_id == self.channel_id {
            info!("Received message: {:?}", msg.content);
            let msg_generic = super::Message {
                subject: "".to_string(),
                body: msg.content.to_string(),
                mime: MIME_DISCORD,
                priority: 0,
            };
            for digestor in self.digestor.iter() {
                match digestor.digest(&msg_generic).await {
                    Err(e) => {
                        info!("Error digesting message: {:?}", e);
                    }
                    Ok(Some(reply)) => {
                        if let Err(why) = msg.channel_id.say(&ctx.http, reply.body).await {
                            info!("Error sending message: {:?}", why);
                        }
                    }
                    Ok(None) => {}
                }
            }
        }
    }

    async fn ready(&self, _: Context, ready: Ready) {
        let mut state = self.state.lock().await;
        state.bot_id = ready.user.id;
    }
}

pub struct DiscordCommunicator {
    client: Client,
    config: &'static DiscordConfig,
}

impl DiscordCommunicator {
    pub async fn new(
        config: &'static DiscordConfig,
        digestor: Arc<Vec<Box<dyn MessageDigestor + Send + Sync>>>,
    ) -> Self {
        let channel_id = ChannelId(config.channel_id);

        let intents = GatewayIntents::non_privileged()
            | GatewayIntents::GUILD_MESSAGES
            | GatewayIntents::DIRECT_MESSAGES
            | GatewayIntents::MESSAGE_CONTENT;

        let handler = Handler::new(channel_id, digestor);

        let mut client = Client::builder(config.token.to_string(), intents)
            .event_handler(handler)
            .await
            .expect("Error creating client");

        tokio::spawn(async move {
            if let Err(why) = client.start().await {
                info!("Client error: {:?}", why);
            }
        });

        let client = Client::builder(config.token.to_string(), intents)
            .await
            .expect("Error creating client");

        Self { client, config }
    }
}

pub const MIME_DISCORD: &'static str = "application/vnd.discord";

#[async_trait]
impl Communicator for DiscordCommunicator {
    fn name(&self) -> &'static str {
        "Discord"
    }
    fn supported_mimes(&self) -> Vec<&'static str> {
        vec![MIME_DISCORD]
    }
    async fn send_message(&self, message: &super::Message) -> Result<()> {
        let channel = ChannelId(self.config.channel_id);
        channel
            .say(&self.client.cache_and_http.http, message.body.to_string())
            .await?;
        Ok(())
    }
}
