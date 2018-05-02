package mysql

import (
	v1 "github.com/yarntime/aiops/pkg/types"
)

type Worker struct {
	//mysql data source name
	Dsn string
}

func NewDBWorker(config *v1.Config) *Worker {
	return &Worker{
		Dsn: config.CustomCfg.Global.MysqlUser + ":" + config.CustomCfg.Global.MysqlPwd + "@tcp(" + config.CustomCfg.Global.MysqlHost + ")/" + config.CustomCfg.Global.MysqlDB,
	}
}


