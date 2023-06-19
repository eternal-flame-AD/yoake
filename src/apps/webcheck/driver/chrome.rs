use rand_core::RngCore;
use std::process::{Child, Command};
use thirtyfour::{prelude::WebDriverResult, ChromeCapabilities, WebDriver};
pub struct ChromeDriver {
    child: Option<Child>,
    port: u16,
    tmp_dir: Option<tempfile::TempDir>,
}

impl ChromeDriver {
    pub fn new() -> Self {
        let port = rand_core::OsRng.next_u32() as u16 + 10000;
        Self::new_port(port)
    }
    pub fn new_port(port: u16) -> Self {
        Self {
            child: None,
            port,
            tmp_dir: None,
        }
    }

    pub fn spawn(&mut self, args: &[&str]) -> anyhow::Result<()> {
        let tmp_dir = tempfile::tempdir()?;
        std::fs::create_dir(tmp_dir.path().join("data"))?;
        std::fs::create_dir(tmp_dir.path().join("crashpad"))?;
        self.tmp_dir = Some(tmp_dir);

        let mut cmd = Command::new("chromedriver");
        cmd.arg(format!("--port={}", self.port));
        cmd.args(args);

        let child = cmd.spawn()?;

        self.child = Some(child);

        Ok(())
    }

    pub async fn connect(&mut self, mut caps: ChromeCapabilities) -> WebDriverResult<WebDriver> {
        let temp_dir = self.tmp_dir.as_ref().unwrap();

        caps.add_chrome_arg(
            format!("--user-data-dir={}", temp_dir.path().join("data").display()).as_str(),
        )?;
        caps.add_chrome_arg(
            format!(
                "--crash-dumps-dir={}",
                temp_dir.path().join("crashpad").display()
            )
            .as_str(),
        )?;

        let addr = format!("http://localhost:{}/", self.port);
        WebDriver::new(addr.as_str(), caps).await
    }
}

impl Drop for ChromeDriver {
    fn drop(&mut self) {
        if let Some(child) = &mut self.child {
            if let Err(e) = child.kill() {
                log::error!("Error killing chrome driver: {}", e);
            }
            if let Err(e) = child.wait() {
                log::error!("Error waiting for chrome driver to exit: {}", e);
            }
        }
    }
}
