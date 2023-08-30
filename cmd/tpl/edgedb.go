package tpl

func EdgeDBToml() []byte {
	return []byte(`[edgedb]
server-version = "3.2"
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
