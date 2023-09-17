use diesel::prelude::*;
use serde::{Deserialize, Serialize};

#[derive(Queryable, Selectable, Insertable, Serialize, Deserialize, Debug, Clone)]
#[diesel(table_name = crate::schema::jpn_wordbook)]
pub struct JpnWordbook {
    pub uuid: String,

    pub ja: String,
    pub altn: String,
    pub jm: String,
    pub fu: String,
    pub en: String,
    pub ex: String,

    pub src: String,

    pub created: chrono::NaiveDateTime,
    pub updated: chrono::NaiveDateTime,
}

#[derive(Serialize, Deserialize, Debug, Clone)]
pub struct JpnWordbookExternal {
    pub uuid: String,

    pub ja: String,
    pub altn: Vec<String>,
    pub jm: Vec<String>,
    pub fu: String,
    pub en: Vec<String>,
    pub ex: Vec<String>,

    pub src: String,

    pub created: chrono::NaiveDateTime,
    pub updated: chrono::NaiveDateTime,
}

impl Into<JpnWordbookExternal> for JpnWordbook {
    fn into(self) -> JpnWordbookExternal {
        JpnWordbookExternal {
            uuid: self.uuid,

            ja: self.ja,
            altn: if self.altn.len() > 0 {
                self.altn.split(',').map(|s| s.to_string()).collect()
            } else {
                vec![]
            },
            jm: if self.jm.len() > 0 {
                self.jm.split('\n').map(|s| s.to_string()).collect()
            } else {
                vec![]
            },
            fu: self.fu,
            en: if self.en.len() > 0 {
                self.en.split('\n').map(|s| s.to_string()).collect()
            } else {
                vec![]
            },
            ex: if self.ex.len() > 0 {
                self.ex.split('\n').map(|s| s.to_string()).collect()
            } else {
                vec![]
            },

            src: self.src,

            created: self.created,
            updated: self.updated,
        }
    }
}
