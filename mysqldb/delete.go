package mysqldb

import (
	"fmt"

	"github.com/yzinkovets/dbex/types"

	_ "github.com/go-sql-driver/mysql"
)

type QueryDelete struct {
	db     types.Db
	logger types.Logger
	Entity string
	Id     uint32
}

func NewQueryDelete(db types.Db, logger types.Logger, entity string, id uint32) QueryDelete {
	query := QueryDelete{
		Entity: entity,
		Id:     id,
		logger: logger,
	}

	if query.logger == nil {
		query.logger = types.NewDefaultLogger()
	}

	return query
}

// Returns final query string
func (this *QueryDelete) GetQuery() string {
	sql := fmt.Sprintf("delete from %s where id=%d", this.Entity, this.Id)
	this.logger.Trace(sql)
	return sql
}

func (this *QueryDelete) Delete() error {
	sql := this.GetQuery()
	if sql == "" { // nothing to delete
		this.logger.Trace("Nothing to delete")
		return nil
	}

	_, err := this.db.Exec(sql)
	if err != nil {
		this.logger.Error(err)
	}
	return err
}
