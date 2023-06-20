use std::sync::Arc;

use askama::Template;
use async_trait::async_trait;
use axum::{body::HttpBody, extract::Query, http::Request, routing::get, Extension, Router};
use log::{debug, info};
use serde::{Deserialize, Serialize};
use tokio::sync::Mutex;

use crate::{
    apps::{
        auth::Role,
        canvas_lms::{
            comm::{GradesTemplate, TemplateGrade},
            grading::CanvasGradingData,
        },
    },
    comm::{Communicator, Message},
    config::Config,
    AppState,
};

use self::{
    grading::Grading,
    graph::{AllCourseData, GraphQuery, GraphResponse, ALL_COURSES_QUERY},
};

use crate::{
    apps::{auth::middleware::AuthInfo, App},
    http::{ApiResponse, ApiResult},
};

mod comm;
pub mod grading;
mod graph;

pub async fn query_grades(
    endpoint: &str,
    token: &str,
    maxn: i32,
) -> Result<GraphResponse<AllCourseData>, reqwest::Error> {
    let client = reqwest::Client::new();
    let query = GraphQuery {
        query: ALL_COURSES_QUERY
            .to_string()
            .replace("$maxn$", &maxn.to_string()),
        operation_name: "gradeQuery".to_string(),
        variables: (),
    };
    let res = client
        .post(endpoint)
        .bearer_auth(token)
        .json(&query)
        .send()
        .await?
        .json::<GraphResponse<AllCourseData>>()
        .await?;
    Ok(res)
}

pub struct CanvasLMSApp {
    state: Mutex<CanvasLMSAppState>,
}

struct CanvasLMSAppState {
    config: Option<&'static Config>,
    grade_cache: GradeCache,
    global_app_state: Option<Arc<Mutex<AppState>>>,
}

#[derive(Debug, Clone, Serialize)]
struct GradeCache {
    last_updated: chrono::DateTime<chrono::Local>,
    response: Option<Vec<Grading>>,
}

#[derive(Debug, Deserialize)]
struct GetGradesOptions {
    force_refresh: Option<bool>,
}

async fn route_get_grades<B: HttpBody>(
    auth: AuthInfo,
    app: Extension<Arc<CanvasLMSApp>>,
    Query(query): Query<GetGradesOptions>,
    _req: Request<B>,
) -> ApiResult<GradeCache>
where
    <B as HttpBody>::Error: std::fmt::Debug,
{
    auth.check_for_any_role(&[Role::Admin])?;

    if Some(true) == query.force_refresh || !app.grade_loaded().await {
        app.refresh_grades().await;
    }
    let state = app.state.lock().await;
    let grade_cache = state.grade_cache.to_owned();
    if grade_cache.response.is_none() {
        return Err(ApiResponse::<()>::error(
            "Grades not available yet".to_string(),
            503,
            None,
        ));
    }
    Ok(ApiResponse::ok(
        "Grades retrieved successfully".to_string(),
        Some(grade_cache),
    ))
}

impl CanvasLMSApp {
    pub fn new() -> Self {
        Self {
            state: Mutex::new(CanvasLMSAppState {
                config: None,
                grade_cache: GradeCache {
                    last_updated: chrono::Local::now(),
                    response: None,
                },
                global_app_state: None,
            }),
        }
    }
    pub(crate) async fn grade_loaded(&self) -> bool {
        let state: tokio::sync::MutexGuard<CanvasLMSAppState> = self.state.lock().await;
        state.grade_cache.response.is_some()
    }
    pub(crate) async fn refresh_grades(&self) {
        let mut state = self.state.lock().await;
        let config = state.config.unwrap();
        let res = query_grades(&config.canvas_lms.endpoint, &config.canvas_lms.token, 50).await;
        match res {
            Ok(res) => {
                let mut res_generalized =
                    CanvasGradingData(res.data).into_iter().collect::<Vec<_>>();
                res_generalized.sort_unstable();
                res_generalized.reverse();
                debug!("Finished refreshing grades");
                if let Some(old_grades) = &state.grade_cache.response {
                    let updates = Grading::find_updates(old_grades, &res_generalized);
                    if !updates.is_empty() {
                        let templated_grades: Vec<TemplateGrade> =
                            res_generalized.iter().map(|g| g.into()).collect();
                        let template_ctx = GradesTemplate {
                            grades: templated_grades.as_ref(),
                        };
                        let template_rendered = template_ctx.render().unwrap();
                        let global_app_state =
                            state.global_app_state.as_ref().unwrap().lock().await;
                        let email_result = global_app_state
                            .comm
                            .send_message(&Message {
                                subject: "New grades available".to_string(),
                                body: template_rendered,
                                mime: "text/html",
                                ..Default::default()
                            })
                            .await;
                        match email_result {
                            Ok(_) => {
                                info!("Sent email notification for new grades");
                            }
                            Err(e) => {
                                log::error!("Error sending email notification: {}", e);
                            }
                        }
                    }
                }
                state.grade_cache.last_updated = chrono::Local::now();
                state.grade_cache.response = Some(res_generalized);
            }
            Err(e) => {
                log::error!("Error querying Canvas LMS: {}", e);
            }
        }
    }
}

#[async_trait]
impl App for CanvasLMSApp {
    async fn initialize(self: Arc<Self>, config: &'static Config, app_state: Arc<Mutex<AppState>>) {
        let self_clone = self.clone();

        let refresh_interval = config.canvas_lms.refresh_interval;
        if refresh_interval == 0 {
            panic!("Canvas LMS refresh interval cannot be 0");
        }

        let mut state = self.state.lock().await;
        state.global_app_state = Some(app_state);
        state.config = Some(config);
        state.grade_cache = GradeCache {
            last_updated: chrono::Local::now(),
            response: None,
        };

        tokio::spawn(async move {
            let mut ticker =
                tokio::time::interval(std::time::Duration::from_secs(refresh_interval));
            ticker.set_missed_tick_behavior(tokio::time::MissedTickBehavior::Delay);
            loop {
                self_clone.refresh_grades().await;
                ticker.tick().await;
            }
        });
    }

    fn api_routes(self: Arc<Self>) -> Router {
        Router::new()
            .route("/canvas_lms/grades", get(route_get_grades))
            .layer(Extension(self.clone()))
    }
}
