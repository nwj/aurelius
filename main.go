package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/julienschmidt/httprouter"
	"html/template"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"time"
)

func main() {
	sectionData := initSectionData()
	router := httprouter.New()
	router.ServeFiles("/assets/*filepath", http.Dir("assets"))
	router.GET("/", IndexHandler)
	router.GET("/rand", ContextDecorator(RandomHandler, sectionData))

	log.Fatal(http.ListenAndServe(":8080", router))
}

type section struct {
	Book    int    `json:"book"`
	Section int    `json:"section"`
	Text    string `json:"text"`
	Length  int    `json:"length"`
}

func initSectionData() []section {
	var sectionData []section

	sectionDataJson, err := ioutil.ReadFile("data/meditations.json")
	if err != nil {
		panic(err)
	}

	err = json.Unmarshal(sectionDataJson, &sectionData)
	if err != nil {
		panic(err)
	}

	return sectionData
}

func ContextDecorator(h httprouter.Handle, sectionData []section) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		ctx := context.WithValue(r.Context(), "sectionData", sectionData)
		h(w, r.WithContext(ctx), ps)
	}
}

func IndexHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	t, err := template.ParseFiles("views/index.html")
	if err != nil {
		panic(err)
	}
	t.Execute(w, nil)
}

func RandomHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	sectionData := r.Context().Value("sectionData").([]section)

	seed := rand.NewSource(time.Now().UnixNano())
	rng := rand.New(seed)
	random := rng.Intn(len(sectionData))
	jsonSection, err := json.MarshalIndent(sectionData[random], "", "    ")
	if err != nil {
		panic(err)
	}
	fmt.Fprint(w, string(jsonSection))
}
