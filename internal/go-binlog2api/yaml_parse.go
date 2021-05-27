package scripts

import (
	"fmt"
	"io/ioutil"

	"github.com/cheshir/go-mq"
	"github.com/siddontang/go-mysql/canal"
	"github.com/siddontang/go-mysql/mysql"
	"gopkg.in/yaml.v2"
)

// Config struct for dbConfig
type Config struct {
	BinlogFileName      string `yaml:"binlogFilename"`
	BinlogFilePos       uint32 `yaml:"binlogPos"`
	DataBaseAddress     string `yaml:"dbAddress"`
	DataBasePort        int    `yaml:"dbPort"`
	DataBaseUser        string `yaml:"dbUser"`
	DataBasePassword    string `yaml:"dbPassword"`
	Flavor              string `yaml:"dbFlavor"`
	APIBaseURL          string `yaml:"apiBaseUrl"`
	FlushPossitionCache bool   `yaml:"flushPossitionCache"`
}

//ReciverConfigs todo
type ReciverConfigs struct {
	RConfig []ReciverConfig `yaml:"Recivers"`
}

//ReciverConfig todo
type ReciverConfig struct {
	Db         string           `yaml:"Db"`
	Table      string           `yaml:"Table"`
	OnField    []string         `yaml:"OnField"`
	ReciverAPI ReciverAPIConfig `yaml:"ReciverAPI"`
	ReciverMQ  ReciverMQConfig  `yaml:"ReciverMQ"`
}

//ReciverAPIConfig todo
type ReciverAPIConfig struct {
	InsertMethod string `yaml:"InsertMethod"`
	UpdateMethod string `yaml:"UpdateMethod"`
	DeleteMethod string `yaml:"DeleteMethod"`
	InsertURL    string `yaml:"InsertURL"`
	UpdateURL    string `yaml:"UpdateURL"`
	DeleteURL    string `yaml:"DeleteURL"`
	Critical     bool   `yaml:"Critical"`
	Format       string `yaml:"Format"`
}

//ReciverMQConfig todo
type ReciverMQConfig struct {
	InsertQueueProducer string `yaml:"InsertQueueProducer"`
	UpdateQueueProducer string `yaml:"UpdateQueueProducer"`
	DeleteQueueProducer string `yaml:"DeleteQueueProducer"`
	Critical            bool   `yaml:"Critical"`
	Format              string `yaml:"Format"`
}

var path = "../configs/config.yaml"
var dfPath = "../configs/distribution_config.yaml"
var rCStatic ReciverConfigs
var rCState = false
var cStatic Config
var cState = false
var mqStatic mq.Config
var mqState = false

const defMth = "POST"
const defFmt = "STD"

var posFF = Position{}

// GetBinlogPosition sets attributes form mysql.Position
func GetBinlogPosition(mysqlPositionStructure *mysql.Position) (err error) {
	config := CConfig()
	err = posFF.Read()
	if err != nil || config.FlushPossitionCache == true {
		mysqlPositionStructure.Name = config.BinlogFileName
		mysqlPositionStructure.Pos = config.BinlogFilePos
		return nil
	}
	mysqlPositionStructure.Name = posFF.BinlogFilename
	mysqlPositionStructure.Pos = posFF.BinlogPos
	return nil
}

// Parse a yaml files
func (config *Config) Parse() (err error) {
	Logger.Info("Parsing Yaml Config")
	source, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	err = yaml.Unmarshal(source, &config)
	if err != nil {
		return err
	}

	return nil
}

// GetDatabaseConfig sets canal.Config
func GetDatabaseConfig(dbConfig *canal.Config) (err error) {
	config := CConfig()
	dbConfig.Addr = fmt.Sprintf("%s:%d", config.DataBaseAddress, config.DataBasePort)
	dbConfig.User = config.DataBaseUser
	dbConfig.Password = config.DataBasePassword
	dbConfig.Flavor = config.Flavor
	dbConfig.Dump.ExecutionPath = ""
	return nil
}

//Parse in ReciverConfigs struct
func (r *ReciverConfigs) Parse() error {
	Logger.Info("Parsing Yaml Reciver Config")
	source, err := ioutil.ReadFile(dfPath)
	if err != nil {
		fmt.Println(err)
		return err
	}
	err = yaml.Unmarshal(source, r)
	if err != nil {
		return err
	}
	return nil
}

