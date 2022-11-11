package canvaslms

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/eternal-flame-AD/yoake/config"
	"github.com/eternal-flame-AD/yoake/internal/auth"
	"github.com/eternal-flame-AD/yoake/internal/comm"
	"github.com/eternal-flame-AD/yoake/internal/echoerror"
	"github.com/labstack/echo/v4"
)

type Handler struct {
	conf           config.CanvasLMS
	maxn           int
	respCache      *responseCache
	respCacheMutex sync.RWMutex
	refreshPeriod  time.Duration
	comm           *comm.CommProvider
}

type GetGradesResponse struct {
	LastRefresh string                    `json:"last_refresh"`
	Grades      []SubmissionScoreResponse `json:"grades"`
}

type SubmissionScoreResponse struct {
	Name               string
	Due                string
	AssignmentID       string
	AssignmentLegacyID string
	AssignmentURL      string
	SubmissionID       string
	SubmissionLegacyID string
	CourseID           string
	CourseLegacyID     string
	CourseName         string
	CourseCode         string

	Graded string
	Posted string

	State string

	Score          float64
	EnteredScore   float64
	PossiblePoints float64
	Grade          string
	EnteredGrade   string
	GradeHidden    bool
	GradedAt       string
	PostedAt       string

	SubmissionUserLegacyID string
	SubmissionUserID       string
	SubmissionUserName     string
	SubmissionUserSISID    string
	SubmissionUserEmail    string
}

func submissionScoreResponseFromQL(courselid, courseid, coursename, coursecode string, coursesubmission GraphSubmissionResponse) SubmissionScoreResponse {
	res := SubmissionScoreResponse{
		Name:                   coursesubmission.Assignment.Name,
		Due:                    "-",
		AssignmentID:           coursesubmission.Assignment.ID,
		AssignmentLegacyID:     coursesubmission.Assignment.IDLegacy,
		AssignmentURL:          coursesubmission.Assignment.HTMLUrl,
		SubmissionID:           coursesubmission.ID,
		SubmissionLegacyID:     coursesubmission.IDLegacy,
		CourseID:               courseid,
		CourseLegacyID:         courselid,
		CourseName:             coursename,
		CourseCode:             coursecode,
		Graded:                 coursesubmission.GradingStatus,
		Posted:                 strconv.FormatBool(coursesubmission.Posted),
		State:                  coursesubmission.State,
		Score:                  -1,
		EnteredScore:           -1,
		PossiblePoints:         coursesubmission.Assignment.PointsPossible,
		Grade:                  "-",
		EnteredGrade:           "-",
		GradeHidden:            coursesubmission.GradeHidden,
		GradedAt:               "-",
		PostedAt:               "-",
		SubmissionUserID:       coursesubmission.User.ID,
		SubmissionUserLegacyID: coursesubmission.User.IDLegacy,
		SubmissionUserName:     coursesubmission.User.Name,
	}
	if coursesubmission.Score != nil {
		res.Score = *coursesubmission.Score
	}
	if coursesubmission.EnteredScore != nil {
		res.EnteredScore = *coursesubmission.EnteredScore
	}
	if coursesubmission.Assignment.DueAt != nil {
		res.Due = *coursesubmission.Assignment.DueAt
	}
	if coursesubmission.Grade != nil {
		res.Grade = *coursesubmission.Grade
	}
	if coursesubmission.GradedAt != nil {
		res.GradedAt = *coursesubmission.GradedAt
	}
	if coursesubmission.PostedAt != nil {
		res.PostedAt = *coursesubmission.PostedAt
	}
	if coursesubmission.User.SISID != nil {
		res.SubmissionUserSISID = *coursesubmission.User.SISID
	}
	if coursesubmission.User.Email != nil {
		res.SubmissionUserEmail = *coursesubmission.User.Email
	}
	return res
}

func (h *Handler) sortResponse(compare GraphSubmissionCompareFunc) (resp []SubmissionScoreResponse, err error) {
	h.respCacheMutex.RLock()
	defer h.respCacheMutex.RUnlock()

	res := make([]GraphSubmissionResponse, h.maxn)
	resF := make([]SubmissionScoreResponse, h.maxn)
	curL := 0

	push := func(pos int, resp GraphSubmissionResponse, respF SubmissionScoreResponse) {
		for i := curL - 1; i > pos; i-- {
			res[i] = res[i-1]
			resF[i] = resF[i-1]
		}
		res[pos] = resp
		resF[pos] = respF
	}
	for _, course := range h.respCache.rawResponse.Data.AllCourses {
		for _, submission := range course.SubmissionsConnection.Nodes {
			pos := curL
			for i := curL - 1; i >= 0; i-- {
				if !compare(submission, res[i]) {
					break
				}
				pos = i
			}
			if pos < curL || curL < h.maxn {
				push(pos, submission, submissionScoreResponseFromQL(course.IDLegacy, course.ID, course.Name, course.CourseCode, submission))
				if curL < h.maxn {
					curL++
				}
			}
		}
	}
	return resF, nil
}

