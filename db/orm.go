package db

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/billyyoyo/microj/config"
	"github.com/billyyoyo/microj/logger"
	"github.com/bwmarrin/snowflake"
	"github.com/pkg/errors"
	"gorm.io/gorm"
	ormlog "gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
	"strconv"
	"strings"
	"time"
)

var (
	dbConf      *DBConfig
	orm         *gorm.DB
	db          *sql.DB
	idGenerater *snowflake.Node
)

type DBConfig struct {
	Host             string       `yaml:"host"`
	Port             int          `yaml:"port"`
	User             string       `yaml:"user"`
	Password         string       `yaml:"password"`
	DB               string       `yaml:"db"`
	Url              string       `yaml:"url"`
	DDL              bool         `yaml:"ddl"`
	Debug            bool         `yaml:"debug"`
	SlowSqlThreshold int64        `yaml:"slowSqlThreshold"`
	TableNamePrefix  string       `yaml:"tableNamePrefix"`
	Pool             DBPoolConfig `yaml:"pool" mapstructure:"pool"`
}

type DBPoolConfig struct {
	MaxOpenConns    int   `yaml:"maxOpenConns"`
	MaxIdleConns    int   `yaml:"maxIdleConns"`
	ConnMaxLifeTime int64 `yaml:"connMaxLifeTime"`
	ConnMaxIdleTime int64 `yaml:"connMaxIdleTime"`
}

func InitDataSource(openFunc func(dsn string) gorm.Dialector, models []interface{}) {
	err := config.Scan("dataSource", &dbConf)
	if err != nil {
		logger.Fatal("db config load error", errors.Wrap(err, ""))
		return
	}
	if dbConf.Host == "" {
		logger.Fatal("db host no value", nil)
		return
	}
	if dbConf.Port == 0 {
		logger.Fatal("db port no value", nil)
	}
	if dbConf.User == "" {
		logger.Fatal("db user no value", nil)
	}
	if dbConf.Password == "" {
		logger.Fatal("db password no value", nil)
	}
	if dbConf.DB == "" {
		logger.Fatal("db dbname no value", nil)
	}
	dbConf.Url = strings.ReplaceAll(dbConf.Url, "${db.host}", dbConf.Host)
	dbConf.Url = strings.ReplaceAll(dbConf.Url, "${db.port}", strconv.Itoa(dbConf.Port))
	dbConf.Url = strings.ReplaceAll(dbConf.Url, "${db.user}", dbConf.User)
	dbConf.Url = strings.ReplaceAll(dbConf.Url, "${db.password}", dbConf.Password)
	dbConf.Url = strings.ReplaceAll(dbConf.Url, "${db.db}", dbConf.DB)

	if dbConf.SlowSqlThreshold == 0 {
		dbConf.SlowSqlThreshold = 3000
	}
	if dbConf.TableNamePrefix == "" {
		dbConf.TableNamePrefix = "t_"
	}

	initIdGenerater()

	orm, err = gorm.Open(openFunc(dbConf.Url), &gorm.Config{
		Logger: &dLogger{},
		NamingStrategy: schema.NamingStrategy{
			TablePrefix:   dbConf.TableNamePrefix, // table name prefix, table for `User` would be `t_users`
			SingularTable: true,                   // use singular table name, table for `User` would be `user` with this option enabled
			NoLowerCase:   false,                  // skip the snake_casing of names
		},
	})
	if err != nil {
		logger.Fatal(err.Error(), err)
	}
	if dbConf.Debug {
		orm.Debug()
	}
	db, err = orm.DB()
	if err != nil {
		logger.Fatal(err.Error(), err)
	}
	fmt.Printf("%+v\n", dbConf)
	if dbConf.Pool.MaxIdleConns > 0 {
		db.SetMaxIdleConns(dbConf.Pool.MaxIdleConns)
	}
	if dbConf.Pool.MaxOpenConns > 0 {
		db.SetMaxOpenConns(dbConf.Pool.MaxOpenConns)
	}
	if dbConf.Pool.ConnMaxIdleTime > 0 {
		db.SetConnMaxIdleTime(time.Duration(dbConf.Pool.ConnMaxIdleTime) * time.Second)
	}
	if dbConf.Pool.ConnMaxLifeTime > 0 {
		db.SetConnMaxLifetime(time.Duration(dbConf.Pool.ConnMaxLifeTime) * time.Second)
	}
	if dbConf.DDL && len(models) > 0 {
		orm.AutoMigrate(models...)
	}

}

func initIdGenerater() {
	var err error
	idGenerater, err = snowflake.NewNode(0)
	if err != nil {
		logger.Fatal("Id generater init failed", errors.Wrap(err, ""))
	}
}

func Orm() *sql.DB {
	return db
}

func NextID() int64 {
	return idGenerater.Generate().Int64()
}

type dLogger struct {
}

func (l *dLogger) LogMode(level ormlog.LogLevel) ormlog.Interface {
	return l
}

func (l *dLogger) Info(ctx context.Context, msg string, infos ...interface{}) {
	logger.Info(msg, infos)
}

func (l *dLogger) Warn(ctx context.Context, msg string, infos ...interface{}) {
	logger.Warn(msg, infos)
}

func (l *dLogger) Error(ctx context.Context, msg string, infos ...interface{}) {
	var kv []logger.Val
	if len(infos) > 0 {
		for i, info := range infos {
			kv = append(kv, logger.Val{
				K: strconv.Itoa(i),
				V: fmt.Sprintf("\t%v\n", info),
			})
		}
	}

	logger.Error("", errors.New(msg), kv...)
}

func (l *dLogger) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	elapsed := time.Since(begin)
	switch {
	case err != nil && !errors.Is(err, gorm.ErrRecordNotFound):
		sql, rows := fc()
		logger.Errorf("Error sql (%s) has error, effected: %d", errors.Wrap(err, ""), sql, rows)
	case elapsed.Milliseconds() > dbConf.SlowSqlThreshold && dbConf.SlowSqlThreshold > 0:
		sql, rows := fc()
		logger.Warnf("Slow sql (%s) spend %dms , effected: %d", sql, elapsed.Milliseconds(), rows)
	case dbConf.Debug:
		sql, rows := fc()
		logger.Infof("Exec sql (%s) spend %d, effected: %d", sql, elapsed.Milliseconds(), rows)
	}
}
