package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
)

type state struct {
	groupQueue            Queue
	nextTokenToBeConsumed int
	previousWritePdaId    string
}

type replica_json struct {
	Pda_code      string   `json:"pda_code"`
	Group_members []string `json:"group_members"`
	groupState    state
}

var s *Store

func (s *state) init() {
	s.groupQueue.init()
	s.nextTokenToBeConsumed = 0
	s.previousWritePdaId = "none"
}

func loadState(destinationPda *PushDownAutomata, sourcePda *PushDownAutomata) {
	copy(destinationPda.pda_queue.elements[0:], sourcePda.pda_queue.elements[0:])
	destinationPda.curr_State = sourcePda.curr_State
	destinationPda.curr_inputToken = sourcePda.curr_inputToken
	destinationPda.nextTokenToBeConsumed = sourcePda.nextTokenToBeConsumed
	destinationPda.pda_stack = sourcePda.pda_stack
	destinationPda.state = sourcePda.state
}

func getAllPdas(w http.ResponseWriter, r *http.Request) {

	var allpdas []string
	pdas := s.getAll()
	for _, pdavalue := range pdas {
		allpdas = append(allpdas, pdavalue.Name)
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	pdaslist, _ := json.Marshal(allpdas)
	w.Write(pdaslist)
}

func getPdaIsAccepted(w http.ResponseWriter, r *http.Request) {

	var pda *PushDownAutomata
	var msg string
	id := mux.Vars(r)["id"]
	pda = s.get(id)
	cookie, _ := r.Cookie("recentpdaid")
	cookieValue := cookie.Value
	if cookieValue == id || cookieValue == "none" {
		pda = s.get(id)
	} else {
		var oldPda *PushDownAutomata
		oldPda = s.get(cookieValue)
		loadState(pda, oldPda)
	}
	if pda.is_accepted() {
		msg = "\nPDA:" + id + " IS ACCEPTED"
	} else {
		msg = "\nPDA:" + id + " IS NOT ACCEPTED"
	}
	replicaGroup := s.getReplica(pda.groupId)
	replicaGroup.groupState.previousWritePdaId = id
	expire := time.Now().Add(20 * time.Minute)
	newCookie := http.Cookie{Name: "recentpdaid", Value: replicaGroup.groupState.previousWritePdaId, Path: "/", Expires: expire}
	http.SetCookie(w, &newCookie)
	// This is actually update opearation
	s.add_replica(pda.groupId, replicaGroup)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(msg))
}

func getPdaTopOfStack(w http.ResponseWriter, r *http.Request) {

	var pda *PushDownAutomata
	id := mux.Vars(r)["id"]
	pda = s.get(id)
	cookie, _ := r.Cookie("recentpdaid")
	cookieValue := cookie.Value
	if cookieValue == id || cookieValue == "none" {
		pda = s.get(id)
	} else {
		var oldPda *PushDownAutomata
		oldPda = s.get(cookieValue)
		loadState(oldPda, pda)
	}
	kstring := mux.Vars(r)["k"]
	k, _ := strconv.Atoi(kstring)

	replicaGroup := s.getReplica(pda.groupId)
	replicaGroup.groupState.previousWritePdaId = id
	expire := time.Now().Add(20 * time.Minute)
	newCookie := http.Cookie{Name: "recentpdaid", Value: replicaGroup.groupState.previousWritePdaId, Path: "/", Expires: expire}
	http.SetCookie(w, &newCookie)
	// This is actually update opearation
	s.add_replica(pda.groupId, replicaGroup)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("\nTop stack elements for PDA " + id + " (starting from left) : " + pda.peek(k)))
}

func getPdaStackLength(w http.ResponseWriter, r *http.Request) {

	var pda *PushDownAutomata
	id := mux.Vars(r)["id"]
	pda = s.get(id)
	cookie, _ := r.Cookie("recentpdaid")
	cookieValue := cookie.Value
	if cookieValue == id || cookieValue == "none" {
		pda = s.get(id)
	} else {
		var oldPda *PushDownAutomata
		oldPda = s.get(cookieValue)
		loadState(pda, oldPda)
	}

	replicaGroup := s.getReplica(pda.groupId)
	replicaGroup.groupState.previousWritePdaId = id
	expire := time.Now().Add(20 * time.Minute)
	newCookie := http.Cookie{Name: "recentpdaid", Value: replicaGroup.groupState.previousWritePdaId, Path: "/", Expires: expire}
	http.SetCookie(w, &newCookie)
	// This is actually update opearation
	s.add_replica(pda.groupId, replicaGroup)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("\nStack size is: " + strconv.Itoa(pda.pda_stack.size())))
}

