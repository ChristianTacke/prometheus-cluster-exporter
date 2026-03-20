// -*- coding: utf-8 -*-
//
// © Copyright 2023 GSI Helmholtzzentrum für Schwerionenforschung
//
// This software is distributed under
// the terms of the GNU General Public Licence version 3 (GPL Version 3),
// copied verbatim in the file "LICENCE".

package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
)

func TestParseLustreMetadataOperations(t *testing.T) {

	var data string = `{"status":"success","data":{"resultType":"vector","result":[
		{"metric":{"jobid":"35044931","target":"hebe-MDT0002"},"value":[1639743019.545,"1"]},
		{"metric":{"jobid":"35070653","target":"hebe-MDT0000"},"value":[1639743019.545,"43"]},
		{"metric":{"jobid":"35189820","target":"hebe-MDT0000"},"value":[1639743019.545,"4"]},
		{"metric":{"jobid":"35166602","target":"hebe-MDT0001"},"value":[1639743019.545,"31"]},
		{"metric":{"jobid":"35189845","target":"hebe-MDT0001"},"value":[1639743019.545,"1"]},
		{"metric":{"jobid":"35048662","target":"hebe-MDT0001"},"value":[1639743019.545,"27"]},
		{"metric":{"jobid":"cp.5689","target":"hebe-MDT0002"},"value":[1639743019.545,"1"]},
		{"metric":{"jobid":"35056989","target":"hebe-OST022d"},"value":[1639743019.545,"5"]},
		{"metric":{"jobid":"touch.6812","target":"hebe-OST020c"},"value":[1639743019.545,"1"]}
		]}}`

	var content []byte = []byte(data)

	var lustreMetadataOperations *[]metadataInfo
	var err error

	lustreMetadataOperations, err = parseLustreMetadataOperations(&content)

	if err != nil {
		t.Error(err)
	}

	var got_count int = len(*lustreMetadataOperations)
	var expected_count int = 7

	if expected_count != got_count {
		t.Errorf("Expected count of metadata operations: %d - got: %d", expected_count, got_count)
	}

	var metadataInfo metadataInfo = (*lustreMetadataOperations)[0]
	var expected_jobid string = "35044931"
	var expected_target string = "hebe-MDT0002"

	if metadataInfo.jobid != expected_jobid {
		t.Errorf("Expected jobid: %s - got: %s", expected_jobid, metadataInfo.jobid)
	}

	if metadataInfo.target != expected_target {
		t.Errorf("Expected target: %s - got: %s", expected_target, metadataInfo.target)
	}

	for _, metadataInfo := range *lustreMetadataOperations {
		if !regexMetadataMDT.MatchString(metadataInfo.target) {
			t.Error("Only MDT as target is allowed:", metadataInfo.target)
		}
	}

}

func TestParseLustreTotalBytes(t *testing.T) {

	var data string = `{"status":"success","data":{"resultType":"vector","result":[
		{"metric":{"jobid":"35652133"},"value":[1640181380.814,"319215.8800539506"]},
		{"metric":{"jobid":"35239994"},"value":[1640181380.814,"125747.2"]},
		{"metric":{"jobid":"35651038"},"value":[1640181380.814,"379697.46153350436"]},
		{"metric":{"jobid":"35683050"},"value":[1640181380.814,"955.7333333333333"]},
		{"metric":{"jobid":"35676304"},"value":[1640181380.814,"893883.7333333333"]},
		{"metric":{"jobid":"35682305"},"value":[1640181380.814,"819.2"]},
		{"metric":{"jobid":"35676288"},"value":[1640181380.814,"689493.3333333334"]},
		{"metric":{"jobid":"35676299"},"value":[1640181380.814,"248627.2"]}
		]}}`

	var content []byte = []byte(data)

	var lustreThroughputInfo *[]throughputInfo
	var err error

	lustreThroughputInfo, err = parseLustreTotalBytes(&content)

	if err != nil {
		t.Error(err)
	}

	var got_count int = len(*lustreThroughputInfo)
	var expected_count int = 8

	if expected_count != got_count {
		t.Errorf("Expected count of metadata operations: %d - got: %d", expected_count, got_count)
	}

	var throughputInfo throughputInfo = (*lustreThroughputInfo)[0]
	var expected_jobid string = "35652133"

	if throughputInfo.jobid != expected_jobid {
		t.Errorf("Expected jobid: %s - got: %s", expected_jobid, throughputInfo.jobid)
	}
}

// getGaugeVecValue returns the sum of all values in a GaugeVec for testing.
func getGaugeVecValue(gv *prometheus.GaugeVec) float64 {
	ch := make(chan prometheus.Metric, 100)
	gv.Collect(ch)
	close(ch)

	var total float64
	for m := range ch {
		var d dto.Metric
		m.Write(&d)
		total += d.GetGauge().GetValue()
	}
	return total
}

