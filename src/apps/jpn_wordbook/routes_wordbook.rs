use std::sync::Arc;

use axum::{extract::Query, Extension, Json};
use chrono::NaiveDateTime;
use hyper::body::Bytes;
use log::error;
use serde::Deserialize;

use crate::{
    apps::auth::{middleware::AuthInfo, Role},
    http::{ApiResponse, ApiResult},
    models::jpn_wordbook::{JpnWordbook, JpnWordbookExternal},
};

use super::{sources::LookupResult, JpnWordBookApp};

#[derive(Deserialize)]
pub struct GetWordbookOptions {
    until: Option<NaiveDateTime>,
    limit: i64,
}

#[derive(Deserialize)]
pub struct GetCsvOptions {
    header: Option<bool>,
}

pub async fn route_get_wordbook_csv(
    Extension(app): Extension<Arc<JpnWordBookApp>>,
    Query(GetCsvOptions { header }): Query<GetCsvOptions>,
) -> Result<Bytes, ApiResponse<()>> {
    let state = app.state.lock().await;
    let state = state.as_ref().unwrap();
    let mut global_app_state = state.global_app_state.lock().await;

    let results = {
        use crate::schema::jpn_wordbook::dsl::*;
        use diesel::prelude::*;

        jpn_wordbook
            .order_by(created.desc())
            .load::<JpnWordbook>(&mut global_app_state.db)
            .map_err(|e| {
                error!("Failed to get wordbook: {}", e);
                ApiResponse::error("Failed to get wordbook".to_string(), 500, None)
            })?
    };

    let mut csv_file = Vec::new();
    let mut csv_writer = csv::Writer::from_writer(&mut csv_file);
    if header.unwrap_or(true) {
        csv_writer
            .write_record(&["uuid", "ja", "altn", "jm", "fu", "en", "ex", "src"])
            .map_err(|e| {
                error!("Failed to write csv header: {}", e);
                ApiResponse::error("Failed to write csv header".to_string(), 500, None)
            })?;
    }

    for result in results {
        csv_writer
            .write_record(&[
                &result.uuid,
                &result.ja,
                &result.altn,
                &result.jm,
                &result.fu,
                &result.en,
                &result.ex,
                &result.src,
            ])
            .map_err(|e| {
                error!("Failed to write csv record: {}", e);
                ApiResponse::error("Failed to write csv record".to_string(), 500, None)
            })?;
    }

    csv_writer.flush().map_err(|e| {
        error!("Failed to flush csv writer: {}", e);
        ApiResponse::error("Failed to flush csv writer".to_string(), 500, None)
    })?;

    drop(csv_writer);

    Ok(Bytes::from(csv_file))
}

pub async fn route_get_wordbook(
    Extension(app): Extension<Arc<JpnWordBookApp>>,
    Query(GetWordbookOptions { until, limit }): Query<GetWordbookOptions>,
) -> ApiResult<Vec<JpnWordbookExternal>> {
    let state = app.state.lock().await;
    let state = state.as_ref().unwrap();
    let mut global_app_state = state.global_app_state.lock().await;

    let results = {
        use crate::schema::jpn_wordbook::dsl::*;
        use diesel::prelude::*;

        match until {
            None => jpn_wordbook
                .limit(limit)
                .order_by(created.desc())
                .load::<JpnWordbook>(&mut global_app_state.db)
                .map_err(|e| {
                    error!("Failed to get wordbook: {}", e);
                    ApiResponse::error("Failed to get wordbook".to_string(), 500, None)
                })?,
            Some(until) => jpn_wordbook
                .limit(limit)
                .order_by(created.desc())
                .filter(created.lt(until))
                .load::<JpnWordbook>(&mut global_app_state.db)
                .map_err(|e| {
                    error!("Failed to get wordbook: {}", e);
                    ApiResponse::error("Failed to get wordbook".to_string(), 500, None)
                })?,
        }
    };

    Ok(ApiResponse::ok(
        "Word lookup successful".to_string(),
        Some(results.into_iter().map(Into::into).collect()),
    ))
}

pub async fn route_store_wordbook(
    auth: AuthInfo,
    Extension(app): Extension<Arc<JpnWordBookApp>>,
    Json(word): Json<LookupResult>,
) -> ApiResult<()> {
    auth.check_for_any_role(&[Role::Admin])?;

    let state = app.state.lock().await;
    let state = state.as_ref().unwrap();
    let mut global_app_state = state.global_app_state.lock().await;

    let existing_uuid = {
        use crate::schema::jpn_wordbook::dsl::*;
        use diesel::prelude::*;

        jpn_wordbook
            .select(uuid)
            .filter(ja.eq(&word.ja))
            .first::<String>(&mut global_app_state.db)
            .optional()
            .map_err(|e| {
                error!("Failed to get word: {}", e);
                ApiResponse::error("Failed to get word".to_string(), 500, None)
            })?
    };

    {
        use crate::schema::jpn_wordbook::dsl::*;
        use diesel::prelude::*;

        diesel::delete(jpn_wordbook)
            .filter(ja.eq(&word.ja))
            .execute(&mut global_app_state.db)
            .map_err(|e| {
                error!("Failed to delete word: {}", e);
                ApiResponse::error("Failed to delete word".to_string(), 500, None)
            })?;

        let mut word = Into::<JpnWordbook>::into(word);
        if let Some(existing_uuid) = existing_uuid {
            word.uuid = existing_uuid;
        }

        diesel::insert_into(jpn_wordbook)
            .values(&word)
            .execute(&mut global_app_state.db)
            .map_err(|e| {
                error!("Failed to insert word: {}", e);
                ApiResponse::error("Failed to insert word".to_string(), 500, None)
            })?;
    }

    Ok(ApiResponse::ok(
        "Word inserted successfully".to_string(),
        None,
    ))
}
