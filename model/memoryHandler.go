package model

import "time"

// 기존의 메모리 맵을 별도 스트럭터로 만듦
type memoryHandler struct {
	todoMap map[int]*Todo
}

// memoryHandler의 메서드
func (m *memoryHandler) GetTodos(sessionId string) []*Todo {
	list := []*Todo{}
	for _, v := range m.todoMap {
		// map의 인덱스 만큼 반복
		// 키는 받지않고 value만 받음
		list = append(list, v)
		// list에 v값을 어팬드
	}
	return list
}

func (m *memoryHandler) AddTodo(sessionId string, name string) *Todo {
	id := len(m.todoMap) + 1                   // map의 갯수만큼 id
	todo := &Todo{id, name, false, time.Now()} // 포인터형 구조체 Todo를 생성
	m.todoMap[id] = todo                       // todo의 데이터를 맵에 넣음
	return todo
}

func (m *memoryHandler) RemoveTodo(id int) bool {
	if _, ok := m.todoMap[id]; ok {
		delete(m.todoMap, id)
		return true
	}
	return false
}

func (m *memoryHandler) CompleteTodo(id int, complete bool) bool {
	if todo, ok := m.todoMap[id]; ok { // 맵에 있는지 확인
		todo.Completed = complete // 맵에 있는경우 변경
		return true
	}
	return false
}

func (m *memoryHandler) Close() {

}

func newMemoryHandler() DBHandler {
	m := &memoryHandler{}
	m.todoMap = make(map[int]*Todo)
	return m
}
