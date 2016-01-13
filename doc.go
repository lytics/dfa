/*
The dfa package implements a deterministic finite automata to define stateful computations
that are easier understood when transitions are specified explicitly. The API is more
interested in using the DFA to clearly define stateful computation, rather than actually
being used to recognize languages.

Importing

    import "github.com/lytics/dfa"

Quick Example

    // Define states.
    Starting = dfa.State("starting")
    Finishing = dfa.State("finishing")

    // Define letters.
    Done = dfa.Letter("done")
    Repeat = dfa.Letter("repeat")

    // Error reporting is done by user code.
    var errors []error

    // Define stateful computations.
    starting := func() dfa.Letter {
        if err := do(); err != nil {
            errors = append(errors, err)
            return Repeat
        } else {
            return Done
        }
    }
    finishing := func() {
        fmt.Println("all finished")
    }

    // Order matter, set start and terminal states first.
    d := dfa.New()
    d.SetStartState(Starting)
    d.SetTerminalStates(Finishing)
    d.SetTransition(Starting, Done, Finishing, finishing)
    d.SetTransition(Starting, Repeat, Starting, starting)

    // Calls the given function in a new go-routine,
    // unless there were initialization errors.
    // Blocks until the DFA accepts its input or
    // its Stop() function is called. The final state
    // is returned, and accepted == true if the final
    // state is a terminal state.
    final, accepted := d.Run(starting)

    // Error handling is up to the user application.
    for _, err := range errors {
        ...
    }
*/
package dfa
