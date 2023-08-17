// main.go에서 db 구조체, json 인코딩 부분

package main

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
)

//DB접속하기 위한 정보 구조체
type testConfig struct {
	DatabaseType     string
	DatabaseIP       string
	DatabasePort     int
	DatabaseID       string
	DataBasePW       string
	DataBaseName     string
	ServerListenPort int
}

//http 핸들러 구조체
type staticHandler struct {
	http.Handler
}

//에러처리를 위한 구조체
type WebErrorResult struct {
	httpStatus int
	resultCode int32
}

//DBconfig 파일(testConfig.json) 생성 또는 확인
func initConfig() (config *testConfig, err error) {
	curPath, err := filepath.Abs(filepath.Dir(os.Args[0]))

	if _, err := os.Stat(curPath + "/testConfig.json"); os.IsNotExist(err) {

		file, err := os.OpenFile(curPath+"/testConfig.json", os.O_CREATE|os.O_RDWR, 0666)
		if err != nil {
			myLogger.Print("Cannot create config file")
			panic(err)
		}

		encoder := json.NewEncoder(file)
		configTemp := &testConfig{
			DatabaseType:     "mssql",
			DatabaseIP:       "127.0.0.1", // "192.168.30.233"
			DatabasePort:     1433,
			DatabaseID:       "sa",
			DataBasePW:       "12345",
			DataBaseName:     "testdb",
			ServerListenPort: 3000,
		}

		err = encoder.Encode(configTemp)
		if err != nil {
			myLogger.Print("Cannot write configuration to file")
			panic(err)
		}
	}
	//textConfig 파일 열어서 객체 생성
	file, err := os.Open(curPath + "/testConfig.json")
	if err != nil {
		myLogger.Print("Cannot open config file")
		panic(err)
	}
	//file(testConfig) 객체를 json객체로 디코딩
	decoder := json.NewDecoder(file)
	//json객체로 만들어 config 변수에 할당
	config = &testConfig{}
	err = decoder.Decode(config)
	if err != nil {

		myLogger.Print("Cannot get configuration from file")
		panic(err)
	}

	myLogger.Print("config initialized success")

	return
}
