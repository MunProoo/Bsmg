// +build windows
package main

///*
import (
	"flag"
	_ "fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/debug"
)

func main() {

	os.Chdir(filepath.Dir(os.Args[0])) // 작업 디렉토리 실행파일 디렉토리로 경로 수정
	const svcName = "DevToolService"

	isIntSess, err := svc.IsAnInteractiveSession()
	if err != nil {

	}
	if !isIntSess {
		runService(svcName, false)
		return
	}

	cmd := flag.String("rtype", "debug", "run type")
	flag.Parse()
	*cmd = strings.ToLower(*cmd)

	switch *cmd {
	case "debug":
		CreateCron() // 일일 업무 주간으로 취합
		runService(svcName, true)
		return
	case "install":
		//err = svr.InstallService(svcName, svcName)
	case "remove":
		//err = svr.RemoveService(svcName)
	case "start":
		//err = svr.StartService(svcName)
	case "stop":
		//err = svr.ControlService(svcName, svc.Stop, svc.Stopped)
	case "pause":
		//err = svr.ControlService(svcName, svc.Pause, svc.Paused)
	case "continue":
		//err = svr.ControlService(svcName, svc.Continue, svc.Running)
	default:

	}

	if err != nil {

	}

	return
}

type csService struct{}

func (m *csService) Execute(args []string, r <-chan svc.ChangeRequest, changes chan<- svc.Status) (ssec bool, errno uint32) {
	const cmdsAccepted = svc.AcceptStop | svc.AcceptShutdown | svc.AcceptPauseAndContinue
	changes <- svc.Status{State: svc.StartPending}

	go StartServer()

	changes <- svc.Status{State: svc.Running, Accepts: cmdsAccepted}
loop:
	for {
		select {
		case c := <-r:
			switch c.Cmd {
			case svc.Interrogate:
				changes <- c.CurrentStatus
				// Testing deadlock from https://code.google.com/p/winsvc/issues/detail?id=4
				time.Sleep(100 * time.Millisecond)
				changes <- c.CurrentStatus
			case svc.Stop, svc.Shutdown:
				break loop
			case svc.Pause:
				changes <- svc.Status{State: svc.Paused, Accepts: cmdsAccepted}
			case svc.Continue:
				changes <- svc.Status{State: svc.Running, Accepts: cmdsAccepted}
			default:
				//out.Printi(out.LogArg{"pn": "main", "fn": "Execute", "text": "unexpected control request", "request": c})
				// elog.Error(1, fmt.Sprintf("unexpected control request #%d", c))
			}
		}
	}
	changes <- svc.Status{State: svc.StopPending}
	return
}

func runService(name string, isDebug bool) {
	//out.Printi(out.LogArg{"pn": "main", "text": "start service", "name": name})

	var err error
	run := svc.Run
	if isDebug {
		run = debug.Run
	}
	err = run(name, &csService{})
	if err != nil {
		// elog.Error(1, fmt.Sprintf("%s service failed: %v", name, err))
		return
	}
	// elog.Info(1, fmt.Sprintf("%s service stopped", name))
}

// writeLog ...
func writeLog(msg string) {
	data, _ := ioutil.ReadFile("D:/log.txt")
	buf := make([]byte, len(data)+len(msg))
	copy(buf, data[:])
	copy(buf[len(data):], []byte(msg))
	ioutil.WriteFile("D:/log.txt", buf, 0)
}

//*/