func getPdaCurrentState(w http.ResponseWriter, r *http.Request) {

	var pda *PushDownAutomata
	id := mux.Vars(r)["id"]
	pda = s.get(id)
	cookie, _ := r.Cookie("recentpdaid")
	cookieValue := cookie.Value
	if cookieValue == id || cookieValue == "none" {
		pda = s.get(id)
	} else {
		var oldPda *PushDownAutomata
		oldPda = s.get(cookieValue)
		loadState(pda, oldPda)
	}

	replicaGroup := s.getReplica(pda.groupId)
	replicaGroup.groupState.previousWritePdaId = id
	expire := time.Now().Add(20 * time.Minute)
	newCookie := http.Cookie{Name: "recentpdaid", Value: replicaGroup.groupState.previousWritePdaId, Path: "/", Expires: expire}
	http.SetCookie(w, &newCookie)
	// This is actually update opearation
	s.add_replica(pda.groupId, replicaGroup)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("\nCurrent state of PDA:" + id + " is " + pda.current_state()))
}

func getPdaQueuedTokens(w http.ResponseWriter, r *http.Request) {

	var pda *PushDownAutomata
	id := mux.Vars(r)["id"]
	pda = s.get(id)
	cookie, _ := r.Cookie("recentpdaid")
	cookieValue := cookie.Value
	if cookieValue == id || cookieValue == "none" {
		pda = s.get(id)
	} else {
		var oldPda *PushDownAutomata
		oldPda = s.get(cookieValue)
		loadState(pda, oldPda)
	}

	replicaGroup := s.getReplica(pda.groupId)
	replicaGroup.groupState.previousWritePdaId = id
	expire := time.Now().Add(20 * time.Minute)
	newCookie := http.Cookie{Name: "recentpdaid", Value: replicaGroup.groupState.previousWritePdaId, Path: "/", Expires: expire}
	http.SetCookie(w, &newCookie)
	// This is actually update opearation
	s.add_replica(pda.groupId, replicaGroup)

	var val []byte = []byte("\nQueue tokens for PDA:" + id + " ")
	val = append(val, pda.queued_tokens()...)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(val)
}

func getPdaSnapshot(w http.ResponseWriter, r *http.Request) {
	var pda *PushDownAutomata
	id := mux.Vars(r)["id"]
	pda = s.get(id)
	cookie, _ := r.Cookie("recentpdaid")
	cookieValue := cookie.Value
	if cookieValue == id || cookieValue == "none" {
		pda = s.get(id)
	} else {
		var oldPda *PushDownAutomata
		oldPda = s.get(cookieValue)
		loadState(pda, oldPda)
	}

	replicaGroup := s.getReplica(pda.groupId)
	replicaGroup.groupState.previousWritePdaId = id

	expire := time.Now().Add(20 * time.Minute)
	newCookie := http.Cookie{Name: "recentpdaid", Value: replicaGroup.groupState.previousWritePdaId, Path: "/", Expires: expire}
	http.SetCookie(w, &newCookie)

	// This is actually update opearation
	s.add_replica(pda.groupId, replicaGroup)

	vstring := "\nSnapshot of PDA:" + id + "\nCurrent State: " +
		pda.current_state() + " Queued Tokens (left to right): " + string(pda.queued_tokens()) + " Top of Stack (left to right): "
	peekk, _ := strconv.Atoi(mux.Vars(r)["k"])
	var val []byte = []byte(vstring)
	val = append(val, pda.peek(peekk)...)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(val)
}

