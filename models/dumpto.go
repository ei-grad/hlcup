package models

import "github.com/mailru/easyjson"

func (v *User) DumpTo(w Writer) {
	easyjson.MarshalToWriter(v, w)
}

func (v *Location) DumpTo(w Writer) {
	easyjson.MarshalToWriter(v, w)
}

func (v *Visit) DumpTo(w Writer) {
	easyjson.MarshalToWriter(v, w)
}

func (v UserVisit) DumpTo(w Writer) {
	easyjson.MarshalToWriter(v, w)
}
