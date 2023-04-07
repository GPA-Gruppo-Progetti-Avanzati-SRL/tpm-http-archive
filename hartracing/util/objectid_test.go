package util_test

import (
	"encoding/binary"
	"fmt"
	"github.com/GPA-Gruppo-Progetti-Avanzati-SRL/tpm-http-archive/hartracing/util"
	"strconv"
	"testing"
	"time"
)

func TestNewTraceId(t *testing.T) {
	t.Log(util.NewTraceId())
}

func TestNewObjectId(t *testing.T) {
	oid := util.NewObjectId()
	t.Log(oid)

	now := time.Now()
	b := make([]byte, 4)
	t.Log("Unix: ", now.Unix())
	binary.BigEndian.PutUint32(b, uint32(now.Unix()))
	t.Log("time-part: ", fmt.Sprintf("%x\n", string(b)))

	now = now.Add(1 * 24 * time.Hour)
	t.Log("Unix: ", now.Unix())
	binary.BigEndian.PutUint32(b, uint32(now.Unix()))
	t.Log("time-part: ", fmt.Sprintf("%x\n", string(b)))

	now = time.Date(2023, 03, 13, 0, 0, 0, 0, time.UTC)
	t.Log("Unix: ", now.Unix())
	binary.BigEndian.PutUint32(b, uint32(now.Unix()))
	t.Log("time-part: ", fmt.Sprintf("%x\n", string(b)))

	now = time.Date(2023, 03, 13, 23, 59, 59, 59, time.UTC)
	t.Log("Unix: ", now.Unix())
	binary.BigEndian.PutUint32(b, uint32(now.Unix()))
	t.Log("time-part: ", fmt.Sprintf("%x\n", string(b)))

	d, _ := strconv.Atoi(now.Format("20060102"))
	t.Log(d, fmt.Sprintf("%x", d))

	v := fmt.Sprintf("%02d", 123)
	t.Log(v)
}
