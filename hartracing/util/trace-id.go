package util

import (
	"github.com/GPA-Gruppo-Progetti-Avanzati-SRL/tpm-common/util"
	"strings"
	"time"
)

func NewTraceId() string {
	var sb strings.Builder
	sb.WriteString(time.Now().Format("021504"))
	sb.WriteString("-")
	sb.WriteString(util.NewObjectId().String())
	return sb.String()
}
