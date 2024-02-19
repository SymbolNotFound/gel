package parser

type PostProcessing interface {
	isPostProc()
}

func (Nothing) isPostProc()          {}
func (StringProjection) isPostProc() {}
func (ItemProjection) isPostProc()   {}
func (ListProjection) isPostProc()   {}
func (RecordProjection) isPostProc() {}
func (PropertyGetter) isPostProc()   {}

type Nothing struct{}

type StringProjection struct {
	value string
}

type ItemProjection struct {
	ref int
}

type ListProjection struct {
	values []PostProcessing
}

type ExpandList struct {
	ItemProjection
}

type ElementGetter struct {
	ItemProjection
	index int
}

type RecordProjection struct {
	name  string
	attrs []attr
}

type attr interface {
	isAttribute()
}

type KeyValue struct {
	key   string
	value PostProcessing
}

func (KeyValue) isAttribute()     {}
func (ExpandRecord) isAttribute() {}

type ExpandRecord struct {
	ItemProjection
}

type PropertyGetter struct {
	ItemProjection
	name string
}
