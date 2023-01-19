package sqlstore

import (
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/magicLian/gocron/pkg/models"
	"github.com/magicLian/gocron/pkg/setting"
	"github.com/magicLian/gocron/pkg/utils"
	"github.com/magicLian/logx"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

const (
	SqlSortDesc = "DESC"
	sqlSortAsc  = "ASC"
)

type SqlStoreInterface interface {
	NodeSqlStore
	TaskSqlStore
	GetDB() *gorm.DB
}

type SqlStore struct {
	cfg *setting.Cfg

	dbCfg models.DatabaseConfig
	db    *gorm.DB

	log logx.Logger
}

func ProvideSqlStore(Cfg *setting.Cfg) (SqlStoreInterface, error) {
	ss := &SqlStore{
		cfg: Cfg,
		log: logx.NewLogx("sqlstore", logx.DebugLevel),
	}

	ss.readConfig()

	db, err := ss.getEngine()
	if err != nil {
		return nil, fmt.Errorf("fail to connect to database: %v", err)
	}
	ss.db = db

	if err := ss.initTables(); err != nil {
		return nil, err
	}

	return ss, nil
}

func (ss *SqlStore) GetDB() *gorm.DB {
	return ss.db
}

func (ss *SqlStore) getEngine() (*gorm.DB, error) {
	connStr, err := ss.buildConnectionString()
	if err != nil {
		return nil, err
	}
	ss.log.Info("Connecting to DB", "dbtype", ss.dbCfg.Type)
	config := &gorm.Config{NamingStrategy: schema.NamingStrategy{
		SingularTable: true,
	}}
	if ss.dbCfg.LogQueries {
		config.Logger = logger.New(
			log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
			logger.Config{
				SlowThreshold:             time.Second, // Slow SQL threshold
				LogLevel:                  logger.Info, // Log level
				IgnoreRecordNotFoundError: true,        // Ignore ErrRecordNotFound error for logger
				Colorful:                  false,
			},
		)
	}
	engine, err := gorm.Open(postgres.Open(connStr), config)
	if err != nil {
		ss.log.Error("Failed to connection to db", "error", err.Error())
		return nil, err
	}

	if ss.dbCfg.SchemaName != "" {
		err = engine.Exec("CREATE SCHEMA IF NOT EXISTS \"" + ss.dbCfg.SchemaName + "\" AUTHORIZATION " + ss.dbCfg.GroupName + ";").Error
		if err != nil {
			ss.log.Error("Failed to create schema.", "schema", ss.dbCfg.SchemaName, "group", ss.dbCfg.GroupName, "error", err.Error())
			return nil, err
		}
	}

	db, err := engine.DB()
	if err != nil {
		return nil, err
	}
	db.SetConnMaxLifetime(time.Second * time.Duration(ss.dbCfg.ConnMaxLifetime))
	db.SetMaxOpenConns(ss.dbCfg.MaxOpenConn)
	db.SetMaxIdleConns(ss.dbCfg.MaxIdleConn)

	return engine, nil
}

func (ss *SqlStore) readConfig() {
	ss.dbCfg.SchemaName = utils.GetEnvOrIniValue(ss.cfg.Raw, "database", "schema_name")
	ss.dbCfg.GroupName = utils.GetEnvOrIniValue(ss.cfg.Raw, "database", "group_name")
	if ss.dbCfg.GroupName != "" {
		err := os.Setenv("group_name", ss.dbCfg.GroupName)
		if err != nil {
			panic(fmt.Sprintf("Failed to set group_name env,error %s", err))
		}
	}
	ss.dbCfg.MaxOpenConn = utils.SetDefaultInt(utils.GetEnvOrIniValue(ss.cfg.Raw, "database", "max_open_conn"), 100)
	ss.dbCfg.MaxIdleConn = utils.SetDefaultInt(utils.GetEnvOrIniValue(ss.cfg.Raw, "database", "max_idle_conn"), 100)
	ss.dbCfg.ConnMaxLifetime = utils.SetDefaultInt(utils.GetEnvOrIniValue(ss.cfg.Raw, "database", "conn_max_life_time"), 14400)
	ss.dbCfg.LogQueries = utils.SetDefaultBool(utils.GetEnvOrIniValue(ss.cfg.Raw, "database", "log_queries"), false)
	ss.dbCfg.SslMode = utils.SetDefaultString(utils.GetEnvOrIniValue(ss.cfg.Raw, "database", "ssl_mode"), "disable")
	ss.dbCfg.CaCertPath = utils.GetEnvOrIniValue(ss.cfg.Raw, "database", "ca_cert_path")
	ss.dbCfg.ClientKeyPath = utils.GetEnvOrIniValue(ss.cfg.Raw, "database", "client_key_path")
	ss.dbCfg.ClientCertPath = utils.GetEnvOrIniValue(ss.cfg.Raw, "database", "client_cert_path")

	//sqlite
	ss.dbCfg.Path = utils.SetDefaultString(utils.GetEnvOrIniValue(ss.cfg.Raw, "database", "path"), "data/asset.db")
	ss.dbCfg.CacheMode = utils.SetDefaultString(utils.GetEnvOrIniValue(ss.cfg.Raw, "database", "cache_mode"), "private")
}

func (ss *SqlStore) buildConnectionString() (string, error) {
	cnnstr := ss.dbCfg.ConnectionString
	if cnnstr != "" {
		return ss.dbCfg.ConnectionString, nil
	}
	switch ss.dbCfg.Type {
	case "postgres":
		addr, err := utils.SplitHostPortDefault(ss.dbCfg.Host, "127.0.0.1", "5432")
		if err != nil {
			return "", fmt.Errorf("invalid host specifier '%s',error %s", ss.dbCfg.Host, err.Error())
		}
		if ss.dbCfg.Password == "" {
			ss.dbCfg.Password = "''"
		}
		if ss.dbCfg.User == "" {
			ss.dbCfg.User = "''"
		}
		cnnstr = fmt.Sprintf("user=%s password=%s host=%s port=%s dbname=%s sslmode=%s ", ss.dbCfg.User, ss.dbCfg.Password, addr.Host, addr.Port, ss.dbCfg.Name, ss.dbCfg.SslMode)
		if ss.dbCfg.SslMode != "disable" {
			cnnstr += fmt.Sprintf(" sslcert=%s sslkey=%s sslrootcert=%s ", ss.dbCfg.ClientCertPath, ss.dbCfg.ClientKeyPath, ss.dbCfg.CaCertPath)
		}
		if ss.dbCfg.SchemaName != "" {
			cnnstr += " search_path=" + ss.dbCfg.SchemaName
		}
	case "sqlite":
		if !filepath.IsAbs(ss.dbCfg.Path) {
			ss.dbCfg.Path = filepath.Join(ss.cfg.DataPath, ss.dbCfg.Path)
		}
		if err := os.MkdirAll(path.Dir(ss.dbCfg.Path), os.ModePerm); err != nil {
			return "", err
		}

		cnnstr = fmt.Sprintf("file:%s?cache=%s&mode=rwc", ss.dbCfg.Path, ss.dbCfg.CacheMode)

		cnnstr += ss.buildExtraConnectionString('&')
	default:
		return "", fmt.Errorf("unknown database type: %s", ss.dbCfg.Type)
	}
	return cnnstr, nil
}

func (ss *SqlStore) buildExtraConnectionString(sep rune) string {
	if ss.dbCfg.UrlQueryParams == nil {
		return ""
	}

	var sb strings.Builder
	for key, values := range ss.dbCfg.UrlQueryParams {
		for _, value := range values {
			sb.WriteRune(sep)
			sb.WriteString(key)
			sb.WriteRune('=')
			sb.WriteString(value)
		}
	}
	return sb.String()
}

func (ss *SqlStore) initTables() error {
	if err := ss.db.AutoMigrate(models.Node{}); err != nil {
		return err
	}
	if err := ss.db.AutoMigrate(models.Task{}); err != nil {
		return err
	}

	return nil
}
