package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

// 로그인 핸들러
func bsmgLoginHandler(writer http.ResponseWriter, request *http.Request) {
	var errResult *WebErrorResult
	fmt.Println("--------------------------------")
	fmt.Println("=== LoginHandler 진입 ===")
	fmt.Println("--------------------------------")

	switch request.Method {
	case "GET": // 로그인 세션체크
		// errResult = getChkLoginRequest(writer, request)
		errResult = getLoginFunc(writer, request)
	case "POST": // 로그인, 로그아웃
		errResult = postLoginFuncs(writer, request)
	case "DELETE": // 사용자 삭제시 사용
		//errResult = deleteLoginRequest(writer, request)
	default:
		errResult = &WebErrorResult{http.StatusOK, ErrorInvalidParameter}
	}

	// 에러 발생 시 처리
	if errResult != nil {
		sendError(writer, request, errResult.httpStatus, errResult.resultCode)
	}
}

func getLoginFunc(writer http.ResponseWriter, request *http.Request) *WebErrorResult {
	reqSlice := strings.Split(request.URL.Path[1:], "/") // 맨 앞에 붙은 "/" 빼고 split

	switch len(reqSlice) {
	case 3: //    /bsmg/login/{id}
		parameter := reqSlice[2]
		switch parameter {
		case "chkLogin": // 세션체크
			return getChkLoginRequest(writer, request)
		default:
			return &WebErrorResult{http.StatusOK, ErrorInvalidParameter}
		}
	default:
		return &WebErrorResult{http.StatusOK, ErrorInvalidParameter}
	}
}
func postLoginFuncs(writer http.ResponseWriter, request *http.Request) *WebErrorResult {
	reqSlice := strings.Split(request.URL.Path[1:], "/")

	switch len(reqSlice) {
	case 3: //    /bsmg/login/{id}
		parameter := reqSlice[2]
		switch parameter {
		case "login":
			return postLoginRequest(writer, request)
		case "logout":
			return postLogoutRequest(writer, request)
		default:
			return &WebErrorResult{http.StatusOK, ErrorInvalidParameter}
		}
	default:
		return &WebErrorResult{http.StatusOK, ErrorInvalidParameter}
	}
}

// 로그인 요청 처리
func postLoginRequest(writer http.ResponseWriter, request *http.Request) *WebErrorResult {
	// 잘못된 요청
	if len(request.URL.Path) < 1 {
		return &WebErrorResult{http.StatusOK, ErrorInvalidParameter}
	}

	// 빈 값 -> 에러처리
	parser := initFormParser(request)
	if parser == nil {
		return &WebErrorResult{http.StatusOK, ErrorInvalidParameter}
	}

	// 서버에서 사용할 변수 선언
	var mem_id string
	var mem_pw string
	var mem_name string
	var mem_rank string
	var mem_part string
	var err error

	// 요청으로 들어온 변수 파싱 및 변수에 할당
	mem_id, err = parser.getStringValue(0, "mem_id", 0)
	if err != nil {
		return &WebErrorResult{http.StatusOK, ErrorInvalidParameter}
	}
	mem_pw, err = parser.getStringValue(0, "mem_pw", 0)
	if err != nil {
		return &WebErrorResult{http.StatusOK, ErrorInvalidParameter}
	}

	var result BsmgMemberResult
	// DB작업 쿼리문 - id, pw 입력
	queryString := fmt.Sprintf("SELECT count(*) FROM bsmgMembers WHERE mem_id = '%s' and mem_pw = '%s'", mem_id, mem_pw)

	row := db.QueryRow(queryString)

	var count int = 4 // 그냥 의미없는 숫자
	err = row.Scan(&count)
	if err != nil {
		fmt.Println("DB데이터 서버변수에 할당 중 오류 ", err)
		return &WebErrorResult{http.StatusOK, ErrorInvalidParameter}
	}

	// 정보가 있으면 직급과 부서를 이름으로 해서 가져오고 정보가 없으면 ResultCode = 1로 해서 리턴
	if count == 1 { // 아이디 비번이 일치함(정보가 있다.)
		queryString = fmt.Sprintf("SELECT m.mem_id,m.mem_name, r.rank_name, p.part_name FROM bsmgMembers m LEFT OUTER JOIN bsmgRank r ON m.mem_rank = r.rank_idx LEFT OUTER JOIN bsmgPart p ON m.mem_part = p.part_idx WHERE m.mem_id = '%s' AND m.mem_pw = '%s'", mem_id, mem_pw)
		row = db.QueryRow(queryString)

		// DB에서 받은 데이터 할당
		err = row.Scan(&mem_id, &mem_name, &mem_rank, &mem_part)
		if err != nil {
			fmt.Println("DB데이터 서버변수에 할당 중 오류 ", err)
			return &WebErrorResult{http.StatusOK, ErrorInvalidParameter}
		}

		result.MemberInfo = &BsmgMemberInfo{}
		result.MemberInfo.Mem_ID = mem_id
		result.MemberInfo.Mem_Name = mem_name
		result.MemberInfo.Mem_Rank = mem_rank
		result.MemberInfo.Mem_Part = mem_part
		result.Result.ResultCode = 0

		var sessionValue sessionValue
		sessionValue.ID = mem_id
		// -------------------------------------------세션 생성 및 value 저장
		CreateSession(writer, request, sessionValue)

		// 세션으로부터 값 받아오기 테스트
		session, _ := store.Get(request, "Member")
		sessionID := session.Values["mem_id"].(string)

		fmt.Println("아이디 : ", sessionID, ", 세션 : ", session)
		// -----------------------------------------------------------
		// bsmgMemberResult 객체(json객체)를 마샬링(xml 변환)
		responseMsg, _ := json.Marshal(result)
		sendResultEx(writer, request, http.StatusOK, responseMsg)
		return nil

	} else { // 아이디 비번이 일치하지 않거나 없는 아이디
		//에러없이 동작하면 ResultCode = 0, 그 외에는 모두 에러
		result.Result.ResultCode = 1

		// bsmgMemberResult 객체(json객체)를 마샬링(xml 변환)
		responseMsg, _ := json.Marshal(result)
		sendResultEx(writer, request, http.StatusOK, responseMsg)
		return nil
	}
}

// 로그아웃 요청 처리
func postLogoutRequest(writer http.ResponseWriter, request *http.Request) *WebErrorResult {

	// 잘못된 요청
	if len(request.URL.Path) < 1 {
		return &WebErrorResult{http.StatusOK, ErrorInvalidParameter}
	}

	var result OnlyResult
	result.Result.ResultCode = 1

	session, _ := store.Get(request, "Member")
	fmt.Println("세션 삭제 전 : ", session)
	// 세션 삭제
	DeleteSession(writer, request)
	fmt.Println("세션 삭제 후 : ", session)

	result.Result.ResultCode = 0

	responseMsg, _ := json.Marshal(result)
	sendResultEx(writer, request, http.StatusOK, responseMsg)

	return nil
}

// 아이디 중복체크 요청처리
func getIdCheckRequest(writer http.ResponseWriter, request *http.Request) *WebErrorResult {

	fmt.Println(" === idCheck Request 처리 ===")
	if len(request.URL.Path) < 1 {
		return &WebErrorResult{http.StatusOK, ErrorInvalidParameter}
	}

	parser := initFormParser(request)
	if parser == nil {
		fmt.Println("데이터 파싱 오류")
		return &WebErrorResult{http.StatusOK, ErrorInvalidParameter}
	}

	var mem_id string
	var count int
	var err error

	// 변수 파싱 및 할당
	mem_id, err = parser.getStringValue(0, "mem_id", 0)
	if err != nil {
		fmt.Println("파싱에러 (Request 데이터 파싱 실패)")
		return &WebErrorResult{http.StatusOK, ErrorInvalidParameter}
	}

	// fmt.Println(mem_id)
	// DB처리
	var result BsmgMemberResult
	result.Result.ResultCode = 0

	queryString := fmt.Sprintf("SELECT count(*) FROM bsmgMembers WHERE mem_id = '%s'", mem_id) // %s에 ' ' 안 써서 에러가 떴었다.
	err = db.QueryRow(queryString).Scan(&count)
	if err != nil {
		fmt.Println("DB결과 변수에 할당 중 오류 ", err)
		//fmt.Print(count) count에 값을 넣는 도중에 에러가 뜨네
		return &WebErrorResult{http.StatusOK, ErrorInvalidParameter}
	}
	// fmt.Println("count : ", count)

	// DB로부터 온 row에서 데이터 할당
	//err = row.Scan(&count)

	if count == 0 { // 아이디 중복이 안됨
		result.Result.ResultCode = 0
	} else { // 중복됨
		result.Result.ResultCode = 1
	}
	responseMsg, _ := json.Marshal(result)
	sendResultEx(writer, request, http.StatusOK, responseMsg)

	return nil
}

// 세팅관련 핸들러 (업무 속성, 직급부서)
func bsmgSettingHandler(writer http.ResponseWriter, request *http.Request) {
	var errResult *WebErrorResult
	fmt.Println(" === settingHandler 진입 === ")

	switch request.Method {
	case "GET": // 속성 트리 세팅
		errResult = getSettingFunc(writer, request)
	default:
		errResult = &WebErrorResult{http.StatusOK, ErrorInvalidParameter}
	}

	if errResult != nil {
		sendError(writer, request, errResult.httpStatus, errResult.resultCode)
	}
}

func getSettingFunc(writer http.ResponseWriter, request *http.Request) *WebErrorResult {
	reqSlice := strings.Split(request.URL.Path[1:], "/")
	switch len(reqSlice) {
	case 3: //    /bsmg/setting/{id}
		parameter := reqSlice[2]
		switch parameter {
		case "attrTree":
			return setAttrTreeRequest(writer, request)
		case "rankPart":
			return setRankPartRequest(writer, request)
		case "weekRptCategory":
			return setWeekRptCategoryRequest(writer, request)
		case "getToRpt":
			return setToRpt(writer, request)
		default:
			return &WebErrorResult{http.StatusOK, ErrorInvalidParameter}
		}
	default:
		return &WebErrorResult{http.StatusOK, ErrorInvalidParameter}
	}
}

// 속성 트리 세팅 요청 처리
func setAttrTreeRequest(writer http.ResponseWriter, request *http.Request) *WebErrorResult {
	if len(request.URL.Path) < 1 {
		fmt.Println("요청 URL 오류 ", request.URL.Path)
		return &WebErrorResult{http.StatusOK, ErrorInvalidParameter}
	}

	var err error
	var cnt sql.NullInt32
	var result BsmgTreeResult
	result.Result.ResultCode = 1

	// 필요한 수 만큼 AttrTreeList 생성
	queryString := `SELECT sum(total.cnt) 
								  FROM 
								  (SELECT COUNT(*) as cnt FROM bsmgAttr1 UNION ALL 
								  SELECT COUNT(*) as cnt FROM bsmgAttr2) total`
	err = db.QueryRow(queryString).Scan(&cnt)
	if err != nil {
		fmt.Println("쿼리문 오류 cnt", err)
		return &WebErrorResult{http.StatusOK, ErrorInvalidParameter}
	}

	if int(cnt.Int32) > 0 {
		result.AttrTreeList = make([]*AttrTree, int(cnt.Int32))
	}

	// 업무속성1 가져와서 통신할 result에 할당
	queryString = "SELECT attr1_category FROM bsmgAttr1"
	rows, err := db.Query(queryString)
	if err != nil {
		fmt.Println("쿼리문 오류", err)
		return &WebErrorResult{http.StatusOK, ErrorInvalidParameter}
	}
	defer rows.Close()

	var ii = 0
	// 업무 속성1 트리에 할당 (부모)
	for rows.Next() {
		var attr1_category sql.NullString

		err := rows.Scan(&attr1_category)
		if err != nil {
			fmt.Println("서버 변수 DB 데이터 할당 오류", err)
			return &WebErrorResult{http.StatusOK, ErrorInvalidParameter}
		}
		result.AttrTreeList[ii] = &AttrTree{}
		result.AttrTreeList[ii].Label = attr1_category.String
		result.AttrTreeList[ii].Value = strconv.Itoa(ii + 1)
		result.AttrTreeList[ii].Parent = "0" // 최상단 트리
		ii++
	}

	queryString = `SELECT ba2.attr2_name , 
	CASE WHEN ba2.attr1_idx = 0 THEN '1-' 
		 WHEN ba2.attr1_idx = 1 THEn '2-' 
		 ELSE 'N/A' 
	END AS [value], 
	ba2.attr1_idx,
	ba2.attr2_idx 
	FROM bsmgAttr2 ba2`
	rows, err = db.Query(queryString)
	if err != nil {
		fmt.Println("쿼리 오류 ", err)
		return &WebErrorResult{http.StatusOK, ErrorInvalidParameter}
	}
	//defer rows.Close() //앞에서 한 번 했으니 안해도 되겠지?

	// 업무 속성2 트리에 할당 (자식)
	var parent = "1"
	for rows.Next() {
		var attr2_name sql.NullString
		var value sql.NullString
		var attr1_idx sql.NullInt32
		var attr2_idx sql.NullInt32

		err := rows.Scan(&attr2_name, &value, &attr1_idx, &attr2_idx)
		if err != nil {
			fmt.Println("데이터 서버변수에 저장중 오류 : ", err)
			return &WebErrorResult{http.StatusOK, ErrorInvalidParameter}
		}

		result.AttrTreeList[ii] = &AttrTree{} // 리스트에는 타입 어떻게 할건지 해줘야지..
		result.AttrTreeList[ii].Label = attr2_name.String
		if int(attr1_idx.Int32) == 0 { // 업무 속성이 제품일때
			result.AttrTreeList[ii].Value = value.String + strconv.Itoa(int(attr2_idx.Int32))
			result.AttrTreeList[ii].Parent = parent
		} else { // 솔루션일때
			result.AttrTreeList[ii].Value = value.String + strconv.Itoa(int(attr2_idx.Int32))
			parent = "2"
			result.AttrTreeList[ii].Parent = parent
		}
		ii++
	}

	result.Result.ResultCode = 0
	responseMsg, _ := json.Marshal(result)
	sendResultEx(writer, request, http.StatusOK, responseMsg)

	return nil
}

