package mysql

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/golang/glog"
	v1 "github.com/yarntime/aiops/pkg/types"
)

var monitorQuery = `
SELECT
    id,
    host,
    instance_name,
    metric,
    monitor_types,
    es_index,
    es_type
FROM
    monitor_obj
`

var paramsQuery = `
SELECT
    param_name,
    param_value
FROM
    job_params
WHERE
    kpi_name='%s'
AND
    monitor_type=%d
`

type Worker struct {
	Dsn string
}

func NewDBWorker(c *v1.Config) *Worker {
	return &Worker{
		Dsn: fmt.Sprintf("%s:%s@tcp(%s)/%s", c.CustomCfg.Global.MysqlUser, c.CustomCfg.Global.MysqlPwd, c.CustomCfg.Global.MysqlHost, c.CustomCfg.Global.MysqlDB),
	}
}

func (w *Worker) List() []*v1.MonitorObject {
	list := []*v1.MonitorObject{}

	db, err := sql.Open("mysql", w.Dsn)
	if err != nil {
		glog.Error("Failed to connect to mysql server.")
		return list
	}
	defer db.Close()

	rows, err := db.Query(monitorQuery)
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	for rows.Next() {
		m := new(v1.MonitorObject)
		if err := rows.Scan(
			&m.ID,
			&m.Host,
			&m.InstanceName,
			&m.Metric,
			&m.MonitorTypes,
			&m.ESIndex,
			&m.ESType,
		); err != nil {
			glog.Errorf(err.Error())
			panic(err)
		}
		list = append(list, m)
	}

	return list
}

func (w *Worker) GetParams(kpiName string, monitorType int) []string {
	params := []string{}
	db, err := sql.Open("mysql", w.Dsn)
	if err != nil {
		glog.Error("Failed to connect to mysql server.")
		return params
	}
	defer db.Close()

	rows, err := db.Query(fmt.Sprintf(paramsQuery, kpiName, monitorType))
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	for rows.Next() {
		name, value := "", ""
		if err := rows.Scan(
			&name,
			&value,
		); err != nil {
			glog.Errorf(err.Error())
			panic(err)
		}
		params = append(params, name+"="+value)
	}

	return params
}
