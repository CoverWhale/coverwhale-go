// Copyright 2023 Cover Whale Insurance Solutions Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package http

import (
	"fmt"
	"testing"
)

func TestError(t *testing.T) {
	tt := []struct {
		name string
		err  ClientError
		want string
	}{
		{name: "simple", err: ClientError{Details: "test", Status: 400}, want: "test"},
	}

	for _, v := range tt {
		t.Run(v.name, func(t *testing.T) {
			if v.err.Error() != v.want {
				t.Errorf("expected %s but got %s", v.want, v.err.Error())
			}

			wantBody := fmt.Sprintf(`{"error": "%s"}`, v.want)
			if string(v.err.Body()) != wantBody {
				t.Errorf("expected %s, but got %s", wantBody, string(v.err.Body()))
			}
		})
	}
}
