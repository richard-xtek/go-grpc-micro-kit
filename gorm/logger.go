package gorm

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/richard-xtek/go-grpc-micro-kit/log"
	"go.uber.org/zap"
)

// DBLogger ...
type DBLogger struct {
	log log.Factory
}

// NewDBLogger ...
func NewDBLogger(log log.Factory) *DBLogger {
	return &DBLogger{log}
}

// Print ...
func (dbl *DBLogger) Print(params ...interface{}) {
	if len(params) <= 1 {
		return
	}

	level := params[0]
	log := dbl.log.With(zap.String("gorm_level", level.(string)), zap.String("db_src", params[1].(string)))

	if level != "sql" {
		log.Bg().Debug("", zap.Any("", params[2:]))
		// log.Debug(params[2:]...)
		return
	}

	dur := params[2].(time.Duration)
	sql := params[3].(string)
	sqlValues := params[4].([]interface{})
	rows := params[5].(int64)

	values := ""
	if valuesJSON, err := json.Marshal(sqlValues); err == nil {
		values = string(valuesJSON)
	} else {
		values = fmt.Sprintf("%+v", sqlValues)
	}

	log.Bg().Debug("", zap.Int64("dur_ns", dur.Nanoseconds()), zap.String("sql", strings.ReplaceAll(sql, `"`, `'`)), zap.String("values", strings.ReplaceAll(values, `"`, `'`)), zap.Int64("rows", rows))
}
