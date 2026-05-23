package png

type Interlacer interface {
	Interlace([]byte) ([]byte, error)
	Unterlace([]byte) ([]byte, error)
}

type Adam7 struct{}

func NewAdam7() Interlacer {
	return &Adam7{}
}

func (*Adam7) Interlace(b []byte) ([]byte, error) { return b, nil }
func (*Adam7) Unterlace(b []byte) ([]byte, error) { return b, nil }

type NoOp struct{}

func NewNoOp() Interlacer {
	return &NoOp{}
}
func (*NoOp) Interlace(a []byte) ([]byte, error) { return a, nil }
func (*NoOp) Unterlace(a []byte) ([]byte, error) { return a, nil }

var _ Interlacer = (*Adam7)(nil)
var _ Interlacer = (*NoOp)(nil)
