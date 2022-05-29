package log

import (
	"bufio"
	"io"
	olog "log"
	"os"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"
)

const (
	Hour          = 3600
	Minute        = 60
	Second        = 1
	OneDay        = 1 * 24 * 60 * 60
	LogNameSuffix = "_clog.log"
)

var std *olog.Logger

type Logger struct {
	filePath    string       // 日志文件路径
	mode        string       // 输出模式
	file        *os.File     // 输出到文件
	stopAutoCut bool         // 禁止自动截断日志文档
	logger      *olog.Logger //
	debugMode   bool         // DEBUG状态
	nextCutDate string       // 上一次切割的日期
	prefix      string       // 日志前缀命名
	path        string       // 日志保存路径
}

func New(path, prefix string, stopAutoCut bool) *Logger {

	var myLog *Logger
	var f *os.File
	var err error
	var mode = "Terminal"

	if len(path) > 0 {
		f, err = os.OpenFile(
			path+prefix+LogNameSuffix,
			os.O_CREATE|os.O_RDWR|os.O_APPEND,
			0777)
		if err != nil {
			Println("功能初始化: Cgo日志系统\t--- [ fail ] : 日志文件配置错误")
			os.Exit(0)
		}
		mode = "File"
		std = olog.New(f, "", olog.LstdFlags)
	}

	myLog = &Logger{
		filePath:    path + prefix + LogNameSuffix,
		file:        f,
		stopAutoCut: stopAutoCut,
		logger:      std,
		debugMode:   false,

		prefix: prefix,
		path:   path,
		mode:   mode}

	if !stopAutoCut {
		go myLog.autoCutOffLogFileHandler()
	}

	Println("功能初始化: Cgo日志系统\t--- [ ok ]")

	return myLog
}

// 设置日志的debug模式, 默认关闭
func (this *Logger) SetDebugMode(key bool) {
	this.debugMode = key
}

func (this *Logger) Print(v ...interface{}) {
	this.logger.Println(splice(v, 0, false, findFileInfos("Normal"))...)
}

func (this *Logger) Println(v ...interface{}) {
	this.logger.Print(splice(v, 0, false, findFileInfos("Normal"))...)
}

func (this *Logger) Info(v ...interface{}) {
	this.logger.Println(splice(v, 0, false, findFileInfos("Info"))...)
}

func (this *Logger) Warn(v ...interface{}) {
	this.logger.Println(splice(v, 0, false, findFileInfos("Warn"))...)
}

func (this *Logger) Error(v ...interface{}) {
	this.logger.Println(splice(v, 0, false, findFileInfos("Error"))...)
}

func (this *Logger) Fatal(v ...interface{}) {
	this.logger.Fatal(splice(v, 0, false, findFileInfos("Fatal"))...)
	os.Exit(1)
}

func (this *Logger) Fatalln(v ...interface{}) {
	this.logger.Fatalln(splice(v, 0, false, findFileInfos("Fatal"))...)
	os.Exit(1)
}

// debug模式下可以使用,设置为非debug 模式则不
func (this *Logger) Debug(v ...interface{}) {
	if !this.debugMode {
		return
	}
	this.logger.Println(splice(v, 0, false, findFileInfos("Debug"))...)
}

// 开一个定时线程执行文件分割, 按天执行
func (this *Logger) autoCutOffLogFileHandler() {
	if this.mode == "Terminal" {
		return
	}

	// 启动时先进行一次分割
	this.cunFile()
	this.Info("下次日志切割,将在", getSurplusSecond()+10*time.Second, "秒后")
	time.Sleep(getSurplusSecond() + 10*time.Second)
	this.autoCutOffLogFileHandler()
}

