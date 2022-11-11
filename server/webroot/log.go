package webroot

import (
	"fmt"
	"log"
	"path"
	"regexp"
	"strconv"

	"github.com/eternal-flame-AD/yoake/internal/auth"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type logEntry struct {
	middleware.RequestLoggerValues
	Categories []string
	CleanPath  string
	Auth       auth.RequestAuth
}

func processLoggerValues(c echo.Context, values middleware.RequestLoggerValues) logEntry {
	status := values.Status
	statusString := []byte(strconv.Itoa(status))
	for i := len(statusString) - 1; i >= 0; i-- {
		logSetRequestCategory(c, fmt.Sprintf("status_%s", statusString))
		statusString[i] = 'x'
	}
	return logEntry{
		RequestLoggerValues: values,
		Categories:          logGetCategories(c),
		CleanPath:           path.Clean(c.Request().URL.Path),
		Auth:                auth.GetRequestAuth(c),
	}
}

type logCompiledFilter struct {
	Negate  bool
	Pattern *regexp.Regexp
}

func logGetCategories(c echo.Context) []string {
	if existingCates, err := c.Get("log_request_categories").([]string); err {
		return existingCates
	} else {
		return []string{}
	}
}

func logCompileFilters(filters []string) []logCompiledFilter {
	var compiledFilters []logCompiledFilter
	for _, filter := range filters {
		negate := false
		if filter[0] == '!' {
			negate = true
			filter = filter[1:]
		}
		log.Printf("Compiling filter: %s negate=%v", filter, negate)
		compiledFilters = append(compiledFilters, logCompiledFilter{negate, regexp.MustCompile(filter)})
	}
	return compiledFilters
}

func logFilterCategories(c echo.Context, filters []logCompiledFilter) bool {
	if filters == nil {
		return true
	}
	for _, category := range logGetCategories(c) {
		for _, filter := range filters {
			matches := filter.Pattern.MatchString(category)
			negate := filter.Negate
			//	log.Printf("Checking category %s against filter %s negate=%v matches=%v", category, filter.Pattern, negate, matches)
			if matches {
				return !negate
			}
		}
	}
	return true
}

func logSetRequestCategory(c echo.Context, category string) {
	if existingCates, ok := c.Get("log_request_categories").([]string); !ok {
		c.Set("log_request_categories", []string{category})
	} else {
		c.Set("log_request_categories", append(existingCates, category))
	}
}

func logRemoveRequestCategory(c echo.Context, category string) {
	if existingCates, ok := c.Get("log_request_categories").([]string); ok {
		for i, existingCate := range existingCates {
			if existingCate == category {
				c.Set("log_request_categories", append(existingCates[:i], existingCates[i+1:]...))
				return
			}
		}
	}
}

func logMiddleware(category string, backend echo.MiddlewareFunc) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			logSetRequestCategory(c, category)
			if backend != nil {
				wrappedNext := func(c echo.Context) error {
					logRemoveRequestCategory(c, category)
					return next(c)
				}
				return backend(wrappedNext)(c)
			}
			return next(c)
		}
	}
}

var (
	loggerConfig = middleware.RequestLoggerConfig{
		LogLatency:       true,
		LogProtocol:      true,
		LogRemoteIP:      true,
		LogHost:          true,
		LogMethod:        true,
		LogURI:           true,
		LogURIPath:       true,
		LogRoutePath:     true,
		LogRequestID:     true,
		LogReferer:       true,
		LogUserAgent:     true,
		LogStatus:        true,
		LogError:         true,
		LogContentLength: true,
		LogResponseSize:  true,
		LogHeaders:       []string{"Content-Type"},
		LogQueryParams:   []string{},
		LogFormValues:    []string{"From", "To"},
	}
)
