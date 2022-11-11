package canvaslms

import (
	"time"
)

const GraphQuery = `query gradeQuery {
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
}`

type GraphResponse struct {
	Data struct {
		AllCourses []struct {
			IDLegacy              string `json:"_id"`
			ID                    string `json:"id"`
			Name                  string `json:"name"`
			State                 string `json:"state"`
			CourseCode            string `json:"courseCode"`
			SubmissionsConnection struct {
				Nodes []GraphSubmissionResponse `json:"nodes"`
			} `json:"submissionsConnection"`
		} `json:"allCourses"`
	} `json:"data"`
}
type GraphSubmissionResponse struct {
	IDLegacy   string `json:"_id"`
	ID         string `json:"id"`
	Assignment struct {
		IDLegacy       string  `json:"_id"`
		ID             string  `json:"id"`
		Name           string  `json:"name"`
		DueAt          *string `json:"dueAt"`
		GradingType    string  `json:"gradingType"`
		PointsPossible float64 `json:"pointsPossible"`
		HTMLUrl        string  `json:"htmlUrl"`
	} `json:"assignment"`
	Score         *float64 `json:"score"`
	EnteredScore  *float64 `json:"enteredScore"`
	Grade         *string  `json:"grade"`
	EnteredGrade  *string  `json:"enteredGrade"`
	GradingStatus string   `json:"gradingStatus"`
	GradeHidden   bool     `json:"gradeHidden"`
	GradedAt      *string  `json:"gradedAt"`
	Posted        bool     `json:"posted"`
	PostedAt      *string  `json:"postedAt"`
	State         string   `json:"state"`
	User          struct {
		IDLegacy string  `json:"_id"`
		ID       string  `json:"id"`
		SISID    *string `json:"sisId"`
		Name     string  `json:"name"`
		Email    *string `json:"email"`
	}
}

type GraphSubmissionCompareFunc func(m1, m2 GraphSubmissionResponse) (m1HasPriority bool)

func parseJSONTime(s string) time.Time {
	t, _ := time.Parse(time.RFC3339, s)
	return t
}

func GraphSubmissionCompareByDue(m1, m2 GraphSubmissionResponse) (m1HasPriority bool) {
	if m1.Assignment.DueAt == nil {
		return false
	}
	if m2.Assignment.DueAt == nil {
		return true
	}

	m1Time, m2Time := parseJSONTime(*m1.Assignment.DueAt), parseJSONTime(*m2.Assignment.DueAt)
	now := time.Now()
	m1IsPast, m2IsPast := now.After(m1Time), now.After(m2Time)
	if m1IsPast && m2IsPast {
		return m1Time.After(m2Time)
	}
	if !m1IsPast && !m2IsPast {
		return m1Time.Before(m2Time)
	}
	return !m1IsPast
}

func laterTime(t1, t2 *string) *string {
	if t1 == nil {
		return t2
	}
	if t2 == nil {
		return t1
	}
	t1T, t2T := parseJSONTime(*t1), parseJSONTime(*t2)

	if t1T.After(t2T) {
		return t1
	}
	return t2
}

func GraphSubmissionCompareByGradeTime(m1, m2 GraphSubmissionResponse) (m1HasPriority bool) {
	m1LastUpdate := laterTime(m1.PostedAt, m1.GradedAt)
	m2LastUpdate := laterTime(m2.PostedAt, m2.GradedAt)
	if m2LastUpdate == nil {
		return true
	}
	if m1LastUpdate == nil {
		return false
	}
	return parseJSONTime(*m1LastUpdate).After(parseJSONTime(*m2LastUpdate))
}
