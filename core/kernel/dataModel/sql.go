package dataModel

import (
	"errors"
	"github.com/kico0909/cgo/core/mysql"
	"log"
	"reflect"
	"strconv"
	"strings"
)

type SQL struct {
	Conn *mysql.DatabaseMysql
}

var startSplit = "__cgo__"

func NewSQL(conn *mysql.DatabaseMysql) *SQL {
	return &SQL{conn}
}

// 数据库语句模型
type sqlStrs struct {
	table                              string
	sqlMode                            string
	where, value, limit, orderBy, keys string
	insertValues                       string
	joinTable                          []string
	joinType                           string
	ignores                            []string
}

func (s *sqlStrs) ToSqlCommand() string {
	var res []string

	switch s.sqlMode {

	case "select":
		if len(s.keys) < 1 {
			s.keys = "*"
		}
		res = append(res, "select", strings.Join(strings.Split(s.keys, startSplit), ","),"from", s.table )
		if len(s.joinType) > 0 {
			res = append(res, s.joinType, strings.Join(s.joinTable, ","))
		}
		break

	case "insert":
		res = append(res, "insert into", s.table, "(", s.keys, ")", "values", s.insertValues)
		break

	case "replace":
		res = append(res, "replace into", s.table, "(", s.keys, ")", "values", s.insertValues)
		break

	case "delete":
		res = append(res, "delete from", s.table)
		break

	case "update":
		res = append(res, "update", s.table, "set", s.value)
		break

	}

	res = append(res, s.where, s.orderBy, s.limit)

	return strings.Join(res, " ") + ";"
}

type DataModle struct {
	conn        *mysql.DatabaseMysql
	tablename   string
	tablestruct interface{}
	sql         sqlStrs
}

func getStructTypeName(ts interface{}) string {
	name := strings.Split(reflect.TypeOf(ts).String(), ".")
	if len(name) < 2 {
		return strings.ToLower(name[0])
	}
	return strings.ToLower(name[len(name)-1])
}

// 注册数据模型
func (d *SQL) RegModel(ts interface{}) *DataModle {
	name := getStructTypeName(ts)
	o := &DataModle{tablename: name, tablestruct: ts, conn: d.Conn, sql: sqlStrs{table: name}}
	return o
}

// 还原数据语句结构
func (t *DataModle) sqlReset() {
	t.sql = sqlStrs{table: t.tablename}
}

// where 条件的设置
func (t *DataModle) Where(command string, tm ...string) *DataModle {
	t.sql.where = "where "
	tc := strings.Split(command, "?")
	if len(tm) < 1 {
		t.sql.where = command
		goto RETURN
	}
	for i, v := range tc {
		val := ""
		if i < len(tm) {
			val = tm[i]
		}
		t.sql.where = t.sql.where + v + val
	}
RETURN:
	return t
}

// 值的写入
func (t *DataModle) Values(v ...string) *DataModle {
	t.sql.value = strings.Join(v, startSplit)
	return t
}

// 忽略值 - 用于插入等操作
func (t *DataModle) IgnoreKeys(v ...string) *DataModle {
	for n, k := range v {
		v[n] = strings.ToLower(k)
	}
	t.sql.ignores = v
	return t
}

// 定义输出KEY
func (t *DataModle) Keys(key ...string) *DataModle {
	t.sql.keys = "`" + strings.Join(key, "`"+startSplit+"`") + "`"
	return t
}

// 排序规则
func (t *DataModle) OrderBy(ob string) *DataModle {
	t.sql.orderBy = "order by " + ob
	return t
}

// 查询条目
func (t *DataModle) Get(v interface{}, num ...int64) error {
	var start string
	var length string

	if len(num) < 1 {
		t.sql.limit = ""
		goto JUMP
	}
	if len(num) == 1 {
		start = "0"
		length = strconv.FormatInt(num[0], 10)
	}
	if len(num) > 1 {
		start = strconv.FormatInt(num[0], 10)
		length = strconv.FormatInt(num[1], 10)
	}

	t.sql.limit = "limit " + start + "," + length

JUMP:
	t.sql.sqlMode = "select"

	err := t.query(v)
	if err != nil {
		return err
	}
	return nil
}

// 删除条目
func (t *DataModle) Delete() (int64, int64, error) {
	t.sql.sqlMode = "delete"
	return t.exec()
}

// 更新条目
func (t *DataModle) Update() (int64, int64, error) {
	t.sql.sqlMode = "update"
	var tmpk []string
	var tmpV []string
	if len(t.sql.keys) > 0 {
		tmpk = strings.Split(t.sql.keys, startSplit)
		tmpV = strings.Split(t.sql.value, startSplit)
	}
	var res []string
	if len(tmpk) != len(tmpV) {
		return 0, 0, errors.New("key and value length Mismatch!")
	}
	for i, v := range tmpk {
		res = append(res, v+"="+tmpV[i])
	}
	t.sql.value = strings.Join(res, ",")
	return t.exec()
}

