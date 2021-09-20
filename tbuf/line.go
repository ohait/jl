package tbuf

import (
	"encoding/json"
	"sort"
	"strconv"
	"time"
)

// change for different pattern
type Schema struct {
	Message string
	Time    string
	Level   string
}

var def = Schema{
	Message: "message",
	Time:    "time",
	Level:   "level",
}
var alt = Schema{
	Message: "msg",
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
	this.m.Lock()
	defer this.m.Unlock()
	this.Lines = append(this.Lines, l)
}

func ParseLine(s string, log func(...interface{})) (out Line) {
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
		if j, exists := out.Tags[def.Message]; exists {
			delete(out.Tags, def.Message)
			out.Short = unmarshalOrString(j)
		} else if j, exists := out.Tags[alt.Message]; exists {
			delete(out.Tags, def.Message)
			out.Short = unmarshalOrString(j)
		}

		// check for "time"
		if j, exists := out.Tags[def.Time]; exists {
			err := json.Unmarshal(j, &out.Time)
			if err == nil {
				delete(out.Tags, def.Time)
			} else {
				epoch, err := strconv.ParseInt(string(j), 10, 64)
				if err == nil {
					out.Time = time.Unix(epoch, 0)
					delete(out.Tags, def.Time)
				}
			}
		}

		// check for "time"
		if j, exists := out.Tags[def.Level]; exists {
			err := json.Unmarshal(j, &out.Level)
			if err == nil {
				delete(out.Tags, def.Level)
			}
		}
	}
	return
}
