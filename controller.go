// main.go에서 로거 실행 부분

package main

import (
	"database/sql" // 데이타베이스를 연결하고, 쿼리를 실행하고, DML 명령을 수행
	"fmt"
	"log"
	"net/http"
	"os"

	_ "github.com/denisenkom/go-mssqldb" //mssql
)

var g_testConfig *testConfig = nil
var myLogger *log.Logger = nil
var mux *http.ServeMux = nil
var db *sql.DB = nil

//func main() {
func StartServer() {

	/*****************************************************/
	// log
	/****************************************************/
	fpLog, err := os.OpenFile("logfile.txt", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		panic(err)
	}
	defer fpLog.Close()

	myLogger = log.New(fpLog, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)

	/****************************************************/
	// config
	/****************************************************/
	// config 파일의 자료를 json 객체화
	g_testConfig, err = initConfig()

	/****************************************************/
	// db
	/****************************************************/
	//"server=127.0.0.1;user id=sa;password=12345;database=testtbl"
	//json화 된 config 객체를 읽어서 확인 및 db 접속
	strConnectString := fmt.Sprintf("server=%s;user id=%s;password=%s;database=%s",
		g_testConfig.DatabaseIP,
		g_testConfig.DatabaseID,
		g_testConfig.DataBasePW,
		g_testConfig.DataBaseName)

	db, err = sql.Open(g_testConfig.DatabaseType, strConnectString)
	if err != nil {
		myLogger.Print("Cannot open sql")
		panic(err)
	}

	defer db.Close()

	/*****************************************************/
	// web server

	strServerAddress := fmt.Sprintf(":%d", g_testConfig.ServerListenPort)
	//mux(multiplexor) - 요청(request)URL을 받아 함수를 연결해주는 객체

	mux = http.NewServeMux()
	//기본 물리적 파일 경로 설정 - 요청(url)이 들어오면 해당 경로로 이동
	fileServer := http.FileServer(http.Dir("./webRoot")) //webRoot 폴더..
	//url을 식별
	/****************************************************/
	// ":3000"
	//서버 주소에 포트주소 붙이기 -> :3000
	mux.Handle("/", http.StripPrefix("/", fileServer))
	// Then, initialize the session manager
	//url 명령어 등록 - 메서드 매칭!!
	//http.HandleFunc("/ws", WebSocketHandler)

	jyHandler()

	//TCP 네트워크 리스너 -> 서버호출
	http.ListenAndServe(strServerAddress, mux)

	myLogger.Println("End of Program")
}
