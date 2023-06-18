use std::io::Write;

use axum::{
    body::HttpBody, extract::Path, http::Request, response::IntoResponse, response::Response,
    routing::get, Router,
};
use flate2::write::GzEncoder;
use hyper::{header, StatusCode};
use rust_embed::RustEmbed;

#[derive(RustEmbed)]
#[folder = "ui/dist"]
struct UiAssets;

pub async fn serve_file(path: &str, accepted_encodings: Vec<&str>) -> Response {
    let ext = path.split('.').last().unwrap_or("");
    let content_type = match ext {
        "html" => "text/html",
        "css" => "text/css",
        "js" => "text/javascript",
        "png" => "image/png",
        "jpg" => "image/jpeg",
        "jpeg" => "image/jpeg",
        "gif" => "image/gif",
        "svg" => "image/svg+xml",
        _ => "text/plain",
    };

    if let Some(asset) = UiAssets::get(path) {
        if accepted_encodings.contains(&"gzip") {
            let mut encoder = GzEncoder::new(Vec::new(), flate2::Compression::default());
            encoder.write_all(asset.data.as_ref()).unwrap();
            let data = encoder.finish().unwrap();
            return (
                StatusCode::OK,
                [
                    (header::CONTENT_TYPE, content_type),
                    (header::CONTENT_ENCODING, "gzip"),
                ],
                data,
            )
                .into_response();
        } else {
            (
                StatusCode::OK,
                [(header::CONTENT_TYPE, content_type)],
                asset.data.to_vec(),
            )
                .into_response()
        }
    } else {
        (
            StatusCode::NOT_FOUND,
            [(header::CONTENT_TYPE, "text/plain")],
            format!("File not found: {}", path),
        )
            .into_response()
    }
}

pub async fn ui_path_handler<B: HttpBody>(Path(path): Path<String>, req: Request<B>) -> Response {
    let mut path = if path.ends_with("/") {
        path + "index.html"
    } else {
        path
    };
    path = path.trim_start_matches('/').to_owned();

    let accepted_encodings = req
        .headers()
        .get(header::ACCEPT_ENCODING)
        .and_then(|v| v.to_str().ok())
        .unwrap_or("")
        .split(", ");

    serve_file(&path, accepted_encodings.collect()).await
}

pub async fn ui_index_handler<B: HttpBody>(req: Request<B>) -> Response {
    let accepted_encodings = req
        .headers()
        .get(header::ACCEPT_ENCODING)
        .and_then(|v| v.to_str().ok())
        .unwrap_or("")
        .split(", ");

    serve_file("index.html", accepted_encodings.collect()).await
}

pub async fn redirect_to_ui() -> Response {
    (
        StatusCode::TEMPORARY_REDIRECT,
        [(header::LOCATION, "/ui/")],
        "",
    )
        .into_response()
}

pub fn ui_router(_dev: bool) -> Router {
    Router::new()
        .route("/*path", get(ui_path_handler))
        .fallback(get(ui_index_handler))
}
