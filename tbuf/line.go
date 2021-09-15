package tbuf

import (
	"encoding/json"
	"sort"
	"time"
)

// change for different pattern
var Schema = struct {
	Message string
	Time    string
	Level   string
}{
	Message: "message",
	Time:    "time",
	Level:   "level",
}

type Line struct {
	Str   string
	Short string
	Tags  map[string]json.RawMessage
	Time  time.Time
	Level string
	Mark  bool
}

func (this Line) SortedTags() (out []string) {
	for n := range this.Tags {
		out = append(out, n)
	}
	sort.Strings(out)
	return
}

func unmarshalOrString(in []byte) string {
	var s string
	err := json.Unmarshal(in, &s)
	if err == nil {
		return s
	} else {
		return string(in)
	}
}

func (this *Buffer) Append(s string, log func(...interface{})) {
	this.m.Lock()
	defer this.m.Unlock()
	if this.Pos == len(this.Lines) {
		this.Pos++
	}
	l := ParseLine(s, log)
	this.Lines = append(this.Lines, l)
	this.Last = time.Now()
}
func (this *Buffer) AppendLine(l Line) {
	this.Lines = append(this.Lines, l)
}

func ParseLine(s string, log func(...interface{})) (out Line) {
	//out.Orig = s
	out.Str = s
	if len(s) < 2 {
		return
	}
	if s[0] == '{' {
		err := json.Unmarshal([]byte(s), &out.Tags)
		if err != nil {
			log("can't unmarshal: %v", err)
			return
		}

		// if message, then use it as main message
		if j, exists := out.Tags[Schema.Message]; exists {
			delete(out.Tags, Schema.Message)
			out.Short = unmarshalOrString(j)
		}

		// check for "time"
		if j, exists := out.Tags[Schema.Time]; exists {
			err := json.Unmarshal(j, &out.Time)
			if err == nil {
				delete(out.Tags, Schema.Time)
			}
		}

		// check for "time"
		if j, exists := out.Tags[Schema.Level]; exists {
			err := json.Unmarshal(j, &out.Level)
			if err == nil {
				delete(out.Tags, Schema.Level)
			}
		}
	}
	return
}
