package models

type Operator string

const (
	OpEquals     = "eq"
	OpNotEquals  = "neq"
	OpContains   = "contains"
	OpStartsWith = "sw"
	OpEndsWith   = "ew"
)