// TestBuildLustreMetadataMetrics_UnknownUidSkipsEntry verifies that an unknown UID
// only skips the affected entry and does not abort processing of remaining entries.
func TestBuildLustreMetadataMetrics_UnknownUidSkipsEntry(t *testing.T) {

	// uid 1000 is known, uid 9999 is NOT in the users map
	var metadataJSON string = `{"status":"success","data":{"resultType":"vector","result":[
		{"metric":{"jobid":"cp.1000","target":"hebe-MDT0000"},"value":[1639743019.545,"10"]},
		{"metric":{"jobid":"cp.9999","target":"hebe-MDT0001"},"value":[1639743019.545,"5"]},
		{"metric":{"jobid":"mv.1000","target":"hebe-MDT0002"},"value":[1639743019.545,"20"]}
		]}}`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, metadataJSON)
	}))
	defer server.Close()

	e := newExporter(5, server.URL, server.URL, server.URL)

	jobs := []jobInfo{{jobid: "12345", account: "acct", user: "testuser"}}
	users := userInfoMap{1000: {user: "alice", uid: 1000, gid: 100}}
	groups := groupInfoMap{100: {group: "staff", gid: 100}}

	err := e.buildLustreMetadataMetrics(jobs, users, groups)
	if err != nil {
		t.Errorf("Expected no error when uid is missing, got: %v", err)
	}

	// Both entries for uid 1000 should still be recorded (10 + 20 = 30).
	got := getGaugeVecValue(e.procMetadataOperationsMetric)
	if got != 30 {
		t.Errorf("Expected proc metadata operations total 30, got: %v", got)
	}
}

// TestBuildLustreThroughputMetrics_UnknownUidSkipsEntry verifies that an unknown UID
// only skips the affected entry and does not abort processing of remaining entries.
func TestBuildLustreThroughputMetrics_UnknownUidSkipsEntry(t *testing.T) {

	// uid 1000 is known, uid 9999 is NOT in the users map
	var throughputJSON string = `{"status":"success","data":{"resultType":"vector","result":[
		{"metric":{"jobid":"cp.1000"},"value":[1640181380.814,"100.0"]},
		{"metric":{"jobid":"cp.9999"},"value":[1640181380.814,"50.0"]},
		{"metric":{"jobid":"mv.1000"},"value":[1640181380.814,"200.0"]}
		]}}`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, throughputJSON)
	}))
	defer server.Close()

	e := newExporter(5, server.URL, server.URL, server.URL)

	jobs := []jobInfo{{jobid: "12345", account: "acct", user: "testuser"}}
	users := userInfoMap{1000: {user: "alice", uid: 1000, gid: 100}}
	groups := groupInfoMap{100: {group: "staff", gid: 100}}

	err := e.buildLustreThroughputMetrics(jobs, users, groups, true)
	if err != nil {
		t.Errorf("Expected no error when uid is missing, got: %v", err)
	}

	// Both entries for uid 1000 should still be recorded (100 + 200 = 300).
	got := getGaugeVecValue(e.procReadThroughputMetric)
	if got != 300 {
		t.Errorf("Expected proc read throughput total 300, got: %v", got)
	}
}

// TestBuildLustreMetadataMetrics_UnknownGidSkipsEntry verifies that an unknown GID
// only skips the affected entry and does not abort processing of remaining entries.
func TestBuildLustreMetadataMetrics_UnknownGidSkipsEntry(t *testing.T) {

	// uid 1000 has gid 100 (known), uid 2000 has gid 999 (NOT in groups map)
	var metadataJSON string = `{"status":"success","data":{"resultType":"vector","result":[
		{"metric":{"jobid":"cp.1000","target":"hebe-MDT0000"},"value":[1639743019.545,"10"]},
		{"metric":{"jobid":"cp.2000","target":"hebe-MDT0001"},"value":[1639743019.545,"5"]},
		{"metric":{"jobid":"mv.1000","target":"hebe-MDT0002"},"value":[1639743019.545,"20"]}
		]}}`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, metadataJSON)
	}))
	defer server.Close()

	e := newExporter(5, server.URL, server.URL, server.URL)

	jobs := []jobInfo{{jobid: "12345", account: "acct", user: "testuser"}}
	users := userInfoMap{
		1000: {user: "alice", uid: 1000, gid: 100},
		2000: {user: "bob", uid: 2000, gid: 999}, // gid 999 not in groups
	}
	groups := groupInfoMap{100: {group: "staff", gid: 100}}

	err := e.buildLustreMetadataMetrics(jobs, users, groups)
	if err != nil {
		t.Errorf("Expected no error when gid is missing, got: %v", err)
	}

	// Both entries for uid 1000 (gid 100) should still be recorded (10 + 20 = 30).
	got := getGaugeVecValue(e.procMetadataOperationsMetric)
	if got != 30 {
		t.Errorf("Expected proc metadata operations total 30, got: %v", got)
	}
}