// 직급 부서 세팅 요청 처리
func setRankPartRequest(writer http.ResponseWriter, request *http.Request) *WebErrorResult {
	if len(request.URL.Path) < 1 {
		fmt.Println("요청 URL 오류 setRankPartRequest", request.URL.Path)
		return &WebErrorResult{http.StatusOK, ErrorInvalidParameter}
	}

	var err error
	/* 받아오는 게 없으니 이건 필요없다.
	parser := initFormParser(request)
	if parser == nil {
		return &WebErrorResult{http.StatusOK, ErrorInvalidParameter}
	}
	*/

	// 직급 개수 세기
	queryString := "SELECT count(*) FROM bsmgRank"
	var count int
	_ = db.QueryRow(queryString).Scan(&count)
	// fmt.Printf("직급 총 %d 개 ", count)

	// 앞단에 보낼 결과 구조체 선언
	var result BsmgRankPartResult
	result.Result.ResultCode = 1

	// 직급 수만큼 RankInfo구조체 배열 선언
	if count > 0 {
		result.RankList = make([]*BsmgRankInfo, count)
	}

	// DB작업
	queryString = "SELECT * FROM bsmgRank"
	rows, err := db.Query(queryString)
	if err != nil {
		fmt.Println("조회 오류 ", err)
		return &WebErrorResult{http.StatusOK, ErrorInvalidParameter}
	}
	defer rows.Close()
	var ii = 0

	// rows에 담긴 데이터 분류(직급이름 배열에 저장)
	for rows.Next() {
		var rank_idx sql.NullInt32
		var rank_name sql.NullString

		err := rows.Scan(&rank_idx, &rank_name)
		if err != nil {
			fmt.Println("데이터 서버변수에 저장중 오류 : ", err)
			return &WebErrorResult{http.StatusOK, ErrorInvalidParameter}
		}

		result.RankList[ii] = &BsmgRankInfo{}
		result.RankList[ii].Rank_Idx = rank_idx.Int32
		result.RankList[ii].Rank_Name = rank_name.String
		ii++
	}

	// 부서 개수 세기
	queryString = "SELECT count(*) FROM bsmgPart"
	_ = db.QueryRow(queryString).Scan(&count)
	// fmt.Printf("부서 총 %d 개", count)

	// 부서 수만큼 PartInfo 배열 선언
	if count > 0 {
		result.PartList = make([]*BsmgPartInfo, count)
	}

	// DB작업
	queryString = "SELECT * FROM bsmgPart"
	rows, err = db.Query(queryString)
	if err != nil {
		fmt.Println("조회 오류 ", err)
		return &WebErrorResult{http.StatusOK, ErrorInvalidParameter}
	}
	// 두번째에는 rows 안꺼도 되겠지
	ii = 0 // 아까 쓰던 변수니까 초기화
	for rows.Next() {
		var part_idx sql.NullInt32
		var part_name sql.NullString

		err := rows.Scan(&part_idx, &part_name)
		if err != nil {
			fmt.Println("데이터 서버변수에 저장중 오류 : ", err)
			return &WebErrorResult{http.StatusOK, ErrorInvalidParameter}
		}

		result.PartList[ii] = &BsmgPartInfo{}
		result.PartList[ii].Part_Idx = part_idx.Int32
		result.PartList[ii].Part_Name = part_name.String
		ii++
	}
	result.Result.ResultCode = 0

	responseMsg, _ := json.Marshal(result)
	sendResultEx(writer, request, http.StatusOK, responseMsg)

	return nil
}

func setWeekRptCategoryRequest(writer http.ResponseWriter, request *http.Request) *WebErrorResult {
	var err error

	// 부서 개수 세기
	queryString := "SELECT COUNT(*) FROM bsmgPart"
	var count int
	_ = db.QueryRow(queryString).Scan(&count)

	var result BsmgTreeResult
	result.Result.ResultCode = 1
	if count > 0 {
		result.PartTreeList = make([]*PartTree, count)
	}

	queryString = "SELECT part_idx,part_name FROM bsmgPart WHERE part_idx != 0" // 관리자는 제외
	rows, err := db.Query(queryString)
	if err != nil {
		fmt.Println("쿼리문 오류", err)
		return &WebErrorResult{http.StatusOK, ErrorInvalidParameter}
	}
	defer rows.Close()

	var ii = 0
	for rows.Next() {
		var part_idx sql.NullInt32
		var part_category sql.NullString

		err := rows.Scan(&part_idx, &part_category)
		if err != nil {
			fmt.Println("서버 변수 DB 데이터 할당 오류", err)
			return &WebErrorResult{http.StatusOK, ErrorInvalidParameter}
		}
		result.PartTreeList[ii] = &PartTree{}
		result.PartTreeList[ii].Label = part_category.String
		result.PartTreeList[ii].Value = "1-" + strconv.Itoa(int(part_idx.Int32))
		result.PartTreeList[ii].Parent = "1"
		ii++
	}
	result.PartTreeList[ii] = &PartTree{}
	result.PartTreeList[ii].Label = "부서별 주간 업무보고"
	result.PartTreeList[ii].Value = "1"
	result.PartTreeList[ii].Parent = "0"

	result.Result.ResultCode = 0
	responseMsg, _ := json.Marshal(result)
	sendResultEx(writer, request, http.StatusOK, responseMsg)
	return nil
}

func setToRpt(writer http.ResponseWriter, request *http.Request) *WebErrorResult {
	var err error
	parser := initFormParser(request)
	if parser == nil {
		fmt.Println("파싱 오류", parser)
		return &WebErrorResult{http.StatusOK, ErrorInvalidParameter}
	}

	var part_idx int32

	part_idx, err = parser.getInt32Value(0, "part_idx", 0)
	if err != nil {
		fmt.Println(err)
		return &WebErrorResult{http.StatusOK, ErrorInvalidParameter}
	}

	var team_leader sql.NullString
	queryString := fmt.Sprintf(`SELECT mem_name FROM bsmgMembers
				INNER JOIN bsmgPart ON part_idx = mem_part
				WHERE mem_part = %d and mem_rank <= 3`, part_idx)
	err = db.QueryRow(queryString).Scan(&team_leader)
	if err != nil {
		fmt.Println(err)
		return &WebErrorResult{http.StatusOK, ErrorInvalidParameter}
	}

	var result BsmgTeamLeaderResult
	result.Part.TeamLeader = team_leader.String
	result.Part.PartIdx = part_idx
	result.Result.ResultCode = 0
	responseMsg, _ := json.Marshal(result)
	sendResultEx(writer, request, http.StatusOK, responseMsg)
	return nil
}

// 유저 등록 핸들러 (필요 시 수정, 삭제기능 )
func bsmgUserHandler(writer http.ResponseWriter, request *http.Request) {
	var errResult *WebErrorResult
	// fmt.Println(" === bsmgUserHandler 진입 ===")

	switch request.Method {
	case "GET":
		errResult = getUserFunc(writer, request)
	case "POST": // 등록
		errResult = postUserRequest(writer, request)
	case "PUT": // 수정
		errResult = putUserListRequest(writer, request)
	case "DELETE": //삭제
		errResult = deleteUserFunc(writer, request)
	default:
		errResult = &WebErrorResult{http.StatusMethodNotAllowed, ErrorInvalidParameter}
	}

	if errResult != nil {
		sendError(writer, request, errResult.httpStatus, errResult.resultCode)
	}
}

func getUserFunc(writer http.ResponseWriter, request *http.Request) *WebErrorResult {
	// 첫 슬래시 제외하고 슬라이싱
	reqSlice := strings.Split(request.URL.Path[1:], "/")

	switch len(reqSlice) {
	case 3:
		parameter := reqSlice[2]
		switch parameter {
		case "userList": // 사용자 리스트 조회
			return getUserListRequest(writer, request)
		case "idCheck": // 아이디 중복체크
			return getIdCheckRequest(writer, request)
		case "userSearch": // 유저 검색 (보고대상,참조대상)
			return getUserSearchRequest(writer, request)
		default:
			return &WebErrorResult{http.StatusOK, ErrorInvalidParameter}
		}
	default:
		return &WebErrorResult{http.StatusOK, ErrorInvalidParameter}
	}
}

func deleteUserFunc(writer http.ResponseWriter, request *http.Request) *WebErrorResult {
	reqSlice := strings.Split(request.URL.Path[1:], "/")

	switch len(reqSlice) {
	case 4:
		parameter := reqSlice[2]
		switch parameter {
		case "deleteUser": // delete method는 url에 idx 포함시켜서 사용한다. (서브미션의 데이터가 제대로 안넘어옴)
			return deleteUserRequest(reqSlice[3], writer, request)
		default:
			return &WebErrorResult{http.StatusOK, ErrorInvalidParameter}
		}
	default:
		return &WebErrorResult{http.StatusOK, ErrorInvalidParameter}
	}
}

func getUserListRequest(writer http.ResponseWriter, request *http.Request) *WebErrorResult {
	// fmt.Println("getUserRequest 호출 확인")
	if len(request.URL.Path) < 1 {
		return &WebErrorResult{http.StatusOK, ErrorInvalidParameter}
	}

	var err error
	queryString := "SELECT count(*) FROM bsmgMembers"

	// 멤버리스트의 열 갯수
	var count int = 0
	err = db.QueryRow(queryString).Scan(&count) // 쿼리문에서 열의 수 파악해서 count에 할당
	if err != nil {
		fmt.Println("count 쿼리 오류", err)
		return &WebErrorResult{http.StatusOK, ErrorInvalidParameter}
	}

	var result BsmgMemberResult
	result.Result.ResultCode = 0

	// count 수만큼 MemberList에 할당할 MemberInfo 구조체 생성
	if count > 0 {
		result.MemberList = make([]*BsmgMemberInfo, count)
	}

	// DB작업
	queryString = `SELECT m.mem_id, m.mem_name, r.rank_name, p.part_name
		FROM bsmgMembers m LEFT OUTER JOIN bsmgRank r
	  	  ON m.mem_rank = r.rank_idx 
	  	LEFT OUTER JOIN bsmgPart p
	  	  ON m.mem_part = p.part_idx`

	rows, err := db.Query(queryString)
	if err != nil {
		fmt.Println("조회 오류", err)
		return &WebErrorResult{http.StatusOK, ErrorInvalidParameter}
	}
	defer rows.Close()

	var ii = 0
	for rows.Next() {
		var mem_id sql.NullString
		var mem_name sql.NullString
		var mem_rank sql.NullString
		var mem_part sql.NullString

		err := rows.Scan(&mem_id, &mem_name, &mem_rank, &mem_part)
		if err != nil {
			fmt.Println("데이터 서버변수에 할당 오류", err)
			return &WebErrorResult{http.StatusOK, ErrorInvalidParameter}
		}

		// 가져온 데이터 구조체에 할당
		result.MemberList[ii] = &BsmgMemberInfo{}
		result.MemberList[ii].Mem_ID = mem_id.String
		result.MemberList[ii].Mem_Name = mem_name.String
		result.MemberList[ii].Mem_Rank = mem_rank.String
		result.MemberList[ii].Mem_Part = mem_part.String

		ii++
	}

	responseMsg, _ := json.Marshal(result)
	sendResultEx(writer, request, http.StatusOK, responseMsg)
	return nil
}

// 유저 등록 요청 처리
func postUserRequest(writer http.ResponseWriter, request *http.Request) *WebErrorResult {
	if len(request.URL.Path) < 1 {
		fmt.Println("요청 URL 오류 ", request.URL.Path)
		return &WebErrorResult{http.StatusOK, ErrorInvalidParameter}
	}

	parser := initFormParser(request)
	if parser == nil {
		fmt.Println("데이터 파싱 오류 ", parser)
		return &WebErrorResult{http.StatusOK, ErrorInvalidParameter}
	}

	// 서버에서 사용할 변수 선언
	var mem_id string
	var mem_pw string
	var mem_name string
	var mem_rank int32
	var mem_part int32
	var err error

	mem_id, err = parser.getStringValue(0, "mem_id", 0)
	if err != nil {
		fmt.Println(err)
		return &WebErrorResult{http.StatusOK, ErrorInvalidParameter}
	}
	mem_pw, err = parser.getStringValue(0, "mem_pw", 0)
	if err != nil {
		fmt.Println(err)
		return &WebErrorResult{http.StatusOK, ErrorInvalidParameter}
	}
	mem_name, err = parser.getStringValue(0, "mem_name", 0)
	if err != nil {
		fmt.Println(err)
		return &WebErrorResult{http.StatusOK, ErrorInvalidParameter}
	}
	mem_rank, err = parser.getInt32Value(0, "mem_rank", 0)
	if err != nil {
		fmt.Println(err)
		return &WebErrorResult{http.StatusOK, ErrorInvalidParameter}
	}
	mem_part, err = parser.getInt32Value(0, "mem_part", 0)
	if err != nil {
		fmt.Println(err)
		return &WebErrorResult{http.StatusOK, ErrorInvalidParameter}
	}

	// fmt.Println(mem_id, "-", mem_pw, "-", mem_name, "-", mem_part, "-", mem_rank)

	// DB작업
	queryString := fmt.Sprintf("INSERT INTO bsmgMembers VALUES('%s', '%s', '%s', %d, %d)", mem_id, mem_pw, mem_name, mem_rank, mem_part)
	// 쿼리실행
	_, err = db.Exec(queryString)
	if err != nil {
		fmt.Println("쿼리문 오류 : ", err)
		return &WebErrorResult{http.StatusOK, ErrorInvalidParameter}
	}

	var result BsmgMemberResult
	result.Result.ResultCode = 0

	responseMsg, _ := json.Marshal(result)
	sendResultEx(writer, request, http.StatusOK, responseMsg)
	return nil
}

