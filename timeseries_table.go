package gocassa

import (
	"time"
)

const bucketFieldName = "bucket"

type timeSeriesT struct {
	t          Table
	timeField  string
	idField    string
	bucketSize time.Duration
}

func (o *timeSeriesT) Table() Table                     { return o.t }
func (o *timeSeriesT) Create() error                    { return o.Table().Create() }
func (o *timeSeriesT) CreateIfNotExist() error          { return o.Table().CreateIfNotExist() }
func (o *timeSeriesT) Name() string                     { return o.Table().Name() }
func (o *timeSeriesT) Recreate() error                  { return o.Table().Recreate() }
func (o *timeSeriesT) CreateStatement() (string, error) { return o.Table().CreateStatement() }
func (o *timeSeriesT) CreateIfNotExistStatement() (string, error) {
	return o.Table().CreateIfNotExistStatement()
}

func (o *timeSeriesT) Set(v interface{}) Op {
	m, ok := toMap(v)
	if !ok {
		panic("Can't set: not able to convert")
	}
	if tim, ok := m[o.timeField].(time.Time); !ok {
		panic("timeField is not actually a time.Time")
	} else {
		m[bucketFieldName] = o.bucket(tim.Unix())
	}
	return o.Table().Set(m)
}

func (o *timeSeriesT) bucket(secs int64) int64 {
	return (secs - secs%int64(o.bucketSize/time.Second)) * 1000
}

func (o *timeSeriesT) Update(timeStamp time.Time, id interface{}, m map[string]interface{}) Op {
	bucket := o.bucket(timeStamp.Unix())
	return o.Table().
		Where(Eq(bucketFieldName, bucket),
			Eq(o.timeField, timeStamp),
			Eq(o.idField, id)).
		Update(m)
}

func (o *timeSeriesT) Delete(timeStamp time.Time, id interface{}) Op {
	bucket := o.bucket(timeStamp.Unix())
	return o.Table().
		Where(Eq(bucketFieldName, bucket),
			Eq(o.timeField, timeStamp),
			Eq(o.idField, id)).
		Delete()
}

func (o *timeSeriesT) Read(timeStamp time.Time, id, pointer interface{}) Op {
	bucket := o.bucket(timeStamp.Unix())
	return o.Table().
		Where(Eq(bucketFieldName, bucket),
			Eq(o.timeField, timeStamp),
			Eq(o.idField, id)).
		ReadOne(pointer)
}

func (o *timeSeriesT) List(startTime time.Time, endTime time.Time, pointerToASlice interface{}) Op {
	buckets := []interface{}{}
	start := o.bucket(startTime.Unix())
	for i := start; ; i += int64(o.bucketSize/time.Second) * 1000 {
		if i >= endTime.Unix()*1000 {
			break
		}
		buckets = append(buckets, i)
	}
	return o.Table().
		Where(In(bucketFieldName, buckets...),
			GTE(o.timeField, startTime),
			LTE(o.timeField, endTime)).
		Read(pointerToASlice)
}

func (o *timeSeriesT) WithOptions(opt Options) TimeSeriesTable {
	return &timeSeriesT{
		t:          o.Table().WithOptions(opt),
		timeField:  o.timeField,
		idField:    o.idField,
		bucketSize: o.bucketSize,
	}
}
