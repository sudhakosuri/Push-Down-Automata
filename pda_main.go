package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
)

type Stack struct {
	elements []byte
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

type PushDownAutomata struct {
	Name             string
	States           []string
	Accepting_states []string
	Transitions      [][]string
	Start_state      string
	//Eos              byte
	Eos string
	//Input_alphabet   []byte
	//Stack_alphabet   []byte
	Input_alphabet  []string
	Stack_alphabet  []string
	curr_State      string
	curr_inputToken string
	pda_stack       Stack
}

func isError(err error) bool {
	if err != nil {
		fmt.Println(err.Error())
	}
	return (err != nil)
}

func (pda *PushDownAutomata) open(jsondata []byte) bool {

	err := json.Unmarshal([]byte(jsondata), &pda)
	if !isError(err) {
		return true
	} else {
		// fmt.Println(pda.Name)
		// fmt.Println(pda.States)
		// fmt.Println(pda.Accepting_states)
		// fmt.Println(pda.Transitions)
		// fmt.Println(pda.Start_state)
		// fmt.Println(pda.Input_alphabet)
		// fmt.Println(pda.Stack_alphabet)
		// fmt.Println(pda.Eos)
		// fmt.Print(pda.Transitions)
		// fmt.Println("\n")
	}
	return false

}

func (pda *PushDownAutomata) current_state() string {

	return pda.curr_State

}

func (pda *PushDownAutomata) close() {

	// Manual Garbage Collection

}

func (pda *PushDownAutomata) reset() {

	pda.curr_State = pda.Start_state
	pda.pda_stack.init()

}

func (pda *PushDownAutomata) peek(k ...int) string {
	// Error handling required. What if length of stack is lesser than k
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
	if pda.curr_inputToken == pda.Eos {
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
	return false
}

func (pda *PushDownAutomata) put(token byte) int {

	eosByte := []byte(pda.Eos)[0]
	var transitionTaken int = -1
	if token == eosByte {
		if pda.pda_stack.is_empty() {
			fmt.Print("Input token string is rejected. \n")
			return transitionTaken
		}
		if pda.pda_stack.top() == eosByte {
			for i := 0; i < len(pda.Transitions); i++ {
				if pda.Transitions[i][2] != "" {
					if []byte(pda.Transitions[i][2])[0] == eosByte {
						pda.curr_State = pda.Transitions[i][3]
						transitionTaken = i
						break
					}
				}
			}
			pda.pda_stack.pop()
		}
		pda.eos()
		if pda.is_accepted() {
			fmt.Print("The input token is accepted. \n")
		} else {
			fmt.Print("Input token string is rejected. \n")
		}
	} else {
		for i := 0; i < len(pda.Transitions); i++ {
			if pda.Transitions[i][0] == pda.curr_State {
				if pda.Transitions[i][1] == "" {
					if pda.Transitions[i][4] == "" {
						if pda.pda_stack.top() == []byte(pda.Transitions[i][2])[0] {
							pda.pda_stack.pop()
						} else {
							//handle invalid case
							fmt.Println("Input token string is rejected.")
							os.Exit(1)
						}
					} else {
						pda.pda_stack.push([]byte(pda.Transitions[i][4])[0])
					}
					pda.curr_State = pda.Transitions[i][3]
					transitionTaken = i
					break
				} else {
					if []byte(pda.Transitions[i][1])[0] == token {
						if pda.Transitions[i][4] == "" {
							if pda.pda_stack.top() == []byte(pda.Transitions[i][2])[0] {
								pda.pda_stack.pop()
							} else {
								//handle invalid case
								fmt.Println("Input token string is rejected.")
								os.Exit(1)
							}
						} else {
							pda.pda_stack.push([]byte(pda.Transitions[i][4])[0])
						}
						pda.curr_State = pda.Transitions[i][3]
						transitionTaken = i
						break
					}
				}
			}
		}
	}

	return transitionTaken
}

func (pda *PushDownAutomata) eos() {
	// Reached end of input token
	pda.curr_inputToken = pda.Eos
	fmt.Println("Reached end of the input")
}

func main() {

	IPStream, JSONSpecFileName := "", ""

	if len(os.Args) < 2 {
		fmt.Println("Missing parameter, provide file name!")
		return
	}

	JSONSpecFileName = os.Args[1]
	if len(os.Args) > 2 {
		IPStream = os.Args[2]
	} else {
		// Input has to be read from stdin console
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("Enter the input stream: ")
		text, _ := reader.ReadString('\n')
		IPStream = strings.TrimSuffix(text, "\n")
	}

	s := PushDownAutomata{}

	jsondata, err := ioutil.ReadFile(JSONSpecFileName)
	if isError(err) {
		return
	}
	open_status := s.open(jsondata)
	if open_status == false {
		return
	}

	// validate the input string
	var subStr bool
	for i := 0; i < len(IPStream); i++ {
		subStr = false
		for j := 0; j < len(s.Input_alphabet); j++ {
			if strings.ContainsAny(s.Input_alphabet[j], string(IPStream[i])) {
				subStr = true
				break
			}
		}
		if subStr == false {
			fmt.Println(string(IPStream[i]) + " not in Input Alphabets Specifications")
			os.Exit(1)
		}
	}

	IPStream = IPStream + s.Eos

	ip := []byte(IPStream)

	s.reset()

	//hack
	s.put('c')

	for i := 0; i < len(ip); i++ {
		//fmt.Println(s.peek(4))
		s.put(ip[i])
	}

}
