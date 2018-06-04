package main

import (
	"encoding/json"
	"flag"
	"github.com/golang/glog"
	"github.com/gorilla/mux"
	"github.com/yarntime/aiops/pkg/controller"
	v1 "github.com/yarntime/aiops/pkg/types"
	io "io/ioutil"
	"net/http"
	"reflect"
)

var (
	apiserverAddress           string
	jobNamespace               string
	successfulJobsHistoryLimit int
	failedJobsHistoryLimit     int
	concurrencyPolicy          string
	imagePullPolicy            string
	triggerJobOnCreation       bool
	/*mysqlHost string
	mysqlUser string
	mysqlPassword string
	mysqlDB  string*/
	globalConfig      string
	applicationConfig string
)

func init() {
	flag.StringVar(&apiserverAddress, "apiserver_address", "", "Kubernetes apiserver address")
	flag.StringVar(&jobNamespace, "job_namespace", "default", "cronjob namespace")
	flag.IntVar(&successfulJobsHistoryLimit, "sucessful_jobs_history_limit", 1, "sucessful jobs history limit")
	flag.IntVar(&failedJobsHistoryLimit, "failed_jobs_history_limit", 2, "failed jobs history limit")
	flag.StringVar(&concurrencyPolicy, "concurrency_policy", "Forbid", "concurrency policy, support Allow, Forbid and Replace, default to Forbid")
	flag.StringVar(&imagePullPolicy, "image_pull_policy", "IfNotPresent", "image pull policy, support Always, Never and IfNotPresent, default to IfNotPresent")
	flag.BoolVar(&triggerJobOnCreation, "trigger_job_on_creation", false, "trigger job on cronjob creation")
	/*flag.StringVar(&mysqlHost, "mysql_host", "192.168.254.44:31786", "mysql host")
	flag.StringVar(&mysqlUser, "mysql_user", "root", "mysql user")
	flag.StringVar(&mysqlPassword, "mysql_password", "password", "mysql password")
	flag.StringVar(&mysqlDB, "mysql_db", "skyform_aiops", "mysql db")*/
	flag.StringVar(&globalConfig, "global_config_file", "/etc/aiops/config.json", "global config file")
	flag.StringVar(&applicationConfig, "applicationConfig", "/etc/aiops/application.json", "application config file")
	flag.Set("alsologtostderr", "true")
	flag.Parse()
}

func main() {
	customConfig := v1.CustomConfig{
		Global: v1.GlobalConfig{
			Namespace:                  jobNamespace,
			SuccessfulJobsHistoryLimit: int32(successfulJobsHistoryLimit),
			FailedJobsHistoryLimit:     int32(failedJobsHistoryLimit),
			ConcurrencyPolicy:          concurrencyPolicy,
			ImagePullPolicy:            imagePullPolicy,
			TriggerJobOnCreation:       triggerJobOnCreation,
		},
	}
	err := LoadConfig(globalConfig, &customConfig)
	if err != nil {
		glog.Fatalf("Failed to load custom config. %s", err.Error())
	}

	appConfig := v1.ApplicationConfig{}
	err = LoadConfig(applicationConfig, &appConfig)
	if err != nil {
		glog.Fatalf("Failed to load application config.%s", err.Error())
	}

	initAppConfig(customConfig, appConfig)

	config := &v1.Config{
		CustomCfg: customConfig,
		AppCfg:    appConfig,
		Host:      apiserverAddress,
	}

	c := controller.NewController(config)

	router := mux.NewRouter()
	router.HandleFunc("/health", health).Methods("GET")
	router.HandleFunc("/create", c.Create).Methods("GET")
	router.HandleFunc("/delete", c.Delete).Methods("GET")

	glog.Info("http server started.")
	glog.Fatal(http.ListenAndServe(":8080", router))
}

func LoadConfig(filename string, v interface{}) error {
	data, err := io.ReadFile(filename)
	if err != nil {
		return err
	}

	dataJson := []byte(data)
	err = json.Unmarshal(dataJson, v)
	if err != nil {
		return err
	}

	return nil
}

func initAppConfig(customConfig v1.CustomConfig, appConfig v1.ApplicationConfig) {
	globalFiled := reflect.TypeOf(customConfig.Global)
	globalValue := reflect.ValueOf(customConfig.Global)
	baseLineParams := []string{}
	capacityParams := []string{}
	for i := 0; i < globalFiled.NumField(); i++ {
		f := globalFiled.Field(i)
		name, exist := f.Tag.Lookup("json")
		if !exist {
			continue
		}
		if name != "params" {
			capacityParams = append(capacityParams, "--"+name+"="+globalValue.Field(i).Interface().(string))
		} else {
			baseLineParams = append(baseLineParams, globalValue.Field(i).Interface().([]string)...)
			capacityParams = append(capacityParams, globalValue.Field(i).Interface().([]string)...)
		}
	}
	for i := 0; i < len(appConfig.App); i++ {
		if appConfig.App[i].Id == 1 {
			appConfig.App[i].Params = append(appConfig.App[i].Params, baseLineParams...)
		} else {
			appConfig.App[i].Params = append(appConfig.App[i].Params, capacityParams...)
		}
	}
}

func health(w http.ResponseWriter, _ *http.Request) {
	w.Write([]byte("ok."))
}
