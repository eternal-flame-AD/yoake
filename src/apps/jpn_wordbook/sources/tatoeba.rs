use std::{
    collections::HashMap,
    io::{BufRead, BufReader},
    path::PathBuf,
};

use super::{Lookup, LookupResult};
use async_trait::async_trait;
use tokio::{fs::File, io::AsyncWriteExt};

pub struct Client {
    client: reqwest::Client,
    data_dir: PathBuf,
    chars_index: HashMap<char, CharIndex>,
}

pub enum CharIndex {
    Hot,
    Cold(Vec<usize>),
}

impl Client {
    pub async fn new(client: reqwest::Client, data_dir: PathBuf) -> anyhow::Result<Self> {
        let mut ret = Self {
            client,
            data_dir,
            chars_index: HashMap::new(),
        };
        ret.download_sentences().await?;
        ret.build_char_index()?;
        Ok(ret)
    }
    pub async fn download_sentences(&self) -> anyhow::Result<()> {
        let mut output_file = File::create(self.data_dir.join("jpn_sentences.tsv.bz2")).await?;
        let url = "https://downloads.tatoeba.org/exports/per_language/jpn/jpn_sentences.tsv.bz2";
        let mut response = self.client.get(url).send().await?;
        while let Some(chunk) = response.chunk().await? {
            output_file.write_all(&chunk).await?;
        }

        Ok(())
    }
    fn open_sentences_file(
        &self,
    ) -> anyhow::Result<BufReader<bzip2::read::BzDecoder<std::fs::File>>> {
        let input_file = std::fs::File::open(self.data_dir.join("jpn_sentences.tsv.bz2"))?;
        let decompressor = bzip2::read::BzDecoder::new(input_file);
        let reader = BufReader::new(decompressor);
        Ok(reader)
    }
    pub fn build_char_index(&mut self) -> anyhow::Result<()> {
        let reader = self.open_sentences_file()?;
        for (line_no, line) in reader.lines().enumerate() {
            let line = line?;
            let mut fields = line.split('\t');
            let _id = fields.next().unwrap();
            let lang = fields.next().unwrap();
            let text = fields.next().unwrap();
            if lang != "jpn" {
                continue;
            }
            for c in text.chars() {
                let entry = self
                    .chars_index
                    .entry(c)
                    .or_insert(CharIndex::Cold(Vec::new()));
                match entry {
                    CharIndex::Hot => {
                        continue;
                    }
                    CharIndex::Cold(v) => {
                        v.push(line_no);
                        if v.len() > 500 {
                            *entry = CharIndex::Hot;
                        }
                    }
                }
            }
        }

        Ok(())
    }
    pub fn search_char_index(&self, word: &str) -> anyhow::Result<Option<Vec<usize>>> {
        let mut result = None;

        for c in word.chars() {
            if let Some(entry) = self.chars_index.get(&c) {
                match entry {
                    CharIndex::Hot => {
                        continue;
                    }
                    CharIndex::Cold(v) => {
                        if result.is_none() {
                            result = Some(v.clone());
                        } else {
                            let mut new_result = Vec::new();
                            for i in result.unwrap() {
                                if v.contains(&i) {
                                    new_result.push(i);
                                }
                            }
                            result = Some(new_result);
                        }
                    }
                }
            } else {
                return Ok(Vec::new().into());
            }
        }

        Ok(result)
    }
    pub fn search_sentences(&self, word: &str) -> anyhow::Result<Vec<String>> {
        let possible_line_nos = self.search_char_index(word)?.map(|v| {
            let mut v = v;
            v.sort();
            v
        });
        let mut next_line_no_idx = 0;

        let reader = self.open_sentences_file()?;
        let mut results = Vec::new();
        for (i, line) in reader.lines().enumerate() {
            if let Some(line_nos) = &possible_line_nos {
                if next_line_no_idx >= line_nos.len() {
                    break;
                }
                if i != line_nos[next_line_no_idx] {
                    continue;
                }
                next_line_no_idx += 1;
            }
            let line = line?;
            let mut fields = line.split('\t');
            let _id = fields.next().unwrap();
            let lang = fields.next().unwrap();
            let text = fields.next().unwrap();
            if lang != "jpn" {
                continue;
            }
            if text.contains(word) {
                results.push(text.to_string());
            }
        }

        Ok(results)
    }
}

#[async_trait]
impl Lookup for Client {
    async fn lookup(&self, word: &str) -> anyhow::Result<Vec<LookupResult>> {
        let examples = tokio::task::block_in_place(|| self.search_sentences(word))?;
        let mut result = LookupResult::new(word.to_string(), "tatoeba");
        result.ex = Some(examples);

        Ok(vec![result])
    }
}
