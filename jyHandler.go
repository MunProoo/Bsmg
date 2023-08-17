package main

//업무 관리
func jyHandler() {
	RegistHandleFunc("/bsmg/login/", bsmgLoginHandler)
	RegistHandleFunc("/bsmg/user/", bsmgUserHandler)
	RegistHandleFunc("/bsmg/setting/", bsmgSettingHandler)
	RegistHandleFunc("/bsmg/report/", bsmgReportHandler)
}
