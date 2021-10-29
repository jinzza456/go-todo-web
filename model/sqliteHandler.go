package model

import (
	"database/sql"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type sqliteHandler struct {
	db *sql.DB
}

func (s *sqliteHandler) GetTodos(sessionId string) []*Todo {
	// 인자로 sessionID 를 받음
	todos := []*Todo{} // 반환값을 가지고 있을 리스트
	rows, err := s.db.Query("SELECT id, name, completed, createdAt FROM todos WHERE sessionId=?", sessionId)
	//SELECT 로 데이터를 가져오고 , FROM 어디서 가져올 껀지
	//id, name, completed를 todos라는 테이블에서 가져옴
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	for rows.Next() { //다음 행이 없을때 까지 반복
		var todo Todo // 가져온 데이터를 담을 Todo 객체
		rows.Scan(&todo.ID, &todo.Name, &todo.Completed, &todo.CreatedAt)
		// 데이터를 읽어 온다
		todos = append(todos, &todo) // todos에 데이터를 저장한다
	}
	return todos
}

func (s *sqliteHandler) AddTodo(name string, sessionId string) *Todo {
	// 인자로 sessionID 도 같이 받음
	stmt, err := s.db.Prepare("INSERT INTO todos (sessionId, name, completed, createdAt) VALUES (?, ?, ?, datetime('now'))")
	//Prepare로 스테이트먼트를 만든다.
	//INSERT INTO todos: todos 테이블에 값을 추가한다.
	//컬럼 값에 sessionID 추가
	//Value값  ? 추가
	if err != nil {
		panic(err)
	}
	rst, err := stmt.Exec(sessionId, name, false)
	//sessionID 를 추가해줌
	if err != nil {
		panic(err)
	}
	id, _ := rst.LastInsertId() // 자동으로 추가된 id의 제일 마지막 값
	var todo Todo
	todo.ID = int(id) // 새로 만든 id는 int64 타입이기 때문에 바꿔줌
	todo.Name = name
	todo.Completed = false
	todo.CreatedAt = time.Now()
	return &todo
}

func (s *sqliteHandler) RemoveTodo(id int) bool {
	stmt, err := s.db.Prepare("DELETE FROM todos WHERE id=?")
	// Prepare 로 스테이트먼트 만들기
	// DELETE FROM todos: todos 테이블의 레코드를 삭제한다.
	// WHERE id=? : 특정 id의 레코드 값만 삭제한다.
	if err != nil {
		panic(err)
	}
	rst, err := stmt.Exec(id)
	if err != nil {
		panic(err)
	}
	cnt, _ := rst.RowsAffected()
	// RowsAffected : 영향 받은 레코드가 있는지 없는지의 여부
	// 업데이트 된 레코드 갯수
	return cnt > 0
}

func (s *sqliteHandler) CompleteTodo(id int, complete bool) bool {
	stmt, err := s.db.Prepare("UPDATE todos SET completed = ? WHERE id=?")
	// Prepare 로 스테이트먼트 만들기
	// UPDATE todos: todos의 값을 변경
	// SET completed: completed 항목을 변경
	// = ?  : 어떻게 업데이트 할것이냐
	// WHERE id=? : 인자로 받은 id를 특정함
	if err != nil {
		panic(err)
	}
	rst, err := stmt.Exec(complete, id)
	if err != nil {
		panic(err)
	}
	cnt, _ := rst.RowsAffected()
	// RowsAffected : 영향 받은 레코드가 있는지 없는지의 여부
	// 업데이트 된 레코드 갯수
	return cnt > 0
	//업데이트된 레코드가 있으면 true
}

// 데이터 베이스를 열면 사라지기전에 닫아줘야함
func (s *sqliteHandler) Close() {
	s.db.Close()
}

func newSqliteHandler(filepath string) DBHandler {
	database, err := sql.Open("sqlite3", filepath)
	if err != nil {
		panic(err)
	}
	statement, _ := database.Prepare(
		`CREATE TABLE IF NOT EXISTS todos (
			id        INTEGER  PRIMARY KEY AUTOINCREMENT, 
			sessionId STRING,
			name      TEXT,
			completed BOOLEAN,
			createdAt DATETIME
		);
		CREATE INDEX IF NOT EXISTS sessionIdIndexOnTodos ON todos (
			sessionId ASC
			)`)
	// 테이블에 sessionID 를 추가
	// GET 할때 빠른 속도를 위해 바이너리 서치 트리(BST) 키 인덱스를 만들어준다.
	statement.Exec()
	return &sqliteHandler{db: database}
}