func putUserListRequest(writer http.ResponseWriter, request *http.Request) *WebErrorResult {
	var err error
	parser := initFormParser(request)
	if parser == nil {
		fmt.Println("parser 에러 ", parser)
		return &WebErrorResult{http.StatusOK, ErrorInvalidParameter}
	}

	var mem_id string
	var mem_name string
	var mem_rank int32
	var mem_part int32

	cnt, err := parser.getValueCount(0, "mem_id")
	if err != nil {
		fmt.Println(err)
		return &WebErrorResult{http.StatusOK, ErrorInvalidParameter}
	}

	fmt.Println("수정할 사용자 수 (수정): ", cnt)
	for i := 0; i < cnt; i++ {
		mem_id, err = parser.getStringValue(0, "mem_id", i)
		if err != nil {
			fmt.Println(err)
			return &WebErrorResult{http.StatusOK, ErrorInvalidParameter}
		}
		mem_name, err = parser.getStringValue(0, "mem_name", i)
		if err != nil {
			fmt.Println(err)
			return &WebErrorResult{http.StatusOK, ErrorInvalidParameter}
		}
		mem_rank, err = parser.getInt32Value(0, "mem_rank", i)
		if err != nil {
			fmt.Println(err)
			return &WebErrorResult{http.StatusOK, ErrorInvalidParameter}
		}
		mem_part, err = parser.getInt32Value(0, "mem_part", i)
		if err != nil {
			fmt.Println(err)
			return &WebErrorResult{http.StatusOK, ErrorInvalidParameter}
		}

		queryString := fmt.Sprintf(`UPDATE bsmgMembers SET mem_name = '%s', mem_rank = %d, mem_part = %d
		WHERE mem_id = '%s'`, mem_name, mem_rank, mem_part, mem_id)
		_, err = db.Exec(queryString)
		if err != nil {
			fmt.Println(err)
			return &WebErrorResult{http.StatusOK, ErrorInvalidParameter}
		}
	}
	var result OnlyResult
	result.Result.ResultCode = 0
	responseMsg, _ := json.Marshal(result)
	sendResultEx(writer, request, http.StatusOK, responseMsg)

	return nil
}

// 로그인 중인지 세션을 통해 확인
func getChkLoginRequest(writer http.ResponseWriter, request *http.Request) *WebErrorResult {

	var result BsmgMemberResult
	chkSess := chkSession(writer, request)
	if !chkSess.Authenticated {
		result.Result.ResultCode = 1

	} else {
		mem_id := chkSess.ID
		var mem_name string
		var mem_rank string
		var mem_part string
		var err error

		queryString := fmt.Sprintf(`SELECT m.mem_name, r.rank_name, p.part_name FROM bsmgMembers m
			LEFT OUTER JOIN bsmgRank r ON m.mem_rank = r.rank_idx
			LEFT OUTER JOIN bsmgPart p ON m.mem_part = p.part_idx 
			WHERE m.mem_id = '%s'`, mem_id)
		row := db.QueryRow(queryString)

		err = row.Scan(&mem_name, &mem_rank, &mem_part)
		if err != nil {
			fmt.Println("DB데이터 서버변수에 할당 중 오류 ", err)
			return &WebErrorResult{http.StatusOK, ErrorInvalidParameter}
		}

		result.MemberInfo = &BsmgMemberInfo{}
		result.MemberInfo.Mem_ID = mem_id
		result.MemberInfo.Mem_Name = mem_name
		result.MemberInfo.Mem_Rank = mem_rank
		result.MemberInfo.Mem_Part = mem_part
		result.Result.ResultCode = 0
	}

	responseMsg, _ := json.Marshal(result)
	sendResultEx(writer, request, http.StatusOK, responseMsg)
	return nil
}

func getUserSearchRequest(writer http.ResponseWriter, request *http.Request) *WebErrorResult {
	// fmt.Println("getUserSearch Request 호출 확인")
	if len(request.URL.Path) < 1 {
		return &WebErrorResult{http.StatusOK, ErrorInvalidParameter}
	}

	var err error
	var SearchCombo string
	var SearchInput string

	parser := initFormParser(request)
	if parser == nil {
		return &WebErrorResult{http.StatusOK, ErrorInvalidParameter}
	}

	SearchCombo, err = parser.getStringValue(0, "search_combo", 0)
	if err != nil {
		fmt.Println(err)
		return &WebErrorResult{http.StatusOK, ErrorInvalidParameter}
	}
	SearchInput, err = parser.getStringValue(0, "search_input", 0)
	if err != nil {
		fmt.Println(err)
		return &WebErrorResult{http.StatusOK, ErrorInvalidParameter}
	}

	// 갯수세는 쿼리 속도 : 조인 > 서브쿼리
	var queryString = "s"
	switch SearchCombo {
	case "0": // 전체
		queryString = fmt.Sprintf(`SELECT count(*) FROM bsmgMembers m
		WHERE m.mem_name like '%%%s%%' or
			  m.mem_part IN (SELECT part_idx FROM bsmgPart WHERE part_name like '%%%s%%') or
			  m.mem_rank IN (SELECT rank_idx FROM bsmgRank WHERE rank_name like '%%%s%%')`, SearchInput, SearchInput, SearchInput)
	case "1": // 이름
		queryString = fmt.Sprintf("SELECT count(*) FROM bsmgMembers WHERE mem_name like '%%%s%%'", SearchInput)
	case "2": // 직급
		queryString = fmt.Sprintf(`SELECT COUNT(*) FROM bsmgMembers m 
			LEFT OUTER JOIN bsmgRank r 
		  	  ON m.mem_rank = r.rank_idx
		   WHERE rank_name like '%%%s%%'`, SearchInput)
	case "3": // 부서
		queryString = fmt.Sprintf(`SELECT COUNT(*) FROM bsmgMembers m 
			LEFT OUTER JOIN bsmgPart p
		      ON m.mem_part = p.part_idx
		   WHERE part_name like '%%%s%%'`, SearchInput)
	}

	var count int = 0
	err = db.QueryRow(queryString).Scan(&count)
	if err != nil {
		fmt.Println("데이터 갯수 쿼리 오류", err)
		return &WebErrorResult{http.StatusOK, ErrorInvalidParameter}
	}

	var result BsmgMemberResult
	result.Result.ResultCode = 1

	if count > 0 {
		result.MemberList = make([]*BsmgMemberInfo, count)
	}

	// DB작업
	switch SearchCombo {
	case "0": // 전체
		queryString = fmt.Sprintf(`SELECT m.mem_id, m.mem_name, r.rank_name, p.part_name FROM bsmgMembers m 
			LEFT OUTER JOIN bsmgRank r ON r.rank_idx = m.mem_rank
			LEFT OUTER JOIN bsmgPart p ON p.part_idx = m.mem_part
			WHERE m.mem_name like '%%%s%%' or
			  m.mem_part IN (SELECT part_idx FROM bsmgPart p WHERE p.part_name like '%%%s%%') or
			  m.mem_rank IN (SELECT rank_idx FROM bsmgRank r WHERE r.rank_name like '%%%s%%')`, SearchInput, SearchInput, SearchInput)
	case "1": // 이름
		queryString = fmt.Sprintf(`SELECT m.mem_id, m.mem_name, r.rank_name, p.part_name FROM bsmgMembers m
			LEFT OUTER JOIN bsmgRank r ON r.rank_idx = m.mem_rank
			LEFT OUTER JOIN bsmgPart p ON p.part_idx = m.mem_part
			WHERE m.mem_name like '%%%s%%'`, SearchInput)
	case "2": // 직급
		queryString = fmt.Sprintf(`SELECT m.mem_id, m.mem_name, r.rank_name, p.part_name FROM bsmgMembers m
			LEFT OUTER JOIN bsmgRank r ON r.rank_idx = m.mem_rank
			LEFT OUTER JOIN bsmgPart p ON p.part_idx = m.mem_part
			WHERE r.rank_name like '%%%s%%'`, SearchInput)
	case "3": // 부서
		queryString = fmt.Sprintf(`SELECT m.mem_id, m.mem_name, r.rank_name, p.part_name FROM bsmgMembers m
			LEFT OUTER JOIN bsmgRank r ON r.rank_idx = m.mem_rank
			LEFT OUTER JOIN bsmgPart p ON p.part_idx = m.mem_part
			WHERE p.part_name like '%%%s%%'`, SearchInput)
	}

	rows, err := db.Query(queryString)
	if err != nil {
		fmt.Println("조회 오류", err)
		return &WebErrorResult{http.StatusOK, ErrorInvalidParameter}
	}
	defer rows.Close()
	var ii = 0

	for rows.Next() {
		var mem_id sql.NullString
		var mem_name sql.NullString
		var mem_rank sql.NullString
		var mem_part sql.NullString

		err := rows.Scan(&mem_id, &mem_name, &mem_rank, &mem_part)
		if err != nil {
			fmt.Println("가져온 데이터 오류(userSearch) : ", err)
			return &WebErrorResult{http.StatusOK, ErrorInvalidParameter}
		}

		result.MemberList[ii] = &BsmgMemberInfo{}
		result.MemberList[ii].Mem_ID = mem_id.String
		result.MemberList[ii].Mem_Name = mem_name.String
		result.MemberList[ii].Mem_Rank = mem_rank.String
		result.MemberList[ii].Mem_Part = mem_part.String
		ii++
	}

	result.Result.ResultCode = 0
	responseMsg, _ := json.Marshal(result)
	sendResultEx(writer, request, http.StatusOK, responseMsg)
	return nil
}

func deleteUserRequest(memID string, writer http.ResponseWriter, request *http.Request) *WebErrorResult {
	var err error
	queryString := fmt.Sprintf(`DELETE FROM bsmgReport WHERE rpt_reporter = '%s'`, memID)
	_, err = db.Exec(queryString)
	if err != nil {
		fmt.Println(err)
		return &WebErrorResult{http.StatusOK, ErrorInvalidParameter}
	}

	queryString = fmt.Sprintf(`DELETE FROM bsmgMembers WHERE mem_id = '%s'`, memID)
	_, err = db.Exec(queryString)
	if err != nil {
		fmt.Println(err)
		return &WebErrorResult{http.StatusOK, ErrorInvalidParameter}
	}
	var result OnlyResult
	result.Result.ResultCode = 0
	responseMsg, _ := json.Marshal(result)
	sendResultEx(writer, request, http.StatusOK, responseMsg)
	return nil
}

func bsmgReportHandler(writer http.ResponseWriter, request *http.Request) {
	var errResult *WebErrorResult

	switch request.Method {
	case "GET":
		errResult = getReportFunc(writer, request)
	case "POST":
		// errResult = postReportRequest(writer, request) // 일일업무보고 등록
		errResult = postReportFunc(writer, request)
	case "PUT":
		errResult = putReportFunc(writer, request)
	case "DELETE":
		errResult = deleteReportFunc(writer, request)
	default:
		errResult = &WebErrorResult{http.StatusOK, ErrorInvalidParameter}
	}

	if errResult != nil {
		sendError(writer, request, errResult.httpStatus, errResult.resultCode)
	}
}

func getReportFunc(writer http.ResponseWriter, request *http.Request) *WebErrorResult {
	reqSlice := strings.Split(request.URL.Path[1:], "/")
	switch len(reqSlice) {
	case 3: //      /bsmg/report/{id}
		parameter := reqSlice[2]
		switch parameter {
		case "reportList":
			return getReportListRequest(writer, request)
		case "reportSearch":
			return getReportSearchRequest(writer, request)
		case "reportAttrSearch":
			return getReportAttrRequest(writer, request)
		case "reportInfo":
			return getReportInfoRequest(writer, request)
		case "getSchdule":
			return getReportSchedule(writer, request)
		case "getWeekRptList":
			return getWeekReportListRequest(writer, request)
		case "getWeekRptSearch":
			return getWeekRptSearchRequest(writer, request)
		case "getWeekRptCategory":
			return getWeekRptCategory(writer, request)
		case "getWeekRptInfo":
			return getWeekRptInfoRequest(writer, request)
		case "confirmRpt":
			return confirmRpt(writer, request)
		default:
			return &WebErrorResult{http.StatusOK, ErrorInvalidParameter}
		}
	default:
		return &WebErrorResult{http.StatusOK, ErrorInvalidParameter}
	}
}

func postReportFunc(writer http.ResponseWriter, request *http.Request) *WebErrorResult {
	reqSlice := strings.Split(request.URL.Path[1:], "/")
	switch len(reqSlice) {
	case 3: //       /bsmg/report/{id}
		parameter := reqSlice[2]
		switch parameter {
		case "report": // 일일 업무보고서 보고
			return postReportRequest(writer, request)
		case "registSch":
			return postScheduleRequest(writer, request)
		default:
			return &WebErrorResult{http.StatusOK, ErrorInvalidParameter}
		}
	default:
		return &WebErrorResult{http.StatusOK, ErrorInvalidParameter}
	}
}

