use std::collections::HashMap;

use anyhow::Result;
use async_trait::async_trait;
use thirtyfour::prelude::*;
use tokio::time;

use super::WebDriverChecker;

pub struct UTDAppChecker {
    config: HashMap<String, String>,
}

impl UTDAppChecker {
    pub fn new() -> Self {
        Self {
            config: HashMap::new(),
        }
    }
}

#[async_trait]
impl WebDriverChecker for UTDAppChecker {
    fn init(&mut self, config: &HashMap<String, String>) -> Result<()> {
        if config.get("username").is_none() || config.get("password").is_none() {
            return Err(anyhow::anyhow!("username or password not set"));
        }

        self.config = config.clone();
        Ok(())
    }
    fn interval(&self) -> u64 {
        let default_interval = "3600".to_string();
        let interval = self.config.get("interval").unwrap_or(&default_interval);
        interval.parse::<u64>().unwrap()
    }

    async fn check(&self, driver: &WebDriver) -> Result<String> {
        let username = self.config.get("username").unwrap();
        let password = self.config.get("password").unwrap();

        driver
            .goto("https://utdallas.my.site.com/TX_SiteLogin?startURL=%2FTargetX_Portal__PB")
            .await?;

        let input_els = driver.find_all(By::Css("input[type='text']")).await?;

        for input_el in input_els {
            let name_attr = input_el.attr("name").await?;
            if name_attr.is_some() && name_attr.unwrap().ends_with(":username") {
                input_el.send_keys(username).await?;
            }
        }

        let password_el = driver.find(By::Css("input[type='password']")).await?;
        password_el.send_keys(password).await?;

        let submit_el = driver.find(By::Css("a.targetx-button")).await?;
        submit_el.click().await?;

        time::sleep(time::Duration::from_secs(10)).await;

        let mut checklist_item_text = Vec::new();
        let checklist_item_text_els = driver.find_all(By::Css("p.checklist-item-text")).await?;

        for checklist_item_text_el in checklist_item_text_els {
            let text = checklist_item_text_el.text().await?;
            checklist_item_text.push(text);
        }

        checklist_item_text.sort();

        Ok(checklist_item_text.join("\n"))
    }
}
