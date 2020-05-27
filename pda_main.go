/*
Deterministic Push Down Automata
Developers:
Kaustubh Dhokte (NB97699)
Sudha Kosuri (AJ34459)

References:
Below websites were refered for stack implementation.
https://stackoverflow.com/questions/28541609/looking-for-reasonable-stack-implementation-in-golang
https://flaviocopes.com/golang-data-structure-stack/
*/
package main

import (
	"fmt"
	"strconv"
)

// Queue ...
type Queue struct {
	elements [50]byte
}

func (q *Queue) init() {
	for i := 0; i < len(q.elements); i++ {
		q.elements[i] = '#'
	}
}

func (q *Queue) insert(t byte, p int) {
	q.elements[p] = t
}

func (q *Queue) getElement(p int) byte {
	return q.elements[p]
}

func (q *Queue) is_empty() bool {
	if len(q.elements) == 0 {
		return true
	}
	return false
}

func (q *Queue) size() int {
	return len(q.elements)
}

func (q *Queue) kth_top_element(k int) byte {
	n := len(q.elements)
	return q.elements[n-k : n-(k-1)][0]
}

// Stack ... to hold PDA tokens
type Stack struct {
	elements []byte
}

type pdaState struct {
	consumedTokens 		  map[int]byte
	acceptedTokens 		  map[int]byte
	nextTokenToBeConsumed int
	currentState          string
}

func (state *pdaState) init() {
	state.consumedTokens = make(map[int]byte)
	state.acceptedTokens = make(map[int]byte)
}

// PushDownAutomata struct
type PushDownAutomata struct {
	Name                  string
	States                []string
	Accepting_states      []string
	Transitions           [][]string
	Start_state           string
	Eos                   string
	Input_alphabet        []string
	Stack_alphabet        []string
	curr_State            string
	curr_inputToken       string
	pda_stack             Stack
	pda_queue             Queue
	clock_tick            int
	nextTokenToBeConsumed int
	Specification         string
	groupId               string
	state                 pdaState
}

func (s *Stack) push(c byte) {
	s.elements = append(s.elements, c)
}

func (s *Stack) pop() byte {
	element := s.elements[len(s.elements)-1]
	s.elements = s.elements[:len(s.elements)-1]
	return element
}

func (s *Stack) top() byte {
	TopElement := s.elements[len(s.elements)-1]
	return TopElement
}

func (s *Stack) init() {
	s.elements = s.elements[len(s.elements):]
}

func (s *Stack) is_empty() bool {
	if len(s.elements) == 0 {
		return true
	}
	return false
}

func (s *Stack) size() int {
	return len(s.elements)
}

func (s *Stack) kth_top_element(k int) byte {
	n := len(s.elements)
	return s.elements[n-k : n-(k-1)][0]
}

func isError(err error) bool {
	if err != nil {
		fmt.Println(err.Error())
	}
	return (err != nil)
}

func (pda *PushDownAutomata) open() bool {
	pda.pda_stack.init()
	pda.pda_queue.init()
	pda.state.init()
	pda.nextTokenToBeConsumed = 0
	return true
}

func (pda *PushDownAutomata) queued_tokens() []byte {
	//returns the tokens that have been presented but not consumed yet
	var temp []byte

	nextConsumedToken := pda.nextTokenToBeConsumed

	for i := nextConsumedToken; i < len(pda.pda_queue.elements); i++ {
		if pda.pda_queue.elements[i] != '#' {
			temp = append(temp, pda.pda_queue.elements[i])
		}
	}
	return temp
}

func (pda *PushDownAutomata) current_state() string {

	return pda.curr_State
}

func (pda *PushDownAutomata) pda_specification() string {

	return pda.Specification
}

func (pda *PushDownAutomata) pdaAdoptReplicaSpecification(js string) {

	pda.Specification = js
}

func (pda *PushDownAutomata) pdaUpdateGroup(groupId string) {

	pda.groupId = groupId
}

func (pda *PushDownAutomata) close() {
	// Manual Garbage Collection
}

func (pda *PushDownAutomata) reset() {
	pda.clock_tick = 0
	pda.curr_State = pda.Start_state
	pda.pda_stack.init()
	pda.nextTokenToBeConsumed = 0
	pda.pda_queue.init()
}

func (pda *PushDownAutomata) peek(k ...int) string {
	var n int
	if len(k) > 0 {
		n = k[0]
	} else {
		n = 1
	}
	if n > pda.pda_stack.size() {
		n = pda.pda_stack.size()
	}
	var KTopElements []byte
	for i := 1; i < n+1; i++ {
		ith_top := pda.pda_stack.kth_top_element(i)
		KTopElements = append(KTopElements, ith_top)
	}
	return string(KTopElements)
}

