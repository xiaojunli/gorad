package gorad

import (
	"strconv"
			"mime/multipart"
				"net/url"
	"github.com/gin-gonic/gin"
	"net/http"
			)

/**
 * 封装Form, PostForm, Param参数
 */
type Query struct {
	Params 		*gin.Params
	Form 		*url.Values
	PostForm 	*url.Values
	MultipartForm	*multipart.Form
}


//////////////////////////////////////////////////////////

/**
 * 获取值
 */
func (self *Query) Get(key string) (result string, err error){
	return self.getSingleValueString(key)
}

/**
 * 获取单一值
 */
func (self *Query) getSingleValueString(key string) (string, error){
	//Form中取值
	var resultValue interface{}

	//先在路径中获取
	var found bool
	if nil != self.Params {
		resultValue, found = self.Params.Get(key)
	}

	//Params中查找不到时查找表单
	if  !found && nil != self.Form {
		resultValue = self.Form.Get(key)
		if "" != resultValue {
			found = true
		}
	}

	//Form中查找不到时查找 PostForm
	if !found && nil != self.PostForm {
		resultValue = self.PostForm.Get(key)
	}

	return resultValue.(string), nil
}

/**
 * 获取字符串
 */
func (self *Query) getString(key string) (result string, err error){

	return self.getSingleValueString(key)
}

/**
 * 获取值取默认值
 */
func (self *Query) MustString(key, defaultValue string) (result string){
	value, err := self.getString(key)

	if nil != err {
		return defaultValue
	}

	return value
}

/**
 * 获取整形值默认值
 */
func (self *Query) MustInt(key string, defaultValue int) (result int){
	value, err := self.getSingleValueString(key)

	if nil != err {
		result = defaultValue
	}

	result, err = strconv.Atoi(value)

	if nil != err {
		result = defaultValue
	}

	return result
}

//////////////////////////////////////////////////

func (self *Query) GetInt64(key string) (result int64, err error) {
	valueStr, err := self.getSingleValueString(key)

	if nil != err {
		return 0, err
	}

	return strconv.ParseInt(valueStr, 10, 8)
}

/**
 * 获取64位整形
 */
func (self *Query) MustInt64(key string, defaultValue int64) (int64) {
	value, err := self.GetInt64(key)

	if nil != err {
		return defaultValue
	}

	return value
}

func (self *Query) MustInt8(key string, defaultValue int8) (int8){
	value, err := self.GetInt64(key)

	if nil != err {
		return defaultValue
	}

	return int8(value)
}

func (self *Query) MustInt16(key string, defaultValue int16) (int16){
	value, err := self.GetInt64(key)

	if nil != err {
		return defaultValue
	}

	return int16(value)
}

func (self *Query) MustInt32(key string, defaultValue int32) (int32){
	value, err := self.GetInt64(key)

	if nil != err {
		return defaultValue
	}

	return int32(value)
}

/////////////////////////////////////////////////////

func (self *Query) GetUint64(key string) (result uint64, err error) {
	valueStr, err := self.getSingleValueString(key)

	if nil != err {
		return 0, err
	}

	return strconv.ParseUint(valueStr, 10, 8)
}

/**
 * 获取uint值
 */
func (self *Query) MustUint(key string, defaultValue uint) (uint) {
	value := self.MustUint64(key, uint64(defaultValue))

	return  uint(value)
}

/**
 * 获取64位整形
 */
func (self *Query) MustUint64(key string, defaultValue uint64) (uint64) {
	value, err := self.GetUint64(key)

	if nil != err {
		return defaultValue
	}

	return value
}

func (self *Query) MustUint8(key string, defaultValue uint8) (uint8){
	value, err := self.GetUint64(key)

	if nil != err {
		return defaultValue
	}

	return uint8(value)
}

func (self *Query) MustUint16(key string, defaultValue uint16) (uint16){
	value, err := self.GetUint64(key)

	if nil != err {
		return defaultValue
	}

	return uint16(value)
}

func (self *Query) MustUint32(key string, defaultValue uint32) (uint32){
	value, err := self.GetUint64(key)

	if nil != err {
		return defaultValue
	}

	return uint32(value)
}
//////////////////////////////////////////////////

/**
 * 获取浮点值
 */
func (self *Query) GetFloat(key string) (result float64, err error) {
	return self.GetFloat64(key)
}

/**
 * 获取64位浮点数值
 */
func (self *Query) GetFloat64(key string) (result float64, err error) {
	valueStr, err := self.getSingleValueString(key)

	if nil != err {
		return 0, err
	}


	return strconv.ParseFloat(valueStr, 64)
}

/**
 * 获取64位服务
 */
func (self *Query) MustFloat(key string, defaultValue float64) (float64) {
	return self.MustFloat64(key, defaultValue);
}

/**
 * 获取64位浮点
 */
func (self *Query) MustFloat64(key string, defaultValue float64) (float64) {
	value, err := self.GetFloat64(key)

	if nil != err {
		return defaultValue
	}

	return value
}

func (self *Query) MustFloat32(key string, defaultValue float32) (float32) {
	value, err := self.GetFloat64(key)

	if nil != err {
		return defaultValue
	}

	return float32(value)
}

//////////////////////////////////////////////////////////

/**
 * 获取bool值
 */
func (self *Query) GetBool(key string) (bool, error) {
	valueStr, err := self.getSingleValueString(key)

	if nil != err {
		return false, err
	}

	return strconv.ParseBool(valueStr)
}

/**
 * 获取bool值
 */
func (self *Query) MustGetBool(key string, defaultValue bool) (bool) {
	result, err := self.GetBool(key)

	if nil != err {
		return defaultValue
	}

	return result
}

/////////////////////////////////

/**
 * 获取文件
 */
func (self *Query) GetFile(key string, req *http.Request) (multipart.File, *multipart.FileHeader, error) {
	return req.FormFile(key)
}

/**
 * 获取多个文件
 */
func (self *Query) GetFiles(key string, req *http.Request) ([]*multipart.FileHeader, error) {
	if files, ok := req.MultipartForm.File[key]; ok {
		return files, nil
	}
	return nil, http.ErrMissingFile
}