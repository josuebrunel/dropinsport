package pbclient

import (
	"bytes"
	"encoding/json"
	"io"
	"strings"

	"github.com/josuebrunel/sportdropin/pkg/xlog"
)

const (
	QTypeData    = "data"
	QTypeExpand  = "expand"
	QTypeFields  = "fields"
	QTypeFilters = "filters"
	QTypeHeaders = "headers"
	QTypePage    = "page"
	QTypeParams  = "params"
	QTypeSort    = "sort"
)

type IQ interface {
	GetQType() string
}

type QHeaders map[string]string

func (q QHeaders) GetQType() string { return QTypeHeaders }

type QParams []any

func (q QParams) GetQType() string { return QTypeParams }

type QData struct {
	Data      any
	DataBytes io.Reader
}

func (q QData) GetQType() string { return QTypeData }

func NewQData(data any) QData {
	return QData{Data: data, DataBytes: jsonMarshal(data)}
}

type QFilters string

func (q QFilters) GetQType() string { return QTypeFilters }

type QSort []string

func (q QSort) GetQType() string { return QTypeSort }

type QFields []string

func (q QFields) GetQType() string { return QTypeFields }

type QPage struct {
	Page      int
	PerPage   int
	SkipTotal bool
}

func (q QPage) GetQType() string { return QTypePage }

type QExpand []string

func (q QExpand) GetQType() string { return QTypeExpand }

func QmListString(ss []string) string {
	var (
		qb strings.Builder
		l  = len(ss)
	)
	for i, s := range ss {
		qb.WriteString(s)
		if i < l {
			qb.WriteString(",")
		}
	}
	return qb.String()
}

func jsonMarshal(d any) io.Reader {
	b, err := json.Marshal(d)
	if err != nil {
		xlog.Error("failed to marshal payload", "payload", d)
	}
	return bytes.NewReader(b)
}
