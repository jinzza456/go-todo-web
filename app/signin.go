package app

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

// 구글에서 보내준 세션정보 형태의 구조체
type GoogleUserId struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email"`
	Picture       string `json:"picture"`
}

// oauth2 패키지를 이용해 config 생성
var googleOauthConfig = oauth2.Config{
	RedirectURL: "http://localhost:3000/auth/google/callback",
	//요청한 콜백을 받을 주소 등록
	ClientID: os.Getenv("GOOGLE_CLIENT_ID"),
	// 환경 변수에서 클라이언트 아이디 가져오기
	ClientSecret: os.Getenv("GOOGLE_SECRET_KEY"),
	// 환경 변수에서 비밀번호 가져오기
	Scopes: []string{"https://www.googleapis.com/auth/userinfo.email"},
	// userinfo.email에 억세스 하는 권한 요청하기
	Endpoint: google.Endpoint,
	// 구글 로그인 요청을 받으면 config를 이용해서
	// 구글의 어떤 경로로 보내야 되는지 (Endpoint)가 나옴
}

// 유저가 EndPoint에 직접 접근해서 로그인 할 수 있게 리다이렉트 해줘야됨
func googleLoginHandler(w http.ResponseWriter, r *http.Request) {
	state := generateStateOauthCookie(w)
	// 경로가 만들어지고 리다이렉트로 calback이 왔을때 키가 맞는지 확이 할 수 있어야됨
	// 유저의 브라우저 쿠키에다가 tmep키를 심고 리다이렉트로 calback이 왔을때 쿠키를 비교하는 방식
	url := googleOauthConfig.AuthCodeURL(state)
	// 어떤 경로로 보내야 되는지 알려줌
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
	// w, r , 리다이렉트 경로 , 왜리다이렉트하는지
	// 유저가 해당경로로 리다이렉트 되고 구글 로그인 폼이 뜬다.
}

func generateStateOauthCookie(w http.ResponseWriter) string {
	expiration := time.Now().Add(1 * 24 * time.Hour)
	// 쿠키 만료시간 설정
	// 현재 시간으로 부터 하루 뒤에 만료
	b := make([]byte, 16) //16바이트 어레이
	rand.Read(b)          // 런덤 숫자로 어레이 채우기
	state := base64.URLEncoding.EncodeToString(b)
	// html 인코딩 하여 스트링으로 바꾸기
	cookie := &http.Cookie{Name: "oauthstate", Value: state, Expires: expiration}
	// 쿠키에 인코딩한 값을 넣어줌
	http.SetCookie(w, cookie)
	// 쿠키를 w에 넘겨줌
	return state
}

func googleAuthCallback(w http.ResponseWriter, r *http.Request) {
	oauthstate, _ := r.Cookie("oauthstate") //쿠키를 읽어옴

	if r.FormValue("state") != oauthstate.Value {
		// 쿠키 값과 state의 값이 다를 경우
		errMsg := fmt.Sprintf("invalid google oauth state cookie: %s state:%s\n", oauthstate.Value, r.FormValue("state"))
		// 잘못된 요청
		log.Printf(errMsg)
		// 로그 남기기
		http.Error(w, errMsg, http.StatusInternalServerError)
		// 에러를 알려줌
	}
	data, err := getGoogleUserInfo(r.FormValue("code"))
	// 코드를 구글에 요청해서 userinfo를 가져온다.
	if err != nil {
		log.Println(err.Error()) // log와 에러를 찍음
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		// 기본 경로로 리다이렉트
		return
	}
	// 세션 정보를 쿠키에 집어넣기
	var userinfo GoogleUserId             // 세션정보를 저장할 객체 선언
	err = json.Unmarshal(data, &userinfo) // 데이터를 언마샬해서 userinfo로 넘김
	if err != nil {
		log.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	session, err := store.Get(r, "session")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	session.Values["id"] = userinfo.ID
	err = session.Save(r, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
}

// 유저정보를 리퀘스트 하는 경로
const oauthGoogleUrlAPI = "https://www.googleapis.com/oauth2/v2/userinfo?access_token="

func getGoogleUserInfo(code string) ([]byte, error) {
	token, err := googleOauthConfig.Exchange(context.Background(), code)
	//config를 통해 토큰 받아오기
	//Exchange로 토큰과 코드 교환
	//context = 쓰래드간에 데이터를 주고받는 쓰래드세이프한 저장소
	if err != nil {
		return nil, fmt.Errorf("Failed to Exchange %s\n", err.Error())
		// 문자열로 에러를 만듦
		// 데이터를 못받았기 때문에 nil을 리턴
	}
	resp, err := http.Get(oauthGoogleUrlAPI + token.AccessToken)
	// 유저정보를 리퀘스트하는 경로에 엑세스 토큰을 붙임
	// GET을 통해 요청
	if err != nil {
		return nil, fmt.Errorf("Failed to Get UserInfo %s\n", err.Error())
	}
	return ioutil.ReadAll(resp.Body)
	// resp의 데이터 반환
}
