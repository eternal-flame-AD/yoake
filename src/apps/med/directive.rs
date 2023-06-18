use std::sync::Arc;

use crate::{
    apps::auth::{middleware::AuthInfo, Role},
    http::{ApiResponse, ApiResult, JsonApiForm},
    models::med::Medication,
};

use axum::{extract::Path, Extension};
use log::error;
use serde::{Deserialize, Serialize};

use super::MedManagementApp;

#[derive(Debug, Serialize, Deserialize)]
pub struct ParseShorthandForm {
    pub shorthand: String,
}

pub async fn route_parse_shorthand(
    JsonApiForm(form): JsonApiForm<ParseShorthandForm>,
) -> ApiResult<Medication> {
    let res = form.shorthand.parse::<Medication>().map_err(|e| {
        error!("Failed to parse shorthand: {}", e);
        ApiResponse::bad_request("Failed to parse".to_string(), None)
    })?;
    Ok(ApiResponse::ok(
        "Parsed successfully".to_string(),
        Some(res),
    ))
}

pub async fn route_format_shorthand(
    JsonApiForm(form): JsonApiForm<crate::models::med::Medication>,
) -> ApiResult<String> {
    let shorthand: String = form.into();
    Ok(ApiResponse::ok(
        "Formatted successfully".to_string(),
        Some(shorthand),
    ))
}

pub async fn route_get_directive(
    auth: AuthInfo,
    app: Extension<Arc<MedManagementApp>>,
) -> ApiResult<Vec<Medication>> {
    auth.check_for_any_role(&[Role::Admin])?;

    let state = app.state.lock().await;
    let state = state.as_ref().unwrap();
    let mut global_app_state = state.global_app_state.lock().unwrap();

    let meds = {
        use crate::schema::medications::dsl::*;
        use diesel::prelude::*;

        medications
            .load::<Medication>(&mut global_app_state.db)
            .map_err(|e| {
                error!("Failed to load meds: {:?}", e);
                ApiResponse::error("Database error".to_string(), 500, None)
            })?
    };

    Ok(ApiResponse::ok(
        "Directives retrieved".to_string(),
        Some(meds),
    ))
}

pub async fn route_post_directive(
    auth: AuthInfo,
    app: Extension<Arc<MedManagementApp>>,
    JsonApiForm(mut form): JsonApiForm<crate::models::med::Medication>,
) -> ApiResult<Medication> {
    auth.check_for_any_role(&[Role::Admin])?;

    let state = app.state.lock().await;
    let state = state.as_ref().unwrap();
    let mut global_app_state = state.global_app_state.lock().unwrap();

    form.uuid = uuid::Uuid::new_v4().to_string();
    form.created = chrono::Utc::now().naive_local();
    form.updated = chrono::Utc::now().naive_local();

    let res = {
        use crate::schema::medications;
        use crate::schema::medications::dsl::*;
        use diesel::prelude::*;

        diesel::insert_into(medications::table)
            .values(&form)
            .execute(&mut global_app_state.db)
            .unwrap();

        medications
            .filter(medications::uuid.eq(&form.uuid))
            .first::<crate::models::med::Medication>(&mut global_app_state.db)
            .unwrap()
    };

    Ok(ApiResponse::ok("Directives posted".to_string(), Some(res)))
}

pub async fn route_patch_directive(
    auth: AuthInfo,
    app: Extension<Arc<MedManagementApp>>,
    JsonApiForm(form): JsonApiForm<crate::models::med::Medication>,
) -> ApiResult<Medication> {
    auth.check_for_any_role(&[Role::Admin])?;

    let state = app.state.lock().await;
    let state = state.as_ref().unwrap();
    let mut global_app_state = state.global_app_state.lock().unwrap();

    let res = {
        use crate::schema::medications;
        use crate::schema::medications::dsl::*;
        use diesel::prelude::*;

        diesel::update(medications.filter(medications::uuid.eq(&form.uuid)))
            .set((
                name.eq(&form.name),
                dosage.eq(&form.dosage),
                dosage_unit.eq(&form.dosage_unit),
                period_hours.eq(&form.period_hours),
                flags.eq(&form.flags),
                options.eq(&form.options),
                updated.eq(chrono::Utc::now().naive_local()),
            ))
            .execute(&mut global_app_state.db)
            .map_err(|e| {
                error!("Failed to update med: {:?}", e);
                ApiResponse::error("Database error".to_string(), 500, None)
            })?;

        medications
            .filter(medications::uuid.eq(&form.uuid))
            .first::<crate::models::med::Medication>(&mut global_app_state.db)
            .map_err(|e| {
                error!("Failed to load med: {:?}", e);
                ApiResponse::error("Database error".to_string(), 500, None)
            })?
    };

    Ok(ApiResponse::ok("Directives updated".to_string(), Some(res)))
}

pub async fn route_delete_directive(
    auth: AuthInfo,
    app: Extension<Arc<MedManagementApp>>,
    Path(med_uuid): Path<String>,
) -> ApiResult<()> {
    auth.check_for_any_role(&[Role::Admin])?;

    let state = app.state.lock().await;
    let state = state.as_ref().unwrap();
    let mut global_app_state = state.global_app_state.lock().unwrap();

    {
        use crate::schema::medications::dsl::medications;
        use diesel::prelude::*;

        diesel::delete(medications.filter(crate::schema::medications::uuid.eq(&med_uuid)))
            .execute(&mut global_app_state.db)
            .map_err(|e| {
                error!("Failed to delete med: {:?}", e);
                ApiResponse::error("Database error".to_string(), 500, None)
            })?;
    };

    Ok(ApiResponse::<()>::ok(
        "Directives deleted".to_string(),
        None,
    ))
}
