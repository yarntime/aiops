package types

type Application struct {
	Application   string   `json:"application"`
	Id            int      `json:"id"`
	Image         string   `json:image`
	Cmd           []string `json:"cmd"`
	Cron          string   `json:"cron"`
	CpuRequest    string   `json:"cpuRequest"`
	MemoryRequest string   `json:"memoryRequest"`
	Params        []string `json:"params"`
}

type ApplicationConfig struct {
	App []Application `json:"app"`
}

type GlobalConfig struct {
	MysqlHost                  string   `json:"mysql_host"`
	MysqlUser                  string   `json:"mysql_user"`
	MysqlPwd                   string   `json:"mysql_pwd"`
	MysqlDB                    string   `json:"mysql_db"`
	Params                     []string `json:"params"`
	Namespace                  string
	SuccessfulJobsHistoryLimit int32
	FailedJobsHistoryLimit     int32
	ConcurrencyPolicy          string
	ImagePullPolicy            string
	TriggerJobOnCreation       bool
}

type CustomConfig struct {
	Global GlobalConfig `json:"global"`
}

type Config struct {
	AppCfg    ApplicationConfig
	CustomCfg CustomConfig
	Host      string
}

type MonitorObject struct {
	ID           int
	Host         string
	InstanceName string
	Metric       string
	MonitorTypes int
	ESIndex      string
	ESType       string
}

type ApiResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}