func putPdaCreate(w http.ResponseWriter, r *http.Request) {
	pda := new(PushDownAutomata)
	err := json.NewDecoder(r.Body).Decode(pda)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
	id := mux.Vars(r)["id"]
	p := s.get(id)
	if p != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("\nPDA with ID: " + id + " already exists in the system. Use new ID to create new PDA. Request ignored\n"))
	} else {
		s.add(id, pda)
		bodyBytes, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Fatal(err)
		}
		pda.open()
		pda.Specification = string(bodyBytes)
		pda.curr_State = pda.Start_state
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)
		w.Write([]byte("\nPDA " + pda.Name + " with ID: " + id + " created.\n"))
	}
}

func putPdaReset(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	pda := s.get(mux.Vars(r)["id"])
	var msg string
	if pda != nil {
		pda.reset()
		msg = "\nPDA with ID: " + id + " reset succesfully!\n"
	} else {
		msg = "\nPDA with ID: " + id + " not found.\n"
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	w.Write([]byte(msg))
}

func putPdaToken(w http.ResponseWriter, r *http.Request) {
	pos := mux.Vars(r)["position"]
	tokenString := mux.Vars(r)["tokens"]
	token := ([]byte(tokenString)[0])
	var pda *PushDownAutomata
	id := mux.Vars(r)["id"]
	pda = s.get(id)
	var msg string
	cookie, _ := r.Cookie("recentpdaid")
	cookieValue := cookie.Value
	//fmt.Println(cookieValue)
	//fmt.Println(id)
	if cookieValue == id || cookieValue == "none" {
		pda = s.get(id)
	} else {
		var oldPda *PushDownAutomata
		oldPda = s.get(cookieValue)
		//fmt.Println(oldPda)
		//fmt.Println(pda)
		loadState(pda, oldPda)
		//fmt.Println(oldPda)
		//fmt.Println(pda)
	}

	if pda != nil {
		positionInteger, _ := strconv.Atoi(pos)
		r, nextToBeConsumedToken := put_present_consume_token(pda, positionInteger, token)
		_, _ = strconv.Atoi(nextToBeConsumedToken)
		if r {
			msg = "\nToken '" + tokenString + "' for position '" + pos + "' was presented to the PDA:" + id + "\n"
		} else {
			msg = "\nToken for position '" + pos + "' had already presented to the PDA:" + id + ". Ignored. \n"
		}

	} else {
		msg = "\nPDA with ID: " + id + " not found.\n"
	}
	replicaGroup := s.getReplica(pda.groupId)
	replicaGroup.groupState.previousWritePdaId = id
	expire := time.Now().Add(20 * time.Minute)
	newCookie := http.Cookie{Name: "recentpdaid", Value: replicaGroup.groupState.previousWritePdaId, Path: "/", Expires: expire}
	http.SetCookie(w, &newCookie)
	// This is actually update opearation
	s.add_replica(pda.groupId, replicaGroup)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(msg))
}

func putPdaEos(w http.ResponseWriter, r *http.Request) {

	var pda *PushDownAutomata
	var msg string
	id := mux.Vars(r)["id"]
	pda = s.get(id)
	cookie, _ := r.Cookie("recentpdaid")
	cookieValue := cookie.Value
	if cookieValue == id || cookieValue == "none" {
		pda = s.get(id)
	} else {
		var oldPda *PushDownAutomata
		oldPda = s.get(cookieValue)
		loadState(pda, oldPda)
	}
	if pda != nil {
		pda.eos()
		msg = "\nEnded the token stream for PDA:" + id + "\n"
	} else {
		msg = "\nPDA with ID: " + id + " not found.\n"
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	w.Write([]byte(msg))
}

func putPdaClose(w http.ResponseWriter, r *http.Request) {
	var pda *PushDownAutomata
	id := mux.Vars(r)["id"]
	pda = s.get(id)
	var msg string
	if pda != nil {
		pda.close()
		msg = "\nPDA with ID: " + id + " closed.\n"
	} else {
		msg = "\nPDA with ID: " + id + " not found.\n"
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	w.Write([]byte(msg))
}

func deletePda(w http.ResponseWriter, r *http.Request) {
	var pda *PushDownAutomata
	id := mux.Vars(r)["id"]
	pda = s.get(id)
	var msg string
	if pda != nil {
		s.remove(id)
		msg = "\nPDA with ID: " + id + " removed succesfully!\n"
	} else {
		msg = "\nPDA with ID: " + id + " not found.\n"
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(msg))
}

func getAllReplicaGroups(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("\nReplicas list - \n"))
	for _, v := range s.getAllReplicas() {
		w.Write([]byte(v + "\n"))
	}
}

func putReplicaGroup(w http.ResponseWriter, r *http.Request) {

	var rp replica_json
	err := json.NewDecoder(r.Body).Decode(&rp)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
	rp.groupState.init()
	gid := mux.Vars(r)["gid"]
	p := s.getReplicaMembers(gid)
	if p != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("\nReplica group with ID: " + gid + " already exists in the system. Use new ID to create new group. Request ignored\n"))
	} else {
		// Create pda member if it does not exist
		for _, v := range rp.Group_members {

			p := s.get(v)
			if p == nil {
				pda := new(PushDownAutomata)
				json.Unmarshal([]byte(rp.Pda_code), pda)

				s.add(v, pda)
				pda.open()
				pda.Specification = rp.Pda_code
				pda.curr_State = pda.Start_state
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusAccepted)
				w.Write([]byte("\nPDA " + pda.Name + " with ID: " + v + " created.\n"))
			}

		}
		s.add_replica(gid, rp)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("\nReplica group - " + gid + " created"))
	}
}

