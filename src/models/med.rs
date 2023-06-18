use std::str::FromStr;

use diesel::prelude::*;
use lazy_static::lazy_static;
use regex::Regex;
use serde::{Deserialize, Serialize};

#[derive(Queryable, Selectable, Insertable, Serialize, Deserialize, Debug)]
#[diesel(table_name = crate::schema::medications)]
pub struct Medication {
    pub uuid: String,
    pub name: String,
    pub dosage: i32,
    pub dosage_unit: String,
    pub period_hours: i32,
    pub flags: String,
    pub options: String,
    pub created: chrono::NaiveDateTime,
    pub updated: chrono::NaiveDateTime,
}

impl Medication {
    pub fn flags_split(&self) -> Vec<String> {
        self.flags.split(' ').map(|s| s.to_string()).collect()
    }
    pub fn options_split(&self) -> Vec<(String, String)> {
        lazy_static! {
            static ref REGEX_MED_OPTION: Regex = Regex::new(r"^([a-zA-Z]+)\((\w+)\)$").unwrap();
        };
        self.options
            .split(' ')
            .map(|s| {
                let caps = REGEX_MED_OPTION.captures(s).unwrap();
                (
                    caps.get(1).unwrap().as_str().to_string(),
                    caps.get(2).unwrap().as_str().to_string(),
                )
            })
            .collect()
    }
}

pub const FLAGS_WITH_IMPLIED_FREQ: [&str; 2] = ["qhs", "qam"];

impl Into<String> for Medication {
    fn into(self) -> String {
        let mut output = String::new();

        output.push_str(&self.name);
        output.push(' ');

        output.push_str(&self.dosage.to_string());
        output.push_str(&self.dosage_unit);
        output.push(' ');

        if !FLAGS_WITH_IMPLIED_FREQ.contains(&self.flags.as_str()) {
            match self.period_hours {
                6 => output.push_str("qid"),
                8 => output.push_str("tid"),
                12 => output.push_str("bid"),
                24 => output.push_str("qd"),
                _ => {
                    if self.period_hours % 24 == 0 {
                        output.push_str(&format!("q{}d", self.period_hours / 24));
                    } else {
                        output.push_str(&format!("q{}h", self.period_hours));
                    }
                }
            }
            output.push(' ');
        }

        output.push_str(&self.flags);
        output.push(' ');
        output.push_str(&self.options);

        output.trim().to_string()
    }
}

impl FromStr for Medication {
    type Err = anyhow::Error;

    fn from_str(s: &str) -> Result<Self, Self::Err> {
        lazy_static! {
            static ref REGEX_NUMBER: Regex = Regex::new(r"^\d+$").unwrap();
            static ref REGEX_NUMBER_WITH_UNIT: Regex = Regex::new(r"^(\d+)(\w+)$").unwrap();
            static ref REGEX_MED_OPTION: Regex = Regex::new(r"^([a-zA-Z]+)\((\w+)\)$").unwrap();
        }
        let mut parts = s.split(' ');

        let mut name = String::new();
        let mut flags = Vec::new();
        let mut options = Vec::new();
        let mut dosage = 0;
        let mut dosage_unit = None;
        for part in parts.by_ref() {
            if REGEX_NUMBER.is_match(part) {
                dosage = part.parse()?;
                break;
            } else if let Some(caps) = REGEX_NUMBER_WITH_UNIT.captures(part) {
                dosage = caps.get(1).unwrap().as_str().parse()?;
                dosage_unit = Some(caps.get(2).unwrap().as_str().to_string());
                break;
            } else {
                name.push_str(part);
                name.push(' ');
            }
        }
        if dosage_unit.is_none() {
            dosage_unit = parts.next().map(|s| s.to_string());
        }
        let period_spec = parts
            .next()
            .ok_or(anyhow::anyhow!("missing period spec"))?
            .to_lowercase();
        let period_hours = match period_spec.as_str() {
            "qd" => 24,
            "bid" => 12,
            "tid" => 8,
            "qid" => 6,
            "qhs" => {
                flags.push("qhs".to_string());
                24
            }
            "qam" => {
                flags.push("qam".to_string());
                24
            }
            _ => {
                if period_spec.starts_with("q") {
                    let period_unit = period_spec.chars().last().unwrap();
                    let period_duration = period_spec[1..period_spec.len() - 1].parse()?;
                    match period_unit {
                        'h' => period_duration,
                        'd' => period_duration * 24,
                        _ => return Err(anyhow::anyhow!("invalid period spec")),
                    }
                } else {
                    return Err(anyhow::anyhow!("invalid period spec"));
                }
            }
        };
        for part in parts {
            if let Some(caps) = REGEX_MED_OPTION.captures(part) {
                let opt_name = caps.get(1).unwrap().as_str();
                let opt_value = caps.get(2).unwrap().as_str();
                options.push((opt_name.to_string(), opt_value.to_string()));
            } else {
                flags.push(part.to_string());
            }
        }
        Ok(Self {
            uuid: uuid::Uuid::new_v4().to_string(),
            name: name.trim().to_string(),
            dosage,
            dosage_unit: dosage_unit.unwrap(),
            period_hours,
            flags: flags.join(" "),
            options: options
                .iter()
                .map(|(k, v)| format!("{}({})", k, v))
                .collect::<Vec<String>>()
                .join(" "),
            created: chrono::Utc::now().naive_utc(),
            updated: chrono::Utc::now().naive_utc(),
        })
    }
}

#[derive(Queryable, Selectable, Insertable, Serialize, Deserialize, Debug, Clone)]
#[diesel(table_name = crate::schema::medication_logs)]
pub struct MedicationLog {
    pub uuid: String,
    pub med_uuid: String,
    pub dosage: i32,
    pub time_actual: chrono::NaiveDateTime,
    pub time_expected: chrono::NaiveDateTime,
    pub dose_offset: f32,
    pub created: chrono::NaiveDateTime,
    pub updated: chrono::NaiveDateTime,
}
