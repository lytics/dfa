package dfa

import (
	"fmt"
	"strings"
	"sync"
	"testing"
)

func NewExec(seq []Letter) *Exec {
	return &Exec{
		Sequence: seq,
		cmu:      &sync.Mutex{},
	}
}

type Exec struct {
	Sequence  []Letter
	log       func(string)
	nextCount int
	lastCount int
	cmu       *sync.Mutex
}

func (e *Exec) Next() Letter {
	e.cmu.Lock()
	defer e.cmu.Unlock()
	l := e.Sequence[0]
	e.Sequence = e.Sequence[1:]
	e.nextCount++
	return l
}

func (e *Exec) Last() {
	e.cmu.Lock()
	defer e.cmu.Unlock()
	e.lastCount++
}

func (e *Exec) NextCount() int {
	e.cmu.Lock()
	defer e.cmu.Unlock()
	return e.nextCount
}

func (e *Exec) LastCount() int {
	e.cmu.Lock()
	defer e.cmu.Unlock()
	return e.lastCount
}

func Print(from State, via Letter, to State) {
	fmt.Printf("%20v + %20v -> %20v\n", from, via, to)
}

func TestStateToString(t *testing.T) {
	if State("starting").String() != "starting" {
		t.Fail()
	}
}

func TestLetterToString(t *testing.T) {
	if Letter("fatal").String() != "fatal" {
		t.Fail()
	}
}

func TestSimple(t *testing.T) {
	// States
	Registering := State("registering")
	Waiting := State("waiting")
	Running := State("running")
	Resending := State("resending")
	Exiting := State("exiting")
	Terminating := State("terminating")

	// Inputs
	RegisterSuccess := Letter("register-success")
	RegisterFailure := Letter("register-failure")
	ExitWanted := Letter("exit-wanted")
	GroupJoinSuccess := Letter("group-join-success")
	GroupJoinFailure := Letter("group-join-failure")
	SendSuccess := Letter("send-success")
	SendFailure := Letter("send-failure")
	DoneSuccess := Letter("done-success")

	e := NewExec([]Letter{
		RegisterSuccess,
		GroupJoinSuccess,
		SendSuccess,
		SendSuccess,
		SendFailure,
		SendFailure,
		SendSuccess,
		ExitWanted,
	})

	d := New()
	d.SetStartState(Registering)
	d.SetTerminalStates(Exiting, Terminating)

	d.SetTransition(Registering, RegisterSuccess, Waiting, e.Next)
	d.SetTransition(Registering, RegisterFailure, Exiting, e.Last)
	d.SetTransition(Registering, ExitWanted, Exiting, e.Last)

	d.SetTransition(Waiting, GroupJoinSuccess, Running, e.Next)
	d.SetTransition(Waiting, GroupJoinFailure, Exiting, e.Last)
	d.SetTransition(Waiting, ExitWanted, Exiting, e.Last)

	d.SetTransition(Running, SendSuccess, Running, e.Next)
	d.SetTransition(Running, SendFailure, Resending, e.Next)
	d.SetTransition(Running, ExitWanted, Exiting, e.Last)
	d.SetTransition(Running, DoneSuccess, Terminating, e.Last)

	d.SetTransition(Resending, SendSuccess, Running, e.Next)
	d.SetTransition(Resending, SendFailure, Resending, e.Next)
	d.SetTransition(Resending, ExitWanted, Exiting, e.Last)

	final, accepted := d.Run(e.Next)
	if !accepted {
		t.Fatalf("failed to recognize Exiting or Terminating as terminal state")
	}

	if final != Exiting {
		t.Fatalf("final state should have been: %v, but was: %v", Exiting, final)
	}

	if e.NextCount() != 8 {
		t.Fatalf("NexCount() expected to be 8, got %d", e.NextCount())
	}

	if e.LastCount() != 1 {
		t.Fatalf("LastCount() expected 1, got %d", e.LastCount())
	}
}

func TestLastState(t *testing.T) {
	// States
	Starting := State("staring")
	// Inputs
	Transition := Letter("transition")

	e := NewExec([]Letter{})

	d := New()
	d.SetStartState(Starting)
	d.SetTerminalStates(Starting)
	d.SetTransition(Starting, Transition, Starting, e.Last)

	// The expectation is that the Next nor Last method will
	// ever be called since the DFA's starting and terminal
	// state are the same.
	final, accepted := d.Run(e.Next)
	if !accepted {
		t.Fatalf("failed to recognize Starting as terminal state")
	}
	if final != Starting {
		t.Fatalf("final state should have been: %v, but was: %v", Starting, final)
	}

	if e.NextCount() != 0 {
		t.Fail()
	}
	if e.LastCount() != 0 {
		t.Fail()
	}
}

