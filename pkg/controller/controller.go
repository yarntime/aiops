package controller

import (
	v1 "github.com/yarntime/aiops/pkg/types"
	"github.com/yarntime/aiops/pkg/mysql"
	"net/http"
	"github.com/golang/glog"
)

type JobController struct {
	dbWorker *mysql.Worker
}

func NewController(c *v1.Config)  *JobController {
	return &JobController{
		dbWorker: mysql.NewDBWorker(c),
	}
}

func (c *JobController) Scan(w http.ResponseWriter, req *http.Request) {
	glog.V(3).Info("scan the monitor object, create the cron jobs")
}