func putReportFunc(writer http.ResponseWriter, request *http.Request) *WebErrorResult {
	reqSlice := strings.Split(request.URL.Path[1:], "/")
	switch len(reqSlice) {
	case 3:
		parameter := reqSlice[2]
		switch parameter {
		case "putRpt":
			return putReportRequest(writer, request)
		case "putSchedule":
			return putScheduleRequest(writer, request)
		case "putWeekRpt":
			return putWeekRptRequest(writer, request)
		default:
			return &WebErrorResult{http.StatusOK, ErrorInvalidParameter}
		}
	default:
		return &WebErrorResult{http.StatusOK, ErrorInvalidParameter}
	}
}

func deleteReportFunc(writer http.ResponseWriter, request *http.Request) *WebErrorResult {
	reqSlice := strings.Split(request.URL.Path[1:], "/")
	switch len(reqSlice) {
	case 4:
		parameter := reqSlice[2]
		switch parameter {
		case "deleteRpt":
			return deleteReportRequest(reqSlice[3], writer, request)
		case "deleteWeekRpt":
			return deleteWeekRptRequest(reqSlice[3], writer, request)
		default:
			return &WebErrorResult{http.StatusOK, ErrorInvalidParameter}
		}
	default:
		return &WebErrorResult{http.StatusOK, ErrorInvalidParameter}
	}
}

func getReportListRequest(writer http.ResponseWriter, request *http.Request) *WebErrorResult {
	// fmt.Println("getReportListRequest 호출 확인")

	var err error
	parser := initFormParser(request)
	if parser == nil {
		fmt.Println("req 파싱 오류", parser)
		return &WebErrorResult{http.StatusOK, ErrorInvalidParameter}
	}

	RptList, err := getPageInfo(request.URL.RawQuery)
	if err != nil {
		fmt.Println("페이징처리 오류", err)
		return &WebErrorResult{http.StatusOK, ErrorInvalidParameter}
	}

	var totalCount int = 0 // 전체 리포트 갯수
	queryString := "SELECT COUNT(*) FROM bsmgReport"
	err = db.QueryRow(queryString).Scan(&totalCount)
	if err != nil {
		fmt.Println("가져온 totalCount 쿼리문 오류 : ", err)
		return &WebErrorResult{http.StatusOK, ErrorInvalidParameter}
	}

	queryString = fmt.Sprintf(`SELECT COUNT(*) FROM 
			(SELECT ROW_NUMBER() OVER (ORDER BY rpt_date DESC) as ROWNUM, * FROM bsmgReport) r WHERE ROWNUM BETWEEN %d AND %d`, RptList.Offset+1, RptList.Limit+RptList.Offset)

	var count int = 0 // 페이지의 행 갯수 행 != column, 행 == row 헷갈ㄴㄴ

	err = db.QueryRow(queryString).Scan(&count)
	if err != nil {
		fmt.Println("가져온 데이터 갯수 오류 : ", err)
		return &WebErrorResult{http.StatusOK, ErrorInvalidParameter}
	}

	var result BsmgReportResult
	result.Result.ResultCode = 1
	result.TotalCount.Count = int32(totalCount)

	if count > 0 {
		result.ReportList = make([]*BsmgReportInfo, count)
	}

	queryString = fmt.Sprintf(`SELECT b.rpt_idx, b.rpt_title, b.rpt_content, b.mem_name, b.rpt_date, b.attr1_category FROM
			(SELECT ROW_NUMBER() OVER (ORDER BY rpt_date DESC) as ROWNUM, * FROM bsmgReport
			INNER JOIN bsmgMembers m ON m.mem_id = rpt_reporter
			INNER JOIN bsmgAttr1 ON attr1_idx = rpt_attr1) b 
			WHERE ROWNUM BETWEEN %d AND %d`, RptList.Offset+1, RptList.Limit+RptList.Offset)

	rows, err := db.Query(queryString)
	if err != nil {
		fmt.Println("조회 오류", err)
		return &WebErrorResult{http.StatusOK, ErrorInvalidParameter}
	}
	defer rows.Close()

	var ii = 0

	for rows.Next() {
		var rpt_idx sql.NullInt32
		var rpt_title sql.NullString
		var rpt_content sql.NullString
		var rpt_reporter sql.NullString
		var rpt_date sql.NullString
		var rpt_dateFormat string
		var rpt_attr1 sql.NullString

		err := rows.Scan(&rpt_idx, &rpt_title, &rpt_content, &rpt_reporter, &rpt_date, &rpt_attr1)
		if err != nil {
			fmt.Println("가져온 데이터 오류", err)
			return &WebErrorResult{http.StatusOK, ErrorInvalidParameter}
		}
		//            0123 4567 8901 23
		// rpt_date = 2022 0525 1101 19
		rpt_dateFormat = rpt_date.String[:4] + "-" + rpt_date.String[4:6] + "-" + rpt_date.String[6:8] + " " + rpt_date.String[8:10] + ":" + rpt_date.String[10:12]

		result.ReportList[ii] = &BsmgReportInfo{}
		result.ReportList[ii].Rpt_Idx = rpt_idx.Int32
		result.ReportList[ii].Rpt_title = rpt_title.String
		result.ReportList[ii].Rpt_content = rpt_content.String
		result.ReportList[ii].Rpt_Reporter = rpt_reporter.String
		result.ReportList[ii].Rpt_date = rpt_dateFormat
		result.ReportList[ii].Rpt_attr1 = rpt_attr1.String
		ii++
	}
	result.Result.ResultCode = 0
	responseMsg, _ := json.Marshal(result)
	sendResultEx(writer, request, http.StatusOK, responseMsg)
	return nil
}

func getReportSearchRequest(writer http.ResponseWriter, request *http.Request) *WebErrorResult {
	// fmt.Println("getReportSearchRequest 호출 확인")
	var err error
	var SearchCombo string
	var SearchInput string

	parser := initFormParser(request)
	if parser == nil {
		fmt.Println(parser)
		return &WebErrorResult{http.StatusOK, ErrorInvalidParameter}
	}

	RptList, err := getPageInfo(request.URL.RawQuery)
	if err != nil {
		fmt.Println("페이징처리 오류", err)
		return &WebErrorResult{http.StatusOK, ErrorInvalidParameter}
	}

	SearchCombo, err = parser.getStringValue(0, "search_combo", 0)
	if err != nil {
		fmt.Println(err)
		return &WebErrorResult{http.StatusOK, ErrorInvalidParameter}
	}
	SearchInput, err = parser.getStringValue(0, "search_input", 0)
	if err != nil {
		fmt.Println(err)
		return &WebErrorResult{http.StatusOK, ErrorInvalidParameter}
	}

	var totalCount int = 0 // 검색에 따른 전체 리포트 갯수
	var queryString string = "s"
	switch SearchCombo {
	case "0": // 전체
		queryString = fmt.Sprintf(`SELECT count(*) FROM bsmgReport
				INNER JOIN bsmgMembers ON mem_id = rpt_reporter
			 	WHERE rpt_title like '%%%s%%' or rpt_content like '%%%s%%' or 
			 	mem_name like '%%%s%%'`, SearchInput, SearchInput, SearchInput)
	case "1": // 제목
		queryString = fmt.Sprintf("SELECT count(*) FROM bsmgReport WHERE rpt_title like '%%%s%%'", SearchInput)
	case "2": // 내용
		queryString = fmt.Sprintf("SELECT count(*) FROM bsmgReport WHERE rpt_content like '%%%s%%'", SearchInput)
	case "3": // 보고자
		queryString = fmt.Sprintf(`SELECT count(*) FROM bsmgReport 
						INNER JOIN bsmgMembers ON mem_id = rpt_reporter
						WHERE mem_name like '%%%s%%'`, SearchInput)
	}

	err = db.QueryRow(queryString).Scan(&totalCount)
	if err != nil {
		fmt.Println("데이터 갯수 쿼리 오류", err)
		return &WebErrorResult{http.StatusOK, ErrorInvalidParameter}
	}

	var count int = 0 // 가져올 rptList
	switch SearchCombo {
	case "0": // 전체
		queryString = fmt.Sprintf(`SELECT COUNT(*) FROM 
					(SELECT ROW_NUMBER() OVER (ORDER BY rpt_date DESC) as ROWNUM, * FROM bsmgReport 
					INNER JOIN bsmgMembers m ON m.mem_id = rpt_reporter
					WHERE (rpt_title like '%%%s%%' or rpt_content like '%%%s%%' or m.mem_name like '%%%s%%')) r 
					  WHERE (ROWNUM BETWEEN %d AND %d) `, SearchInput, SearchInput, SearchInput, RptList.Offset+1, RptList.Limit+RptList.Offset)
	case "1": // 제목
		queryString = fmt.Sprintf(`SELECT COUNT(*) FROM 
					(SELECT ROW_NUMBER() OVER (ORDER BY rpt_date DESC) as ROWNUM, * FROM bsmgReport WHERE (rpt_title like '%%%s%%')) r 
					WHERE (ROWNUM BETWEEN %d AND %d)`, SearchInput, RptList.Offset+1, RptList.Limit+RptList.Offset)
	case "2": // 내용
		queryString = fmt.Sprintf(`SELECT COUNT(*) FROM 
					(SELECT ROW_NUMBER() OVER (ORDER BY rpt_date DESC) as ROWNUM, * FROM bsmgReport WHERE (rpt_content like '%%%s%%')) r 
					WHERE (ROWNUM BETWEEN %d AND %d) `, SearchInput, RptList.Offset+1, RptList.Limit+RptList.Offset)
	case "3": // 보고자
		queryString = fmt.Sprintf(`SELECT COUNT(*) FROM 
					(SELECT ROW_NUMBER() OVER (ORDER BY rpt_date DESC) as ROWNUM, * FROM bsmgReport 
					INNER JOIN bsmgMembers m ON m.mem_id = rpt_reporter
					WHERE (m.mem_name like '%%%s%%')) r 
					WHERE (ROWNUM BETWEEN %d AND %d)`, SearchInput, RptList.Offset+1, RptList.Limit+RptList.Offset)
	}
	err = db.QueryRow(queryString).Scan(&count)
	if err != nil {
		fmt.Println("데이터 갯수 쿼리 오류", err)
		return &WebErrorResult{http.StatusOK, ErrorInvalidParameter}
	}

	var result BsmgReportResult
	result.TotalCount.Count = int32(totalCount)
	result.Result.ResultCode = 1

	if count > 0 {
		result.ReportList = make([]*BsmgReportInfo, count)
	}

	switch SearchCombo {
	case "0": // 전체
		queryString = fmt.Sprintf(`SELECT r.rpt_idx, r.rpt_title, r.rpt_content, r.mem_name, r.rpt_date, r.attr1_category FROM 
				(SELECT ROW_NUMBER() OVER (ORDER BY rpt_date DESC) as ROWNUM, * FROM bsmgReport 
				INNER JOIN bsmgMembers m ON m.mem_id = rpt_reporter
				INNER JOIN bsmgAttr1 ON attr1_idx = rpt_attr1
				WHERE (rpt_title like '%%%s%%' or rpt_content like '%%%s%%' or m.mem_name like '%%%s%%')) r 
				WHERE (ROWNUM BETWEEN %d AND %d) `, SearchInput, SearchInput, SearchInput, RptList.Offset+1, RptList.Limit+RptList.Offset)
	case "1": // 제목
		queryString = fmt.Sprintf(`SELECT r.rpt_idx, r.rpt_title, r.rpt_content, r.mem_name, r.rpt_date, r.attr1_category FROM 
				(SELECT ROW_NUMBER() OVER (ORDER BY rpt_date DESC) as ROWNUM, * FROM bsmgReport 
				INNER JOIN bsmgMembers m ON m.mem_id = rpt_reporter
				INNER JOIN bsmgAttr1 ON attr1_idx = rpt_attr1
				WHERE (rpt_title like '%%%s%%')) r 
				WHERE (ROWNUM BETWEEN %d AND %d)`, SearchInput, RptList.Offset+1, RptList.Limit+RptList.Offset)
	case "2": // 내용
		queryString = fmt.Sprintf(`SELECT r.rpt_idx, r.rpt_title, r.rpt_content, r.mem_name, r.rpt_date, r.attr1_category FROM 
				(SELECT ROW_NUMBER() OVER (ORDER BY rpt_date DESC) as ROWNUM, * FROM bsmgReport 
				INNER JOIN bsmgMembers m ON m.mem_id = rpt_reporter
				INNER JOIN bsmgAttr1 ON attr1_idx = rpt_attr1
				WHERE (rpt_content like '%%%s%%')) r 
				WHERE (ROWNUM BETWEEN %d AND %d) `, SearchInput, RptList.Offset+1, RptList.Limit+RptList.Offset)
	case "3": // 보고자
		queryString = fmt.Sprintf(`SELECT r.rpt_idx, r.rpt_title, r.rpt_content, r.mem_name, r.rpt_date, r.attr1_category FROM 
				(SELECT ROW_NUMBER() OVER (ORDER BY rpt_date DESC) as ROWNUM, * FROM bsmgReport 
				INNER JOIN bsmgMembers m ON m.mem_id = rpt_reporter
				INNER JOIN bsmgAttr1 ON attr1_idx = rpt_attr1
				WHERE (m.mem_name like '%%%s%%')) r 
				WHERE (ROWNUM BETWEEN %d AND %d)`, SearchInput, RptList.Offset+1, RptList.Limit+RptList.Offset)
	}
	rows, err := db.Query(queryString)
	if err != nil {
		fmt.Println("조회 오류", err)
		return &WebErrorResult{http.StatusOK, ErrorInvalidParameter}
	}
	defer rows.Close()
	var ii = 0

	for rows.Next() {
		var rpt_idx sql.NullInt32
		var rpt_title sql.NullString
		var rpt_content sql.NullString
		var rpt_reporter sql.NullString
		var rpt_date sql.NullString
		var rpt_dateFormat string
		var rpt_attr1 sql.NullString

		err := rows.Scan(&rpt_idx, &rpt_title, &rpt_content, &rpt_reporter, &rpt_date, &rpt_attr1)
		if err != nil {
			fmt.Println("가져온 데이터 오류", err)
			return &WebErrorResult{http.StatusOK, ErrorInvalidParameter}
		}
		//            0123 4567 8901 23
		// rpt_date = 2022 0525 1101 19
		rpt_dateFormat = rpt_date.String[:4] + "-" + rpt_date.String[4:6] + "-" + rpt_date.String[6:8] + " " + rpt_date.String[8:10] + ":" + rpt_date.String[10:12]

		result.ReportList[ii] = &BsmgReportInfo{}
		result.ReportList[ii].Rpt_Idx = rpt_idx.Int32
		result.ReportList[ii].Rpt_title = rpt_title.String
		result.ReportList[ii].Rpt_content = rpt_content.String
		result.ReportList[ii].Rpt_Reporter = rpt_reporter.String
		result.ReportList[ii].Rpt_date = rpt_dateFormat
		result.ReportList[ii].Rpt_attr1 = rpt_attr1.String
		ii++
	}
	result.Result.ResultCode = 0
	responseMsg, _ := json.Marshal(result)
	sendResultEx(writer, request, http.StatusOK, responseMsg)
	return nil
}