// 新增条目
func (t *DataModle) Insert(vv interface{}) (int64, int64, error) {
	if vv == nil {
		return 0, 0, errors.New("none data insert")
	}
	t.sql.sqlMode = "insert"
	ks, vs, _ := handlerArgumentInterface(vv, t.sql.ignores)
	if len(ks) < 1 {
		t.sql.keys = ""
	} else {
		t.sql.keys = "`" + strings.Join(ks, "`,`") + "`"
	}
	t.sql.insertValues = strings.Join(vs, ",")
	return t.exec()
}

// 覆盖新增条目
func (t *DataModle) Replace(vv interface{}) (int64, int64, error) {
	if vv == nil {
		return 0, 0, errors.New("none data insert")
	}
	t.sql.sqlMode = "replace"
	ks, vs, err := handlerArgumentInterface(vv, t.sql.ignores)
	if err != nil {
		return 0, 0, err
	}
	t.sql.keys = strings.Join(ks, ",")
	t.sql.insertValues = strings.Join(vs, ",")
	return t.exec()
}

// 连表
func (t *DataModle) Join(table interface{}, name ...string) *DataModle {
	tableName := getStructTypeName(table)
	if len(name) > 0 {
		tableName = tableName + " " + name[0]
	}
	t.sql.joinTable = append(t.sql.joinTable, tableName)
	t.sql.joinType = "join"
	return t
}
func (t *DataModle) JoinLeft(table interface{}, name ...string) *DataModle {
	tableName := getStructTypeName(table)
	if len(name) > 0 {
		tableName = tableName + " " + name[0]
	}
	t.sql.joinTable = append(t.sql.joinTable, tableName)
	t.sql.joinType = "left join"
	return t
}
func (t *DataModle) JoinRight(table interface{}, name ...string) *DataModle {
	tableName := getStructTypeName(table)
	if len(name) > 0 {
		tableName = tableName + " " + name[0]
	}
	t.sql.joinTable = append(t.sql.joinTable, tableName)
	t.sql.joinType = "right join"
	return t
}

// 整理struct
func handlerArgumentInterface(data interface{}, ignores []string) (keys, vals []string, err error) {
	var k reflect.Type
	var v []reflect.Value

	ts := reflect.TypeOf(data)
	vs := reflect.ValueOf(data)

	switch ts.Kind() {
	case reflect.Slice:
		k = ts.Elem()
		for i := 0; i < vs.Len(); i++ {
			v = append(v, vs.Index(i))
		}
		break

	case reflect.Struct:
		k = ts
		v = append(v, vs)
		break

	default:
		return keys, vals, errors.New("argument type must struct or slice")
	}

	// 整理key
	for i := 0; i < k.NumField(); i++ {
		keys = append(keys, k.Field(i).Name)
	}

	// 整理忽略的key 及 记录忽略的key 下标
	for i := 0; i < len(keys); i++ {
		for _, v := range ignores {
			if strings.ToLower(keys[i]) == v {
				keys = append(keys[:i], keys[i+1:]...)
			}
		}
	}

	// 整理value
	for i := 0; i < len(v); i++ {
		var tmpVS []string
		for _, name := range keys {
			tmp_v := v[i].FieldByName(name)
			switch tmp_v.Kind() {
			case reflect.Int64, reflect.Int32, reflect.Int16, reflect.Int8:
				tmpVS = append(tmpVS, strconv.FormatInt(tmp_v.Int(), 10))
				break
			case reflect.Float64, reflect.Float32:
				tmpVS = append(tmpVS, strconv.FormatFloat(tmp_v.Float(), 'E', -1, 10))
				break
			case reflect.String:
				tmpVS = append(tmpVS, "'"+tmp_v.String()+"'")
				break
			}
		}
		vals = append(vals, "("+strings.Join(tmpVS, ",")+")")
	}

	// 进行忽略数据的设置设置
	return keys, vals, nil
}

// sql 执行
func (t *DataModle) query(v interface{}) error {
	log.Print("sql command line [query] => ", t.sql.ToSqlCommand())
	command := t.sql.ToSqlCommand()
	t.sqlReset()
	return t.conn.Query(v, command)
}
func (t *DataModle) exec(args ...interface{}) (int64, int64, error) {
	log.Print("sql command line [exec] => ", t.sql.ToSqlCommand())
	command := t.sql.ToSqlCommand()
	t.sqlReset()
	res, err := t.conn.Exec(command, args...)
	if err != nil {
		return 0, 0, err
	}
	lastInsertId, err := res.LastInsertId()
	if err != nil {
		return 0, 0, err
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return lastInsertId, 0, err
	}
	return lastInsertId, rowsAffected, nil
}