func TestGraphViz(t *testing.T) {

	// States
	Starting := State("starting")
	Running := State("running")
	Resending := State("resending")
	Finishing := State("finishing")
	Exiting := State("exiting")
	Terminating := State("terminating")
	// Letters
	Failure := Letter("failure")
	SendFailure := Letter("send-failure")
	SendSuccess := Letter("send-success")
	EverybodyStarted := Letter("everybody-started")
	EverybodyFinished := Letter("everybody-finished")
	ProducersFinished := Letter("producers-finished")
	Exit := Letter("exit")

	d := New()
	d.SetStartState(Starting)
	d.SetTerminalStates(Exiting, Terminating)

	next := func() Letter { return Exit }
	exit := func() {}

	d.SetTransition(Starting, EverybodyStarted, Running, next)
	d.SetTransition(Starting, Failure, Exiting, exit)
	d.SetTransition(Starting, Exit, Exiting, exit)

	d.SetTransition(Running, SendFailure, Resending, next)
	d.SetTransition(Running, ProducersFinished, Finishing, next)
	d.SetTransition(Running, Failure, Exiting, exit)
	d.SetTransition(Running, Exit, Exiting, exit)

	d.SetTransition(Resending, SendSuccess, Running, next)
	d.SetTransition(Resending, SendFailure, Resending, next)
	d.SetTransition(Resending, Failure, Exiting, exit)
	d.SetTransition(Resending, Exit, Exiting, exit)

	d.SetTransition(Finishing, EverybodyFinished, Terminating, exit)
	d.SetTransition(Finishing, Failure, Exiting, exit)
	d.SetTransition(Finishing, Exit, Exiting, exit)

	viz := d.GraphViz()

	// Should be something like below, but order of
	// nodes is not consistent:
	//
	// digraph {
	//     "resending" -> "resending"[label="send-failure"];
	//     "starting" -> "running"[label="everybody-started"];
	//     "starting" -> "exiting"[label="failure"];
	//     "running" -> "resending"[label="send-failure"];
	//     "resending" -> "exiting"[label="failure"];
	//     "finishing" -> "exiting"[label="failure"];
	//     "running" -> "finishing"[label="producers-finished"];
	//     "running" -> "exiting"[label="failure"];
	//     "resending" -> "running"[label="send-success"];
	//     "finishing" -> "terminating"[label="everybody-finished"];
	//     "finishing" -> "exiting"[label="exit"];
	//     "starting" -> "exiting"[label="exit"];
	//     "running" -> "exiting"[label="exit"];
	//     "resending" -> "exiting"[label="exit"];
	// }

	if !strings.Contains(viz, `"resending" -> "resending"[label="send-failure"];`) {
		t.Fatalf("expected string: `%v`", `"resending" -> "resending"[label="send-failure"];`)
	}

	if !strings.Contains(viz, `"finishing" -> "exiting"[label="exit"];`) {
		t.Fatalf("expected string: `%v`", `"finishing" -> "exiting"[label="exit"];`)
	}
}

func TestRunSimple(t *testing.T) {
	// States
	Registering := State("registering")
	Waiting := State("waiting")
	Running := State("running")
	Resending := State("resending")
	Exiting := State("exiting")
	Terminating := State("terminating")

	// Inputs
	RegisterSuccess := Letter("register-success")
	RegisterFailure := Letter("register-failure")
	ExitWanted := Letter("exit-wanted")
	GroupJoinSuccess := Letter("group-join-success")
	GroupJoinFailure := Letter("group-join-failure")
	SendSuccess := Letter("send-success")
	SendFailure := Letter("send-failure")
	DoneSuccess := Letter("done-success")

	e := NewExec([]Letter{
		RegisterSuccess,
		GroupJoinSuccess,
		SendSuccess,
		SendSuccess,
		SendFailure,
		SendFailure,
		SendSuccess,
		ExitWanted,
	})

	d := New()
	d.SetStartState(Registering)
	d.SetTerminalStates(Exiting, Terminating)

	d.SetTransition(Registering, RegisterSuccess, Waiting, e.Next)
	d.SetTransition(Registering, RegisterFailure, Exiting, e.Last)
	d.SetTransition(Registering, ExitWanted, Exiting, e.Last)

	d.SetTransition(Waiting, GroupJoinSuccess, Running, e.Next)
	d.SetTransition(Waiting, GroupJoinFailure, Exiting, e.Last)
	d.SetTransition(Waiting, ExitWanted, Exiting, e.Last)

	d.SetTransition(Running, SendSuccess, Running, e.Next)
	d.SetTransition(Running, SendFailure, Resending, e.Next)
	d.SetTransition(Running, ExitWanted, Exiting, e.Last)
	d.SetTransition(Running, DoneSuccess, Terminating, e.Last)

	d.SetTransition(Resending, SendSuccess, Running, e.Next)
	d.SetTransition(Resending, SendFailure, Resending, e.Next)
	d.SetTransition(Resending, ExitWanted, Exiting, e.Last)

	final, accepted := d.RunSynchronous(e.Next)
	if !accepted {
		t.Fatalf("failed to recognize Exiting or Terminating as terminal state")
	}

	if final != Exiting {
		t.Fatalf("final state should have been: %v, but was: %v", Exiting, final)
	}

	if e.NextCount() != 8 {
		t.Fatalf("NexCount() expected to be 8, got %d", e.NextCount())
	}

	if e.LastCount() != 1 {
		t.Fatalf("LastCount() expected 1, got %d", e.LastCount())
	}
}
