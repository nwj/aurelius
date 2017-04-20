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
	"strconv"
)

func main() {
	sectionData := initSectionData()
	router := httprouter.New()
	router.ServeFiles("/assets/*filepath", http.Dir("assets"))
	router.GET("/", ContextDecorator(IndexHandler, sectionData))
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

func getRandomSection(sectionData []section, maxLength int) section {
	if maxLength >= 0 {
		var filteredSections []section
		for _, s := range(sectionData) {
			if s.Length <= maxLength {
				filteredSections = append(filteredSections, s)
			}
		}
		sectionData = filteredSections
	}

	seed := rand.NewSource(time.Now().UnixNano())
	rng := rand.New(seed)
	random := rng.Intn(len(sectionData))
	return sectionData[random]
}

func ContextDecorator(h httprouter.Handle, sectionData []section) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		ctx := context.WithValue(r.Context(), "sectionData", sectionData)
		h(w, r.WithContext(ctx), ps)
	}
}

func IndexHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	sectionData := r.Context().Value("sectionData").([]section)
	section := getRandomSection(sectionData, 350)
	t, err := template.ParseFiles("views/index.html")
	if err != nil {
		panic(err)
	}
	t.Execute(w, &section)
}

func RandomHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	maxLength := -1
	if len(r.URL.Query()["maxLength"]) == 1 {
		maxLengthQueryParam, err := strconv.Atoi(r.URL.Query()["maxLength"][0])
		if err != nil {
			panic(err)
		}
		maxLength = maxLengthQueryParam
	}

	sectionData := r.Context().Value("sectionData").([]section)
	jsonSection, err := json.MarshalIndent(getRandomSection(sectionData, maxLength), "", "    ")
	if err != nil {
		panic(err)
	}
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprint(w, string(jsonSection))
}
