package mysqldb

import (
	"fmt"
	"regexp"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	"github.com/yzinkovets/dbex/types"
)

type QuerySelect struct {
	db          types.Db
	logger      types.Logger
	Body        string
	Params      map[string]any
	Where       string
	Limit       string
	OrderBy     string
	OrderByRex  *regexp.Regexp
	OrParamsRex *regexp.Regexp
	OrValues    map[string]OrValue
}

type OrValue struct {
	Field string
	Value any
}

func NewQuerySelect(db types.Db, logger types.Logger, body string, params map[string]interface{}) QuerySelect {
	if _, ok := params["limit"]; !ok {
		params["limit"] = "200"
	}

	where := "where 1=1"
	if v, ok := params["where"]; ok {
		where = "where " + v.(string)
	}

	q := QuerySelect{
		db:          db,
		logger:      logger,
		Body:        body,
		Params:      params,
		Where:       where,
		Limit:       params["limit"].(string),
		OrderBy:     "",
		OrderByRex:  regexp.MustCompile(`^sort\[([^\]]+)\]$`),
		OrParamsRex: regexp.MustCompile(`^or\[([^\]]+)\]$`),
		OrValues:    map[string]OrValue{},
	}

	if q.logger == nil {
		q.logger = types.NewDefaultLogger()
	}

	q.OrderBy = q.getSortString(params)

	return q
}

// Returns sort string for query
// Example: "order by main.name asc, main.created_at desc"
// Sorting fields and orders are passed in "sort" parameter
// Example: "?sort[name]=asc&sort[created_at]=desc"
func (this *QuerySelect) getSortString(params map[string]interface{}) string {
	ret := ""

	for key, value := range params {
		if this.OrderByRex.MatchString(key) {
			matches := this.OrderByRex.FindStringSubmatch(key)
			if len(matches) == 2 {
				if ret != "" {
					ret += ", "
				}

				field := matches[1]
				if !strings.Contains(field, ".") {
					field = "main." + field
				}

				ret += fmt.Sprintf("%s %s", field, value.(string))
			}
		}
	}

	return ret
}

func (this *QuerySelect) Filter(param string) {
	this.FilterEx(param, "main."+param)
}

func (this *QuerySelect) FilterEx(param string, field string) {
	if this.processOrParam(param, field) {
		return
	}

	if _, ok := this.Params[param]; ok {
		this.Where += fmt.Sprintf(" and %s like :%s", field, param)
	}
}

// Returns true if param is OR parameter
// OR fields passed in "or" parameter
// Example: "?or[first_name]=*Yuri*&or[last_name]=*Zinko*"
func (this *QuerySelect) processOrParam(param string, field string) bool {
	orParam := "or[" + param + "]"
	if value, ok := this.Params[orParam]; ok && value != "" && value != "*" && value != "**" {
		this.OrValues[param] = OrValue{
			Field: field,
			Value: value,
		}
		return true
	}
	return false
}

func (this *QuerySelect) CustomFilter(condition string) {
	this.Where += " " + condition
}

// Returns final query string
func (this *QuerySelect) GetQuery() string {
	limit := ""

	// default API limit is 200 records
	if this.Limit == "" {
		this.Limit = "limit 200"
	} else {
		limit = " limit " + this.Limit
	}

	if len(this.OrValues) > 0 {
		this.Where += " and ("
		for param, ov := range this.OrValues {
			this.Where += fmt.Sprintf("%s like :%s or ", ov.Field, param)
			this.Params[param] = ov.Value
		}
		// remove last " or "
		this.Where = this.Where[:len(this.Where)-4] + ")"
	}

	sql := this.Body + " " + this.Where
	if this.OrderBy != "" {
		sql += " order by " + this.OrderBy
	}

	if limit != "" {
		sql += " " + limit
	}

	this.logger.Trace(sql)
	this.logger.Trace(fmt.Sprintf("Params: %v", this.Params))

	return sql
}

func SelectTyped[T any](db types.Db, sql string, params map[string]interface{}, ret *T) error {
	pq, err := db.PrepareNamed(sql)
	if err != nil {
		return err
	}

	defer pq.Close()

	replaceAsterisksInStrings(params)

	pqu := pq.Unsafe()

	err = pqu.Select(ret, params)
	if err != nil {
		return err
	}

	return nil
}

func SelectTypedWithTotal[T any](db types.Db, sql string, params map[string]interface{}, ret *T, total *uint32) error {
	// main query
	err := SelectTyped(db, sql, params, ret)
	if err != nil {
		return err
	}

	// get total amount of records
	sqlTotal := GetTotalSelectQuery(sql)
	totalData := []uint32{0}
	err = SelectTyped(db, sqlTotal, params, &totalData)
	if err != nil {
		return err
	}
	*total = totalData[0]

	return nil
}

// Returns a query to get total amount of records for original query
// ( Replaces fields list to "count(0) cnt" and removes limit|offset
func GetTotalSelectQuery(sql string) string {
	r := regexp.MustCompile("(?i)from")
	loc := r.FindStringIndex(sql)
	if loc == nil {
		return sql
	}

	fromIndex := loc[1]
	toIndex := len(sql)

	// remove offset if is
	r = regexp.MustCompile("(?i)(limit|offset).*$")
	loc = r.FindStringIndex(sql)
	if loc != nil {
		toIndex = loc[0]
	}

	return "select count(0) cnt from" + sql[fromIndex:toIndex]
}

func replaceAsterisksInStrings(params map[string]interface{}) {
	for key, value := range params {
		switch v := value.(type) {
		case string:
			params[key] = strings.Replace(v, "*", "%", -1)
		}
	}
}
