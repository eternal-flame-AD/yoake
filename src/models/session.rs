use diesel::prelude::*;

#[derive(Queryable, Selectable, Insertable)]
#[diesel(table_name = crate::schema::sessions)]
pub struct Session {
    pub uuid: String,
    pub expiry: chrono::NaiveDateTime,
    pub content: String,
}
