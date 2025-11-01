package shared

import (
    "os"
    "testing"
)

// withStdin replaces os.Stdin with a pipe providing the given input for the duration of fn.
func withStdin(input string, fn func()) {
    r, w, err := os.Pipe()
    if err != nil {
        panic(err)
    }
    old := os.Stdin
    os.Stdin = r
    // write the input in a goroutine then close the writer
    go func() {
        _, _ = w.Write([]byte(input))
        _ = w.Close()
    }()
    defer func() {
        os.Stdin = old
        _ = r.Close()
    }()
    fn()
}

func TestAskYesNo_AcceptsYesAndNo(t *testing.T) {
    cases := []struct {
        name   string
        input  string
        expect bool
    }{
        {"yes lower", "y\n", true},
        {"yes upper", "Y\n", true},
        {"no lower", "n\n", false},
        {"no upper", "N\n", false},
    }
    for _, tc := range cases {
        t.Run(tc.name, func(t *testing.T) {
            withStdin(tc.input, func() {
                got := AskYesNo("Confirm?")
                if got != tc.expect {
                    t.Fatalf("AskYesNo => %v; want %v (input %q)", got, tc.expect, tc.input)
                }
            })
        })
    }
}

func TestAskYesNo_InvalidThenValid(t *testing.T) {
    // provide an invalid response first, then a valid 'y'
    withStdin("maybe\ny\n", func() {
        got := AskYesNo("Proceed?")
        if !got {
            t.Fatalf("AskYesNo should return true after invalid then 'y'")
        }
    })
}

func TestAskString_ReturnsInputBeforeNewline(t *testing.T) {
    withStdin("hello world\n", func() {
        got := AskString("Enter value: ")
        // Scanln reads up to first space, so only 'hello' is captured
        if got != "hello" {
            t.Fatalf("AskString = %q; want %q", got, "hello")
        }
    })
}
