use std::{collections::HashMap, sync::Arc};

use axum::{extract::Query, Extension};
use futures::StreamExt;
use log::{error, info};
use serde::Deserialize;

use crate::http::{ApiResponse, ApiResult};

use super::{
    sources::{Lookup, LookupResult},
    JpnWordBookApp,
};

#[derive(Deserialize)]
pub struct SearchQuery {
    pub query: String,
}

pub async fn route_jisho_search(
    Extension(app): Extension<Arc<JpnWordBookApp>>,
    Query(query): Query<SearchQuery>,
) -> ApiResult<Vec<LookupResult>> {
    let state = app.state.lock().await;
    let state = state.as_ref().unwrap();

    let results = state.jisho.lookup(&query.query).await.map_err(|e| {
        error!("Failed to lookup word: {}", e);
        ApiResponse::error("Failed to lookup word".to_string(), 500, None)
    })?;

    Ok(ApiResponse::ok(
        "Jisho search successful".to_string(),
        Some(results),
    ))
}

pub async fn route_jisho_search_top(
    Extension(app): Extension<Arc<JpnWordBookApp>>,
    Query(query): Query<SearchQuery>,
) -> ApiResult<LookupResult> {
    let state = app.state.lock().await;
    let state = state.as_ref().unwrap();

    let result = state.jisho.lookup_top(&query.query).await.map_err(|e| {
        error!("Failed to lookup word: {}", e);
        ApiResponse::error("Failed to lookup word".to_string(), 500, None)
    })?;

    Ok(ApiResponse::ok(
        "Jisho search successful".to_string(),
        Some(result),
    ))
}

pub async fn route_goo_search(
    Extension(app): Extension<Arc<JpnWordBookApp>>,
    Query(query): Query<SearchQuery>,
) -> ApiResult<Vec<LookupResult>> {
    let state = app.state.lock().await;
    let state = state.as_ref().unwrap();

    let results = state.goo.lookup(&query.query).await.map_err(|e| {
        error!("Failed to lookup word: {}", e);
        ApiResponse::error("Failed to lookup word".to_string(), 500, None)
    })?;

    Ok(ApiResponse::ok(
        "Goo search successful".to_string(),
        Some(results),
    ))
}

pub async fn route_tatoeba_search(
    Extension(app): Extension<Arc<JpnWordBookApp>>,
    Query(query): Query<SearchQuery>,
) -> ApiResult<Vec<LookupResult>> {
    let state = app.state.lock().await;
    let state = state.as_ref().unwrap();

    let results = state.tatoeba.lookup(&query.query).await.map_err(|e| {
        error!("Failed to lookup word: {}", e);
        ApiResponse::error("Failed to lookup word".to_string(), 500, None)
    })?;

    Ok(ApiResponse::ok(
        "Tatoeba search successful".to_string(),
        Some(results),
    ))
}

pub async fn route_combo_search(
    Extension(app): Extension<Arc<JpnWordBookApp>>,
    Query(query): Query<SearchQuery>,
) -> ApiResult<Vec<LookupResult>> {
    let state = app.state.lock().await;
    let state = state.as_ref().unwrap();

    let start = std::time::Instant::now();

    let results_jisho = state.jisho.lookup(&query.query).await.map_err(|e| {
        error!("Failed to lookup word: {}", e);
        ApiResponse::error("Failed to lookup word".to_string(), 500, None)
    })?;

    let time = start.elapsed();
    info!("Jisho lookup took {}ms", time.as_millis());
    let start = std::time::Instant::now();

    let results_goo = state.goo.lookup(&query.query).await.map_err(|e| {
        error!("Failed to lookup word: {}", e);
        ApiResponse::error("Failed to lookup word".to_string(), 500, None)
    })?;

    let time = start.elapsed();
    info!("Goo lookup took {}ms", time.as_millis());

    let mut combined_results = HashMap::new();
    for result in results_jisho {
        combined_results.insert(result.ja.clone(), result);
    }
    for result in results_goo {
        if let Some(existing_result) = combined_results.get_mut(&result.ja) {
            existing_result.merge(result);
        } else {
            combined_results.insert(result.ja.clone(), result);
        }
    }

    let combined_results = combined_results
        .into_iter()
        .map(|(_, v)| v)
        .collect::<Vec<_>>();

    let start = std::time::Instant::now();

    let combined_results_stream = tokio_stream::iter(combined_results.into_iter());
    let mut combined_results = combined_results_stream
        .map(|mut r| async move {
            r.merge(state.tatoeba.lookup_top(&r.ja).await.unwrap());
            r.clone()
        })
        .buffer_unordered(10)
        .collect::<Vec<_>>()
        .await;

    let time = start.elapsed();
    info!("Tatoeba lookup took {}ms", time.as_millis());

    combined_results.sort_by(|a, b| {
        b.match_score(&query.query)
            .cmp(&a.match_score(&query.query))
    });

    Ok(ApiResponse::ok(
        "Combo search successful".to_string(),
        Some(combined_results),
    ))
}

pub async fn route_combo_search_top(
    Extension(app): Extension<Arc<JpnWordBookApp>>,
    Query(query): Query<SearchQuery>,
) -> ApiResult<LookupResult> {
    let state = app.state.lock().await;
    let state = state.as_ref().unwrap();

    let results_jisho = state.jisho.lookup_top(&query.query).await.map_err(|e| {
        error!("Failed to lookup word: {}", e);
        ApiResponse::error("Failed to lookup word".to_string(), 500, None)
    })?;

    let results_goo = state.goo.lookup_top(&query.query).await.map_err(|e| {
        error!("Failed to lookup word: {}", e);
        ApiResponse::error("Failed to lookup word".to_string(), 500, None)
    })?;

    let mut combined_result = if results_jisho.ja != results_goo.ja {
        if results_jisho.match_score(&query.query) > results_goo.match_score(&query.query) {
            results_jisho
        } else {
            results_goo
        }
    } else {
        let mut combined_result = results_jisho;
        combined_result.merge(results_goo);
        combined_result
    };

    combined_result.merge(state.tatoeba.lookup_top(&combined_result.ja).await.unwrap());

    Ok(ApiResponse::ok(
        "Combo search successful".to_string(),
        Some(combined_result),
    ))
}
