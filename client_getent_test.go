// -*- coding: utf-8 -*-
//
// © Copyright 2023 GSI Helmholtzzentrum für Schwerionenforschung
//
// This software is distributed under
// the terms of the GNU General Public Licence version 3 (GPL Version 3),
// copied verbatim in the file "LICENCE".

package main

import "testing"

func TestParseGetentPasswdLine(t *testing.T) {
	info, err := parseGetentPasswdLine("alice:x:1001:100:Alice:/home/alice:/bin/bash")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if info.user != "alice" || info.uid != 1001 || info.gid != 100 {
		t.Fatalf("unexpected parsed info: %#v", info)
	}

	if _, err := parseGetentPasswdLine("alice:x:1001"); err == nil {
		t.Fatalf("expected error for insufficient fields")
	}

	if _, err := parseGetentPasswdLine("alice:x:not-a-number:100"); err == nil {
		t.Fatalf("expected error for invalid uid")
	}

	if _, err := parseGetentPasswdLine("alice:x:1001:not-a-number"); err == nil {
		t.Fatalf("expected error for invalid gid")
	}
}

func TestParseGetentGroupLine(t *testing.T) {
	info, err := parseGetentGroupLine("staff:x:100:alice,bob")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if info.group != "staff" || info.gid != 100 {
		t.Fatalf("unexpected parsed info: %#v", info)
	}

	if _, err := parseGetentGroupLine("staff:x"); err == nil {
		t.Fatalf("expected error for insufficient fields")
	}

	if _, err := parseGetentGroupLine("staff:x:not-a-number"); err == nil {
		t.Fatalf("expected error for invalid gid")
	}
}
