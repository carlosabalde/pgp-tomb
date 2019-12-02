package query

type boolean bool

func (b boolean) Eval(p Params) bool {
	return bool(b)
}

func (b boolean) String() string {
	if bool(b) {
		return "true"
	}
	return "false"
}

var (
	True  = boolean(true)
	False = boolean(false)
)