func getReportAttrRequest(writer http.ResponseWriter, request *http.Request) *WebErrorResult {
	// fmt.Println("getReportAttrRequest 호출 확인")
	var err error
	var AttrValue int32
	var AttrCategory int32

	parser := initFormParser(request)
	if parser == nil {
		fmt.Println(parser)
		return &WebErrorResult{http.StatusOK, ErrorInvalidParameter}
	}

	RptList, err := getPageInfo(request.URL.RawQuery)
	if err != nil {
		fmt.Println("페이징처리 오류", err)
		return &WebErrorResult{http.StatusOK, ErrorInvalidParameter}
	}

	AttrValue, err = parser.getInt32Value(0, "attrValue", 0)
	if err != nil {
		fmt.Println(err)
		return &WebErrorResult{http.StatusOK, ErrorInvalidParameter}
	}
	AttrCategory, err = parser.getInt32Value(0, "attrCategory", 0)
	if err != nil {
		fmt.Println(err)
		return &WebErrorResult{http.StatusOK, ErrorInvalidParameter}
	}

	var totalCount int = 0
	var queryString string = "s"
	switch AttrCategory {
	case 0: // 업무속성1
		queryString = fmt.Sprintf(`SELECT COUNT(*) FROM bsmgReport WHERE rpt_attr1 = %d`, AttrValue)
	case 1: // 업무속성2
		queryString = fmt.Sprintf(`SELECT COUNT(*) FROM bsmgReport WHERE rpt_attr2 = %d`, AttrValue)
	}

	err = db.QueryRow(queryString).Scan(&totalCount)
	if err != nil {
		fmt.Println("데이터 갯수 쿼리 오류", err)
		return &WebErrorResult{http.StatusOK, ErrorInvalidParameter}
	}

	var count int = 0
	switch AttrCategory {
	case 0: // 업무속성1
		queryString = fmt.Sprintf(`SELECT COUNT(*) FROM 
			(SELECT ROW_NUMBER() OVER(ORDER BY rpt_date DESC) as ROWNUM, * FROM bsmgReport WHERE rpt_attr1 = %d) r
			  WHERE (ROWNUM BETWEEN %d AND %d)`, AttrValue, RptList.Offset+1, RptList.Limit+RptList.Offset)
	case 1: // 업무속성2
		queryString = fmt.Sprintf(`SELECT COUNT(*) FROM 
			(SELECT ROW_NUMBER() OVER(ORDER BY rpt_date DESC) as ROWNUM, * FROM bsmgReport WHERE rpt_attr2 = %d) r
		 	  WHERE (ROWNUM BETWEEN %d AND %d)`, AttrValue, RptList.Offset+1, RptList.Limit+RptList.Offset)
	}
	err = db.QueryRow(queryString).Scan(&count)
	if err != nil {
		fmt.Println("데이터 갯수 쿼리 오류", err)
		return &WebErrorResult{http.StatusOK, ErrorInvalidParameter}
	}

	var result BsmgReportResult
	result.Result.ResultCode = 1
	result.TotalCount.Count = int32(totalCount)
	if count > 0 {
		result.ReportList = make([]*BsmgReportInfo, count)
	}

	switch AttrCategory {
	case 0:
		queryString = fmt.Sprintf(`SELECT r.rpt_idx, r.rpt_title, r.rpt_content, r.mem_name, r.rpt_date, r.attr1_category FROM 
		(SELECT ROW_NUMBER() OVER(ORDER BY rpt_date DESC) as ROWNUM, * FROM bsmgReport 
		INNER JOIN bsmgMembers m ON m.mem_id = rpt_reporter
		INNER JOIN bsmgAttr1 ON attr1_idx = rpt_attr1
		WHERE rpt_attr1 = %d) r
		WHERE (ROWNUM BETWEEN %d AND %d)`, AttrValue, RptList.Offset+1, RptList.Limit+RptList.Offset)
	case 1:
		queryString = fmt.Sprintf(`SELECT r.rpt_idx, r.rpt_title, r.rpt_content, r.mem_name, r.rpt_date, r.attr1_category FROM 
		(SELECT ROW_NUMBER() OVER(ORDER BY rpt_date DESC) as ROWNUM, * FROM bsmgReport 
		INNER JOIN bsmgMembers m ON m.mem_id = rpt_reporter
		INNER JOIN bsmgAttr1 ON attr1_idx = rpt_attr1
		WHERE rpt_attr2 = %d) r
		  WHERE (ROWNUM BETWEEN %d AND %d)`, AttrValue, RptList.Offset+1, RptList.Limit+RptList.Offset)
	}
	rows, err := db.Query(queryString)
	if err != nil {
		fmt.Println("조회 오류", err)
		return &WebErrorResult{http.StatusOK, ErrorInvalidParameter}
	}
	defer rows.Close()
	var ii = 0

	for rows.Next() {
		var rpt_idx sql.NullInt32
		var rpt_title sql.NullString
		var rpt_content sql.NullString
		var rpt_reporter sql.NullString
		var rpt_date sql.NullString
		var rpt_dateFormat string
		var rpt_attr1 sql.NullString

		err := rows.Scan(&rpt_idx, &rpt_title, &rpt_content, &rpt_reporter, &rpt_date, &rpt_attr1)
		if err != nil {
			fmt.Println("가져온 데이터 오류", err)
			return &WebErrorResult{http.StatusOK, ErrorInvalidParameter}
		}
		rpt_dateFormat = rpt_date.String[:4] + "-" + rpt_date.String[4:6] + "-" + rpt_date.String[6:8] + " " + rpt_date.String[8:10] + ":" + rpt_date.String[10:12]

		result.ReportList[ii] = &BsmgReportInfo{}
		result.ReportList[ii].Rpt_Idx = rpt_idx.Int32
		result.ReportList[ii].Rpt_title = rpt_title.String
		result.ReportList[ii].Rpt_content = rpt_content.String
		result.ReportList[ii].Rpt_Reporter = rpt_reporter.String
		result.ReportList[ii].Rpt_date = rpt_dateFormat
		result.ReportList[ii].Rpt_attr1 = rpt_attr1.String
		ii++
	}

	result.Result.ResultCode = 0
	responsMsg, _ := json.Marshal(result)
	sendResultEx(writer, request, http.StatusOK, responsMsg)
	return nil
}

func getReportInfoRequest(writer http.ResponseWriter, request *http.Request) *WebErrorResult {
	// fmt.Println("getReportInfoRequest 호출 확인")

	var err error
	parser := initFormParser(request)
	if parser == nil {
		fmt.Println("parser오류", parser)
		return &WebErrorResult{http.StatusOK, ErrorInvalidParameter}
	}

	var rpt_idx int32
	rpt_idx, err = parser.getInt32Value(0, "rpt_idx", 0)
	if err != nil {
		fmt.Println("파싱에러", err)
		return &WebErrorResult{http.StatusOK, ErrorInvalidParameter}
	}

	var rpt_reporter sql.NullString
	var rpt_date sql.NullString
	var rpt_toRpt sql.NullString
	var rpt_ref sql.NullString
	var rpt_title sql.NullString
	var rpt_content sql.NullString
	var rpt_etc sql.NullString
	var rpt_attr1 sql.NullString
	var rpt_attr2 sql.NullString
	var rpt_confirm sql.NullBool

	var result BsmgReportResult
	result.Result.ResultCode = 1 // 중간에 끊기면 오류로 가게

	queryString := fmt.Sprintf(`SELECT r.rpt_idx, m1.mem_name, r.rpt_date, m2.mem_name, r.rpt_ref, r.rpt_title, 
					r.rpt_content, r.rpt_etc, ba1.attr1_category, ba2.attr2_name, r.rpt_confirm
					FROM bsmgReport r
				   INNER JOIN bsmgAttr1 ba1
	  				  ON ba1.attr1_idx = r.rpt_attr1
				   INNER JOIN bsmgAttr2 ba2
	  				  ON ba2.attr2_idx = r.rpt_attr2
					INNER JOIN bsmgMembers m1 ON m1.mem_id = r.rpt_reporter 
					INNER JOIN bsmgMembers m2 ON m2.mem_id = r.rpt_toRpt
   				   WHERE rpt_idx = %d`, rpt_idx)
	row := db.QueryRow(queryString)
	err = row.Scan(&rpt_idx, &rpt_reporter, &rpt_date, &rpt_toRpt, &rpt_ref, &rpt_title, &rpt_content, &rpt_etc, &rpt_attr1, &rpt_attr2, &rpt_confirm)
	if err != nil {
		fmt.Println("DB데이터 변수에 할당 중 오류", err)
		return &WebErrorResult{http.StatusOK, ErrorInvalidParameter}
	}

	result.ReportInfo = &BsmgReportInfo{}
	result.ReportInfo.Rpt_Idx = rpt_idx
	result.ReportInfo.Rpt_Reporter = rpt_reporter.String
	result.ReportInfo.Rpt_date = rpt_date.String
	result.ReportInfo.Rpt_toRpt = rpt_toRpt.String
	result.ReportInfo.Rpt_ref = rpt_ref.String
	result.ReportInfo.Rpt_title = rpt_title.String
	result.ReportInfo.Rpt_content = rpt_content.String
	result.ReportInfo.Rpt_etc = rpt_etc.String
	result.ReportInfo.Rpt_attr1 = rpt_attr1.String
	result.ReportInfo.Rpt_attr2 = rpt_attr2.String
	result.ReportInfo.Rpt_confirm = rpt_confirm.Bool
	result.Result.ResultCode = 0

	responseMsg, _ := json.Marshal(result)
	sendResultEx(writer, request, http.StatusOK, responseMsg)
	return nil
}

func getReportSchedule(writer http.ResponseWriter, request *http.Request) *WebErrorResult {
	// fmt.Println("getReportSchedule 호출 확인")

	var err error
	parser := initFormParser(request)
	if parser == nil {
		fmt.Println("parser오류", parser)
		return &WebErrorResult{http.StatusOK, ErrorInvalidParameter}
	}

	var rpt_idx int32
	rpt_idx, err = parser.getInt32Value(0, "rpt_idx", 0)
	if err != nil {
		fmt.Println("파싱에러", err)
		return &WebErrorResult{http.StatusOK, ErrorInvalidParameter}
	}

	var result BsmgReportResult
	result.Result.ResultCode = 1

	var count int = 0
	queryString := fmt.Sprintf(`SELECT COUNT(*) FROM bsmgSchedule WHERE rpt_idx = %d`, rpt_idx)
	err = db.QueryRow(queryString).Scan(&count)
	if err != nil {
		fmt.Println("데이터 갯수 쿼리 오류", err)
		return &WebErrorResult{http.StatusOK, ErrorInvalidParameter}
	}
	if count > 0 {
		result.ScheduleList = make([]*BsmgScheduleInfo, count)
	}

	queryString = fmt.Sprintf(`SELECT sc_content FROM bsmgSchedule WHERE rpt_idx = %d`, rpt_idx)
	rows, err := db.Query(queryString)
	if err != nil {
		fmt.Println("조회 오류", err)
		return &WebErrorResult{http.StatusOK, ErrorInvalidParameter}
	}
	defer rows.Close()

	var ii = 0

	for rows.Next() {
		var sc_content sql.NullString

		err := rows.Scan(&sc_content)
		if err != nil {
			fmt.Println("가져온 데이터 오류", err)
			return &WebErrorResult{http.StatusOK, ErrorInvalidParameter}
		}

		result.ScheduleList[ii] = &BsmgScheduleInfo{}
		result.ScheduleList[ii].Sc_Content = sc_content.String
		ii++
	}
	result.Result.ResultCode = 0

	responseMsg, _ := json.Marshal(result)
	sendResultEx(writer, request, http.StatusOK, responseMsg)
	return nil
}