func putResetReplicaGroup(w http.ResponseWriter, r *http.Request) {
	var pda *PushDownAutomata
	var rp replica_json
	gid := mux.Vars(r)["gid"]
	var m = s.getReplicaMembers(gid)
	for _, value := range m {
		pda = s.get(value)
		pda.reset()
	}
	// expire := time.Now().Add(20 * time.Minute)
	// newCookie := http.Cookie{Name: "recentpdaid", Value: "none", Path: "/", Expires: expire}
	// http.SetCookie(w, &newCookie)
	rp = s.getReplica(gid)
	// Reset is equivalent to init
	rp.groupState.init()

	expire := time.Now().Add(20 * time.Minute)
	newCookie := http.Cookie{Name: "recentpdaid", Value: rp.groupState.previousWritePdaId, Path: "/", Expires: expire}
	http.SetCookie(w, &newCookie)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("\nReplica group has been reset"))
}

func getMemberAddresses(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	gid := mux.Vars(r)["gid"]
	//fmt.Println(reflect.TypeOf(gid))
	var m = s.getReplicaMembers(gid)
	w.Write([]byte("\nGroup members are: \n"))
	for _, value := range m {
		w.Write([]byte(value + "\n"))
	}
}

func getRandomMemberAddress(w http.ResponseWriter, r *http.Request) {

	gid := mux.Vars(r)["gid"]
	var replicaGroup replica_json
	var m = s.getReplicaMembers(gid)
	rand.Seed(time.Now().Unix())
	var randomPdaAddress = m[rand.Intn(len(m))]
	replicaGroup = s.getReplica(gid)
	expire := time.Now().Add(20 * time.Minute)
	cookie := http.Cookie{Name: "recentpdaid", Value: replicaGroup.groupState.previousWritePdaId, Path: "/", Expires: expire}
	http.SetCookie(w, &cookie)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(randomPdaAddress))
}

func putCloseMembers(w http.ResponseWriter, r *http.Request) {
	var pda *PushDownAutomata
	gid := mux.Vars(r)["gid"]
	var m = s.getReplicaMembers(gid)
	for _, value := range m {
		pda = s.get(value)
		pda.close()
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("\nReplica group closed"))
}

func deleteReplicaGroup(w http.ResponseWriter, r *http.Request) {
	gid := mux.Vars(r)["gid"]
	s.removeReplica(gid)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("\nReplica " + gid + " deleted !"))
}

func putPdaToReplicaGroup(w http.ResponseWriter, r *http.Request) {

	keys, _ := r.URL.Query()["replica_group"]
	gid := (string(keys[0]))
	id := mux.Vars(r)["id"]
	var arr = s.getReplicaMembers(gid)

	arr = append(arr, id)
	v, _ := s.c_replica.Get(gid)
	value, _ := v.(replica_json)
	var rp1 replica_json
	rp1.Pda_code = value.Pda_code
	rp1.Group_members = arr
	rp1.groupState.init()
	s.add_replica(gid, rp1)

	//PDA adopting the JSON specification of the replica group
	var pda *PushDownAutomata
	pda = s.get(id)
	pda.pdaAdoptReplicaSpecification(value.Pda_code)
	pda.pdaUpdateGroup(gid)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("\nPDA with address " + id + " joined the replica group " + gid))
}

