package port

//go:generate go tool mockgen -destination=mock_${GOFILE} -package=${GOPACKAGE} -source=${GOFILE}

// IDGenerator generates unique identifiers for aggregates.
type IDGenerator interface {
	Generate() string
}
