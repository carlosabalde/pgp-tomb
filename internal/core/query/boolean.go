package query

type boolean bool

func (self boolean) Eval(identifiers Identifiers) bool {
	return bool(self)
}

func (self boolean) String() string {
	if bool(self) {
		return "true"
	}
	return "false"
}

var (
	True  = boolean(true)
	False = boolean(false)
)
