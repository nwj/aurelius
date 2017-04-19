package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/julienschmidt/httprouter"
	"io/ioutil"
	"log"
	"net/http"
)

func main() {
	sectionData := initSectionData()
	router := httprouter.New()
	router.GET("/rand", ContextDecorator(RandomHandler, &sectionData))

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

func ContextDecorator(h httprouter.Handle, sectionData *[]section) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		ctx := context.WithValue(r.Context(), "sectionData", sectionData)
		h(w, r.WithContext(ctx), ps)
	}
}

func RandomHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	fmt.Fprint(w, "TESTING")
}
