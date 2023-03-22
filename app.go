package main

import (
	"encoding/json"
	"fmt"
	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
	"log"
	"net/http"
)

type App struct {
	Router     *mux.Router
	Middleware *Middleware
	Config     *Conf
	Storage
}

type ShortenReq struct {
	URL                 string `json:"url" validate:"required"`
	ExpirationInMinutes int64  `json:"expiration_in_minutes" validate:"min=0"`
}

type shortLinkResp struct {
	ShortLink string `json:"short_link"`
}

// Initialize app
func (a *App) Initialize() {
	// set log formatter
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	a.Router = mux.NewRouter()
	a.Middleware = &Middleware{}
	a.Config = InitConfig()
	a.Storage = NewRedisClient(a.Config.Redis)
	a.initializeRoutes()
}

func (a *App) initializeRoutes() {
	a.Router.Use(a.Middleware.LoggingHandler, a.Middleware.RecoverHandler)

	a.Router.HandleFunc("/api/shorten", a.createShortLink).Methods("POST")
	a.Router.HandleFunc("/api/info", a.getShortLinkInfo).Methods("GET")
	a.Router.HandleFunc("/{shortlink:[a-zA-Z0-9]{1,11}}", a.redirect).Methods("GET")
}

func (a *App) createShortLink(w http.ResponseWriter, r *http.Request) {
	var req ShortenReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		responseWithError(w, StatusError{
			Code: http.StatusBadRequest,
			Err:  fmt.Errorf("parse parameters failed: %v", req),
		})
		return
	}
	// validate request params
	validate := validator.New()
	if err := validate.Struct(&req); err != nil {
		responseWithError(w, StatusError{
			Code: http.StatusBadRequest,
			Err:  fmt.Errorf("validate parameters failed: %v", req),
		})
		return
	}
	defer r.Body.Close()

	encodeId, err := a.Storage.Shorten(req.URL, req.ExpirationInMinutes)
	if err != nil {
		responseWithError(w, err)
	} else {
		responseWithJson(w, http.StatusOK, shortLinkResp{ShortLink: encodeId})
	}
}

func (a *App) getShortLinkInfo(w http.ResponseWriter, r *http.Request) {
	vals := r.URL.Query()
	encodeId := vals.Get("shortlink")

	info, err := a.Storage.ShortLinkInfo(encodeId)
	if err != nil {
		responseWithError(w, err)
	} else {
		responseWithJson(w, http.StatusOK, info)
	}
}

// redirect to original url
func (a *App) redirect(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	// 返回的是字典类型
	encodeId := vars["shortlink"]
	url, err := a.Storage.UnShorten(encodeId)
	if err != nil {
		responseWithError(w, err)
	} else {
		http.Redirect(w, r, url, http.StatusTemporaryRedirect)
	}

}

// Run starts to listen on server
func (a *App) Run(addr string) {
	log.Fatal(http.ListenAndServe(addr, a.Router))
}
