use clap::Parser;
use simple_logger::SimpleLogger;
use yoake::config;

#[derive(Debug, Parser)]
#[clap(name = "yoake_server", version, author, about)]
struct Cli {
    #[arg(short, long, default_value = "config.yaml")]
    config: String,
    #[arg(short, long)]
    dev: bool,
}

#[tokio::main]
async fn main() {
    let args = Cli::parse();

    SimpleLogger::new()
        .with_module_level(
            "yoake",
            if args.dev {
                log::LevelFilter::Debug
            } else {
                log::LevelFilter::Info
            },
        )
        .with_level(if args.dev {
            log::LevelFilter::Info
        } else {
            log::LevelFilter::Warn
        })
        .init()
        .unwrap();

    let config = config::Config::load_yaml_file(args.config);
    unsafe {
        config::set_config(config);
    }

    yoake::main_server(args.dev).await;
}
