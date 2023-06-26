package tpl

func EdgeDBToml() []byte {
	return []byte(`[edgedb]
server-version = "2.14"
`)
}

func DefaultEsdl() []byte {
	return []byte(`using extension graphql;

module default {

    type DriverAge {
        required property min_age -> int64;
        required property max_age -> int64;
        required property factor -> float64;
    }

    type State {
        required property abbr -> str;
    }

    type Coverage {
        property base_rate -> int64;
        property effective_date -> cal::local_date;
        property coverage_type -> str;
        property carrier -> int64;
        multi link states -> State;
        multi link driver_ages -> DriverAge;
    }
}
`)
}

func FutureEsdl() []byte {
	return []byte(`# Disable the application of access policies within access policies
# themselves. This behavior will become the default in EdgeDB 3.0.
# See: https://www.edgedb.com/docs/reference/ddl/access_policies#nonrecursive
using future nonrecursive_access_policies;
`)
}