func getWeekReportListRequest(writer http.ResponseWriter, request *http.Request) *WebErrorResult {
	var err error
	parser := initFormParser(request)
	if parser == nil {
		fmt.Println("req 파싱 오류", parser)
		return &WebErrorResult{http.StatusOK, ErrorInvalidParameter}
	}

	weekRptList, err := getPageInfo(request.URL.RawQuery)
	if err != nil {
		fmt.Println("페이징처리 오류", err)
		return &WebErrorResult{http.StatusOK, ErrorInvalidParameter}
	}

	var totalCount int = 0 // 전체 리포트 갯수
	queryString := "SELECT COUNT(*) FROM bsmgWeekRpt"
	err = db.QueryRow(queryString).Scan(&totalCount)
	if err != nil {
		fmt.Println("가져온 totalCount 쿼리문 오류 : ", err)
		return &WebErrorResult{http.StatusOK, ErrorInvalidParameter}
	}
	var result BsmgWeekRptResult
	result.Result.ResultCode = 1
	result.TotalCount.Count = int32(totalCount)

	var count int = 0 // 페이지의 행 갯수. 행 == row == 데이터 한줄. 열 == column
	queryString = fmt.Sprintf(`SELECT COUNT(*) FROM 
			(SELECT ROW_NUMBER() OVER (ORDER BY wRpt_idx DESC) as ROWNUM, * FROM bsmgWeekRpt) w WHERE ROWNUM 
			BETWEEN %d AND %d`, weekRptList.Offset+1, weekRptList.Limit+weekRptList.Offset)
	err = db.QueryRow(queryString).Scan(&count)
	if err != nil {
		fmt.Println("가져온 데이터 갯수 오류 : ", err)
		return &WebErrorResult{http.StatusOK, ErrorInvalidParameter}
	}

	if count > 0 {
		result.WeekReportList = make([]*BsmgWeekRptInfo, count)
	}

	queryString = fmt.Sprintf(`SELECT b.wRpt_idx, b.wRpt_title, b.wRpt_content, b.wRpt_toRpt, b.mem_name FROM
		(SELECT ROW_NUMBER() OVER (ORDER BY wRpt_idx DESC) as ROWNUM, * FROM bsmgWeekRpt
		INNER JOIN bsmgMembers m ON m.mem_id = wRpt_reporter
		) b 
		  WHERE ROWNUM BETWEEN %d AND %d`, weekRptList.Offset+1, weekRptList.Limit+weekRptList.Offset)
	rows, err := db.Query(queryString)
	if err != nil {
		fmt.Println("조회 오류", err)
		return &WebErrorResult{http.StatusOK, ErrorInvalidParameter}
	}
	defer rows.Close()

	var ii = 0
	for rows.Next() {
		var wRpt_idx sql.NullInt32
		var wRpt_title sql.NullString
		var wRpt_content sql.NullString
		var wRpt_reporter sql.NullString
		var wRpt_toRpt sql.NullString

		err := rows.Scan(&wRpt_idx, &wRpt_title, &wRpt_content, &wRpt_toRpt, &wRpt_reporter)
		if err != nil {
			fmt.Println("가져온 데이터 오류", err)
			return &WebErrorResult{http.StatusOK, ErrorInvalidParameter}
		}
		result.WeekReportList[ii] = &BsmgWeekRptInfo{}
		result.WeekReportList[ii].WRpt_Idx = wRpt_idx.Int32
		result.WeekReportList[ii].WRpt_Title = wRpt_title.String
		result.WeekReportList[ii].WRpt_Content = wRpt_content.String
		result.WeekReportList[ii].WRpt_Reporter = wRpt_reporter.String
		result.WeekReportList[ii].WRpt_ToRpt = wRpt_toRpt.String
		ii++
	}
	result.Result.ResultCode = 0
	responseMsg, _ := json.Marshal(result)
	sendResultEx(writer, request, http.StatusOK, responseMsg)
	return nil
}

func getWeekRptSearchRequest(writer http.ResponseWriter, request *http.Request) *WebErrorResult {
	var err error
	var SearchCombo string
	var SearchInput string

	parser := initFormParser(request)
	if parser == nil {
		fmt.Println(parser)
		return &WebErrorResult{http.StatusOK, ErrorInvalidParameter}
	}

	RptList, err := getPageInfo(request.URL.RawQuery)
	if err != nil {
		fmt.Println("페이징처리 오류", err)
		return &WebErrorResult{http.StatusOK, ErrorInvalidParameter}
	}

	SearchCombo, err = parser.getStringValue(0, "search_combo", 0)
	if err != nil {
		fmt.Println(err)
		return &WebErrorResult{http.StatusOK, ErrorInvalidParameter}
	}
	SearchInput, err = parser.getStringValue(0, "search_input", 0)
	if err != nil {
		fmt.Println(err)
		return &WebErrorResult{http.StatusOK, ErrorInvalidParameter}
	}

	var totalCount int = 0 // 검색에 따른 전체 리포트 갯수
	var queryString string = "s"
	switch SearchCombo {
	case "0": // 전체
		queryString = fmt.Sprintf(`SELECT count(*) FROM bsmgWeekRpt INNER JOIN bsmgMembers m ON m.mem_id = wRpt_reporter
				WHERE wRpt_title like '%%%s%%' or wRpt_content like '%%%s%%' or 
					  m.mem_name like '%%%s%%'`, SearchInput, SearchInput, SearchInput)
	case "1": // 제목
		queryString = fmt.Sprintf("SELECT count(*) FROM bsmgWeekRpt WHERE wRpt_title like '%%%s%%'", SearchInput)
	case "2": // 내용
		queryString = fmt.Sprintf("SELECT count(*) FROM bsmgWeekRpt WHERE wRpt_content like '%%%s%%'", SearchInput)
	case "4": // 보고자
		queryString = fmt.Sprintf(`SELECT count(*) FROM bsmgWeekRpt
							INNER JOIN bsmgMembers ON mem_id = wRpt_reporter
							WHERE (mem_name like '%%%s%%')`, SearchInput)
	}
	err = db.QueryRow(queryString).Scan(&totalCount)
	if err != nil {
		fmt.Println("데이터 갯수 쿼리 오류", err)
		return &WebErrorResult{http.StatusOK, ErrorInvalidParameter}
	}

	var count int = 0 // 가져올 rptList
	switch SearchCombo {
	case "0": // 전체
		queryString = fmt.Sprintf(`SELECT COUNT(*) FROM 
						(SELECT ROW_NUMBER() OVER (ORDER BY wRpt_idx DESC) as ROWNUM, * FROM bsmgWeekRpt 
						INNER JOIN bsmgMembers ON mem_id = wRpt_reporter
						WHERE (wRpt_title like '%%%s%%' or wRpt_content like '%%%s%%' or mem_name like '%%%s%%')) r 
						  WHERE (ROWNUM BETWEEN %d AND %d) `, SearchInput, SearchInput, SearchInput, RptList.Offset+1, RptList.Limit+RptList.Offset)
	case "1": // 제목
		queryString = fmt.Sprintf(`SELECT COUNT(*) FROM 
						(SELECT ROW_NUMBER() OVER (ORDER BY wRpt_idx DESC) as ROWNUM, * FROM bsmgWeekRpt  
						WHERE (wRpt_title like '%%%s%%')) r 
						WHERE (ROWNUM BETWEEN %d AND %d)`, SearchInput, RptList.Offset+1, RptList.Limit+RptList.Offset)
	case "2": // 보고대상
		queryString = fmt.Sprintf(`SELECT COUNT(*) FROM 
						(SELECT ROW_NUMBER() OVER (ORDER BY wRpt_idx DESC) as ROWNUM, * FROM bsmgWeekRpt 
						WHERE (wRpt_content like '%%%s%%')) r 
						WHERE (ROWNUM BETWEEN %d AND %d) `, SearchInput, RptList.Offset+1, RptList.Limit+RptList.Offset)
	case "4": // 보고자
		queryString = fmt.Sprintf(`SELECT COUNT(*) FROM 
						(SELECT ROW_NUMBER() OVER (ORDER BY wRpt_idx DESC) as ROWNUM, * FROM bsmgWeekRpt  
						INNER JOIN bsmgMembers ON mem_id = wRpt_reporter
						WHERE (mem_name like '%%%s%%')) r 
						WHERE (ROWNUM BETWEEN %d AND %d)`, SearchInput, RptList.Offset+1, RptList.Limit+RptList.Offset)
	}
	err = db.QueryRow(queryString).Scan(&count)
	if err != nil {
		fmt.Println("데이터 갯수 쿼리 오류", err)
		return &WebErrorResult{http.StatusOK, ErrorInvalidParameter}
	}

	var result BsmgWeekRptResult
	result.TotalCount.Count = int32(totalCount)
	result.Result.ResultCode = 1

	if count > 0 {
		result.WeekReportList = make([]*BsmgWeekRptInfo, count)
	}

	switch SearchCombo {
	case "0": // 전체
		queryString = fmt.Sprintf(`SELECT r.wRpt_idx,r.wRpt_title, r.wRpt_content, r.wRpt_toRpt, mem_name FROM 
						(SELECT ROW_NUMBER() OVER (ORDER BY wRpt_idx DESC) as ROWNUM, * FROM bsmgWeekRpt 
						INNER JOIN bsmgMembers ON mem_id = wRpt_reporter
						WHERE (wRpt_title like '%%%s%%' or wRpt_content like '%%%s%%' or mem_name like '%%%s%%')) r 
						  WHERE (ROWNUM BETWEEN %d AND %d) `, SearchInput, SearchInput, SearchInput, RptList.Offset+1, RptList.Limit+RptList.Offset)
	case "1": // 제목
		queryString = fmt.Sprintf(`SELECT r.wRpt_idx,r.wRpt_title, r.wRpt_content, r.wRpt_toRpt, mem_name FROM 
						(SELECT ROW_NUMBER() OVER (ORDER BY wRpt_idx DESC) as ROWNUM, * FROM bsmgWeekRpt  
						INNER JOIN bsmgMembers ON mem_id = wRpt_reporter
						WHERE (wRpt_title like '%%%s%%')) r 
						WHERE (ROWNUM BETWEEN %d AND %d)`, SearchInput, RptList.Offset+1, RptList.Limit+RptList.Offset)
	case "2": // 보고대상
		queryString = fmt.Sprintf(`SELECT r.wRpt_idx,r.wRpt_title, r.wRpt_content, r.wRpt_toRpt, mem_name FROM 
						(SELECT ROW_NUMBER() OVER (ORDER BY wRpt_idx DESC) as ROWNUM, * FROM bsmgWeekRpt 
						INNER JOIN bsmgMembers ON mem_id = wRpt_reporter
						WHERE wRpt_content like '%%%s%%') r 
						WHERE (ROWNUM BETWEEN %d AND %d) `, SearchInput, RptList.Offset+1, RptList.Limit+RptList.Offset)
	case "4": // 보고자
		queryString = fmt.Sprintf(`SELECT r.wRpt_idx,r.wRpt_title, r.wRpt_content, r.wRpt_toRpt, mem_name FROM 
						(SELECT ROW_NUMBER() OVER (ORDER BY wRpt_idx DESC) as ROWNUM, * FROM bsmgWeekRpt  
						INNER JOIN bsmgMembers ON mem_id = wRpt_reporter
						WHERE (mem_name like '%%%s%%')) r 
						WHERE (ROWNUM BETWEEN %d AND %d)`, SearchInput, RptList.Offset+1, RptList.Limit+RptList.Offset)
	}
	rows, err := db.Query(queryString)
	if err != nil {
		fmt.Println("조회 오류", err)
		return &WebErrorResult{http.StatusOK, ErrorInvalidParameter}
	}
	defer rows.Close()
	var ii = 0

	for rows.Next() {
		var wRpt_idx sql.NullInt32
		var wRpt_title sql.NullString
		var wRpt_content sql.NullString
		var wRpt_toRpt sql.NullString
		var wRpt_reporter sql.NullString

		err := rows.Scan(&wRpt_idx, &wRpt_title, &wRpt_content, &wRpt_toRpt, &wRpt_reporter)
		if err != nil {
			fmt.Println("가져온 데이터 오류", err)
			return &WebErrorResult{http.StatusOK, ErrorInvalidParameter}
		}

		result.WeekReportList[ii] = &BsmgWeekRptInfo{}
		result.WeekReportList[ii].WRpt_Idx = wRpt_idx.Int32
		result.WeekReportList[ii].WRpt_Title = wRpt_title.String
		result.WeekReportList[ii].WRpt_Content = wRpt_content.String
		result.WeekReportList[ii].WRpt_ToRpt = wRpt_toRpt.String
		result.WeekReportList[ii].WRpt_Reporter = wRpt_reporter.String
		ii++
	}
	result.Result.ResultCode = 0
	responseMsg, _ := json.Marshal(result)
	sendResultEx(writer, request, http.StatusOK, responseMsg)
	return nil
}

