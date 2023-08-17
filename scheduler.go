package main

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/shinYeongHyeon/go-times"
	"gopkg.in/robfig/cron.v3"
)

const (
	//            Minute   Hour   Day      Month                  Day of Week
	//		 		0-59   0-23   1-31   1-12 or JAN-DEC		  0-6 or SUN-SAT
	// CronSpec = "47 15 * * FRI"
	CronSpec = "01 16 * * MON"
	// CronSpec = "0 11 * * THU"
)

var (
	Data int
)

func CreateCron() {
	c := cron.New()
	c.AddFunc(CronSpec, makeWeekRpt)
	c.Start()
}

func PrintReport(mem_name string, mem_id string, mem_part int32, now string, week_content string, omissionDate string) {
	_, _, _, t := getDate()
	fmt.Println("주간보고 제목 : ", getWeekRptTitle(mem_name, t))
	fmt.Println("주간보고 보고자 : ", mem_name)
	fmt.Println("주간보고 대상 : ", getWeekRpttoRpt(mem_id))
	fmt.Println("주간보고 파트 : ", mem_part)
	fmt.Println("주간보고 날짜 : ", now)
	fmt.Println("week_content : ", week_content)
	fmt.Println("omissionDate : ", omissionDate)
	fmt.Println("-----------------------------절취선-------------------------------")
}

func insertWeekRpt(mem_name string, mem_id string, mem_part int32, now string, week_content string, omissionDate string) {
	_, _, _, t := getDate()
	queryString := fmt.Sprintf(`INSERT INTO bsmgWeekRpt VALUES('%s','%s','%s','%s',N'%s',%d,'%s')`,
		mem_id, now, getWeekRpttoRpt(mem_id), getWeekRptTitle(mem_name, t), week_content, mem_part, omissionDate)
	_, err := db.Exec(queryString)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("일일 업무보고 취합 완료")
}

// 일일 업무보고 취합할 기간
func getDate() (string, string, string, time.Time) {
	// t := time.Now()
	t := time.Now().AddDate(0, 0, 3)
	now := t.Format("20060102000000")
	bef7d := t.AddDate(0, 0, -7).Format("20060102000000") //7일전 (저번주목욜)
	bef1d := t.AddDate(0, 0, -1).Format("20060102000000") // 1일전 (이번주수욜)
	return bef7d, bef1d, now, t
}

// 주간보고로 취합
func makeWeekRpt() {
	bef7d, bef1d, now, t := getDate()
	var count int = 0 // 멤버 수
	queryString := `SELECT top 1 mem_idx FROM bsmgMembers ORDER BY mem_idx DESC`
	err := db.QueryRow(queryString).Scan(&count)
	if err != nil {
		fmt.Println(err)
		return
	}
	if count <= 0 {
		return
	}

	for i := 0; i <= count; i++ {
		// 보고자 ID get
		var mem_id sql.NullString
		var mem_name sql.NullString
		var mem_part sql.NullInt32
		var mem_rank sql.NullInt32
		queryString = fmt.Sprintf(`SELECT mem_id,mem_part,mem_name,mem_rank FROM bsmgMembers WHERE mem_idx =%d `, i)
		_ = db.QueryRow(queryString).Scan(&mem_id, &mem_part, &mem_name, &mem_rank)

		if mem_rank.Int32 >= 3 {
			// 기간에 맞는 일일 보고서 get
			queryString = fmt.Sprintf(`SELECT rpt_date, rpt_title, rpt_content FROM bsmgReport WHERE rpt_reporter='%s' 
				and CONVERT(numeric, rpt_date) >= '%s' and CONVERT(numeric, rpt_date) <= '%s'`, mem_id.String, bef7d, bef1d)
			rows, err := db.Query(queryString)
			if err != nil {
				fmt.Println(err)
				return
			}
			defer rows.Close()

			// 보고가 빠진 날짜 찾기 세팅
			findOmission := setOmissionMap(t)

			// 주간보고 내용
			var week_content string = ""
			for rows.Next() {
				var rpt_date sql.NullString
				var rpt_title sql.NullString
				var rpt_content sql.NullString

				err = rows.Scan(&rpt_date, &rpt_title, &rpt_content)
				if err != nil {
					fmt.Println("변수에 데이터 할당 오류", err)
				}
				week_content += "📆" + rpt_date.String[:8] + "\n"
				week_content += rpt_content.String + "\n"
				setRptDate(findOmission, rpt_date.String)
			}
			if week_content != "" {
				omissionDate := extractMap(findOmission)
				PrintReport(mem_name.String, mem_id.String, mem_part.Int32, now, week_content, omissionDate)
				insertWeekRpt(mem_name.String, mem_id.String, mem_part.Int32, now, week_content, omissionDate)
				// fmt.Println(findOmission)
			}
		}
	}

}

// 주간 업무 기간 세팅 (빠진 날짜 찾기위해)
func setOmissionMap(t time.Time) map[string]bool {
	findOmission := map[string]bool{}
	// 초기화 : make()를 써도 된다. 맵을 초기화하지않고 사용하면 assingment to entry in nil map 에러가 발생한다
	for i := 0; i < 7; i++ {
		date := t.AddDate(0, 0, -7+i)
		if date.Weekday() == 6 || date.Weekday() == 0 { // 토요일이거나 일요일이면 true
			findOmission[t.AddDate(0, 0, -7+i).Format("20060102")] = true
		} else {
			findOmission[t.AddDate(0, 0, -7+i).Format("20060102")] = false
		}

	}
	return findOmission
}

// 보고한 날짜 = true로
func setRptDate(findOmission map[string]bool, rpt_date string) {
	rpt_date = rpt_date[:8]
	_, exists := findOmission[rpt_date] // key값, 존재여부 반환
	if exists {
		findOmission[rpt_date] = true
	}
}

// map에서 value가 false인 key 추출
func extractMap(findOmission map[string]bool) string {
	var omissionDate string
	for key, value := range findOmission {
		if !value {
			omissionDate += key + ", "
		}
	}
	omissionDate = omissionDate[:len(omissionDate)-2]
	return omissionDate
}

// 주간업무보고 제목
func getWeekRptTitle(mem_name string, t time.Time) string {
	t = t.AddDate(0, 0, -1)
	// t := time.Now().AddDate(0, 0, -1)
	month := t.Format("200601020000")[4:6]
	nWeek := times.GetNthWeekOfMonth(t)
	return fmt.Sprintf("%s월 %d주차 %s 주간 업무보고", month, nWeek, mem_name)
}

// 각 팀의 팀장님 (보고대상)
func getWeekRpttoRpt(mem_id string) string {
	var toRptName sql.NullString

	queryString := fmt.Sprintf(`SELECT a.mem_name FROM bsmgMembers a
		INNER JOIN bsmgRank ON a.mem_rank = rank_idx AND rank_name = '팀장'
		INNER JOIN bsmgPart ON a.mem_part = part_idx 
		INNER JOIN (SELECT mem_part FROM bsmgPart 
		INNER JOIN bsmgMembers ON mem_part = part_idx AND mem_id = '%s') b ON a.mem_part = b.mem_part`, mem_id)
	err := db.QueryRow(queryString).Scan(&toRptName)
	if err != nil {
		fmt.Println(err)
		return ""
	}
	return toRptName.String
}
