use std::sync::{Arc, Mutex};

use apps::{med, webcheck};
use axum::{http::Request, middleware::Next, response::Response, routing::get, Extension, Router};
use axum_server::tls_rustls::RustlsConfig;
use diesel::{sqlite, Connection};
use diesel_migrations::{embed_migrations, EmbeddedMigrations, MigrationHarness};
use hyper::Method;
use log::info;
use tower_http::cors::{self, CorsLayer};

use crate::{
    apps::{auth, canvas_lms, server_info, App},
    comm::GlobalCommunicator,
};

pub mod apps;
pub mod config;

pub mod session;

pub mod models;
pub mod schema;

pub mod http;

pub mod comm;

pub mod ui;

async fn log_middleware<B>(req: Request<B>, next: Next<B>) -> Response {
    info!("Request: {} {}", req.method(), req.uri());
    let res = next.run(req).await;
    info!("Response: {:?}", res);
    res
}

const MIGRATIONS: EmbeddedMigrations = embed_migrations!("migrations");

pub struct AppState {
    pub db: sqlite::SqliteConnection,
    pub comm: comm::GlobalCommunicator,
}

pub fn establish_db_connection() -> sqlite::SqliteConnection {
    let db_url = &config::get_config().db.url;

    let mut conn = sqlite::SqliteConnection::establish(db_url)
        .unwrap_or_else(|_| panic!("Error connecting to {}", db_url));

    conn.run_pending_migrations(MIGRATIONS).unwrap();

    conn
}

pub async fn server_listen(router: Router) {
    let config = config::get_config();

    let listen_addr = config
        .listen
        .addr
        .parse()
        .expect("Failed to parse listen address");
    if config.listen.cert.is_some() && config.listen.key.is_some() {
        let tls_config = RustlsConfig::from_pem_file(
            config.listen.cert.as_ref().unwrap(),
            config.listen.key.as_ref().unwrap(),
        )
        .await
        .expect("Failed to load TLS certificate and key");
        info!("Listening on https://{}", config.listen.addr);
        axum_server::bind_rustls(listen_addr, tls_config)
            .serve(router.into_make_service())
            .await
            .unwrap();
    } else {
        info!("Listening on http://{}", config.listen.addr);
        axum_server::bind(listen_addr)
            .serve(router.into_make_service())
            .await
            .unwrap();
    }
}

pub async fn main_server(dev: bool) {
    let config = config::get_config();

    let apps: &mut [Arc<dyn App>] = &mut [
        Arc::new(server_info::ServerInfoApp::new()),
        Arc::new(auth::AuthApp::new()),
        Arc::new(canvas_lms::CanvasLMSApp::new()),
        Arc::new(med::MedManagementApp::new()),
        Arc::new(webcheck::WebcheckApp::new()),
    ];

    let mut comm = GlobalCommunicator::new();
    comm.add_communicator(Arc::new(comm::gotify::GotifyCommunicator::new(config)));
    comm.add_communicator(Arc::new(comm::email::EmailCommunicator::new(config)));
    let app_state = Arc::new(Mutex::new(AppState {
        db: establish_db_connection(),
        comm: comm,
    }));

    for app in &mut *apps {
        app.clone().initialize(config, app_state.clone()).await;
    }

    let mut api_router = axum::Router::new();
    for app in apps {
        api_router = api_router.merge(app.clone().api_routes());
    }

    let router = axum::Router::new()
        .nest("/api", api_router)
        .route("/", get(ui::redirect_to_ui))
        .route("/ui/", get(ui::ui_index_handler))
        .route("/ui/*path", get(ui::ui_path_handler))
        .layer(axum::middleware::from_fn(session::middleware))
        .layer(Extension(config))
        .layer(Extension(app_state))
        .layer(axum::middleware::from_fn(log_middleware));

    let router = if dev {
        router.layer(
            CorsLayer::new()
                .allow_methods([Method::GET, Method::POST])
                .allow_origin(cors::Any),
        )
    } else {
        router
    };

    server_listen(router).await;
}