func getPdaSpecification(w http.ResponseWriter, r *http.Request) {
	var pda *PushDownAutomata
	id := mux.Vars(r)["id"]
	pda = s.get(id)
	var msg string
	msg = pda.pda_specification()
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(msg))
}

func getStateInformation(w http.ResponseWriter, r *http.Request) {

	var pda *PushDownAutomata
	id := mux.Vars(r)["id"]
	pda = s.get(id)

	consumed := make(map[int]string)
	accepted := make(map[int]string)

	for key, value := range pda.state.consumedTokens {

		consumed[key] = string([]byte{value})
	}

	for key, value := range pda.state.acceptedTokens {

		accepted[key] = string([]byte{value})
	}

	stateInfo := "ConsumedTokens: " + fmt.Sprint(consumed) +
		"\nAcceptedTokens: " + fmt.Sprint(accepted) +
		"\nNextTokenToBeConsumedPosition: " + strconv.Itoa(pda.state.nextTokenToBeConsumed) +
		"\nCurrentState: " + pda.state.currentState

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(stateInfo))

}

func main() {
	r := mux.NewRouter()
	s = new(Store)
	s.init()
	api := r.PathPrefix("").Subrouter()
	api.HandleFunc("/pdas/{id}/delete", deletePda).Methods(http.MethodDelete)
	api.HandleFunc("/pdas", getAllPdas).Methods(http.MethodGet)
	api.HandleFunc("/pdas/{id}/is_accepted", getPdaIsAccepted).Methods(http.MethodGet)
	api.HandleFunc("/pdas/{id}/stack/top/{k}", getPdaTopOfStack).Methods(http.MethodGet)
	api.HandleFunc("/pdas/{id}/stack/len", getPdaStackLength).Methods(http.MethodGet)
	api.HandleFunc("/pdas/{id}/state", getPdaCurrentState).Methods(http.MethodGet)
	api.HandleFunc("/pdas/{id}/tokens", getPdaQueuedTokens).Methods(http.MethodGet)
	api.HandleFunc("/pdas/{id}/snapshot/{k}", getPdaSnapshot).Methods(http.MethodGet)
	api.HandleFunc("/pdas/{id}", putPdaCreate).Methods(http.MethodPut)
	api.HandleFunc("/pdas/{id}/reset", putPdaReset).Methods(http.MethodPut)
	api.HandleFunc("/pdas/{id}/{tokens}/{position}", putPdaToken).Methods(http.MethodPut)
	api.HandleFunc("/pdas/{id}/eos", putPdaEos).Methods(http.MethodPut)
	api.HandleFunc("/pdas/{id}/close", putPdaClose).Methods(http.MethodPut)

	api.HandleFunc("/replica_pdas", getAllReplicaGroups).Methods(http.MethodGet)
	api.HandleFunc("/replica_pdas/{gid}", putReplicaGroup).Methods(http.MethodPut)
	api.HandleFunc("/replica_pdas/{gid}/reset", putResetReplicaGroup).Methods(http.MethodPut)
	api.HandleFunc("/replica_pdas/{gid}/members", getMemberAddresses).Methods(http.MethodGet)
	api.HandleFunc("/replica_pdas/{gid}/connect", getRandomMemberAddress).Methods(http.MethodGet)
	api.HandleFunc("/replica_pdas/{gid}/close", putCloseMembers).Methods(http.MethodPut)
	api.HandleFunc("/replica_pdas/{gid}/delete", deleteReplicaGroup).Methods(http.MethodDelete)
	api.HandleFunc("/pdas/{id}/join", putPdaToReplicaGroup).Methods(http.MethodPut)
	api.HandleFunc("/pdas/{id}/code", getPdaSpecification).Methods(http.MethodGet)
	api.HandleFunc("/pdas/{id}/c3state", getStateInformation).Methods(http.MethodGet)

	fmt.Println("Server started listening on port 8080. Press Ctl-C to stop the server.")
	log.Print(http.ListenAndServe(":8080", r))

}