func (r *ReciverConfigs) validate() {
	for k, v := range r.RConfig {
		if IsSet(v.Db) == false {
			Logger.Fatalln("Db name in yaml not set")
			panic("DB name not set")
		}
		if IsSet(v.Table) == false {
			Logger.Fatalln("Table name in yaml not set")
			panic("Table name not set")
		}
		if IsSet(v.ReciverAPI.InsertMethod) == false {
			v.ReciverAPI.InsertMethod = defMth
			r.RConfig[k].ReciverAPI.InsertMethod = defMth
		}
		if IsSet(v.ReciverAPI.DeleteMethod) == false {
			v.ReciverAPI.DeleteMethod = defMth
			r.RConfig[k].ReciverAPI.DeleteMethod = defMth
		}
		if IsSet(v.ReciverAPI.UpdateMethod) == false {
			r.RConfig[k].ReciverAPI.UpdateMethod = defMth
			v.ReciverAPI.UpdateMethod = defMth
		}
		if IsSet(v.ReciverAPI.Format) == false {
			r.RConfig[k].ReciverAPI.Format = defFmt
			v.ReciverAPI.Format = defFmt
		}
		if IsSet(v.ReciverMQ.Format) == false {
			r.RConfig[k].ReciverMQ.Format = defFmt
			v.ReciverMQ.Format = defFmt
		}

		if IsSet(v.ReciverAPI.InsertURL) == true && ValidateURL(v.ReciverAPI.InsertURL) == false {
			Logger.Fatalln("Wrong URL for Insert in Config : ", v.ReciverAPI.InsertURL)
			panic("Wrong URL for Insert in Config")
		}
		if IsSet(v.ReciverAPI.DeleteURL) == true && ValidateURL(v.ReciverAPI.DeleteURL) == false {
			Logger.Fatalln("Wrong URL for Delete in Config : ", v.ReciverAPI.DeleteURL)
			panic("Wrong URL for Delete in Config")
		}
		if IsSet(v.ReciverAPI.UpdateURL) == true && ValidateURL(v.ReciverAPI.UpdateURL) == false {
			Logger.Fatalln("Wrong URL for Update in Config : ", v.ReciverAPI.UpdateURL)
			panic("Wrong URL for Update in Config")
		}
		if ValidateHTTPMethod(v.ReciverAPI.InsertMethod) == false {
			Logger.Fatalln("Wrong Http Method for Insert in Config : ", v.ReciverAPI.InsertMethod)
			panic("Wrong Http Method for Insert in Config")
		}
		if ValidateHTTPMethod(v.ReciverAPI.DeleteMethod) == false {
			Logger.Fatalln("Wrong Http Method for Delete in Config : ", v.ReciverAPI.DeleteMethod)
			panic("Wrong Http Method for Delete in Config")
		}
		if ValidateHTTPMethod(v.ReciverAPI.UpdateMethod) == false {
			Logger.Fatalln("Wrong Http Method for Update in Config : ", v.ReciverAPI.UpdateMethod)
			panic("Wrong Http Method for Update in Config")
		}
		if ValidateFmt(v.ReciverAPI.Format) == false {
			Logger.Fatalln("Wrong format in Config for API: ", v.ReciverAPI.Format)
			panic("Wrong format in Config for API")
		}
		if ValidateFmt(v.ReciverMQ.Format) == false {
			Logger.Fatalln("Wrong format in Config for RabbitMQ: ", v.ReciverMQ.Format)
			panic("Wrong format in Config for RabbitMQ")
		}
	}
}

//RConfig TODO :DOROBIĆ BŁĘDY
func RConfig() ReciverConfigs {
	if rCState == true {
		return rCStatic
	}
	cf := ReciverConfigs{}
	err := cf.Parse()
	cf.validate()
	if err != nil {
		panic("nie udało się sparsować konfiguracji")
	}
	rCStatic = cf
	rCState = true
	return cf

}

//CConfig TODO :DOROBIĆ BŁĘDY
func CConfig() Config {
	if cState == true {
		return cStatic
	}
	cf := Config{}
	err := cf.Parse()
	if err != nil {
		panic("nie udało się sparsować konfiguracji")
	}
	cStatic = cf
	cState = true
	return cf

}

func MqConfig() mq.Config {
	if mqState == true {
		return mqStatic
	}
	err := mqParse()
	if err != nil {
		panic("nie udało się sparsować konfiguracji")
	}
	mqState = true
	return mqStatic
}

func mqParse() error {
	Logger.Info("Parsing Yaml mqConfig")
	source, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	err = yaml.Unmarshal(source, &mqStatic)
	if err != nil {
		return err
	}

	return nil
}
