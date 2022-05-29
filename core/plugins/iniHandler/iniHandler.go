package iniHandler

import (
	"bufio"
	"encoding/json"
	"errors"
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"
)

var (
	exp_RemLine    = regexp.MustCompile(`^[\s|#].*`)
	exp_Rem        = regexp.MustCompile(`#.*$`)
	exp_rn         = regexp.MustCompile(`[\n|\t|\r|\s]*`)
	exp_AreaKey    = regexp.MustCompile(`^\[.*\]$`)
	exp_AreaKeyTag = regexp.MustCompile(`\[|\]`)
	exp_Keyname    = regexp.MustCompile(`^[a-z|A-Z]{1}[a-z|A-Z|0-9]*`) // 首字符必须是字母

	// 数据转换
	exp_Int         = regexp.MustCompile(`^\d*$`)     // 全部是数字
	exp_Float       = regexp.MustCompile(`^\d*.\d*$`) // 全部是数字
	exp_forceString = regexp.MustCompile(`^".*"$`)    // 强转字符串类型
)

type handlerStrarrToStructChip struct {
	KeyName string                               `json:"key_name"` // 当前数据类型
	Values  interface{}                          `json:"values"`   // 当前数据的值
	Child   map[string]handlerStrarrToStructChip `json:"child"`    // 当前数据子级
}

type jsonDataType map[string]interface{}

func InitFile(path string, v interface{}) error {
	if len(path) < 1 {
		return errors.New("无效的path")
	}

	baseData, err := loadConfig_ini(path)
	if err != nil {
		return err
	}
	processData := handleStringForStruct(baseData)

	jsonData := make(jsonDataType)
	handlerDataConvertInterface(&processData, &jsonData)

	b, err := json.Marshal(jsonData)
	if err != nil {
		return err
	}

	err = json.Unmarshal(b, v)
	if err != nil {
		return err
	}
	b, _ = json.Marshal(v)
	return nil
}

// 读取INI 并初步整理 -> 去空白,去注释 -> 标准化数据
func loadConfig_ini(path string) ([]string, error) {

	var res []string

	f, err := os.Open(path)

	if err != nil {
		return res, err
	}
	defer f.Close()

	b := bufio.NewReader(f)
	for {
		d, _, msg := b.ReadLine()
		if msg == io.EOF {
			break
		}
		tmp := exp_rn.ReplaceAllString(string(d), "")
		if !exp_RemLine.MatchString(tmp) {
			tmp = exp_Rem.ReplaceAllString(tmp, "")
			res = append(res, tmp)
		}
	}
	return res, nil
}

// 整理读取的标准化数据
func handleStringForStruct(d []string) map[string]handlerStrarrToStructChip {
	res := make(map[string]handlerStrarrToStructChip)

	var nowKeyName string
	for _, v := range d {
		if exp_AreaKey.MatchString(v) { // 判断是否是JSON的 object key
			nowKeyName = exp_AreaKeyTag.ReplaceAllString(v, "")
			res[nowKeyName] = handlerStrarrToStructChip{KeyName: nowKeyName, Child: make(map[string]handlerStrarrToStructChip)}
		} else {
			tmp := strings.Split(v, "=")
			if len(tmp) == 2 {
				if res[nowKeyName].Child != nil {
					res[nowKeyName].Child[tmp[0]] = handlerStrarrToStructChip{KeyName: tmp[0], Values: handlerDataType(tmp[1])}
				} else {
					res[tmp[0]] = handlerStrarrToStructChip{KeyName: tmp[0], Values: handlerDataType(tmp[1])}
				}
			}
		}
	}
	return res
}

// 整理初步整理后的INI 数据, 返回可json化的 interface{} 类型数据
func handlerDataConvertInterface(d *map[string]handlerStrarrToStructChip, data *jsonDataType, keyName ...string) {

	if len(keyName) < 1 { // 未传参
		for k, v := range *d {
			kn := strings.Split(k, "-")

			if v.Values == nil { // 不是 key-value 键值对 => 嵌套的object
				if (*data)[kn[0]] == nil {
					(*data)[kn[0]] = make(jsonDataType)
				}
				tmp := (*data)[kn[0]].(jsonDataType)
				handlerDataConvertInterface(&v.Child, &tmp, kn[1:]...)
			} else {
				kk := strings.Split(k, ":")
				if len(kk) < 2 {
					(*data)[k] = v.Values
				} else {
					switch strings.ToLower(kk[1]) {
					case "array":
						(*data)[kk[0]] = strings.Split(v.Values.(string), ",")
						break
					case "int64":
						(*data)[kk[0]], _ = v.Values.(int64)
						break
					case "int32":
						(*data)[kk[0]], _ = v.Values.(int32)
						break
					case "int16":
						(*data)[kk[0]], _ = v.Values.(int16)
						break
					case "int":
						(*data)[kk[0]], _ = v.Values.(int)
						break
					case "float64":
						(*data)[kk[0]], _ = strconv.ParseFloat(v.Values.(string), 64)
						break
					case "float32":
						(*data)[kk[0]], _ = strconv.ParseFloat(v.Values.(string), 32)
						break
					case "float16":
						(*data)[kk[0]], _ = strconv.ParseFloat(v.Values.(string), 16)
						break
					case "bool":
						(*data)[kk[0]], _ = strconv.ParseBool(v.Values.(string))
						break
					}
				}
			}
		}

	} else {
		if (*data)[keyName[0]] == nil {
			(*data)[keyName[0]] = make(jsonDataType)
		}
		tmp := (*data)[keyName[0]].(jsonDataType)
		handlerDataConvertInterface(d, &tmp, keyName[1:]...)
	}
}

// 判断数据格式转换
func handlerDataType(d string) interface{} {
	// 判断强转字符串类型
	if exp_forceString.MatchString(d) {
		return d[1 : len(d)-1]
	}

	// 判断整形
	if exp_Int.MatchString(d) {
		dint, err := strconv.ParseInt(d, 10, 64)
		if err == nil {
			return dint
		}
	}

	// 判断浮点型
	if exp_Float.MatchString(d) {
		dfloat, err := strconv.ParseFloat(d, 10)
		if err == nil {
			return dfloat
		}
	}

	// boolen 类型
	if d == "true" {
		return true
	}
	if d == "false" {
		return false
	}

	// 非可解析类型 的 返回字符串
	return d
}
