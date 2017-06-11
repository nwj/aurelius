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
	"strconv"
	"time"
	"errors"
)

func main() {
	sectionData, err := initSectionData("data/meditations.json")
	if err != nil {
		panic(err)
	}
	router := httprouter.New()
	router.ServeFiles("/assets/*filepath", http.Dir("assets"))
	router.GET("/", ContextDecorator(IndexHandler, sectionData))
	router.GET("/faq", FaqHandler)
	router.GET("/meditations/:bookIndex/:sectionIndex", ContextDecorator(MeditationsHandler, sectionData))
	router.GET("/random", ContextDecorator(RandomHandler, sectionData))

	log.Fatal(http.ListenAndServe(":8080", router))
}

type section struct {
	Book    int    `json:"book"`
	Section int    `json:"section"`
	Text    string `json:"text"`
	Length  int    `json:"length"`
}

func initSectionData(pathToSectionData string) ([]section, error) {
	var sectionData []section

	sectionDataJson, err := ioutil.ReadFile(pathToSectionData)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(sectionDataJson, &sectionData)
	if err != nil {
		return nil, err
	}

	return sectionData, nil
}

func getRandomSection(sectionData []section, maxLength int) (section, error) {
	if maxLength <= 0 && maxLength != -1 {
		return section{}, errors.New("maxLength must be greater than 0")
	}

	if maxLength != -1 {
		var filteredSections []section
		for _, s := range sectionData {
			if s.Length <= maxLength {
				filteredSections = append(filteredSections, s)
			}
		}
		sectionData = filteredSections
		if len(sectionData) == 0 {
			return section{}, errors.New("No sections shorter than maxLength")
		}
	}

	seed := rand.NewSource(time.Now().UnixNano())
	rng := rand.New(seed)
	random := rng.Intn(len(sectionData))
	return sectionData[random], nil
}

func ContextDecorator(h httprouter.Handle, sectionData []section) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		ctx := context.WithValue(r.Context(), "sectionData", sectionData)
		h(w, r.WithContext(ctx), ps)
	}
}

func IndexHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	sectionData := r.Context().Value("sectionData").([]section)
	section, err := getRandomSection(sectionData, 350)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, err)
		return
	}
	t, err := template.ParseFiles("views/index.html")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, err)
		return
	}
	t.Execute(w, &section)
}

func FaqHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	t, err := template.ParseFiles("views/faq.html")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, err)
		return
	}
	t.Execute(w, nil)
}

func MeditationsHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	sectionData := r.Context().Value("sectionData").([]section)
	bookIndex, err := strconv.Atoi(ps.ByName("bookIndex"))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, "Invalid book parameter")
		return
	}
	sectionIndex, err := strconv.Atoi(ps.ByName("sectionIndex"))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, "Invalid section parameter")
		return
	}

	var foundSection section
	for _, s := range sectionData {
		if s.Book == bookIndex && s.Section == sectionIndex {
			foundSection = s
			break
		}
	}

	if foundSection.Section == 0 {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, "404 meditation not found")
		return
	}

	jsonSection, err := json.MarshalIndent(foundSection, "", "    ")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprint(w, string(jsonSection))
}

func RandomHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	maxLength := -1
	if len(r.URL.Query()["maxLength"]) == 1 {
		maxLengthQueryParam, err := strconv.Atoi(r.URL.Query()["maxLength"][0])
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, "Invalid maxLength parameter")
			return
		}
		maxLength = maxLengthQueryParam
	}

	sectionData := r.Context().Value("sectionData").([]section)
	randomSection, err := getRandomSection(sectionData, maxLength)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, err)
		return
	}
	jsonSection, err := json.MarshalIndent(randomSection, "", "    ")
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprint(w, string(jsonSection))
}
