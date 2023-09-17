use chrono::{DateTime, Utc};
use serde::Serialize;

#[derive(Debug, Clone, Serialize)]
pub struct Grading {
    pub name: String,
    pub course_name: String,
    pub course_code: String,

    pub submission_username: String,

    pub assignment_id: String,
    pub assignment_legacy_id: String,
    pub assignment_url: String,
    pub submission_id: String,
    pub submission_legacy_id: String,
    pub course_id: String,
    pub course_legacy_id: String,

    pub due_at: Option<DateTime<Utc>>,
    pub state: String,
    pub score: Option<f64>,
    pub entered_score: Option<f64>,
    pub possible_points: f64,
    pub grade_hidden: bool,
    pub grade: String,
    pub entered_grade: String,

    pub graded_at: Option<DateTime<Utc>>,
    pub posted_at: Option<DateTime<Utc>>,
}

impl Grading {
    pub fn last_updated(&self) -> Option<DateTime<Utc>> {
        [self.graded_at, self.posted_at]
            .iter()
            .filter_map(|d| *d)
            .max()
    }
}

impl Grading {
    pub fn find_updates<'a>(before: &[Grading], after: &'a [Grading]) -> Vec<&'a Grading> {
        let mut updates = Vec::new();
        for new in after {
            if let Some(old) = before
                .iter()
                .find(|old| old.submission_id == new.submission_id)
            {
                if old.last_updated() < new.last_updated() {
                    updates.push(new);
                }
            } else if new.last_updated().is_some() {
                updates.push(new);
            }
        }
        updates
    }
}

impl PartialEq for Grading {
    fn eq(&self, other: &Self) -> bool {
        self.submission_id == other.submission_id && self.last_updated() == other.last_updated()
    }
}

impl PartialOrd for Grading {
    fn partial_cmp(&self, other: &Self) -> Option<std::cmp::Ordering> {
        if self == other {
            Some(std::cmp::Ordering::Equal)
        } else {
            self.last_updated().partial_cmp(&other.last_updated())
        }
    }
}

impl Eq for Grading {
    fn assert_receiver_is_total_eq(&self) {}
}

impl Ord for Grading {
    fn cmp(&self, other: &Self) -> std::cmp::Ordering {
        self.partial_cmp(other).unwrap()
    }
}

pub struct CanvasGradingData(pub super::graph::AllCourseData);

impl IntoIterator for CanvasGradingData {
    type Item = Grading;
    type IntoIter = std::vec::IntoIter<Self::Item>;

    fn into_iter(self) -> Self::IntoIter {
        let mut vec = Vec::new();
        for course in self.0.all_courses {
            for submission in course.submissions_connection.nodes {
                vec.push(Grading {
                    name: submission.assignment.name,
                    course_name: course.name.clone(),
                    course_code: course.course_code.clone(),
                    due_at: submission.assignment.due_at.map(|d| d.into()),

                    submission_username: submission.user.name.clone(),

                    assignment_id: submission.assignment.id.clone(),
                    assignment_legacy_id: submission.assignment.id_legacy.clone(),
                    assignment_url: submission.assignment.html_url.clone(),
                    submission_id: submission.id.clone(),
                    submission_legacy_id: submission.id_legacy.clone(),
                    course_id: course.id.clone(),
                    course_legacy_id: course.id_legacy.clone(),

                    state: course.state.clone(),
                    score: submission.score,
                    entered_score: submission.entered_score,
                    possible_points: submission.assignment.points_possible,
                    grade_hidden: submission.grade_hidden,
                    grade: submission.grade.unwrap_or_else(|| "".to_string()),
                    entered_grade: submission.entered_grade.unwrap_or_else(|| "".to_string()),
                    graded_at: submission.graded_at.map(|d| d.into()),
                    posted_at: submission.posted_at.map(|d| d.into()),
                });
            }
        }
        vec.into_iter()
    }
}
