package scripts

import (
	"net/url"
	"strings"
)

var httpMth = []string{"OPTIONS", "GET", "HEAD", "POST", "PUT", "DELETE", "TRACE", "CONNECT"}
var fmtOpt = []string{"RAW", "STD"}

//ValidateHTTPMethod checks httpMethod
func ValidateHTTPMethod(m string) bool {
	is := StringInSlice(m, httpMth)
	if is == false {
		Logger.Info("Incorect HTTP Method: ", m)
		return false
	}
	return true
}

//ValidateFmt checks dataFormat
func ValidateFmt(f string) bool {
	is := StringInSlice(f, fmtOpt)
	if is == false {
		Logger.Info("Incorect Format ", f)
		return false
	}
	return true
}

//ValidateURL check URL
func ValidateURL(u string) bool {
	_, err := url.ParseRequestURI(u)
	if err != nil {
		Logger.Info("Wrong URL: ", u)
		return false
	}
	return true
}

//IsSet check if string or []string is set
func IsSet(i interface{}) bool {
	switch v := i.(type) {
	case string:
		if strings.TrimSpace(v) != "" {
			return true
		}
		return false
	case []string:
		if len(v) != 0 {
			return true
		}
		return false
	default:
		panic("not supported type")
	}
}

//StringInSlice checks if
func StringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}
