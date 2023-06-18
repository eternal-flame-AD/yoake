use axum::{
    async_trait,
    body::{Bytes, HttpBody},
    extract::FromRequest,
    http::Request,
    response::{IntoResponse, Response},
    BoxError, Json,
};
use hyper::StatusCode;
use serde::{de::DeserializeOwned, Deserialize, Serialize};

pub struct JsonApiForm<T>(pub T);

#[async_trait]
impl<T, S, B> FromRequest<S, B> for JsonApiForm<T>
where
    T: DeserializeOwned,
    S: Send + Sync,
    B: HttpBody + Send + 'static,
    B::Data: Send,
    B::Error: Into<BoxError>,
{
    type Rejection = ApiResponse<()>;

    async fn from_request(req: Request<B>, state: &S) -> Result<Self, Self::Rejection> {
        if req.headers().get("content-type") != Some(&"application/json".parse().unwrap()) {
            return Err(ApiResponse::error(
                "invalid content-type".to_string(),
                400,
                None,
            ));
        }
        let bytes = Bytes::from_request(req, state)
            .await
            .map_err(|_| ApiResponse::error("failed reading request".to_string(), 400, None))?;

        let des = &mut serde_json::Deserializer::from_slice(&bytes);

        Ok(JsonApiForm(T::deserialize(des).map_err(|e| {
            ApiResponse::error(format!("failed parsing json: {}", e), 400, None)
        })?))
    }
}

#[derive(Clone, Debug, Serialize, Deserialize)]
pub struct ApiResponse<T> {
    pub code: u16,
    pub status: ApiStatus,
    pub message: String,
    pub data: Option<T>,
}

pub type ApiResult<T> = Result<ApiResponse<T>, ApiResponse<()>>;

#[derive(Clone, Debug, Serialize, Deserialize)]
pub enum ApiStatus {
    Ok,
    Error,
}

impl<T> ApiResponse<T> {
    pub const fn ok(message: String, data: Option<T>) -> Self {
        Self {
            code: 200,
            status: ApiStatus::Ok,
            message,
            data,
        }
    }

    pub const fn error(message: String, code: u16, data: Option<T>) -> Self {
        Self {
            code,
            status: ApiStatus::Error,
            message,
            data,
        }
    }

    pub const fn unauthorized(message: String, data: Option<T>) -> Self {
        Self {
            code: 401,
            status: ApiStatus::Error,
            message,
            data,
        }
    }

    pub const fn bad_request(message: String, data: Option<T>) -> Self {
        Self {
            code: 400,
            status: ApiStatus::Error,
            message,
            data,
        }
    }
}

impl<T> IntoResponse for ApiResponse<T>
where
    T: Serialize,
{
    fn into_response(self) -> Response {
        let resp = (StatusCode::from_u16(self.code).unwrap(), Json(self)).into_response();

        resp
    }
}