// 检测并切割文件
func (this *Logger) cunFile() {

	this.Info("STEP-1: 准备执行切割")

	f, err := os.OpenFile(
		this.filePath,
		os.O_CREATE|os.O_RDWR|os.O_APPEND,
		0777)

	if err != nil {
		this.Error("STEP-ERR: 读取日志文件错误 >", err)
	}

	finfo, _ := f.Stat()
	if finfo.Size() <= 0 { // 日志文件尺寸为0 直接跳过切割执行
		this.Info("STEP-END: 日志文件为空无需进行切割")
		return
	}

	// 当天和昨天的日志变量
	var logArr []string
	var newFileContent []string // 前一天日志文件

	// 前一天日志的正则
	regexpStr, _ := regexp.Compile("^" + yesterdayDate("/") + "[\\S|\\s]*$")

	logFileChip := bufio.NewReader(f)
	for {
		content, _, eof := logFileChip.ReadLine()
		if eof == io.EOF {
			break
		}

		if len(content) > 1 {
			if regexpStr.MatchString(string(content)) {
				newFileContent = append(newFileContent, string(content))
			} else {
				logArr = append(logArr, string(content))
			}
		}
	}

	// 写入日志分割
	if len(newFileContent) > 0 {
		if !saveCutoffLogFile(this, newFileContent) {
			return
		}
		if !saveRecreatedLogFile(this, logArr) {
			return
		}
	}
}

func saveCutoffLogFile(this *Logger, newFileByte []string) bool {

	f, err := os.OpenFile(this.path+yesterdayDate("-")+"_"+this.prefix+LogNameSuffix, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0664)
	defer f.Close()

	if err != nil {
		this.Error("Cgo LOG SYSTEM ==> Cgo log auto cut create new log file error!(" + err.Error() + ")")
		return false
	}

	_, err = f.Write([]byte(strings.Join(newFileByte, "\n") + "\n"))
	if err != nil {
		this.Error("Cgo LOG SYSTEM ==> Cgo log auto cut create new log file error!(" + err.Error() + ")")
		return false
	}

	return true
}

func saveRecreatedLogFile(this *Logger, logArr []string) bool {
	// 重构老日志
	f, err := os.OpenFile(this.path+this.prefix+LogNameSuffix, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0664)
	defer f.Close()

	if err != nil {
		this.Error("Cgo LOG SYSTEM ==> Cgo log recreate actived log file error!(" + err.Error() + ")")
		return false
	}

	_, err = f.Write([]byte(strings.Join(logArr, "\n") + "\n"))
	if err != nil {
		this.Error("Cgo LOG SYSTEM ==> Cgo log recreate actived log file error!(" + err.Error() + ")")
		return false
	}

	return true
}

func yesterdayDate(tag string) string {
	return time.Now().AddDate(0, 0, -1).Format("2006" + tag + "01" + tag + "02")
}

func getSurplusSecond() time.Duration {
	now := time.Now()
	//return 10 * time.Second
	return time.Duration(int64(OneDay-now.Hour()*Hour-now.Minute()*Minute-now.Second()*Second)) * time.Second
}

// 数组增删字符串
func splice(arr []interface{}, index int64, replace bool, insertValue interface{}) []interface{} {
	var res []interface{}

	if insertValue != nil {
		res = append(res, insertValue)
	}

	if replace {
		res = append(res, arr[index+1:]...)
	} else {
		res = append(res, arr[index:]...)
	}

	res = append(arr[:index], res...)

	return res

}

//
func Println(v ...interface{}) {
	std.Println(splice(v, 0, false, findFileInfos("Normal"))...)
}

func Info(v ...interface{}) {
	std.Println(splice(v, 0, false, findFileInfos("Info"))...)
}

func Warn(v ...interface{}) {
	std.Println(splice(v, 0, false, findFileInfos("Warn"))...)
}

func Error(v ...interface{}) {
	std.Println(splice(v, 0, false, findFileInfos("Error"))...)
}

// debug模式下可以使用,设置为非debug 模式则不
func Debug(v ...interface{}) {
	std.Println(splice(v, 0, false, findFileInfos("Debug"))...)
}

// debug模式下可以使用,设置为非debug 模式则不
func Fatalln(v ...interface{}) {
	std.Println(splice(v, 0, false, findFileInfos("Fatal"))...)
}

func Fatal(v ...interface{}) {
	std.Println(splice(v, 0, false, findFileInfos("Fatal"))...)
}

func init() {
	std = olog.New(os.Stderr, "", olog.LstdFlags)
}

func findFileInfos(mode string) string {
	_, b, c, d := runtime.Caller(2)
	if !d {
		return " "
	}
	sb := strings.Split(b, "/")

	return "[ " + mode + "( " + sb[len(sb)-1] + ":" + strconv.FormatInt(int64(c), 10) + " )] -> \b"
}