func (h *Handler) refresh() (hasUpdate bool, err error) {
	h.respCacheMutex.Lock()
	defer h.respCacheMutex.Unlock()

	client := http.Client{
		Timeout: h.refreshPeriod / 2,
	}
	buf := bytes.NewBufferString("")
	e := json.NewEncoder(buf)
	e.Encode(struct {
		Query         string    `json:"query"`
		OperationName string    `json:"operationName"`
		Variables     *struct{} `json:"variables"`
	}{strings.ReplaceAll(GraphQuery, "$maxn$", h.conf.MaxN), "gradeQuery", nil})
	now := time.Now()
	req, err := http.NewRequest("POST", h.conf.Endpoint, buf)
	if err != nil {
		return false, err
	}
	req.Header.Set("content-type", "application/json")
	req.Header.Set("Authorization", "Bearer "+h.conf.Token)
	resp, err := client.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {

		return false, fmt.Errorf("remote returned with %d", resp.StatusCode)
	}
	dec := json.NewDecoder(resp.Body)
	respStruct := new(GraphResponse)
	if err := dec.Decode(respStruct); err != nil {
		return false, err
	}

	hasUpdate = false
	lastUpdateTime := make(map[string]time.Time)
	for _, course := range respStruct.Data.AllCourses {
		for _, submission := range course.SubmissionsConnection.Nodes {
			lastUpdForSubmission := laterTime(submission.PostedAt, submission.GradedAt)
			if lastUpdForSubmission != nil {
				newUpdTime := parseJSONTime(*lastUpdForSubmission)
				lastUpdateTime[submission.ID] = newUpdTime

				if h.respCache != nil && (h.conf.SubmissionName == "" || submission.User.Name == h.conf.SubmissionName) {
					if lastUpdateTime, ok := h.respCache.submissionLastUpdate[submission.ID]; !ok || lastUpdateTime.UnixNano() != newUpdTime.UnixNano() {
						hasUpdate = true
					}
				}
			}
		}
	}
	h.respCache = &responseCache{
		rawResponse:          *respStruct,
		requestTime:          now,
		submissionLastUpdate: lastUpdateTime,
	}
	if hasUpdate {
		go h.SendGradeMessage(h.conf.Message.OnUpdate)
	}
	return hasUpdate, nil
}

func (h *Handler) GetInformation(key string) (data interface{}, err error) {
	switch key {
	case "recent-graded":
		return h.sortResponse(GraphSubmissionCompareByGradeTime)
	case "recent-due":
		return h.sortResponse(GraphSubmissionCompareByDue)
	}
	return nil, errors.New("unknown info request type")
}

type responseCache struct {
	rawResponse          GraphResponse
	requestTime          time.Time
	submissionLastUpdate map[string]time.Time
}

func Register(g *echo.Group, comm *comm.CommProvider) (h *Handler, err error) {
	h = &Handler{conf: config.Config().CanvasLMS, comm: comm}
	if h.conf.Token == "" {
		return nil, errors.New("canvas token not set")
	}
	maxn, err := strconv.Atoi(h.conf.MaxN)
	if err != nil {
		return nil, err
	}
	h.maxn = maxn
	refreshperiod, err := time.ParseDuration(h.conf.Frequency)
	if err != nil {
		return nil, err
	}
	h.refreshPeriod = refreshperiod

	checkForUpdates := make(chan bool)
	go func() {
		if _, err := h.refresh(); err != nil {
			log.Panicf("cannot access graphql endpoint: %v", err)
		} else {
			go h.SendGradeMessage(h.conf.Message.OnStartup)
		}
		for forced := range checkForUpdates {
			_ = forced
			h.refresh()
			// TODO: notify if there is an update
		}
	}()
	go func() {
		for range time.NewTicker(h.refreshPeriod).C {
			checkForUpdates <- false
		}
	}()

	gradesG := g.Group("/grades", auth.RequireMiddleware(auth.RoleAdmin), func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if h.respCache == nil {
				return echoerror.NewHttp(http.StatusServiceUnavailable, errors.New("not yet initialized"))
			}
			return next(c)
		}
	})
	{
		gradesG.GET("", func(c echo.Context) error {
			if c.QueryParam("refresh") == "1" {
				if _, err := h.refresh(); err != nil {
					return fmt.Errorf("cannot access graphql endpoint: %v", err)
				}
			}
			sortQuery := c.QueryParam("sort")
			if sortQuery == "" {
				sortQuery = "recent-graded"
			}
			var res GetGradesResponse
			res.LastRefresh = h.respCache.requestTime.Format(time.RFC3339)
			switch sortQuery {
			case "recent-graded":
				if grades, err := h.sortResponse(GraphSubmissionCompareByGradeTime); err != nil {
					return err
				} else {
					res.Grades = grades
				}
				return c.JSON(http.StatusOK, res)
			case "recent-due":
				if grades, err := h.sortResponse(GraphSubmissionCompareByDue); err != nil {
					return err
				} else {
					res.Grades = grades
				}
				return c.JSON(http.StatusOK, res)
			}
			return errors.New("unknown info request type")
		})
	}

	return h, nil
}
