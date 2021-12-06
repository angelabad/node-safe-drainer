/*
 * Copyright (c) 2021 Angel Abad. All Rights Reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *        http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package utils

import (
	"testing"
)

func TestAdd(t *testing.T) {
	q := NewQueue(10)
	defer q.Close()
	n := 5
	for i := 0; i != n; i++ {
		q.Add()
		go func(c int) {
		}(i)
	}
	if jobs := q.Current(); jobs != n {
		t.Errorf("Expected %d got %d", n, jobs)
		t.Fail()
	}
}

func TestWait(t *testing.T) {
	q := NewQueue(10)
	defer q.Close()
	n := 5
	for i := 0; i != n; i++ {
		q.Add()
		go func(c int) {
			defer q.Done()
		}(i)
	}
	// wait for the end of the all jobs
	q.Wait()
	if jobs := q.Current(); jobs != 0 {
		t.Errorf("Expected %d got %d", 0, jobs)
		t.Fail()
	}
}

func TestDone(t *testing.T) {
	q := NewQueue(10)
	defer q.Close()
	n := 5
	for i := 0; i != n; i++ {
		q.Add()
		go func(c int) {
			// let all the jobs done
			defer q.Done()
		}(i)
	}
	// wait for the end of the all jobs
	q.Wait()
	if jobs := q.Current(); jobs != 0 {
		t.Errorf("Expected %d got %d", 0, jobs)
		t.Fail()
	}
}

func TestCurrent(t *testing.T) {
	q := NewQueue(10)
	defer q.Close()
	n := 5
	for i := 0; i != n; i++ {
		q.Add()
		go func(c int) {
			defer q.Done()
		}(i)
	}
	q.Wait()
	// current should be 0
	if jobs := q.Current(); jobs != 0 {
		t.Errorf("Expected %d got %d", 0, jobs)
		t.Fail()
	}
}
