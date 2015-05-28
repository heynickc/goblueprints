package trace

import (
	"fmt"
	"io"
)

// tracer is the interface that describes an object capable
// of tracing events throughout code

// the ...interface{} means that the Trace
// method takes any number of arguments
// of any type
type Tracer interface {
	Trace(...interface{})
}

func New(w io.Writer) Tracer {
	return &tracer{out: w}
}

type tracer struct {
	out io.Writer
}

func (t *tracer) Trace(a ...interface{}) {
	t.out.Write([]byte(fmt.Sprint(a...)))
	t.out.Write([]byte("\n"))
}
