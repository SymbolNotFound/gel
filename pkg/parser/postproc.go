package parser

type PostProcessing interface {
	isPostProc()
}

func (Nothing) isPostProc()          {}
func (Ident) isPostProc()            {}
func (StringProjection) isPostProc() {}
func (ListProjection) isPostProc()   {}
func (DictProjection) isPostProc()   {}

type Nothing struct{}

type Ident struct{}

type StringProjection struct {
	value string
}

type ListProjection struct {
	values []PostProcessing
}

func projlist(values ...PostProcessing) ListProjection {
	return ListProjection{
		values: values,
	}
}

type DictProjection struct {
	name  string
	attrs []attr
}

type attr interface {
	isDictKV()
}

type keyval struct {
	key   string
	value PostProcessing
}

func (keyval) isDictKV() {}

func projdict(name string, attrs ...attr) DictProjection {
	return DictProjection{
		name:  name,
		attrs: attrs,
	}
}

type ItemProjection struct {
	Reference int
}

func (ItemProjection) isPostProc() {}

type ref = ItemProjection

type Expand struct {
	Reference int
}

func (Expand) isPostProc() {}
func (Expand) isDictKV()   {}
