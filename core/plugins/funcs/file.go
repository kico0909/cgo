/*
操作文件系统
*/
package funcs

import (
	"errors"
	"io"
	"io/ioutil"
	"mime/multipart"
	"os"
	"path/filepath"
	"runtime"
)

/*
获得当前的系统绝对路径
*/
func GetMyPath() (string, error) {

	switch runtime.GOOS {

	case "darwin", "windows":
		return os.Getwd()
		break

	case "linux":
		return filepath.Abs(filepath.Dir(os.Args[0]))
		break

	}
	return "", errors.New("不支持的操作系统! [" + runtime.GOOS + "]")
}

/*
读取指定文件
*/
func ReadFile(filePath string) ([]byte, error) {
	return ioutil.ReadFile(filePath)
}

/*
上传文件到指定路径
*/

func UploadFile(path string, fileContent multipart.File) (bool, error) {
	//写入文件
	file, err := os.Create(path)
	if err != nil {
		return false, err
	}

	defer fileContent.Close()
	defer file.Close()
	n, err := io.Copy(file, fileContent)

	return n > 0, nil
}
