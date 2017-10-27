/**
 * Copyright (C) 2016 Deepin Technology Co., Ltd.
 *
 * This program is free software; you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation; either version 3 of the License, or
 * (at your option) any later version.
 **/

package launcher

import (
	"fmt"
	"regexp"
	"strings"
	"sync"
	"unicode"
)

type searchTask struct {
	chars        []rune
	fuzzyMatcher *regexp.Regexp
	stack        *searchTaskStack

	result MatchResults

	isCanceled      bool
	isCanceledMutex sync.RWMutex
	isFinished      bool
	isFinishedMutex sync.RWMutex
}

func (t *searchTask) IsCanceled() bool {
	t.isCanceledMutex.RLock()
	defer t.isCanceledMutex.RUnlock()
	return t.isCanceled
}

func (t *searchTask) Cancel() {
	t.isCanceledMutex.Lock()
	t.isCanceled = true
	t.isCanceledMutex.Unlock()
}

func (t *searchTask) IsFinished() bool {
	t.isFinishedMutex.RLock()
	defer t.isFinishedMutex.RUnlock()
	return t.isFinished
}

func (t *searchTask) Finish() {
	t.isFinishedMutex.Lock()
	t.isFinished = true
	t.isFinishedMutex.Unlock()
	logger.Debug("finish", t)
}

func (t *searchTask) Index() int {
	for i, task := range t.stack.tasks {
		if t == task {
			return i
		}
	}
	return -1
}

func (t *searchTask) next() *searchTask {
	tasks := t.stack.tasks
	index := t.Index()
	if index == -1 {
		return nil
	}
	index++
	if 0 <= index && index < len(tasks) {
		return tasks[index]
	}
	return nil
}

func (t *searchTask) String() string {
	if t == nil {
		return "<nil>"
	}
	canceled := t.IsCanceled()
	finished := t.IsFinished()
	return fmt.Sprintf("<Task %s count=%v canceled=%v finished=%v>", string(t.chars), len(t.result), canceled, finished)
}

func newSearchTask(c rune, stack *searchTaskStack, prev *searchTask) *searchTask {
	t := &searchTask{
		stack: stack,
	}

	if prev != nil {
		// copy chars
		t.chars = prev.chars[:]
	}
	t.chars = append(t.chars, c)

	// init fuzzyMatcher
	var metaQuotedChars []string
	for _, char := range t.chars {
		metaQuotedChars = append(metaQuotedChars, regexp.QuoteMeta(string(char)))
	}
	regStr := strings.Join(metaQuotedChars, ".*?")
	logger.Debug("regexp Str:", regStr)
	var err error
	t.fuzzyMatcher, err = regexp.Compile(regStr)
	if err != nil {
		logger.Warning(err)
	}

	return t
}

func (t *searchTask) doSearch(prev *searchTask) {
	if prev == nil {
		go t.doFirst()
	} else {
		if prev.IsFinished() {
			logger.Debug("start", t, "doSearch prev finished")
			go t.doWithBase(prev.result)
		}
	}
}

func (t *searchTask) doFirst() {
	logger.Debug("start first", t)
	for _, item := range t.stack.items {
		t.matchItem(item)
		if t.IsCanceled() {
			logger.Debug("matchItem stop canceled", t)
			return
		}
	}
	t.done()
}

func (st *searchTask) doWithBase(result MatchResults) {
	for _, mResult := range result {
		st.matchItem(mResult.item)
		if st.IsCanceled() {
			logger.Debug("matchItem stop canceled", st)
			return
		}
	}
	st.done()
}

const (
	Poor         = 50
	BelowAverage = 60
	Average      = 70
	AboveAverage = 75
	Good         = 80
	VeryGood     = 85
	Excellent    = 90
	Highest      = 100
)

func (st *searchTask) match(item *Item) *MatchResult {
	var score SearchScore
	for v, vScore := range item.searchTargets {
		key := string(st.chars)
		index := strings.Index(v, key)
		if index != -1 {
			// key is substr of v
			score += 2 * vScore
			if len(key) == len(v) {
				// ^query$
				score += Highest
			} else if index == 0 {
				// ^query
				score += Excellent
			} else {
				prev := v[:index]
				var prevChar rune
				for _, r := range prev {
					prevChar = r
				}
				//logger.Debugf("prevChar %c", prevChar)
				if prevChar != 0 && !unicode.IsLetter(prevChar) {
					// \bquery
					score += AboveAverage
				} else {
					// xqueryx
					score += BelowAverage
				}
			}
			continue
		}

		if st.fuzzyMatcher != nil {
			loc := st.fuzzyMatcher.FindStringIndex(v)
			if loc != nil {
				score += vScore
				score += BelowAverage
			}
		}
	}

	if score == 0 {
		return nil
	}
	mResult := &MatchResult{
		item:  item,
		score: score,
	}
	return mResult
}

func (st *searchTask) matchItem(item *Item) {
	mResult := st.match(item)
	if mResult != nil {
		st.result = append(st.result, mResult)
	}
}

func (st *searchTask) emitResult() {
	st.stack.manager.emitSearchDone(st.result)
}

func (st *searchTask) done() {
	if st.IsCanceled() {
		logger.Debug("no done canceled", st)
		return
	}
	next := st.next()
	if next != nil {
		// notify next task
		logger.Debug("start", next, "next")
		go next.doWithBase(st.result)
		st.Finish()
	} else {
		// if no next task, emit SearchDone signal
		st.Finish()
		st.emitResult()
	}
}
