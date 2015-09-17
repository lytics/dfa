dfa
===

The `dfa` package implements a deterministic finite automata for use to define stateful
computations that are easier understood when transitions are specified explicitly. The
API is more interested in using the DFA to clearly define stateful computation, rather
than actually being used to recognize languages.

The implementation has all the normal compnents:

 * a finite set of states (Q)
 * a finite set of input symbols called the alphabet (Σ)
 * a transition function (δ : Q × Σ → Q)
 * a start state (q0 ∈ Q)
 * a set of terminal states (F ⊆ Q)

The one addition is that for each transition, Q × Σ → Q, you _must_ specify a function
to run that will supply the next letter.

### Importing

    import "github.com/lytics/dfa"

### Quick Example

```go
    // Define states.
    Starting = dfa.State("starting")
    Finishing = dfa.State("finishing")

    // Define letters.
    Done = dfa.Letter("done")
    Repeat = dfa.Letter("repeat")

    // Define stateful computations.
    starting := func() dfa.Letter {
        if err := do(); err != nil {
            return Repeat
        } else {
            return Done
        }
    }
    finishing := func() {
        fmt.Println("all finished")
    }

    d := dfa.New()
    d.SetStartState(Starting)
    d.SetTerminalStates(Finishing)
    d.SetTransition(Starting, Done, Finishing, finishing)
    d.SetTransition(Starting, Repeat, Starting, starting)

    // Calls the given function in a new go-routine,
    // unless there were initialization errors.
    err := d.Run(starting)
    ...

    // Blocks until the DFA accepts its input or
    // its Stop() function is called. If the DFA
    // stopped in a non-terminal state an err is
    // reported. The final state is always given.
    final, err := d.Done()
    ...
```

### Stateful Computations

The functions given when defining the transitions must have one of
two types:

 * `func()`
 * `func() dfa.Letter`

Functions associated with a transition ending in a terminal state must
give a function that returns no values. Functions that transition to
non-terminal states must return a letter.
