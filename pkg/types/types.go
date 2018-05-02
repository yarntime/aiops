package types

type Application struct {
	Application string   `json:"application"`
	Id          int      `json:"id"`
	Cmd         string   `json:"cmd"`
	Params      []string `json:"params"`
}

type ApplicationConfig struct {
	App []Application `json:"app"`
}

type GlobalConfig struct {
	Image     string `json:"image"`
	ESHosts   string `json:"es_hosts"`
	Index     string `json:"index"`
	DocType   string `json:"doc_type"`
	Timename  string `json:"timename"`
	MysqlHost string `json:"mysql_host"`
	MysqlUser string `json:"mysql_user"`
	MysqlPwd  string `json:"mysql_pwd"`
	MysqlDB   string `json:"mysql_db"`
}

type CapacityPredictionConfig struct {
	ID                 int    `json:"id"`
	Name               string `json:"name"`
	Freq               string `json:"freq"`
	Timedelta          string `json:"timedelta"`
	Period             string `json:"period"`
	ConfidenceInterval string `json:"confidence_interval"`
	Threshold          string `json:"threshold"`
	TriggerTime        string `json:"trigger_time"`
	TriggerFreq        string `json:"trigger_freq"`
}

type BaseLineConfig struct {
	ID                 int    `json:"id"`
	Name               string `json:"name"`
	Freq               string `json:"freq"`
	Period             string `json:"period"`
	ConfidenceInterval string `json:"confidence_interval"`
	TriggerTime        string `json:"trigger_time"`
	TriggerFreq        string `json:"trigger_freq"`
}

type CustomConfig struct {
	Global             GlobalConfig             `json:"global"`
	CapacityPrediction CapacityPredictionConfig `json:"capacity_prediction"`
	BaseLine           BaseLineConfig           `json:"baseline"`
}

type Config struct {
	AppCfg    ApplicationConfig
	CustomCfg CustomConfig
	Host      string
}

type MonitorObject struct {
	Id           int
	Host         string
	InstanceName string
	Metric       string
	MonitorTypes int
}

type ApiResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}
