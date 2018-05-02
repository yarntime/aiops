package mysql

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/golang/glog"
	v1 "github.com/yarntime/aiops/pkg/types"
)

var peopleQuery = `
SELECT
    host,
    instance_name,
    metric,
    monitor_types
FROM
    monitor_obj
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

	rows, err := db.Query(peopleQuery)
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	for rows.Next() {
		m := new(v1.MonitorObject)
		if err := rows.Scan(
			&m.Host,
			&m.InstanceName,
			&m.Metric,
			&m.MonitorTypes,
		); err != nil {
			glog.Errorf(err.Error())
			panic(err)
		}
		list = append(list, m)
	}

	return list
}
