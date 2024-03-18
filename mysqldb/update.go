package mysqldb

import (
	"database/sql"
	"fmt"
	"strings"

	_ "github.com/go-sql-driver/mysql"

	"github.com/yzinkovets/dbex/types"
	"github.com/yzinkovets/dbex/utils"
)

type QueryUpdate struct {
	db     types.Db
	logger types.Logger
	Entity string
	Id     uint32
	Params map[string]any
	Fields []string
}

func NewQueryUpdate(db types.Db, logger types.Logger, entity string, id uint32, params map[string]interface{}) QueryUpdate {
	query := QueryUpdate{
		db:     db,
		logger: logger,
		Entity: entity,
		Id:     id,
		Params: params,
		Fields: []string{},
	}

	if query.logger == nil {
		query.logger = types.NewDefaultLogger()
	}

	query.setNullValuesInParams()

	return query
}

func (this *QueryUpdate) setNullValuesInParams() {
	for k, v := range this.Params {
		if "null" == utils.GetStringFromAny(v) {
			this.Params[k] = sql.NullString{String: "", Valid: false}
		}
	}
}

func (this *QueryUpdate) AddField(param string) {
	if _, ok := this.Params[param]; ok {
		this.Fields = append(this.Fields, param+" = :"+param)
	}
}

// Returns final query string
func (this *QueryUpdate) GetQuery() string {
	if len(this.Fields) == 0 {
		return ""
	}

	fieldsList := strings.Join(this.Fields[:], ",")

	sql := "update " + this.Entity + " set " + fieldsList + " where id=" + fmt.Sprintf("%d", this.Id)
	this.logger.Trace(sql)
	this.logger.Trace(fmt.Sprintf("Params: %v", this.Params))

	return sql
}

// Returns id of inserted record or 0 in case of an error
func (this *QueryUpdate) Update(params map[string]interface{}) error {
	sql := this.GetQuery()
	if sql == "" { // nothing to update
		this.logger.Trace("Nothing to update")
		return nil
	}

	pq, err := this.db.PrepareNamed(sql)
	if err != nil {
		return err
	}

	defer pq.Close()

	pqu := pq.Unsafe()

	_, err = pqu.Exec(params)
	if err != nil {
		return err
	}

	return nil
}
