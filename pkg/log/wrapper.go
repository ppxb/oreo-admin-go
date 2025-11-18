package log

type Wrapper struct {
	log    Interface
	fields map[string]interface{}
}

func NewWrapper(l Interface) *Wrapper {
	return &Wrapper{
		log:    l,
		fields: map[string]interface{}{},
	}
}