func (pda *PushDownAutomata) is_accepted() bool {
	var flag int = 0
	if pda.pda_stack.is_empty() {
		for i := 0; i < len(pda.Accepting_states); i++ {
			if pda.curr_State == pda.Accepting_states[i] {
				flag = 1
				break
			}
		}
		if flag == 1 {
			return true
		} else {
			return false
		}
	} else {
		return false
	}
}

func put_present_consume_token(pda *PushDownAutomata, position int, token byte) (bool, string) {
	if pda.pda_queue.getElement(position) != '#' {
		return false, ""
	}
	pda.pda_queue.insert(token, position)
	pda.state.acceptedTokens[position] = token
	c := make(chan string)
	go consumeToken(c, pda, position, token)
	nextToBeConsumedToken := <-c
	return true, nextToBeConsumedToken
}

func (pda *PushDownAutomata) put(position int, token byte) int {

	var old_clock_tick int = pda.clock_tick
	eosByte := []byte(pda.Eos)[0]

	if pda.curr_State == pda.Start_state {
		pda.pda_stack.push(eosByte)
		for i := 0; i < len(pda.Transitions); i++ {
			if (pda.Transitions[i][0] == pda.Start_state) && ([]byte(pda.Transitions[i][4])[0] == eosByte) {
				pda.curr_State = pda.Transitions[i][3]
				pda.clock_tick = pda.clock_tick + 1
				pda.state.currentState = pda.curr_State
				break
			}
		}
	}
	for i := 0; i < len(pda.Transitions); i++ {
		if pda.Transitions[i][0] == pda.curr_State {
			if pda.Transitions[i][1] == "" {
				if pda.Transitions[i][4] == "" {
					if pda.pda_stack.top() == []byte(pda.Transitions[i][2])[0] {
						pda.pda_stack.pop()
					} else {
						pda.pda_stack.push([]byte(pda.Transitions[i][1])[0])
					}
				} else {
					pda.pda_stack.push([]byte(pda.Transitions[i][4])[0])
				}
				pda.curr_State = pda.Transitions[i][3]
				pda.clock_tick = pda.clock_tick + 1
				pda.state.currentState = pda.curr_State
				break
			} else {
				if []byte(pda.Transitions[i][1])[0] == token {
					if pda.Transitions[i][4] == "" {
						if pda.pda_stack.top() == []byte(pda.Transitions[i][2])[0] {
							pda.pda_stack.pop()
						} else {
							pda.pda_stack.push([]byte(pda.Transitions[i][1])[0])
						}
					} else {
						pda.pda_stack.push([]byte(pda.Transitions[i][4])[0])
					}
					pda.curr_State = pda.Transitions[i][3]
					pda.clock_tick = pda.clock_tick + 1
					pda.state.currentState = pda.curr_State
					break
				}
			}
		}
	}
	return (old_clock_tick - pda.clock_tick)
}

func consumeToken(c chan string, pda *PushDownAutomata, pos int, token byte) {

	var msg = "\nToken " + strconv.Itoa(pda.nextTokenToBeConsumed) + " consumed"
	if pos == pda.nextTokenToBeConsumed {
		pda.put(pos, token)
		pda.state.consumedTokens[pos] = token
		pda.nextTokenToBeConsumed = pda.nextTokenToBeConsumed + 1
		pda.state.nextTokenToBeConsumed = pda.nextTokenToBeConsumed
	}
	for {
		if pda.pda_queue.getElement(pda.nextTokenToBeConsumed) != '#' {
			pda.put(pda.nextTokenToBeConsumed, pda.pda_queue.getElement(pda.nextTokenToBeConsumed))
			pda.state.consumedTokens[pda.nextTokenToBeConsumed] = pda.pda_queue.getElement(pda.nextTokenToBeConsumed)
			pda.nextTokenToBeConsumed = pda.nextTokenToBeConsumed + 1
			pda.state.nextTokenToBeConsumed = pda.nextTokenToBeConsumed
		} else {
			msg = strconv.Itoa(pda.nextTokenToBeConsumed)
			break
		}
	}
	c <- msg
}

func (pda *PushDownAutomata) eos() {
	for i := 0; i < len(pda.Transitions); i++ {
		if pda.pda_stack.top() == []byte(pda.Eos)[0] {
			if pda.curr_State == pda.Transitions[i][0] {
				if pda.Transitions[i][2] == pda.Eos {
					pda.pda_stack.pop()
					pda.curr_State = pda.Transitions[i][3]
					pda.clock_tick = pda.clock_tick + 1
					pda.state.currentState = pda.curr_State
				}
			}
		}
	}
}
