package types

import "main/pkg/constants"

type QueryInfo struct {
	Success   bool
	QueryType constants.QueryType
	Node      string
}
