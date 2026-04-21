// -*- coding: utf-8 -*-
//
// © Copyright 2023 GSI Helmholtzzentrum für Schwerionenforschung
//
// This software is distributed under
// the terms of the GNU General Public Licence version 3 (GPL Version 3),
// copied verbatim in the file "LICENCE".

package main

import (
	"bytes"
	"errors"
	"io/ioutil"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

const GETENT = "getent"

type userInfo struct {
	user string
	uid  int
	gid  int
}

type groupInfo struct {
	group string
	gid   int
}

type userInfoMap map[int]userInfo
type groupInfoMap map[int]groupInfo

type userInfoMapResult struct {
	elapsed float64
	users   userInfoMap
	err     error
}

type groupInfoMapResult struct {
	elapsed float64
	groups  groupInfoMap
	err     error
}

func parseGetentPasswdLine(line string) (userInfo, error) {
	fields := strings.SplitN(line, ":", 5)

	if len(fields) < 4 {
		return userInfo{}, errors.New("insufficient field count found in line: " + line)
	}

	uid, err := strconv.Atoi(fields[2])
	if err != nil {
		return userInfo{}, err
	}

	gid, err := strconv.Atoi(fields[3])
	if err != nil {
		return userInfo{}, err
	}

	return userInfo{fields[0], uid, gid}, nil
}

func parseGetentGroupLine(line string) (groupInfo, error) {
	fields := strings.SplitN(line, ":", 4)

	if len(fields) < 3 {
		return groupInfo{}, errors.New("insufficient field count found in line: " + line)
	}

	gid, err := strconv.Atoi(fields[2])
	if err != nil {
		return groupInfo{}, err
	}

	return groupInfo{fields[0], gid}, nil
}

func createUserInfoMap(channel chan<- userInfoMapResult) {

	start := time.Now()

	userInfoMap := make(userInfoMap)

	cmd := exec.Command(GETENT, "passwd")

	pipe, err := cmd.StdoutPipe()

	if err != nil {
		channel <- userInfoMapResult{0, nil, err}
		return
	}

	if err := cmd.Start(); err != nil {
		channel <- userInfoMapResult{0, nil, err}
		return
	}

	out, err := ioutil.ReadAll(pipe)

	if err != nil {
		channel <- userInfoMapResult{0, nil, err}
		return
	}

	// TODO Timeout handling?
	if err := cmd.Wait(); err != nil {
		channel <- userInfoMapResult{0, nil, err}
		return
	}

	// TrimSpace on []bytes is more efficient than calling TrimSpace on a string since it creates a copy
	content := string(bytes.TrimSpace(out))

	if len(content) == 0 {
		channel <- userInfoMapResult{0, nil, errors.New("retrieved content in createUserInfoMap() is empty")}
		return
	}

	lines := strings.Split(content, "\n")

	for _, line := range lines {
		info, err := parseGetentPasswdLine(line)
		if err != nil {
			channel <- userInfoMapResult{0, nil, err}
			return
		}

		userInfoMap[info.uid] = info
	}

	elapsed := time.Since(start).Seconds()

	channel <- userInfoMapResult{elapsed, userInfoMap, nil}
}

func createGroupInfoMap(channel chan<- groupInfoMapResult) {

	start := time.Now()

	groupInfoMap := make(groupInfoMap)

	cmd := exec.Command(GETENT, "group")

	pipe, err := cmd.StdoutPipe()

	if err != nil {
		channel <- groupInfoMapResult{0, nil, err}
		return
	}

	if err := cmd.Start(); err != nil {
		channel <- groupInfoMapResult{0, nil, err}
		return
	}

	out, err := ioutil.ReadAll(pipe)

	if err != nil {
		channel <- groupInfoMapResult{0, nil, err}
		return
	}

	// TODO Timeout handling?
	if err := cmd.Wait(); err != nil {
		channel <- groupInfoMapResult{0, nil, err}
		return
	}

	// TrimSpace on []bytes is more efficient than calling TrimSpace on a string since it creates a copy
	content := string(bytes.TrimSpace(out))

	if len(content) == 0 {
		channel <- groupInfoMapResult{0, nil, errors.New("retrieved content in createGroupInfoMap() is empty")}
		return
	}

	lines := strings.Split(content, "\n")

	for _, line := range lines {
		info, err := parseGetentGroupLine(line)
		if err != nil {
			channel <- groupInfoMapResult{0, nil, err}
			return
		}

		groupInfoMap[info.gid] = info
	}

	elapsed := time.Since(start).Seconds()

	channel <- groupInfoMapResult{elapsed, groupInfoMap, nil}
}
