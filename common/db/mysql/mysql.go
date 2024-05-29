package mysql

import (
	"github.com/ambitiousmice/go-one/common/log"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"sync"
)

var once sync.Once
var config *Config

var shareDB *gorm.DB

type Config struct {
	Addr                                     string `yaml:"addr"`
	Database                                 string `yaml:"database"`
	Username                                 string `yaml:"username"`
	Password                                 string `yaml:"password"`
	DefaultStringSize                        uint   `yaml:"default-string-size"`
	DisableDatetimePrecision                 bool   `yaml:"disable-datetime-precision"`
	DontSupportRenameIndex                   bool   `yaml:"disable-rename-index"`
	DontSupportRenameColumn                  bool   `yaml:"disable-rename-column"`
	SkipInitializeWithVersion                bool   `yaml:"skip-initialize-with-version"`
	DisableForeignKeyConstraintWhenMigrating bool   `yaml:"disable-foreign-key-constraint-when-migrating"`
	MaxIdleConns                             int    `yaml:"max-idle-conns"`
	MaxOpenConns                             int    `yaml:"max-open-conns"`
}

func (c *Config) BuildMysqlConfig() mysql.Config {
	mysqlConfig := &mysql.Config{}

	dsn := c.Username
	if c.Password != "" {
		dsn = dsn + ":" + c.Password
	}
	dsn = dsn + "@tcp(" + c.Addr + ")/" + c.Database + "?charset=utf8mb4&parseTime=true&loc=Local"
	mysqlConfig.DSN = dsn
	mysqlConfig.DefaultStringSize = c.DefaultStringSize
	mysqlConfig.DisableDatetimePrecision = c.DisableDatetimePrecision
	mysqlConfig.DontSupportRenameIndex = c.DontSupportRenameIndex
	mysqlConfig.DontSupportRenameColumn = c.DontSupportRenameColumn
	mysqlConfig.SkipInitializeWithVersion = c.SkipInitializeWithVersion
	return *mysqlConfig
}

func (c *Config) BuildGormConfig() *gorm.Config {
	gormConfig := &gorm.Config{}
	gormConfig.DisableForeignKeyConstraintWhenMigrating = c.DisableForeignKeyConstraintWhenMigrating
	return gormConfig
}

func InitMysql(c *Config) {
	if c == nil || c.Addr == "" {
		return
	}

	config = c
	GetMysqlClient()

	log.Infof("Mysql init success:%s", c.Addr)
}

func GetMysqlClient() *gorm.DB {
	once.Do(func() {
		// 创建 Mysql 客户端，只会执行一次
		db, err := gorm.Open(mysql.New(config.BuildMysqlConfig()), config.BuildGormConfig())

		if err != nil {
			log.Fatal(err)
			return
		}
		sqlDB, _ := db.DB()
		if config.MaxIdleConns > 0 {
			sqlDB.SetMaxIdleConns(config.MaxIdleConns)
		}
		if config.MaxOpenConns > 0 {
			sqlDB.SetMaxOpenConns(config.MaxOpenConns)
		}
		shareDB = db
	})

	return shareDB
}
