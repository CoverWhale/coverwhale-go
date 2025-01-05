// Copyright 2025 Sencillo
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

package tpl

func EdgeDBToml() []byte {
	return []byte(`[edgedb]
server-version = "4.0"
`)
}

func DefaultEsdl() []byte {
	return []byte(`using extension graphql;

module default {
    scalar type StateAbbr extending enum<NY, PA, SC>;

    type DriverAge {
        required min_age: int64;
        required max_age: int64;
        required factor: float64;
    }

    type State {
        required abbr: StateAbbr;
    }

    type Coverage {
        property base_rate: int64;
        property effective_date: cal::local_date;
        property coverage_type: str;
        property carrier: int64;
        multi link states: State;
        multi link driver_ages: DriverAge;
    }
}
`)
}
