// main.go .

package main

import (
	"encoding/json"
	"net/http"
)

//url 요청을 mux에 등록하고 함수를 매칭하는 함수
func RegistHandleFunc(pattern string, handler func(http.ResponseWriter, *http.Request)) {
	if mux == nil {
		mux = http.NewServeMux()
	}
	//url패턴과 매핑할 handler 메서드를 매개변수로 둘을 매칭
	mux.HandleFunc(pattern, handler)
}

//핸들러 에러 발생 시 처리
func sendError(writer http.ResponseWriter, request *http.Request, httpStatus int, resultCode int32) {

	var result WCResult
	result.Result.ResultCode = resultCode

	responseData, _ := json.Marshal(result)

	writer.Header().Set("Content-type", "application/json")
	writer.WriteHeader(httpStatus)
	writer.Write(responseData)
}

//결과 객체를 response에 setting한다
func sendResultEx(writer http.ResponseWriter, request *http.Request, resultCode int, data []byte) {

	if resultCode == 0 { // TODO 정의 되지 않은 HTTP Status code가 넘어오는 경우 체크 필요.
		resultCode = 500
	}
	writer.Header().Set("Content-type", "application/json")
	writer.WriteHeader(resultCode)
	writer.Write(data)
}
