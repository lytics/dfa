package dfa

import (
	"fmt"
	"testing"
)

type Exec struct {
	Sequence  []Letter
	NextCount int
	LastCount int
}

func (e *Exec) Next() Letter {
	l := e.Sequence[0]
	e.Sequence = e.Sequence[1:]
	e.NextCount++
	return l
}

func (e *Exec) Last() {
	e.LastCount++
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

	e := &Exec{
		Sequence: []Letter{
			RegisterSuccess,
			GroupJoinSuccess,
			SendSuccess,
			SendSuccess,
			SendFailure,
			SendFailure,
			SendSuccess,
			ExitWanted,
		},
	}

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

	err := d.Run(e.Next)
	if err != nil {
		t.Fatalf("failed to run dfa: %v", err)
	}

	final, err := d.Done()
	if err != nil {
		t.Fatalf("failed to terminate dfa: %v", err)
	}
	if final != Exiting {
		t.Fatalf("final state should have been: %v, but was: %v", Exiting, final)
	}

	if e.NextCount != 8 {
		t.Fail()
	}
	if e.LastCount != 1 {
		t.Fail()
	}
}

func TestLastState(t *testing.T) {
	// States
	Starting := State("staring")
	// Inputs
	Transition := Letter("transition")

	e := &Exec{
		Sequence: []Letter{},
	}

	d := New()
	d.SetStartState(Starting)
	d.SetTerminalStates(Starting)
	d.SetTransition(Starting, Transition, Starting, e.Last)

	// The expectation is that the Next nor Last method will
	// ever be called since the DFA's starting and terminal
	// state are the same.
	err := d.Run(e.Next)
	if err != nil {
		t.Fatalf("failed to run dfa: %v", err)
	}
	final, err := d.Done()
	if err != nil {
		t.Fatalf("received error: %v", err)
	}
	if final != Starting {
		t.Fatalf("final state should have been: %v, but was: %v", Starting, final)
	}

	if e.NextCount != 0 {
		t.Fail()
	}
	if e.LastCount != 0 {
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

	fmt.Println(d.GraphViz())
}
