package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm/logger"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	config "go-auth/internal/config"
)

// AppRepository defines a interface for access the database.
type AppRepository interface {
	Driver() *gorm.DB
	Model(value interface{}) *gorm.DB
	Select(query interface{}, args ...interface{}) *gorm.DB
	Find(out interface{}, where ...interface{}) *gorm.DB
	Exec(sql string, values ...interface{}) *gorm.DB
	First(out interface{}, where ...interface{}) *gorm.DB
	Raw(sql string, values ...interface{}) *gorm.DB
	Create(value interface{}) *gorm.DB
	// gorm insert or update all fields
	Save(value interface{}) *gorm.DB
	// gorm update non-zero fields by default
	Updates(value interface{}) *gorm.DB
	Delete(value interface{}) *gorm.DB
	Where(query interface{}, args ...interface{}) *gorm.DB
	Preload(column string, conditions ...interface{}) *gorm.DB
	Scopes(funcs ...func(*gorm.DB) *gorm.DB) *gorm.DB
	ScanRows(rows *sql.Rows, result interface{}) error
	Transaction(fc func(tx AppRepository) error) (err error)
	Close() error
	DropTableIfExists(value interface{}) error
	AutoMigrate(value interface{}) error
}

// repository defines a repository for access the database.
type repository struct {
	db *gorm.DB
}

type mainRepository struct {
	*repository
}

func MustNewRepository(config *config.AppConfig,

// logger loggerX.AppLogger
) AppRepository {

	db, err := connectDatabase(config) // logger

	if err != nil {
		panic(err)
	}

	dbSQL, _ := db.DB()

	err = dbSQL.Ping()

	if err != nil {
		panic(err)
	}

	if config.DB.MaxOpen > 0 {
		dbSQL.SetMaxOpenConns(config.DB.MaxOpen)
	}

	if config.DB.MaxIdle > 0 {
		dbSQL.SetMaxIdleConns(config.DB.MaxIdle)
	}

	if config.DB.IdleTime > 0 {
		dbSQL.SetConnMaxIdleTime(time.Duration(config.DB.IdleTime) * time.Second)
	}

	res := &mainRepository{&repository{
		db: db,
	}}

	return res
}

const (
	SQLITE   = "sqlite"
	POSTGRES = "postgres"
)

func connectDatabase(
	appConfig *config.AppConfig,
	// logger loggerX.AppLogger,
) (*gorm.DB, error) {
	cfg := &appConfig.DB
	var dsn string
	gormConfig := &gorm.Config{
		SkipDefaultTransaction: true,
		// Logger:                 logger,
		// NamingStrategy: schema.NamingStrategy{
		// 	// SingularTable: true, // not *s as table suffix // use singular table name, table for `User` would be `user` with this option enabled
		// 	// NoLowerCase:   true, // as-is // skip the snake_casing of names
		// },
		// Logger: loggerGorm.Default.LogMode(logger_gorm.Info),
	}

	if cfg.Debug {
		// init logger before connect
		gormConfig.Logger = logger.Default.LogMode(logger.Info)
		gormConfig.Logger.Info(context.TODO(), "gorm logger mode: info/debug")
	}

	if cfg.Dialect == POSTGRES {
		sslmode := "disable" // TODO "require" if cfg.SSL
		dsn = fmt.Sprintf("host=%s port=%s user=%s dbname=%s password=%s connect_timeout=%s application_name=%s sslmode=%s ",
			cfg.Host, cfg.Port, cfg.User,
			cfg.Name, cfg.Password, "10", appConfig.Name /*split*/, sslmode)

		if cfg.Schema != "" {
			dsn += " search_path=" + cfg.Schema //  search_path=your_schema
		}

		return gorm.Open(postgres.Open(dsn), gormConfig)
	} else if cfg.Dialect == SQLITE {
		return gorm.Open(sqlite.Open(cfg.Host), gormConfig)
	}

	panic("undefined db dialect")
}

func (rep *repository) Driver() *gorm.DB {
	return rep.db
}

// Model specify the model you would like to run db operations
func (rep *repository) Model(value interface{}) *gorm.DB {
	return rep.db.Model(value)
}

// Select specify fields that you want to retrieve from database when querying, by default, will select all fields;
func (rep *repository) Select(query interface{}, args ...interface{}) *gorm.DB {
	return rep.db.Select(query, args...)
}

// Find find records that match given conditions.
func (rep *repository) Find(out interface{}, where ...interface{}) *gorm.DB {
	return rep.db.Find(out, where...)
}

// Exec exec given SQL using by gorm.DB.
func (rep *repository) Exec(sql string, values ...interface{}) *gorm.DB {
	return rep.db.Exec(sql, values...)
}

// First returns first record that match given conditions, order by primary key.
func (rep *repository) First(out interface{}, where ...interface{}) *gorm.DB {
	return rep.db.First(out, where...)
}

// Raw returns the record that executed the given SQL using gorm.DB.
func (rep *repository) Raw(sql string, values ...interface{}) *gorm.DB {
	return rep.db.Raw(sql, values...)
}

// Create insert the value into database.
func (rep *repository) Create(value interface{}) *gorm.DB {
	return rep.db.Create(value)
}

// Save update value in database, if the value doesn't have primary key, will insert it.
func (rep *repository) Save(value interface{}) *gorm.DB {
	return rep.db.Save(value) // insert or update
}

// Update update value in database, update non-zero fields by default
func (rep *repository) Updates(value interface{}) *gorm.DB {
	return rep.db.Updates(value)
}

// Delete delete value match given conditions.
func (rep *repository) Delete(value interface{}) *gorm.DB {
	return rep.db.Delete(value)
}

// Where returns a new relation.
func (rep *repository) Where(query interface{}, args ...interface{}) *gorm.DB {
	return rep.db.Where(query, args...)
}

// Preload preload associations with given conditions.
func (rep *repository) Preload(column string, conditions ...interface{}) *gorm.DB {
	return rep.db.Preload(column, conditions...)
}

// Scopes pass current database connection to arguments `func(*DB) *DB`,
// which could be used to add conditions dynamically
func (rep *repository) Scopes(funcs ...func(*gorm.DB) *gorm.DB) *gorm.DB {
	return rep.db.Scopes(funcs...)
}

// ScanRows scan `*sql.Rows` to give struct
func (rep *repository) ScanRows(rows *sql.Rows, result interface{}) error {
	return rep.db.ScanRows(rows, result)
}

// Close close current db connection. If database connection is not an io.Closer, returns an error.
func (rep *repository) Close() error {
	sqlDB, _ := rep.db.DB()
	return sqlDB.Close()
}

// DropTableIfExists drop table if it is exist
func (rep *repository) DropTableIfExists(value interface{}) error {
	return rep.db.Migrator().DropTable(value)
}

// AutoMigrate run auto migration for given models, will only add missing fields, won't delete/change current data
func (rep *repository) AutoMigrate(value interface{}) error {
	return rep.db.AutoMigrate(value)
}

// Transaction start a transaction as a block.
// If it is failed, will rollback and return error.
// If it is sccuessed, will commit.
// ref: https://github.com/jinzhu/gorm/blob/master/main.go#L533
func (rep *repository) Transaction(fc func(tx AppRepository) error) (err error) {
	panicked := true
	tx := rep.db.Begin()
	defer func() {
		if panicked || err != nil {
			tx.Rollback()
		}
	}()

	txrep := &repository{}
	txrep.db = tx
	err = fc(txrep)

	if err == nil {
		err = tx.Commit().Error
	}

	panicked = false
	return
}
