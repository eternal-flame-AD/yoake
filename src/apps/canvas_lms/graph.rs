use serde::{Deserialize, Serialize};

#[derive(Debug, Serialize)]
pub struct GraphQuery<T> {
    pub query: String,
    #[serde(rename = "operationName")]
    pub operation_name: String,
    pub variables: T,
}

#[derive(Debug, Clone, Deserialize, Serialize)]
pub struct GraphNodes<T> {
    pub nodes: Vec<T>,
}

#[derive(Debug, Clone, Deserialize)]
pub struct GraphResponse<T> {
    pub data: T,
}

pub const ALL_COURSES_QUERY: &str = "query gradeQuery {
    allCourses {
      _id
      id
      name
      state
      courseCode
      submissionsConnection(first: $maxn$, orderBy: {field: gradedAt, direction: descending}) {
        nodes {
          _id
          id
          assignment {
            _id
            id
            name
            dueAt
            gradingType
            pointsPossible
            htmlUrl
          }
          score
          enteredScore
          grade
          enteredGrade
          gradingStatus
          gradeHidden
          gradedAt
          posted
          postedAt
          state
          user {
              _id
              id
              name
              sisId
              email
          }
        }
      }
    }
  }";

#[derive(Debug, Clone, Deserialize, Serialize)]
pub struct AllCourseData {
    #[serde(rename = "allCourses")]
    pub all_courses: Vec<GraphCourse>,
}

#[derive(Debug, Clone, Deserialize, Serialize)]
pub struct GraphCourse {
    #[serde(rename = "_id")]
    pub id_legacy: String,
    pub id: String,
    pub name: String,
    pub state: String,
    #[serde(rename = "courseCode")]
    pub course_code: String,
    #[serde(rename = "submissionsConnection")]
    pub submissions_connection: GraphNodes<GraphSubmission>,
}

#[derive(Debug, Clone, Deserialize, Serialize)]
pub struct GraphSubmission {
    #[serde(rename = "_id")]
    pub id_legacy: String,
    pub id: String,
    pub assignment: GraphAssignment,
    pub score: Option<f64>,
    #[serde(rename = "enteredScore")]
    pub entered_score: Option<f64>,
    pub grade: Option<String>,
    #[serde(rename = "enteredGrade")]
    pub entered_grade: Option<String>,
    #[serde(rename = "gradingStatus")]
    pub grading_status: Option<String>,
    #[serde(rename = "gradeHidden")]
    pub grade_hidden: bool,
    #[serde(rename = "gradedAt")]
    pub graded_at: Option<chrono::DateTime<chrono::FixedOffset>>,
    pub posted: bool,
    #[serde(rename = "postedAt")]
    pub posted_at: Option<chrono::DateTime<chrono::FixedOffset>>,
    pub state: String,
    pub user: GraphUser,
}

#[derive(Debug, Clone, Deserialize, Serialize)]
pub struct GraphAssignment {
    #[serde(rename = "_id")]
    pub id_legacy: String,
    pub id: String,
    pub name: String,
    #[serde(rename = "dueAt")]
    pub due_at: Option<chrono::DateTime<chrono::FixedOffset>>,
    #[serde(rename = "gradingType")]
    pub grading_type: String,
    #[serde(rename = "pointsPossible")]
    pub points_possible: f64,
    #[serde(rename = "htmlUrl")]
    pub html_url: String,
}

#[derive(Debug, Clone, Deserialize, Serialize)]
pub struct GraphUser {
    #[serde(rename = "_id")]
    pub id_legacy: String,
    pub id: String,
    #[serde(rename = "sisId")]
    pub sis_id: Option<String>,
    pub name: String,
    pub email: Option<String>,
}
