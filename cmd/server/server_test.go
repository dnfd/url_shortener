package main

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/fasthttp/router"
	lru "github.com/hashicorp/golang-lru"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/valyala/fasthttp"
	"github.com/valyala/fasthttp/fasthttputil"
)

// serve serves http request using provided fasthttp handler
func serve(handler fasthttp.RequestHandler, req *http.Request) (*http.Response, error) {
	cache, _ = lru.New(2)

	ln := fasthttputil.NewInmemoryListener()
	defer ln.Close()

	go func() {
		err := fasthttp.Serve(ln, handler)
		if err != nil {
			panic(fmt.Errorf("failed to serve: %v", err))
		}
	}()

	client := http.Client{
		Transport: &http.Transport{
			DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				return ln.Dial()
			},
		},
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	return client.Do(req)
}

func mockDatabase(mocker func(sqlmock.Sqlmock)) {
	dbMock, mock, err := sqlmock.New()
	if err != nil {
		panic("An error '%s' was not expected when opening a stub database connection" + err.Error())
	}

	mocker(mock)

	db = sqlx.NewDb(dbMock, "sqlmock")
}

func concreteHandler(ctx *fasthttp.RequestCtx) {
	ctx.WriteString("concrete")
}

func defaultHandler(ctx *fasthttp.RequestCtx) {
	ctx.WriteString("default")
}

func TestServer_Route(t *testing.T) {
	r := router.New()
	r.GET("/concrete", concreteHandler)

	t.Run("for GET /concrete calls concrete handler", func(t *testing.T) {
		req, err := http.NewRequest("GET", "http://test/concrete", nil)
		assert.NoError(t, err)

		res, err := serve(route(defaultHandler, r), req)
		assert.NoError(t, err)

		body, err := ioutil.ReadAll(res.Body)
		assert.NoError(t, err)
		assert.Equal(t, "concrete", string(body))
	})

	t.Run("for GET /any calls default handler", func(t *testing.T) {
		req, err := http.NewRequest("GET", "http://test/any", nil)
		assert.NoError(t, err)

		res, err := serve(route(defaultHandler, r), req)
		assert.NoError(t, err)

		body, err := ioutil.ReadAll(res.Body)
		assert.NoError(t, err)
		assert.Equal(t, "default", string(body))
	})
}

func TestServer_AddURL(t *testing.T) {
	t.Run("error without post body", func(t *testing.T) {
		req, err := http.NewRequest("GET", "http://test/", nil)
		assert.NoError(t, err)

		res, err := serve(addURL, req)
		assert.NoError(t, err)

		body, err := ioutil.ReadAll(res.Body)
		assert.NoError(t, err)

		assert.Equal(t, 422, res.StatusCode)
		assert.Equal(t, "unexpected end of JSON input", string(body))
	})

	t.Run("error with incorrect URL", func(t *testing.T) {
		postBody := `{"url": "not an url"}`
		req, err := http.NewRequest("GET", "http://test/", strings.NewReader(postBody))
		assert.NoError(t, err)

		res, err := serve(addURL, req)
		assert.NoError(t, err)

		body, err := ioutil.ReadAll(res.Body)
		assert.NoError(t, err)

		assert.Equal(t, 422, res.StatusCode)
		assert.Equal(t, "URL is not valid.\n", string(body))
	})

	t.Run("creates newshort URL", func(t *testing.T) {
		mockDatabase(func(mock sqlmock.Sqlmock) {
			values := []driver.Value{"https://www.google.com"}
			mock.ExpectExec("INSERT IGNORE INTO urls").
				WithArgs(values...).
				WillReturnResult(sqlmock.NewResult(1, 1))

			mock.ExpectQuery("SELECT").
				WithArgs(values...).
				WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("1"))
		})

		postBody := `{"url": "https://www.google.com"}`
		req, err := http.NewRequest("POST", "http://test/", strings.NewReader(postBody))
		assert.NoError(t, err)

		res, err := serve(addURL, req)
		assert.NoError(t, err)

		body, err := ioutil.ReadAll(res.Body)
		assert.NoError(t, err)

		assert.Equal(t, 201, res.StatusCode)
		assert.Equal(t, "test/b\n", string(body))
	})
}

func TestServer_Redirect(t *testing.T) {
	t.Run("sets redirect when url found", func(t *testing.T) {
		mockDatabase(func(mock sqlmock.Sqlmock) {
			values := []driver.Value{1}
			mock.ExpectQuery("SELECT").
				WithArgs(values...).
				WillReturnRows(sqlmock.NewRows([]string{"url"}).AddRow("https://www.google.com"))
		})

		req, err := http.NewRequest("GET", "http://test/b", nil)
		assert.NoError(t, err)

		res, err := serve(redirect, req)
		assert.NoError(t, err)
		assert.Equal(t, 302, res.StatusCode)

		loc, err := res.Location()
		assert.NoError(t, err)
		assert.Equal(t, loc.String(), "https://www.google.com/")
	})

	t.Run("sets 404 when url not found", func(t *testing.T) {
		mockDatabase(func(mock sqlmock.Sqlmock) {
			values := []driver.Value{1}
			mock.ExpectQuery("SELECT").
				WithArgs(values...).
				WillReturnError(sql.ErrNoRows)
		})

		req, err := http.NewRequest("GET", "http://test/b", nil)
		assert.NoError(t, err)

		res, err := serve(redirect, req)
		assert.NoError(t, err)
		assert.Equal(t, 404, res.StatusCode)
	})
}
