package mysqldb

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/go-sql-driver/mysql"
	"github.com/yzinkovets/dbex/types"
	"github.com/yzinkovets/dbex/utils"
)

type QueryCreate struct {
	db     types.Db
	logger types.Logger
	Entity string
	Params map[string]any
	Fields []string
	Values []string
}

func NewQueryCreate(db types.Db, logger types.Logger, entity string, params map[string]interface{}) QueryCreate {
	query := QueryCreate{
		db:     db,
		Entity: entity,
		Params: params,
		Fields: []string{},
		Values: []string{},
		logger: logger,
	}

	if query.logger == nil {
		query.logger = types.NewDefaultLogger()
	}

	query.setNullValuesInParams()

	return query
}

func (this *QueryCreate) setNullValuesInParams() {
	for k, v := range this.Params {
		if "null" == utils.GetStringFromAny(v) {
			this.Params[k] = sql.NullString{String: "", Valid: false}
		}
	}
}

func (this *QueryCreate) AddField(param string) {
	if _, ok := this.Params[param]; ok {
		this.Fields = append(this.Fields, param)
		this.Values = append(this.Values, ":"+param)
	}
}

// Returns final query string
func (this *QueryCreate) GetQuery() string {
	if len(this.Fields) == 0 {
		return ""
	}

	fieldsList := strings.Join(this.Fields[:], ",")
	valuesList := strings.Join(this.Values[:], ",")

	sql := "insert into " + this.Entity + "(" + fieldsList + ") values(" + valuesList + ")"
	this.logger.Trace(sql)
	this.logger.Trace(fmt.Sprintf("Params: %v", this.Params))

	return sql
}

// Returns id of inserted record or 0 in case of an error
func (this *QueryCreate) Insert(params map[string]interface{}) (uint32, error) {
	sql := this.GetQuery()
	if sql == "" { // nothing to insert
		this.logger.Trace("Nothing to create")
		return 0, nil
	}

	pq, err := this.db.PrepareNamed(sql)
	if err != nil {
		return 0, err
	}

	defer pq.Close()

	pqu := pq.Unsafe()

	r, err := pqu.Exec(params)
	if err != nil {
		if e, ok := err.(*mysql.MySQLError); ok && e.Number == 1062 {
			return 0, errors.New("Object already exists")
		}
		return 0, err
	}

	id, err := r.LastInsertId()
	if err != nil {
		return 0, err
	}

	return uint32(id), nil
}
