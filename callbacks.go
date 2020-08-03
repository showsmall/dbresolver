package dbresolver

import (
	"strings"

	"gorm.io/gorm"
)

func (dr *DBResolver) registerCallbacks(db *gorm.DB) {
	dr.Callback().Create().Before("*").Register("gorm:db_resolver", dr.switchSource)
	dr.Callback().Query().Before("*").Register("gorm:db_resolver", dr.switchSource)
	dr.Callback().Update().Before("*").Register("gorm:db_resolver", dr.switchSource)
	dr.Callback().Delete().Before("*").Register("gorm:db_resolver", dr.switchReplica)
	dr.Callback().Row().Before("*").Register("gorm:db_resolver", dr.switchGuess)
}

func (dr *DBResolver) switchSource(db *gorm.DB) {
	if !isTransaction(db.Statement.ConnPool) {
		db.Statement.ConnPool = dr.resolve(db.Statement, Write)
	}
}

func (dr *DBResolver) switchReplica(db *gorm.DB) {
	if !isTransaction(db.Statement.ConnPool) {
		db.Statement.ConnPool = dr.resolve(db.Statement, Read)
	}
}

func (dr *DBResolver) switchGuess(db *gorm.DB) {
	if !isTransaction(db.Statement.ConnPool) {
		if rawSQL := db.Statement.SQL.String(); len(rawSQL) > 6 && strings.EqualFold(rawSQL[:6], "select") {
			db.Statement.ConnPool = dr.resolve(db.Statement, Read)
		} else {
			db.Statement.ConnPool = dr.resolve(db.Statement, Write)
		}
	}
}

func isTransaction(connPool gorm.ConnPool) bool {
	switch connPool.(type) {
	case gorm.TxBeginner:
		return true
	case gorm.ConnPoolBeginner:
		return true
	default:
		return false
	}
}