func getWeekRptCategory(writer http.ResponseWriter, request *http.Request) *WebErrorResult {
	var err error
	var PartValue int32

	parser := initFormParser(request)
	if parser == nil {
		fmt.Println(parser)
		return &WebErrorResult{http.StatusOK, ErrorInvalidParameter}
	}
	RptList, err := getPageInfo(request.URL.RawQuery)
	if err != nil {
		fmt.Println("페이징처리 오류", err)
		return &WebErrorResult{http.StatusOK, ErrorInvalidParameter}
	}
	PartValue, err = parser.getInt32Value(0, "part_value", 0)
	if err != nil {
		fmt.Println(err)
		return &WebErrorResult{http.StatusOK, ErrorInvalidParameter}
	}

	var totalCount int = 0
	var queryString string = "s"

	if PartValue == 0 {
		queryString = `SELECT COUNT(*) FROM bsmgWeekRpt`
	} else {
		queryString = fmt.Sprintf(`SELECT COUNT(*) FROM bsmgWeekRpt WHERE wRpt_part = %d`, PartValue)
	}

	err = db.QueryRow(queryString).Scan(&totalCount)
	if err != nil {
		fmt.Println("데이터 갯수 쿼리 오류", err)
		return &WebErrorResult{http.StatusOK, ErrorInvalidParameter}
	}

	var count int = 0
	if PartValue == 0 {
		queryString = fmt.Sprintf(`SELECT COUNT(*) FROM 
			(SELECT ROW_NUMBER() OVER (ORDER BY wRpt_idx DESC) as ROWNUM, * FROM bsmgWeekRpt) r 
			WHERE (ROWNUM BETWEEN %d AND %d)`, RptList.Offset+1, RptList.Offset+RptList.Limit)
	} else {
		queryString = fmt.Sprintf(`SELECT COUNT(*) FROM 
			(SELECT ROW_NUMBER() OVER (ORDER BY wRpt_idx DESC) as ROWNUM, * FROM bsmgWeekRpt  
			WHERE wRpt_part = %d) r 
			WHERE (ROWNUM BETWEEN %d AND %d)`, PartValue, RptList.Offset+1, RptList.Offset+RptList.Limit)
	}
	err = db.QueryRow(queryString).Scan(&count)
	if err != nil {
		fmt.Println("데이터 갯수 쿼리 오류2", err)
		return &WebErrorResult{http.StatusOK, ErrorInvalidParameter}
	}

	var result BsmgWeekRptResult
	result.Result.ResultCode = 1
	result.TotalCount.Count = int32(totalCount)
	if count > 0 {
		result.WeekReportList = make([]*BsmgWeekRptInfo, count)
	}

	if PartValue == 0 {
		queryString = fmt.Sprintf(`SELECT r.wRpt_idx,r.wRpt_title, r.wRpt_content, r.wRpt_toRpt, mem_name FROM
			(SELECT ROW_NUMBER() OVER (ORDER BY wRpt_idx DESC) as ROWNUM, * FROM bsmgWeekRpt  
			INNER JOIN bsmgMembers ON mem_id = wRpt_reporter) r 
			WHERE (ROWNUM BETWEEN %d AND %d)`, RptList.Offset+1, RptList.Offset+RptList.Limit)
	} else {
		queryString = fmt.Sprintf(`SELECT r.wRpt_idx,r.wRpt_title, r.wRpt_content, r.wRpt_toRpt, mem_name FROM
			(SELECT ROW_NUMBER() OVER (ORDER BY wRpt_idx DESC) as ROWNUM, * FROM bsmgWeekRpt  
			INNER JOIN bsmgMembers ON mem_id = wRpt_reporter WHERE wRpt_part = %d) r 
			WHERE (ROWNUM BETWEEN %d AND %d)`, PartValue, RptList.Offset+1, RptList.Offset+RptList.Limit)
	}
	rows, err := db.Query(queryString)
	if err != nil {
		fmt.Println("조회 오류", err)
		return &WebErrorResult{http.StatusOK, ErrorInvalidParameter}
	}
	defer rows.Close()
	var ii = 0
	for rows.Next() {
		var wRpt_idx sql.NullInt32
		var wRpt_title sql.NullString
		var wRpt_content sql.NullString
		var wRpt_toRpt sql.NullString
		var wRpt_reporter sql.NullString

		err := rows.Scan(&wRpt_idx, &wRpt_title, &wRpt_content, &wRpt_toRpt, &wRpt_reporter)
		if err != nil {
			fmt.Println("가져온 데이터 오류", err)
			return &WebErrorResult{http.StatusOK, ErrorInvalidParameter}
		}

		result.WeekReportList[ii] = &BsmgWeekRptInfo{}
		result.WeekReportList[ii].WRpt_Idx = wRpt_idx.Int32
		result.WeekReportList[ii].WRpt_Title = wRpt_title.String
		result.WeekReportList[ii].WRpt_Content = wRpt_content.String
		result.WeekReportList[ii].WRpt_ToRpt = wRpt_toRpt.String
		result.WeekReportList[ii].WRpt_Reporter = wRpt_reporter.String
		ii++
	}
	result.Result.ResultCode = 0
	responsMsg, _ := json.Marshal(result)
	sendResultEx(writer, request, http.StatusOK, responsMsg)
	return nil
}

func getWeekRptInfoRequest(writer http.ResponseWriter, request *http.Request) *WebErrorResult {
	var err error
	parser := initFormParser(request)
	if parser == nil {
		fmt.Println("parser오류", parser)
		return &WebErrorResult{http.StatusOK, ErrorInvalidParameter}
	}

	var wRpt_idx int32
	wRpt_idx, err = parser.getInt32Value(0, "wRpt_idx", 0)
	if err != nil {
		fmt.Println("파싱에러", err)
		return &WebErrorResult{http.StatusOK, ErrorInvalidParameter}
	}

	var wRpt_reporter sql.NullString
	var wRpt_date sql.NullString
	var wRpt_toRpt sql.NullString
	var wRpt_title sql.NullString
	var wRpt_content sql.NullString
	var wRpt_part sql.NullInt32
	var wRpt_omissionDate sql.NullString

	var result BsmgWeekRptResult
	result.Result.ResultCode = 1

	queryString := fmt.Sprintf(`SELECT m.mem_name, w.wRpt_date, w.wRpt_toRpt, w.wRpt_title, w.wRpt_content, w.wRpt_part, w.wRpt_omissionDate
					FROM bsmgWeekRpt w
   					INNER JOIN bsmgMembers m ON m.mem_id = w.wRpt_reporter 
					WHERE wRpt_idx = %d`, wRpt_idx)
	err = db.QueryRow(queryString).Scan(&wRpt_reporter, &wRpt_date, &wRpt_toRpt, &wRpt_title, &wRpt_content, &wRpt_part, &wRpt_omissionDate)
	if err != nil {
		fmt.Println("DB데이터 변수에 할당 중 오류", err)
		return &WebErrorResult{http.StatusOK, ErrorInvalidParameter}
	}

	result.WeekReportInfo = &BsmgWeekRptInfo{}
	result.WeekReportInfo.WRpt_Idx = wRpt_idx
	result.WeekReportInfo.WRpt_Reporter = wRpt_reporter.String
	result.WeekReportInfo.WRpt_Date = wRpt_date.String
	result.WeekReportInfo.WRpt_ToRpt = wRpt_toRpt.String
	result.WeekReportInfo.WRpt_Title = wRpt_title.String
	result.WeekReportInfo.WRpt_Content = wRpt_content.String
	result.WeekReportInfo.WRpt_Part = wRpt_part.Int32
	result.WeekReportInfo.WRpt_OmissionDate = wRpt_omissionDate.String
	result.Result.ResultCode = 0
	responseMsg, _ := json.Marshal(result)
	sendResultEx(writer, request, http.StatusOK, responseMsg)
	return nil
}

func confirmRpt(writer http.ResponseWriter, request *http.Request) *WebErrorResult {
	var err error
	parser := initFormParser(request)
	if parser == nil {
		fmt.Println("request 파싱 오류")
		return &WebErrorResult{http.StatusOK, ErrorInvalidParameter}
	}

	rptIdx, err := parser.getInt32Value(0, "rpt_idx", 0)
	if err != nil {
		fmt.Println(err)
		return &WebErrorResult{http.StatusOK, ErrorInvalidParameter}
	}

	queryString := fmt.Sprintf("UPDATE bsmgReport SET rpt_confirm = 1 WHERE rpt_idx = %d", rptIdx)
	_, err = db.Exec(queryString)
	if err != nil {
		fmt.Println(err)
		return &WebErrorResult{http.StatusOK, ErrorInvalidParameter}
	}

	var result OnlyResult
	result.Result.ResultCode = 0
	responseMsg, _ := json.Marshal(result)
	sendResultEx(writer, request, http.StatusOK, responseMsg)
	return nil
}

func postReportRequest(writer http.ResponseWriter, request *http.Request) *WebErrorResult {
	// fmt.Println("postReportRequest 호출 확인")
	if len(request.URL.Path) < 1 {
		return &WebErrorResult{http.StatusOK, ErrorInvalidParameter}
	}

	parser := initFormParser(request)
	if parser == nil {
		fmt.Println("parser 에러 ", parser)
		return &WebErrorResult{http.StatusOK, ErrorInvalidParameter}
	}

	var rpt_reporter string
	var rpt_date string
	var rpt_toRpt string
	var rpt_ref string
	var rpt_title string
	var rpt_content string
	var rpt_etc string
	var rpt_attr1 int32
	var rpt_attr2String string
	var rpt_attr2 int32
	var err error

	chkSess := chkSession(writer, request) // 세션을 통해 reporter확인
	rpt_reporter = chkSess.ID

	rpt_date, err = parser.getStringValue(0, "rpt_date", 0)
	if err != nil {
		fmt.Println(err)
		return &WebErrorResult{http.StatusOK, ErrorInvalidParameter}
	}
	rpt_toRpt, err = parser.getStringValue(0, "rpt_toRptID", 0)
	if err != nil {
		fmt.Println(err)
		return &WebErrorResult{http.StatusOK, ErrorInvalidParameter}
	}
	rpt_ref, err = parser.getStringValue(0, "rpt_ref", 0)
	if err != nil {
		fmt.Println(err)
		return &WebErrorResult{http.StatusOK, ErrorInvalidParameter}
	}
	rpt_title, err = parser.getStringValue(0, "rpt_title", 0)
	if err != nil {
		fmt.Println(err)
		return &WebErrorResult{http.StatusOK, ErrorInvalidParameter}
	}
	rpt_content, err = parser.getStringValue(0, "rpt_content", 0)
	if err != nil {
		fmt.Println(err)
		return &WebErrorResult{http.StatusOK, ErrorInvalidParameter}
	}
	rpt_etc, err = parser.getStringValue(0, "rpt_etc", 0)
	if err != nil {
		fmt.Println(err)
		return &WebErrorResult{http.StatusOK, ErrorInvalidParameter}
	} else if rpt_etc == "" {
		rpt_etc = " "
	}
	rpt_attr1, err = parser.getInt32Value(0, "rpt_attr1", 0)
	if err != nil {
		fmt.Println(err)
		return &WebErrorResult{http.StatusOK, ErrorInvalidParameter}
	}
	rpt_attr2String, err = parser.getStringValue(0, "rpt_attr2", 0)
	if err != nil {
		fmt.Println(err)
		return &WebErrorResult{http.StatusOK, ErrorInvalidParameter}
	}

	rpt_attr1 = rpt_attr1 - 1 // 트리메뉴를 구성하며 value에 +1을 해줬었으므로 실제 idx는 -1해줘야한다
	rpt_attr2 = getAttr2Idx(rpt_attr2String)
	// fmt.Println(rpt_reporter, rpt_date, rpt_toRpt, rpt_ref, rpt_title, rpt_content, rpt_attr1, rpt_attr2, rpt_etc)

	//DB 연동
	queryString := fmt.Sprintf(`INSERT INTO bsmgReport VALUES('%s', '%s', '%s', '%s', '%s', '%s',
			%d, %d, '%s', default)`, rpt_reporter, rpt_date, rpt_toRpt, rpt_ref, rpt_title, rpt_content, rpt_attr1, rpt_attr2, rpt_etc)

	_, err = db.Exec(queryString)
	if err != nil {
		fmt.Println("INSERT문 오류", err)
		return &WebErrorResult{http.StatusOK, ErrorInvalidParameter}
	}

	// 일정 추가를 위해 idx 받아오기
	var rpt_idx sql.NullInt32
	queryString = fmt.Sprintf(`SELECT top 1 rpt_idx FROM bsmgReport WHERE rpt_reporter = '%s'
		ORDER BY rpt_idx DESC`, rpt_reporter)
	err = db.QueryRow(queryString).Scan(&rpt_idx)
	if err != nil {
		fmt.Println("DB데이터 변수에 할당 중 오류", err)
		return &WebErrorResult{http.StatusOK, ErrorInvalidParameter}
	}

	var result BsmgReportResult
	result.Result.ResultCode = 0
	result.ReportInfo = &BsmgReportInfo{}
	result.ReportInfo.Rpt_Idx = rpt_idx.Int32

	responseMsg, _ := json.Marshal(result)
	sendResultEx(writer, request, http.StatusOK, responseMsg)

	return nil
}

func postScheduleRequest(writer http.ResponseWriter, request *http.Request) *WebErrorResult {
	fmt.Println("postReportRequest 호출 확인")
	if len(request.URL.Path) < 1 {
		return &WebErrorResult{http.StatusOK, ErrorInvalidParameter}
	}

	parser := initFormParser(request)
	if parser == nil {
		fmt.Println("parser 에러 ", parser)
		return &WebErrorResult{http.StatusOK, ErrorInvalidParameter}
	}

	// 일정에 관한 부분은 이곳에 작성

	var sc_content string
	var rpt_idxString string
	var err error

	// 다른 ds혹은 dm이면 index를 주의해서 넣어줘야한다. index : eXBuilder에서의 요청데이터 순서
	rpt_idxString, err = parser.getStringValue(1, "rpt_idx", 0)
	if err != nil {
		fmt.Println(err)
		return &WebErrorResult{http.StatusOK, ErrorInvalidParameter}
	}

	cnt, err := parser.getValueCount(0, "sc_content")
	if err != nil {
		fmt.Println(err)
		return &WebErrorResult{http.StatusOK, ErrorInvalidParameter}
	}
	fmt.Println("들어온 행수 (수정): ", cnt)
	rpt_idx, err := strconv.Atoi(rpt_idxString)
	if err != nil {
		fmt.Println(err)
		return &WebErrorResult{http.StatusOK, ErrorInvalidParameter}
	}
	for i := 0; i < cnt; i++ {
		// 일정 파싱 및 할당
		sc_content, err = parser.getStringValue(0, "sc_content", i)
		if err != nil {
			fmt.Println(err)
			return &WebErrorResult{http.StatusOK, ErrorInvalidParameter}
		}
		// fmt.Println("일정 : " + sc_content)
		queryString := fmt.Sprintf(" INSERT INTO bsmgSchedule VALUES(%d,'%s') ", rpt_idx, sc_content)
		_, err = db.Exec(queryString)
		if err != nil {
			fmt.Println(err)
			return &WebErrorResult{http.StatusOK, ErrorInvalidParameter}
		}
	}

	var result OnlyResult
	result.Result.ResultCode = 0
	responseMsg, _ := json.Marshal(result)
	sendResultEx(writer, request, http.StatusOK, responseMsg)
	return nil
}

