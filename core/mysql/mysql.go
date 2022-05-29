package mysql

/*
 Cgo的mysql方法实现, 可读写分离, 读库要使用Query方法,写库要使用Exec方法, 否则会导致读写错乱
*/

import (
	_ "github.com/kico0909/cgo/lib/mysql"
	"github.com/kico0909/cgo/core/kernel/config"
	"github.com/kico0909/cgo/core/kernel/logger"
	"database/sql"
	"encoding/json"
	"errors"
	"os"
	"strconv"
)

// 数据库链接信息
type dbConnectionInfoType struct {
	config.MysqlSetOpt
}

// 数据库分类
type dbConnectionsInfoType struct {
	r     dbConnectionInfoType
	w     dbConnectionInfoType
	links map[string]dbConnectionInfoType
}

// 链接分类
type connType struct {
	r *sql.DB
	w *sql.DB
}

type DatabaseMysql struct {
	sqlmode           string // 数据库模式: 单库, 读写分离
	dbConnectionsInfo dbConnectionsInfoType
	conn              connType
	Links             map[string]*DatabaseMysql // 子数据链接
}

func createConnectionInfo(conf dbConnectionInfoType) string {

	if len(conf.Charset) < 1 {
		conf.Charset = "utf8"
	}

	//  链接写库
	_, err := os.Stat(conf.Socket)
	// 存在套字链接的路径, 优先使用套子链接
	if err == nil {
		return conf.Username + `:` + conf.Password + `@unix(` + conf.Socket + `)/` + conf.Dbname + "?charset=" + conf.Charset
	} else {
		if (conf.Host == "localhost" || conf.Host == "127.0.0.1") && conf.Port == 3306 {
			return conf.Username + `:` + conf.Password + `@/` + conf.Dbname + "?charset=" + conf.Charset
		} else {
			return conf.Username + `:` + conf.Password + `@tcp(` + conf.Host + `:` + strconv.FormatInt(conf.Port, 10) + `)/` + conf.Dbname + "?charset=" + conf.Charset
		}
	}
}

// 连接数据库
func (_self *DatabaseMysql) connectionDB(open, idle int, linkname string) *DatabaseMysql {

	// 连接写库
	_dbw, _err := sql.Open("mysql", createConnectionInfo(_self.dbConnectionsInfo.w))
	if _err != nil {
		log.Fatalln("数据库连接["+linkname+"]出现错误: ", _err)
	}
	// 最大连接
	_dbw.SetMaxOpenConns(open)

	// 保持连接
	_dbw.SetMaxIdleConns(idle)

	dbPing_w := _dbw.Ping()
	if dbPing_w != nil {
		log.Fatalln("数据库连接["+linkname+"]无法Ping通: ", dbPing_w)
	}

	_self.conn.w = _dbw

	if _self.sqlmode == "default" {
		_self.conn.r = _dbw
		return _self
	}

	// 连接读库
	_dbr, _err := sql.Open("mysql", createConnectionInfo(_self.dbConnectionsInfo.r))

	if _err != nil {
		log.Fatalln("数据库连接["+linkname+"]出现错误: ", _err)
	}

	// 最大连接
	_dbr.SetMaxOpenConns(open)

	// 保持连接
	_dbr.SetMaxIdleConns(idle)

	dbPing := _dbr.Ping()
	if dbPing != nil {
		log.Fatalln("数据库连接["+linkname+"]无法Ping通: ", dbPing)
	}

	_self.conn.r = _dbr

	return _self

}

/*
私有方法, 用于关闭数据库
*/
func (_self *DatabaseMysql) closeDB() {
	_self.conn.w.Close()
	_self.conn.r.Close()
}

