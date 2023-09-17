use async_trait::async_trait;
use futures::StreamExt;
use log::error;
use scraper::{Html, Selector};

use super::{match_len, Lookup, LookupResult};

pub fn strip_url_hash<S: AsRef<str>>(url: S, origin: &'static str) -> String {
    let url_with_host = if url.as_ref().starts_with("//") {
        format!("https:{}", url.as_ref())
    } else if url.as_ref().starts_with("/") {
        format!("{}{}", origin, url.as_ref())
    } else {
        url.as_ref().to_string()
    };
    let mut url = url::Url::parse(&url_with_host).unwrap();
    if url.host().is_none() {
        url.set_host(Some(origin)).unwrap();
    }
    url.set_fragment(None);
    url.into()
}

pub struct Client {
    client: reqwest::Client,
}

pub struct Candidate {
    pub url: String,
    pub title: String,
    pub text: Option<String>,
}

impl Client {
    pub fn new(client: reqwest::Client) -> Self {
        Self { client }
    }
    async fn fetch_html<S: AsRef<str>>(&self, url: S) -> anyhow::Result<Html> {
        let response = self.client.get(url.as_ref()).send().await?;
        let text = response.text().await?;
        Ok(Html::parse_document(&text))
    }
    async fn lookup_candidates<S: AsRef<str>>(&self, prefix: S) -> anyhow::Result<Vec<Candidate>> {
        let resp = self
            .client
            .get(format!(
                "https://dictionary.goo.ne.jp/srch/jn/{}/m0u/",
                urlencoding::encode(prefix.as_ref())
            ))
            .send()
            .await?;
        if let Some(redir) = resp.headers().get("Location") {
            return Ok(vec![Candidate {
                url: redir.to_str()?.to_string(),
                title: prefix.as_ref().to_string(),
                text: None,
            }]);
        }
        let mut candidates = Vec::new();
        let response_html = Html::parse_document(&resp.text().await?);
        let clist = response_html
            .select(&Selector::parse("div.section ul.content_list").unwrap())
            .next()
            .unwrap();

        for c in clist.select(&Selector::parse("li").unwrap()) {
            let url = c
                .select(&Selector::parse("a").unwrap())
                .next()
                .unwrap()
                .value()
                .attr("href")
                .unwrap();
            let real_url = strip_url_hash(url, "https://dictionary.goo.ne.jp");
            let title = c.select(&Selector::parse("p.title").unwrap()).next();
            if title.is_none() {
                continue;
            }
            let title = title.unwrap().text();
            let text = c
                .select(&Selector::parse("p.text").unwrap())
                .next()
                .unwrap()
                .text();
            candidates.push(Candidate {
                url: real_url,
                title: title.fold(String::new(), |mut acc, s| {
                    acc.push_str(s);
                    acc
                }),
                text: Some(text.fold(String::new(), |mut acc, s| {
                    acc.push_str(s);
                    acc
                })),
            });
        }

        Ok(candidates)
    }
    pub async fn lookup_definition<S: AsRef<str>>(
        &self,
        url: S,
        query: &str,
    ) -> anyhow::Result<LookupResult> {
        let response_html = self.fetch_html(url).await?;

        if let Some(error_ele) = response_html
            .select(&Selector::parse("div#NR-main div.error").unwrap())
            .next()
        {
            let error = error_ele.text().fold(String::new(), |mut acc, s| {
                acc.push_str(s);
                acc
            });
            return Err(anyhow::anyhow!("Error: {}", error));
        }

        let keyword_ele = response_html
            .select(&Selector::parse("div#NR-main h1").unwrap())
            .next()
            .unwrap();
        let keyword = keyword_ele
            .text()
            .next()
            .unwrap()
            .replace("\n", "")
            .replace("(", "")
            .replace("（", "")
            .replace(")", "")
            .replace("）", "");

        let keywords = keyword.split("／").collect::<Vec<_>>();
        let keyword = keywords
            .iter()
            .max_by(|a, b| match_len(query, a).cmp(&match_len(query, b)))
            .unwrap()
            .to_string();
        let altn_keywords = keywords
            .iter()
            .filter(|k| ***k != keyword)
            .map(|k| k.to_string())
            .collect::<Vec<_>>();

        let yomi_ele = keyword_ele
            .select(&Selector::parse("span.yomi").unwrap())
            .next();
        let yomi = yomi_ele.map(|e| {
            e.text()
                .next()
                .unwrap()
                .replace("(", "")
                .replace("（", "")
                .replace(")", "")
                .replace("）", "")
        });

        let tense_list_ele = response_html
            .select(&Selector::parse("div.section").unwrap())
            .next()
            .unwrap();

        let mut meanings = Vec::new();
        for t in tense_list_ele.select(&Selector::parse("ol.meaning").unwrap()) {
            let mut meaning = String::new();
            for (i, m) in t.select(&Selector::parse(".text").unwrap()).enumerate() {
                if i != 0 {
                    meaning.push_str("\n");
                }
                meaning.push_str(&m.text().fold(String::new(), |mut acc, s| {
                    acc.push_str(s);
                    acc
                }));
            }
            meanings.push(meaning);
        }
        if meanings.len() == 0 {
            for ele in
                response_html.select(&Selector::parse("div.meaning_area div.contents").unwrap())
            {
                let mut meaning = String::new();
                for (i, m) in ele.select(&Selector::parse(".text").unwrap()).enumerate() {
                    if i != 0 {
                        meaning.push_str("\n");
                    }
                    meaning.push_str(&m.text().fold(String::new(), |mut acc, s| {
                        acc.push_str(s);
                        acc
                    }));
                }
                meanings.push(meaning);
            }
        }

        let mut result = LookupResult::new(keyword, "goo_jp");
        result.altn = if altn_keywords.len() > 0 {
            Some(altn_keywords)
        } else {
            None
        };
        result.fu = yomi;
        result.jm = Some(meanings);
        Ok(result)
    }
}

#[async_trait]
impl Lookup for Client {
    async fn lookup(&self, word: &str) -> anyhow::Result<Vec<LookupResult>> {
        let candiates = self.lookup_candidates(word).await?;
        let candidates_stream = tokio_stream::iter(candiates.into_iter());
        let results = candidates_stream
            .map(|c| async move { self.lookup_definition(c.url, word).await })
            .buffer_unordered(10)
            .collect::<Vec<_>>()
            .await
            .into_iter()
            .map(|r| {
                if let Err(e) = r {
                    error!("Failed to lookup definition: {}", e);
                    return None;
                }
                Some(r.unwrap())
            })
            .filter(|r| r.is_some())
            .map(|r| r.unwrap())
            .collect::<Vec<_>>();

        Ok(results)
    }
    async fn lookup_top(&self, word: &str) -> anyhow::Result<LookupResult> {
        let candidates = self.lookup_candidates(word).await?;
        if candidates.len() == 0 {
            return Err(anyhow::anyhow!("No candidates found"));
        }
        let result = self
            .lookup_definition(candidates[0].url.clone(), word)
            .await?;
        Ok(result)
    }
}
