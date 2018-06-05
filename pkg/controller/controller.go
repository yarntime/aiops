package controller

import (
	"encoding/json"
	"github.com/golang/glog"
	"github.com/yarntime/aiops/pkg/mysql"
	v1 "github.com/yarntime/aiops/pkg/types"
	"net/http"
)

type Controller struct {
	dbWorker      *mysql.Worker
	jobController *JobController
	customConfig  v1.CustomConfig
	appConfig     v1.ApplicationConfig
}

func NewController(c *v1.Config) *Controller {
	return &Controller{
		dbWorker:      mysql.NewDBWorker(c),
		jobController: NewJobController(c),
		customConfig:  c.CustomCfg,
		appConfig:     c.AppCfg,
	}
}

func (c *Controller) Run(stopCh <-chan struct{}) {
	c.jobController.Run(stopCh)
}

func (c *Controller) Create(w http.ResponseWriter, req *http.Request) {
	glog.V(3).Info("scan the monitor objects, create cron jobs")

	c.jobController.DeleteCronJob(c.customConfig)
	monitorObjects := c.dbWorker.List()
	for _, monitorObject := range monitorObjects {
		glog.Infof("creating cronjob for monitor object: %v\n", monitorObject)
		for _, appConf := range c.appConfig.App {
			if monitorObject.MonitorTypes&appConf.Id != 0 {
				objParams := c.dbWorker.GetParams(monitorObject.Metric, appConf.Id)
				job, err := c.jobController.CreateCronJob(monitorObject, c.customConfig, appConf, objParams)
				if err != nil {
					continue
				}
				if c.customConfig.Global.TriggerJobOnCreation {
					c.jobController.CreateJobFromCronJob(job)
				}
			}
		}
	}

	res := &v1.ApiResponse{
		Code:    200,
		Message: "Successful to scan monitor objects.",
	}

	r, _ := json.Marshal(res)
	w.Write(r)
}

func (c *Controller) Delete(w http.ResponseWriter, req *http.Request) {
	c.jobController.DeleteCronJob(c.customConfig)
	res := &v1.ApiResponse{
		Code:    200,
		Message: "Successful to delete all cron jobs.",
	}

	r, _ := json.Marshal(res)
	w.Write(r)
}
