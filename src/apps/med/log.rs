use std::sync::Arc;

use crate::{
    apps::auth::{middleware::AuthInfo, Role},
    http::{ApiResponse, ApiResult, JsonApiForm},
    models::med::{Medication, MedicationLog},
};

use anyhow::Result;
use axum::{
    extract::{Path, Query},
    Extension,
};
use chrono::{Duration, TimeZone, Utc};
use log::error;
use serde::{Deserialize, Serialize};

use super::MedManagementApp;

pub fn effective_last_dose(
    med: &Medication,
    mut med_logs: Vec<MedicationLog>,
) -> Option<MedicationLog> {
    let mut remaining_dose = med.dosage;
    med_logs.sort_by(|a, b| a.time_actual.cmp(&b.time_actual));
    med_logs.reverse();

    for log in med_logs {
        if log.dosage >= remaining_dose {
            return Some(log);
        } else {
            remaining_dose -= log.dosage;
        }
    }

    None
}

pub fn project_next_dose(med: &Medication, med_logs: Vec<MedicationLog>) -> MedicationLog {
    let effective_last = effective_last_dose(med, med_logs);

    let now = Utc::now().naive_utc();

    let next_time = match effective_last {
        Some(last) => last.time_actual + Duration::hours(med.period_hours as i64),
        None => now,
    };

    let offset = (now.timestamp_millis() - next_time.timestamp_millis()) as f64
        / (med.period_hours * 60 * 60 * 1000) as f64;

    MedicationLog {
        uuid: uuid::Uuid::new_v4().to_string(),
        med_uuid: med.uuid.clone(),
        dosage: med.dosage,
        time_actual: now,
        time_expected: next_time,
        dose_offset: offset as f32,
        created: now,
        updated: now,
    }
}

pub async fn project_next_doses(
    app: &MedManagementApp,
) -> Result<Vec<(Medication, MedicationLog)>> {
    let state = app.state.lock().await;
    let state = state.as_ref().unwrap();
    let mut global_app_state = state.global_app_state.lock().await;

    let meds = {
        use crate::schema::medications::dsl::*;
        use diesel::prelude::*;

        medications
            .load::<Medication>(&mut global_app_state.db)
            .map_err(|e| {
                error!("Failed to load meds: {:?}", e);
                e
            })?
    };

    let mut med_logs = Vec::with_capacity(meds.len());
    for med in &meds {
        let logs = {
            use crate::schema::medication_logs::dsl::*;
            use diesel::prelude::*;

            medication_logs
                .order(time_actual.desc())
                .filter(med_uuid.eq(&med.uuid))
                .limit(10)
                .load::<MedicationLog>(&mut global_app_state.db)
                .map_err(|e| {
                    error!("Failed to load logs: {:?}", e);
                    e
                })?
        };

        med_logs.push(logs);
    }

    let mut res = Vec::with_capacity(meds.len());

    for (med, logs) in meds.into_iter().zip(med_logs.into_iter()) {
        let next_dose = project_next_dose(&med, logs);
        res.push((med, next_dose));
    }

    Ok(res)
}

pub async fn route_project_next_dose(
    auth: AuthInfo,
    app: Extension<Arc<MedManagementApp>>,
    Path(med_uuid_form): Path<String>,
) -> ApiResult<MedicationLog> {
    auth.check_for_any_role(&[Role::Admin])?;

    let state = app.state.lock().await;
    let state = state.as_ref().unwrap();
    let mut global_app_state = state.global_app_state.lock().await;

    let med = {
        use crate::schema::medications::dsl::*;
        use diesel::prelude::*;

        medications
            .filter(uuid.eq(&med_uuid_form))
            .first::<Medication>(&mut global_app_state.db)
            .map_err(|e| {
                error!("Failed to get med: {}", e);
                ApiResponse::error("Failed to get med".to_string(), 500, None)
            })?
    };

    let med_logs = {
        use crate::schema::medication_logs::dsl::*;
        use diesel::prelude::*;

        medication_logs
            .order_by(time_actual.desc())
            .limit(10)
            .filter(med_uuid.eq(&med_uuid_form))
            .load::<MedicationLog>(&mut global_app_state.db)
            .map_err(|e| {
                error!("Failed to get med logs: {}", e);
                ApiResponse::error("Failed to get med logs".to_string(), 500, None)
            })?
    };

    let next_dose = project_next_dose(&med, med_logs);
    Ok(ApiResponse::ok(
        "Next dose projected successfully".to_string(),
        Some(next_dose),
    ))
}

#[derive(Serialize, Deserialize)]
pub struct GetLogOptions {
    until: Option<i64>,
    limit: i64,
}

