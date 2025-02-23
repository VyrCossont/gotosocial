package gtsmodel

// ParsedQuery represents the results of parsing the search operator terms within a query.
// It's not currently persisted in the DB, but needs to be available
// at similar places in the build to GTS DB models.
type ParsedQuery struct {
	// Query is the original search query text with operator terms removed.
	Query string

	// ClassicScope enables vanilla GtS search scope restrictions.
	ClassicScope bool

	// FromAccountID is the account from a successfully resolved `from:` operator, if present.
	FromAccountID string

	// ToAccountID is the account from a successfully resolved `to:` operator, if present.
	ToAccountID string

	IsReply     ParsedQueryTernary
	IsSensitive ParsedQueryTernary
	IsPublic    ParsedQueryTernary
	IsUnlisted  ParsedQueryTernary
	IsPrivate   ParsedQueryTernary
	IsDirect    ParsedQueryTernary
	IsLocal     ParsedQueryTernary
	IsLocalOnly ParsedQueryTernary
	IsNote      ParsedQueryTernary
	IsArticle   ParsedQueryTernary
	IsBot       ParsedQueryTernary
	HasCW       ParsedQueryTernary
	HasMedia    ParsedQueryTernary
	HasAudio    ParsedQueryTernary
	HasImage    ParsedQueryTernary
	HasVideo    ParsedQueryTernary
	HasPoll     ParsedQueryTernary
	HasLink     ParsedQueryTernary
	HasTag      ParsedQueryTernary
}

// ParsedQueryTernary represents a filter that may include or exclude.
type ParsedQueryTernary int8

const (
	ParsedQueryTernaryIgnore ParsedQueryTernary = iota
	ParsedQueryTernaryInclude
	ParsedQueryTernaryExclude
)
