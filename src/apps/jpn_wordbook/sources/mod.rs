use async_trait::async_trait;
use serde::{Deserialize, Serialize};

pub mod goo;
pub mod jisho;
pub mod tatoeba;

pub fn match_len(query: &str, word: &str) -> usize {
    let mut query_chars = query.chars();
    let mut word_chars = word.chars();

    let mut match_len = 0;
    loop {
        let query_char = query_chars.next();
        let word_char = word_chars.next();

        if query_char.is_none() || word_char.is_none() {
            break;
        }

        if query_char == word_char {
            match_len += 1;
        } else {
            break;
        }
    }

    match_len
}

#[derive(Default, Debug, Clone, Serialize, Deserialize)]
pub struct LookupResult {
    pub ja: String,
    pub altn: Option<Vec<String>>,
    pub jm: Option<Vec<String>>,
    pub fu: Option<String>,
    pub en: Option<Vec<String>>,
    pub ex: Option<Vec<String>>,

    pub src: String,
}

impl Into<crate::models::jpn_wordbook::JpnWordbook> for LookupResult {
    fn into(self) -> crate::models::jpn_wordbook::JpnWordbook {
        crate::models::jpn_wordbook::JpnWordbook {
            uuid: uuid::Uuid::new_v4().to_string(),

            ja: self.ja,
            altn: if self.altn.is_some() {
                self.altn.unwrap().join(",")
            } else {
                String::new()
            },
            jm: if self.jm.is_some() {
                self.jm.unwrap().join("\n")
            } else {
                String::new()
            },
            fu: if self.fu.is_some() {
                self.fu.unwrap()
            } else {
                String::new()
            },
            en: if self.en.is_some() {
                self.en.unwrap().join("\n")
            } else {
                String::new()
            },
            ex: if self.ex.is_some() {
                self.ex.unwrap().join("\n")
            } else {
                String::new()
            },

            src: self.src,

            created: chrono::Local::now().naive_utc(),
            updated: chrono::Local::now().naive_utc(),
        }
    }
}

impl LookupResult {
    pub fn new(ja: String, src: &'static str) -> Self {
        Self {
            ja,
            altn: None,
            jm: None,
            fu: None,
            en: None,
            ex: None,

            src: src.to_string(),
        }
    }
    pub fn merge(&mut self, other: Self) {
        if self.altn.is_none() {
            self.altn = other.altn;
        } else {
            if other.altn.is_some() {
                self.altn.as_mut().unwrap().extend(other.altn.unwrap());
            }
        }
        if self.jm.is_none() {
            self.jm = other.jm;
        }
        if self.fu.is_none() {
            self.fu = other.fu;
        }
        if self.en.is_none() {
            self.en = other.en;
        }
        if self.ex.is_none() {
            self.ex = other.ex;
        }

        if self.src != other.src {
            self.src = format!("{},{}", self.src, other.src);
        }
    }
}

impl LookupResult {
    pub fn match_score<S: AsRef<str>>(&self, word: S) -> usize {
        if self.ja == word.as_ref() {
            return 100;
        }
        if self.ja.starts_with(word.as_ref()) {
            return 95;
        }

        (match_len(word.as_ref(), &self.ja) as f64 * 100.0 / self.ja.len() as f64) as usize
    }
}

#[async_trait]
pub trait Lookup {
    async fn lookup(&self, word: &str) -> anyhow::Result<Vec<LookupResult>>;
    async fn lookup_top(&self, word: &str) -> anyhow::Result<LookupResult> {
        let results = self.lookup(word).await?;
        let top_result = match results
            .into_iter()
            .max_by(|a, b| a.match_score(word).cmp(&b.match_score(word)))
        {
            Some(r) => r,
            None => {
                return Err(anyhow::anyhow!("No results found"));
            }
        };

        Ok(top_result)
    }
}
