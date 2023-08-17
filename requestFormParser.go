package main

import (
	"errors"
	"net/http"
	"strconv"
)

//FormParser ...
type FormParser struct {
	request   *http.Request
	prefix    []string
	prefixLen int
}

func initFormParser(request *http.Request) (parser *FormParser) {

	request.ParseForm()

	if len(request.Form) == 0 {
		return nil
	}

	parser = &FormParser{
		request: request,
	}

	parser.prefix = request.Form["@d#"]
	parser.prefixLen = len(parser.prefix)
	return
}

func (parser *FormParser) getValueCount(index int, keyword string) (count int, err error) {
	if parser.prefixLen != 0 && index >= parser.prefixLen {
		err = errors.New("prefix index out of range")
		return
	}
	var key string
	if parser.prefixLen > 0 {
		key = parser.prefix[index] + keyword
	} else {
		key = keyword
	}

	count = len(parser.request.Form[key])
	return
}

func (parser *FormParser) getInt32Value(index int, keyword string, subIndex int) (value int32, err error) {
	if parser.prefixLen != 0 && index >= parser.prefixLen {
		err = errors.New("prefix index out of range")
		return
	}
	var key string
	if parser.prefixLen > 0 {
		key = parser.prefix[index] + keyword
	} else {
		key = keyword
	}

	if len(parser.request.Form[key]) > 0 {
		var temp int64
		temp, err = strconv.ParseInt(parser.request.Form[key][subIndex], 10, 32)
		if err != nil {
			return
		}
		value = int32(temp)
	}
	return
}

func (parser *FormParser) getInt64Value(index int, keyword string, subIndex int) (value int64, err error) {
	if parser.prefixLen != 0 && index >= parser.prefixLen {
		err = errors.New("prefix index out of range")
		return
	}
	var key string
	if parser.prefixLen > 0 {
		key = parser.prefix[index] + keyword
	} else {
		key = keyword
	}

	if len(parser.request.Form[key]) > 0 {
		value, err = strconv.ParseInt(parser.request.Form[key][subIndex], 10, 64)
		if err != nil {
			return
		}
	}
	return
}

func (parser *FormParser) getStringArray(index int, keyword string) (values []string, err error) {
	if parser.prefixLen != 0 && index >= parser.prefixLen {
		err = errors.New("prefix index out of range")
		return
	}
	var key string
	if parser.prefixLen > 0 {
		key = parser.prefix[index] + keyword
	} else {
		key = keyword
	}

	if len(parser.request.Form[key]) > 0 {
		return parser.request.Form[key], nil
	}
	return
}

func (parser *FormParser) getStringValue(index int, keyword string, subIndex int) (value string, err error) {
	if parser.prefixLen != 0 && index >= parser.prefixLen {
		err = errors.New("prefix index out of range")
		return
	}
	var key string
	if parser.prefixLen > 0 {
		key = parser.prefix[index] + keyword
	} else {
		key = keyword
	}

	if len(parser.request.Form[key]) > 0 {
		value = parser.request.Form[key][subIndex]
	}
	return
}