func putReportRequest(writer http.ResponseWriter, request *http.Request) *WebErrorResult {
	fmt.Println("putReportRequest 호출 확인")
	parser := initFormParser(request)
	if parser == nil {
		fmt.Println("parser오류", parser)
		return &WebErrorResult{http.StatusOK, ErrorInvalidParameter}
	}

	var rpt_idx int32
	var rpt_title string
	var rpt_content string
	var rpt_etc string
	var rpt_attr1 int32
	var rpt_attr2String string
	var rpt_attr2 int32
	var err error

	rpt_idx, err = parser.getInt32Value(0, "rpt_idx", 0)
	if err != nil {
		fmt.Println(err)
		return &WebErrorResult{http.StatusOK, ErrorInvalidParameter}
	}
	rpt_title, err = parser.getStringValue(0, "rpt_title", 0)
	if err != nil {
		fmt.Println(err)
		return &WebErrorResult{http.StatusOK, ErrorInvalidParameter}
	}
	rpt_content, err = parser.getStringValue(0, "rpt_content", 0)
	if err != nil {
		fmt.Println(err)
		return &WebErrorResult{http.StatusOK, ErrorInvalidParameter}
	}
	rpt_etc, err = parser.getStringValue(0, "rpt_etc", 0)
	if err != nil {
		fmt.Println(err)
		return &WebErrorResult{http.StatusOK, ErrorInvalidParameter}
	}
	rpt_attr1, err = parser.getInt32Value(0, "rpt_attr1", 0)
	if err != nil {
		fmt.Println(err)
		return &WebErrorResult{http.StatusOK, ErrorInvalidParameter}
	}
	rpt_attr2String, err = parser.getStringValue(0, "rpt_attr2", 0)
	if err != nil {
		fmt.Println(err)
		return &WebErrorResult{http.StatusOK, ErrorInvalidParameter}
	}

	rpt_attr1 = rpt_attr1 - 1 // 트리 구성하며 +1해준거 다시 -1
	rpt_attr2 = getAttr2Idx(rpt_attr2String)

	queryString := fmt.Sprintf(`UPDATE bsmgReport SET rpt_title = '%s', rpt_content = '%s',
		rpt_etc = '%s', rpt_attr1 = %d, rpt_attr2 = %d 
		WHERE rpt_idx = %d`, rpt_title, rpt_content, rpt_etc, rpt_attr1, rpt_attr2, rpt_idx)
	_, err = db.Exec(queryString)
	if err != nil {
		fmt.Println("INSERT문 오류", err)
		return &WebErrorResult{http.StatusOK, ErrorInvalidParameter}
	}

	var result OnlyResult
	result.Result.ResultCode = 0
	responseMsg, _ := json.Marshal(result)
	sendResultEx(writer, request, http.StatusOK, responseMsg)
	return nil
}

func putScheduleRequest(writer http.ResponseWriter, request *http.Request) *WebErrorResult {
	fmt.Println("putSchedule 호출 확인")
	parser := initFormParser(request)
	if parser == nil {
		fmt.Println("parser 에러 ", parser)
		return &WebErrorResult{http.StatusOK, ErrorInvalidParameter}
	}

	var sc_content string
	var rpt_idx int32
	var err error
	var queryString string = ""

	rpt_idx, err = parser.getInt32Value(0, "rpt_idx", 0)
	if err != nil {
		fmt.Println(err)
		return &WebErrorResult{http.StatusOK, ErrorInvalidParameter}
	}

	queryString = fmt.Sprintf(`DELETE FROM bsmgSchedule WHERE rpt_idx = %d `, rpt_idx)
	_, err = db.Exec(queryString)
	if err != nil {
		fmt.Println(err)
		return &WebErrorResult{http.StatusOK, ErrorInvalidParameter}
	}

	cnt, err := parser.getValueCount(1, "sc_content")
	if err != nil {
		fmt.Println(err)
		return &WebErrorResult{http.StatusOK, ErrorInvalidParameter}
	}

	fmt.Println("들어온 행수 (수정): ", cnt)

	for i := 0; i < cnt; i++ {
		// 일정 파싱 및 할당
		sc_content, err = parser.getStringValue(1, "sc_content", i)
		if err != nil {
			fmt.Println(err)
			return &WebErrorResult{http.StatusOK, ErrorInvalidParameter}
		}
		// 일정은 update하기에는 유니크 키가 없으므로 삭제후 다시 생성으로..
		queryString = fmt.Sprintf(`INSERT INTO bsmgSchedule VALUES (%d, '%s') `, rpt_idx, sc_content)
		_, err = db.Exec(queryString)
		if err != nil {
			fmt.Println(err)
			return &WebErrorResult{http.StatusOK, ErrorInvalidParameter}
		}
	}
	var result OnlyResult
	result.Result.ResultCode = 0
	responseMsg, _ := json.Marshal(result)
	sendResultEx(writer, request, http.StatusOK, responseMsg)
	return nil
}

func putWeekRptRequest(writer http.ResponseWriter, request *http.Request) *WebErrorResult {
	var err error

	parser := initFormParser(request)
	if parser == nil {
		fmt.Println("parser오류", parser)
		return &WebErrorResult{http.StatusOK, ErrorInvalidParameter}
	}

	var wRpt_idx int32
	var wRpt_date string
	var wRpt_title string
	var wRpt_content string
	var wRpt_toRpt string
	var wRpt_part int32
	var wRpt_omissionDate string

	wRpt_idx, err = parser.getInt32Value(0, "wRpt_idx", 0)
	if err != nil {
		fmt.Println("파싱에러", err)
		return &WebErrorResult{http.StatusOK, ErrorInvalidParameter}
	}
	wRpt_date, err = parser.getStringValue(0, "wRpt_date", 0)
	if err != nil {
		fmt.Println(err)
		return &WebErrorResult{http.StatusOK, ErrorInvalidParameter}
	}
	wRpt_title, err = parser.getStringValue(0, "wRpt_title", 0)
	if err != nil {
		fmt.Println(err)
		return &WebErrorResult{http.StatusOK, ErrorInvalidParameter}
	}
	wRpt_content, err = parser.getStringValue(0, "wRpt_content", 0)
	if err != nil {
		fmt.Println(err)
		return &WebErrorResult{http.StatusOK, ErrorInvalidParameter}
	}
	wRpt_toRpt, err = parser.getStringValue(0, "wRpt_toRpt", 0)
	if err != nil {
		fmt.Println(err)
		return &WebErrorResult{http.StatusOK, ErrorInvalidParameter}
	}

	wRpt_part, err = parser.getInt32Value(0, "wRpt_part", 0)
	if err != nil {
		fmt.Println(err)
		return &WebErrorResult{http.StatusOK, ErrorInvalidParameter}
	}
	wRpt_omissionDate, err = parser.getStringValue(0, "wRpt_omissionDate", 0)
	if err != nil {
		fmt.Println(err)
		return &WebErrorResult{http.StatusOK, ErrorInvalidParameter}
	}

	queryString := fmt.Sprintf(`UPDATE bsmgWeekRpt SET  wRpt_date = '%s', wRpt_title = '%s', wRpt_content = N'%s', wRpt_toRpt = '%s',
		wRpt_part = %d, wRpt_omissionDate = '%s'
		WHERE wRpt_idx = %d`, wRpt_date, wRpt_title, wRpt_content, wRpt_toRpt, wRpt_part, wRpt_omissionDate, wRpt_idx)

	_, err = db.Exec(queryString)
	if err != nil {
		fmt.Println("INSERT문 오류", err)
		return &WebErrorResult{http.StatusOK, ErrorInvalidParameter}
	}

	var result OnlyResult
	result.Result.ResultCode = 0
	responseMsg, _ := json.Marshal(result)
	sendResultEx(writer, request, http.StatusOK, responseMsg)
	return nil
}

func deleteReportRequest(rptIdx string, writer http.ResponseWriter, request *http.Request) *WebErrorResult {
	fmt.Println("deleteReportRequest 호출 확인")

	// parser := initFormParser(request)
	// if parser == nil {
	// 	fmt.Println("request 파싱 오류")
	// 	return &WebErrorResult{http.StatusOK, ErrorInvalidParameter}
	// }

	var err error
	rptIdx_int, err := strconv.Atoi(rptIdx)
	if err != nil {
		fmt.Println(err)
		return &WebErrorResult{http.StatusOK, ErrorInvalidParameter}
	}
	// rpt_idx, err = parser.getInt32Value(0, "rpt_idx", 0)
	// if err != nil {
	// 	fmt.Println(err)
	// 	return &WebErrorResult{http.StatusOK, ErrorInvalidParameter}
	// }

	queryString := fmt.Sprintf("DELETE FROM bsmgSchedule WHERE rpt_idx = %d", rptIdx_int)
	_, err = db.Exec(queryString)
	if err != nil {
		fmt.Println(err)
		return &WebErrorResult{http.StatusOK, ErrorInvalidParameter}
	}

	queryString = fmt.Sprintf("DELETE FROM bsmgReport WHERE rpt_idx = %d", rptIdx_int)
	_, err = db.Exec(queryString)
	if err != nil {
		fmt.Println(err)
		return &WebErrorResult{http.StatusOK, ErrorInvalidParameter}
	}

	var result OnlyResult
	result.Result.ResultCode = 0
	responseMsg, _ := json.Marshal(result)
	sendResultEx(writer, request, http.StatusOK, responseMsg)
	return nil
}

func deleteWeekRptRequest(wRptIdx string, writer http.ResponseWriter, request *http.Request) *WebErrorResult {
	var err error
	wRptIdx_int, err := strconv.Atoi(wRptIdx)
	if err != nil {
		fmt.Println(err)
		return &WebErrorResult{http.StatusOK, ErrorInvalidParameter}
	}

	queryString := fmt.Sprintf("DELETE FROM bsmgWeekRpt WHERE wRpt_idx = %d", wRptIdx_int)
	_, err = db.Exec(queryString)
	if err != nil {
		fmt.Println(err)
		return &WebErrorResult{http.StatusOK, ErrorInvalidParameter}
	}

	var result OnlyResult
	result.Result.ResultCode = 0
	responseMsg, _ := json.Marshal(result)
	sendResultEx(writer, request, http.StatusOK, responseMsg)
	return nil
}

// 트리에서 가져온 value를 다시 attr2_idx로
func getAttr2Idx(rpt_attr2String string) int32 {
	idx := strings.Index(rpt_attr2String, "-")
	rpt_attr2String = rpt_attr2String[idx+1:]
	rpt_attr2, err := strconv.ParseInt(rpt_attr2String, 10, 32)
	if err != nil {
		return 0
	}

	return int32(rpt_attr2)
}

//
//
//
//
//
//
//-------------------------------------------------------페이징
func getPageInfo(query string) (PageInfo PageInfo, err error) {
	values, err := ParseQuery(query)
	if err != nil {
		return
	}

	var tempInt int64

	tempInt, err = strconv.ParseInt(values.Get("limit"), 10, 32)
	if err != nil {
		return
	}
	PageInfo.Limit = int32(tempInt)
	if PageInfo.Limit < 1 {
		err = errors.New("Limit must bigger than 0")
		return
	}

	tempInt, err = strconv.ParseInt(values.Get("offset"), 10, 32)
	if err != nil {
		return
	}
	PageInfo.Offset = int32(tempInt)

	return
}

// 쿼리를 파싱 (호출해주는 역할)
func ParseQuery(query string) (Values, error) {
	m := make(Values)
	err := parseQuery(m, query)
	return m, err
}

// 쿼리를 파싱 (실제 알고리즘)
func parseQuery(m Values, query string) (err error) {
	for query != "" {
		key := query
		if i := strings.IndexAny(key, "&;"); i >= 0 {
			key, query = key[:i], key[i+1:]
		} else {
			query = ""
		}
		if key == "" {
			continue
		}
		value := ""
		if i := strings.Index(key, "="); i >= 0 {
			key, value = key[:i], key[i+1:]
		}
		m[key] = append(m[key], value)
	}
	return err
}

type Values map[string][]string // 흠..?

func (v Values) Get(key string) string {
	if v == nil {
		return ""
	}
	vs := v[key]
	if len(vs) == 0 {
		return ""
	}
	return vs[0]
}

/*
	페이징 처리를 하려면?
	1. getPageInfo -> limit, offset 설정
	2. 테이블의 전체 갯수 알기 (totalCount)
	3. 모든 쿼리에서 ROW_NUMBER().... BETWEEN 설정
	4. result에 totalCount 넣어주기

*/
