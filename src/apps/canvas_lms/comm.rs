use askama::Template;
use chrono::{Local, TimeZone};

use super::grading::Grading;

#[derive(Template)]
#[template(path = "grades.html")]
pub struct GradesTemplate<'a> {
    pub grades: &'a [TemplateGrade],
}

pub struct TemplateGrade {
    pub submission_username: String,
    pub course_code: String,
    pub assignment_url: String,
    pub name: String,
    pub due: String,
    pub grade: String,
    pub graded_at: String,
    pub posted_at: String,
}

impl From<&Grading> for TemplateGrade {
    fn from(grading: &Grading) -> Self {
        Self {
            submission_username: grading.submission_username.clone(),
            course_code: grading.course_code.clone(),
            assignment_url: grading.assignment_url.clone(),
            name: grading.name.clone(),
            due: grading
                .due_at
                .map(|d| {
                    Local
                        .from_utc_datetime(&d.naive_utc())
                        .format("%Y-%m-%d %H:%M")
                        .to_string()
                })
                .unwrap_or_else(|| "N/A".to_string()),
            grade: if grading.grade_hidden {
                "Hidden".to_string()
            } else if grading.score.is_none() {
                "Not Graded".to_string()
            } else {
                format!(
                    "{:.2} ({})/ {:.2}",
                    grading.score.unwrap(),
                    grading.grade,
                    grading.possible_points
                )
            },
            graded_at: grading
                .graded_at
                .map(|d| {
                    Local
                        .from_utc_datetime(&d.naive_utc())
                        .format("%Y-%m-%d %H:%M")
                        .to_string()
                })
                .unwrap_or_else(|| "N/A".to_string()),
            posted_at: grading
                .posted_at
                .map(|d| {
                    Local
                        .from_utc_datetime(&d.naive_utc())
                        .format("%Y-%m-%d %H:%M")
                        .to_string()
                })
                .unwrap_or_else(|| "N/A".to_string()),
        }
    }
}
