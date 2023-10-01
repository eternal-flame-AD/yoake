use std::collections::HashMap;

use super::ReqwestChecker;
use async_trait::async_trait;
use serde::Deserialize;

#[derive(Deserialize, Debug)]
pub struct SaveCaseCheckResp {
    #[serde(rename = "caseNumber")]
    case_number: String,
    #[serde(rename = "createdDate")]
    created_date: String,
    #[serde(rename = "agencyName")]
    agency_name: String,
    #[serde(rename = "caseStatus")]
    case_status: String,
}

pub struct SaveChecker {
    config: HashMap<String, String>,
}

impl SaveChecker {
    pub fn new() -> Self {
        Self {
            config: HashMap::new(),
        }
    }
}

#[async_trait]
impl ReqwestChecker for SaveChecker {
    fn init(&mut self, config: &HashMap<String, String>) -> anyhow::Result<()> {
        if config.get("case_id").is_none() {
            return Err(anyhow::anyhow!("case_id not set"));
        }

        self.config = config.clone();
        Ok(())
    }
    fn interval(&self) -> u64 {
        let default_interval = "3600".to_string();
        let interval = self.config.get("interval").unwrap_or(&default_interval);
        interval.parse::<u64>().unwrap()
    }
    async fn check(&self, client: &reqwest::Client) -> anyhow::Result<String> {
        let case_id = self.config.get("case_id").unwrap();
        let req = client
            .request(
                reqwest::Method::GET,
                format!(
                    "https://save.uscis.gov/api/save/read/cases/check/{}",
                    case_id
                ),
            )
            .header("Accept", "application/json")
            .header("Content-Type", "application/json")
            .header(
                "User-Agent",
                "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:109.0) Gecko/20100101 Firefox/117.0",
            )
            .header(
                "Referer",
                format!(
                    "https://save.uscis.gov/save/app/client/ui/case-check/detail/{}",
                    case_id
                ),
            )
            .build()?;
        let resp = client.execute(req).await?;
        let res = resp.json::<SaveCaseCheckResp>().await?;
        Ok(format!("Case {} is {}", res.case_number, res.case_status,))
    }
}