// 根据连接信息 初始化数据库
func New(conf *config.ConfigMysqlOptions, childName interface{}, recall func()) *DatabaseMysql {
	// 连接信息生成
	var wDBinfo dbConnectionInfoType
	var rDBinfo dbConnectionInfoType

	// 模式判断
	var sqlMode = "default"

	if childName == nil {
		if len(conf.Default.Host) < 1 {
			sqlMode = "rw"
		}
	} else {
		sqlMode = "childs"
	}
	switch sqlMode {
	case "default":
		wDBinfo = dbConnectionInfoType{conf.Default}
		rDBinfo = dbConnectionInfoType{conf.Default}
		break
	case "rw":
		wDBinfo = dbConnectionInfoType{conf.Write}
		rDBinfo = dbConnectionInfoType{conf.Read}
		break
	}

	// 其他数据库连接
	moreLinks := make(map[string]dbConnectionInfoType)
	for k, v := range conf.Childs {
		moreLinks[k] = dbConnectionInfoType{*v}
	}

	links := make(map[string]*DatabaseMysql)
	for k, v := range moreLinks {
		tmp := &DatabaseMysql{sqlmode: sqlMode, dbConnectionsInfo: dbConnectionsInfoType{w: v, r: v}, Links: nil}
		links[k] = tmp.connectionDB(conf.MaxOpen, conf.MaxIdle, k)
	}

	// 创建实例
	tmp := &DatabaseMysql{sqlmode: sqlMode, dbConnectionsInfo: dbConnectionsInfoType{w: wDBinfo, r: rDBinfo}, Links: links}
	recall()
	return tmp.connectionDB(conf.MaxOpen, conf.MaxIdle, sqlMode)
}

// 数据库查询操作
func (_self *DatabaseMysql) Query(v interface{}, sql string) (err error) {

	rows, err := _self.conn.r.Query(sql)
	if err != nil {
		return errors.New("sql query error[" + err.Error() + "]")
	}

	defer rows.Close()

	//读出查询出的列字段名
	cols, _ := rows.Columns()
	colsTypes, _ := rows.ColumnTypes()

	//values是每个列的值，这里获取到byte里
	values := make([][]byte, len(cols))

	//query.Scan的参数，因为每次查询出来的列是不定长的，用len(cols)确定当次查询的长度
	scans := make([]interface{}, len(cols))

	//让每一行数据都填充到[][]byte里面
	for i := range values {
		scans[i] = &values[i]
	}

	var results []map[string]interface{}
	for rows.Next() { //循环，让游标往下推
		if err := rows.Scan(scans...); err != nil { //query.Scan查询出来的不定长值放到scans[i] = &values[i],也就是每行都放在values里
			return errors.New("MYSQL rows.Scan ERROR=>" + err.Error())
		}

		row := make(map[string]interface{}) //每行数据
		for i, v := range values {          //每行数据是放在values里面，现在把它挪到row里
			switch colsTypes[i].ScanType().Name() {
			case "uint64", "int64", "NullInt64":
				row[cols[i]], _ = strconv.ParseInt(string(v), 10, 64)
				break
			case "uint32", "int32", "NullInt32":
				row[cols[i]], _ = strconv.ParseInt(string(v), 10, 32)
				break
			case "uint16", "int16", "NullInt16":
				row[cols[i]], _ = strconv.ParseInt(string(v), 10, 16)
				break
			case "uint8", "int8", "NullInt8":
				row[cols[i]], _ = strconv.ParseInt(string(v), 10, 8)
				break

			case "RawBytes", "NullTime":
				row[cols[i]] = string(v)
				break
			default:
				log.Println("位置sql返回判断类型:", cols[i], "==>", colsTypes[i].ScanType().Name())
			}

		}
		results = append(results, row)
	}
	b, err := json.Marshal(results)
	if err != nil {
		return err
	}
	err = json.Unmarshal(b, v)
	if err != nil {
		return err
	}
	return nil
}

// 非查询类数据库操作
func (_self *DatabaseMysql) Exec(query string, args ...interface{}) (sql.Result, error) {
	return _self.conn.w.Exec(query, args...)
}
