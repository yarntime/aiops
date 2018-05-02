package main

import (
	v1 "github.com/yarntime/aiops/pkg/types"
	"github.com/yarntime/aiops/pkg/controller"
	io "io/ioutil"
	"encoding/json"
	"github.com/golang/glog"
	"net/http"

	"github.com/gorilla/mux"
)

func main() {
	customConfig := v1.CustomConfig{}
	err := LoadConfig("G:/opt/config.json", &customConfig)
	if err != nil {
		glog.Fatal("Failed to load custom config.")
	}

	appConfig := v1.ApplicationConfig{}
	err = LoadConfig("G:/opt/application.json", &appConfig)
	if err != nil {
		glog.Fatal("Failed to load application config." + err.Error())
	}

	config := &v1.Config{
		CustomCfg: customConfig,
		AppCfg: appConfig,
	}

	c := controller.NewController(config)

	router := mux.NewRouter()
	router.HandleFunc("/scan", c.Scan).Methods("GET")

	glog.Fatal(http.ListenAndServe(":8080", router))
}

func LoadConfig(filename string, v interface{}) error {
	data, err := io.ReadFile(filename)
	if err != nil{
		return err
	}

	dataJson := []byte(data)
	err = json.Unmarshal(dataJson, v)
	if err != nil{
		return err
	}

	return nil
}