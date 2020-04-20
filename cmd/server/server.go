package main

import (
	"database/sql"
	"encoding/json"
	"flag"
	"log"

	"github.com/asaskevich/govalidator"
	"github.com/fasthttp/router"
	lru "github.com/hashicorp/golang-lru"
	"github.com/jmoiron/sqlx"
	"github.com/valyala/fasthttp"

	"github.com/dnfd/url_shortener/internal/urlconverter"

	_ "github.com/go-sql-driver/mysql"
)

var (
	port = flag.String("port", "8080", "Port to bind")

	db    *sqlx.DB
	cache *lru.Cache
)

func route(defaultHandler fasthttp.RequestHandler, r *router.Router) fasthttp.RequestHandler {
	return fasthttp.RequestHandler(func(ctx *fasthttp.RequestCtx) {
		handler, _ := r.Lookup(string(ctx.Method()), string(ctx.Path()), ctx)

		if handler == nil {
			defaultHandler(ctx)
		} else {
			handler(ctx)
		}
	})
}

func redirect(ctx *fasthttp.RequestCtx) {
	url, err := getURL(ctx.Path()[1:]) // skip '/'
	if err == sql.ErrNoRows {
		ctx.SetStatusCode(404)
		return
	}

	if err != nil {
		ctx.SetStatusCode(422)
		log.Println(err)
		return
	}

	ctx.Redirect(url, 302)
}

type addURLParams struct {
	URL string `json:"url"`
}

func addURL(ctx *fasthttp.RequestCtx) {
	params := addURLParams{}
	err := json.Unmarshal(ctx.PostBody(), &params)
	if err != nil {
		ctx.SetStatusCode(422)
		ctx.WriteString(err.Error())
		return
	}

	url := params.URL
	if !govalidator.IsURL(url) {
		ctx.SetStatusCode(422)
		ctx.WriteString("URL is not valid.\n")
		return
	}

	id, err := persistURL(url)
	if err != nil {
		ctx.SetStatusCode(500)
		ctx.WriteString(err.Error())
		return
	}

	ctx.SetStatusCode(201)
	ctx.WriteString("localhost:8080/")
	ctx.Write(urlconverter.IDToURL(id))
	ctx.WriteString("\n")
}

func persistURL(url string) (int64, error) {
	var id int64

	_, err := db.Exec(`INSERT IGNORE INTO urls (url) VALUES (?)`, url)
	if err != nil {
		return id, err
	}

	row := db.QueryRow(`SELECT id from urls WHERE url = ?`, url)
	err = row.Scan(&id)

	return id, err
}

func getURL(shortURL []byte) (string, error) {
	cached, ok := cache.Get(string(shortURL))
	if ok {
		return cached.(string), nil
	}

	var url string
	id, err := urlconverter.URLToID(shortURL)
	if err != nil {
		return url, err
	}

	err = db.Get(&url, "SELECT url from urls WHERE id = ?", id)
	if err == nil {
		cache.Add(string(shortURL), url)
	}

	return url, err
}

func main() {
	var err error
	flag.Parse()

	db = sqlx.MustConnect("mysql", "root:@(localhost:3306)/url_shortener")
	cache, err = lru.New(128)
	if err != nil {
		log.Panic(err)
	}

	r := router.New()

	r.POST("/urls/new", addURL)

	err = fasthttp.ListenAndServe(":"+*port, route(redirect, r))
	log.Println(err)
}
