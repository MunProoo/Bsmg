package main

import (
	"fmt"
	"net/http"

	"github.com/gorilla/sessions"
)

var (
	key   = []byte("super-secret-key")
	store = sessions.NewCookieStore(key)
)

// 세션에 들어가는 값 (ID)
type sessionValue struct {
	ID         string
	RememberID int32
}

// 세션 컨트롤의 결과를 한 눈에 알아보기 위해 생성 (다른 파일에서 세션 새로 만들어서 보기 어려우니?)
type chkSess struct {
	Authenticated bool
	ID            string
	RememberedID  string
}

func secret(w http.ResponseWriter, r *http.Request) { // 사용자등록 버튼 혹시 눌릴 수 있으니 권한설정? 열람?
	session, _ := store.Get(r, "Member")

	// Check if user is authenticated
	if auth, ok := session.Values["authenticated"].(bool); !ok || !auth { // 값을 받는 auth, (타입이 맞는지? || 타입변환이 됐는지?) 확인하는 ok
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	fmt.Fprintln(w, "접근이 제한되어있습니다.")
}

func CreateSession(w http.ResponseWriter, r *http.Request, sessionValue sessionValue) {
	session, _ := store.Get(r, "Member")

	// Set session
	session.Values["authenticated"] = true
	session.Values["mem_id"] = sessionValue.ID

	// ID 기억하기에 체크 되어있다면. rememberID에 sessionValue.ID 기입
	var rememberedID string
	if sessionValue.RememberID == 1 {
		rememberedID = sessionValue.ID
	} else {
		rememberedID = ""
	}

	session.Values["rememberedID"] = rememberedID

	session.Save(r, w)
}

func chkSession(w http.ResponseWriter, r *http.Request) chkSess {
	session, _ := store.Get(r, "Member")

	if session.Values["mem_id"] == nil {
		session.Values["authenticated"] = false
		session.Values["mem_id"] = ""
		session.Values["rememberedID"] = ""
	}

	var chksess chkSess
	if auth, ok := session.Values["authenticated"].(bool); ok && auth {
		chksess.Authenticated = true
		chksess.ID = session.Values["mem_id"].(string)
	} else {
		chksess.Authenticated = false
	}
	chksess.RememberedID = session.Values["rememberedID"].(string)
	return chksess
}

func DeleteSession(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "Member")

	// Revoke member auth
	session.Values["authenticated"] = false
	session.Values["mem_id"] = ""
	session.Save(r, w)
}
