use async_trait::async_trait;
use serde::{Deserialize, Serialize};

use super::{Lookup, LookupResult};

pub struct Client {
    client: reqwest::Client,
}

#[derive(Default, Debug, Clone, Serialize, Deserialize)]
pub struct Response<T> {
    pub meta: ResponseMeta,
    pub data: T,
}

#[derive(Default, Debug, Clone, Serialize, Deserialize)]
pub struct ResponseMeta {
    pub status: u16,
}

#[derive(Default, Debug, Clone, Serialize, Deserialize)]
pub struct WordResponse {
    pub slug: String,
    pub is_common: bool,
    pub tags: Vec<String>,
    pub jlpt: Vec<String>,
    pub japanese: Vec<Japanese>,
    pub senses: Vec<Sense>,
    pub attribution: Attribution,
}

impl Into<LookupResult> for WordResponse {
    fn into(self) -> LookupResult {
        let first_meaning = &self.japanese[0];

        let mut result = LookupResult::new(
            first_meaning
                .word
                .as_ref()
                .or_else(|| Some(&first_meaning.reading))
                .unwrap()
                .to_string(),
            "Jisho.org",
        );

        result.fu = Some(first_meaning.reading.clone());
        let mut en = Vec::new();
        for sense in self.senses {
            let mut en_chunk = String::new();
            for (i, def) in sense.english_definitions.into_iter().enumerate() {
                if i != 0 {
                    en_chunk.push_str("; ");
                }
                en_chunk.push_str(&def);
            }
            en.push(en_chunk);
        }
        result.en = Some(en);

        result
    }
}

#[derive(Default, Debug, Clone, Serialize, Deserialize)]
pub struct Japanese {
    pub word: Option<String>,
    pub reading: String,
}

#[derive(Default, Debug, Clone, Serialize, Deserialize)]
pub struct Sense {
    pub english_definitions: Vec<String>,
    pub parts_of_speech: Vec<String>,
    pub links: Vec<Link>,
    pub tags: Vec<String>,
    pub see_also: Vec<String>,
    pub info: Vec<String>,
}

#[derive(Default, Debug, Clone, Serialize, Deserialize)]
pub struct Link {
    pub text: String,
    pub url: String,
}

#[derive(Default, Debug, Clone, Serialize, Deserialize)]
pub struct Attribution {
    pub jmdict: bool,
    pub jmnedict: bool,
}

#[async_trait]
impl Lookup for Client {
    async fn lookup(&self, word: &str) -> anyhow::Result<Vec<LookupResult>> {
        let query = url::form_urlencoded::Serializer::new(String::new())
            .append_pair("keyword", word)
            .finish();
        let url = format!("https://jisho.org/api/v1/search/words?{}", query.as_str());
        let response = self.client.get(&url).send().await?;
        let response: Response<Vec<WordResponse>> = response.json().await?;
        Ok(response.data.into_iter().map(|r| r.into()).collect())
    }
}

impl Client {
    pub fn new(client: reqwest::Client) -> Self {
        Self { client }
    }
    pub async fn lookup_word<S: AsRef<str>>(&self, word: S) -> anyhow::Result<Vec<WordResponse>> {
        let query = url::form_urlencoded::Serializer::new(String::new())
            .append_pair("keyword", word.as_ref())
            .finish();
        let url = format!("https://jisho.org/api/v1/search/words?{}", query.as_str());
        let response = self.client.get(&url).send().await?;
        let response: Response<Vec<WordResponse>> = response.json().await?;
        Ok(response.data)
    }
}