pub async fn route_get_log(
    auth: AuthInfo,
    app: Extension<Arc<MedManagementApp>>,
    Path(med_uuid_form): Path<String>,
    Query(GetLogOptions { until, limit }): Query<GetLogOptions>,
) -> ApiResult<Vec<MedicationLog>> {
    auth.check_for_any_role(&[Role::Admin])?;

    let state = app.state.lock().await;
    let state = state.as_ref().unwrap();
    let mut global_app_state = state.global_app_state.lock().await;

    let med_logs = {
        use crate::schema::medication_logs::dsl::*;
        use diesel::prelude::*;

        match until {
            None => medication_logs
                .limit(limit)
                .order_by(time_actual.desc())
                .filter(med_uuid.eq(&med_uuid_form))
                .load::<MedicationLog>(&mut global_app_state.db)
                .map_err(|e| {
                    error!("Failed to get med logs: {}", e);
                    ApiResponse::error("Failed to get med logs".to_string(), 500, None)
                })?,
            Some(until) => {
                let until = Utc.timestamp_millis_opt(until).unwrap().naive_utc();
                medication_logs
                    .limit(limit)
                    .order_by(time_actual.desc())
                    .filter(med_uuid.eq(&med_uuid_form).and(time_actual.lt(until)))
                    .load::<MedicationLog>(&mut global_app_state.db)
                    .map_err(|e| {
                        error!("Failed to get med logs: {}", e);
                        ApiResponse::error("Failed to get med logs".to_string(), 500, None)
                    })?
            }
        }
    };

    Ok(ApiResponse::ok(
        "Med logs retrieved successfully".to_string(),
        Some(med_logs),
    ))
}

pub async fn route_post_log(
    auth: AuthInfo,
    app: Extension<Arc<MedManagementApp>>,
    Path(med_uuid_form): Path<String>,
    JsonApiForm(form): JsonApiForm<MedicationLog>,
) -> ApiResult<MedicationLog> {
    auth.check_for_any_role(&[Role::Admin])?;

    let state = app.state.lock().await;
    let state = state.as_ref().unwrap();
    let mut global_app_state = state.global_app_state.lock().await;

    let med = {
        use crate::schema::medications::dsl::*;
        use diesel::prelude::*;

        medications
            .filter(uuid.eq(&med_uuid_form))
            .first::<Medication>(&mut global_app_state.db)
            .map_err(|e| {
                error!("Failed to get med: {}", e);
                ApiResponse::error("Failed to get med".to_string(), 500, None)
            })?
    };

    let med_logs = {
        use crate::schema::medication_logs::dsl::*;
        use diesel::prelude::*;

        medication_logs
            .order_by(time_actual.desc())
            .limit(10)
            .filter(
                med_uuid
                    .eq(&med_uuid_form)
                    .and(time_actual.lt(form.time_actual)),
            )
            .load::<MedicationLog>(&mut global_app_state.db)
            .map_err(|e| {
                error!("Failed to get med logs: {}", e);
                ApiResponse::error("Failed to get med logs".to_string(), 500, None)
            })?
    };

    let projected_next_dose = project_next_dose(&med, med_logs);
    let mut form = form;

    form.med_uuid = med_uuid_form;
    form.time_expected = projected_next_dose.time_expected;
    form.dose_offset = projected_next_dose.dose_offset;
    let now = Utc::now().naive_utc();
    form.created = now;
    form.updated = now;

    let log = {
        use crate::schema::medication_logs::dsl::*;
        use diesel::prelude::*;

        diesel::insert_into(medication_logs)
            .values(form.clone())
            .execute(&mut global_app_state.db)
            .map_err(|e| {
                error!("Failed to insert log: {}", e);
                ApiResponse::error("Failed to insert log".to_string(), 500, None)
            })?;

        medication_logs
            .filter(uuid.eq(form.uuid))
            .first::<MedicationLog>(&mut global_app_state.db)
            .map_err(|e| {
                error!("Failed to get log: {}", e);
                ApiResponse::error("Failed to get log".to_string(), 500, None)
            })?
    };

    Ok(ApiResponse::ok(
        "Log inserted successfully".to_string(),
        Some(log),
    ))
}

pub async fn route_delete_log(
    auth: AuthInfo,
    app: Extension<Arc<MedManagementApp>>,
    Path((_med_uuid, log_uuid)): Path<(String, String)>,
) -> ApiResult<()> {
    auth.check_for_any_role(&[Role::Admin])?;

    let state = app.state.lock().await;
    let state = state.as_ref().unwrap();
    let mut global_app_state = state.global_app_state.lock().await;

    {
        use crate::schema::medication_logs::dsl::*;
        use diesel::prelude::*;

        diesel::delete(medication_logs.filter(uuid.eq(log_uuid)))
            .execute(&mut global_app_state.db)
            .map_err(|e| {
                error!("Failed to delete log: {}", e);
                ApiResponse::error("Failed to delete log".to_string(), 500, None)
            })?;
    }

    Ok(ApiResponse::ok(
        "Log deleted successfully".to_string(),
        None,
    ))
}
