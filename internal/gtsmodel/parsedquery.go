package gtsmodel

import (
	"github.com/uptrace/bun"
	bundialect "github.com/uptrace/bun/dialect"
	"github.com/uptrace/bun/schema"
)

// ParsedQuery represents the results of parsing the search operator terms within a query.
// It's not currently persisted in the DB, but needs to be available
// at similar places in the build to GTS DB models.
type ParsedQuery struct {
	// Query is the search query text with operator terms removed.
	Query string

	// Operators is the list of operators found in the original search query text.
	Operators []QueryOperator
}

type QueryOperator interface {
	// Modify a search query with this operator.
	Modify(dialect schema.Dialect, q *bun.SelectQuery) *bun.SelectQuery
}

// region ClassicScopeOperator

type ClassicScopeOperator struct {
	requestingAccountID string
}

func (o *ClassicScopeOperator) Modify(dialect schema.Dialect, q *bun.SelectQuery) *bun.SelectQuery {
	return q.WhereGroup(" AND ", func(r *bun.SelectQuery) *bun.SelectQuery {
		return r.
			Where("? = ?", bun.Ident("status.account_id"), o.requestingAccountID).
			WhereOr("? = ?", bun.Ident("status.in_reply_to_account_id"), o.requestingAccountID)
	})
}

func NewClassicScopeOperator(requestingAccountID string) QueryOperator {
	return &ClassicScopeOperator{requestingAccountID: requestingAccountID}
}

// endregion

// region NotNullFilterOperator

type NotNullFilterOperator struct {
	negated bool
	column  bun.Ident
}

func (o *NotNullFilterOperator) Modify(dialect schema.Dialect, q *bun.SelectQuery) *bun.SelectQuery {
	query := "? IS NOT NULL"
	if o.negated {
		query = "? IS NULL"
	}
	return q.Where(query, o.column)
}

func NewIsReplyOperator(negated bool) QueryOperator {
	return &NotNullFilterOperator{
		negated: negated,
		column:  "status.in_reply_to_id",
	}
}

func NewHasPollOperator(negated bool) QueryOperator {
	return &NotNullFilterOperator{
		negated: negated,
		column:  "status.poll_id",
	}
}

// endregion

// region BoolFilterOperator

type BoolFilterOperator struct {
	negated bool
	column  bun.Ident
}

func (o *BoolFilterOperator) Modify(dialect schema.Dialect, q *bun.SelectQuery) *bun.SelectQuery {
	query := "?"
	if o.negated {
		query = "NOT ?"
	}
	return q.Where(query, o.column)
}

func NewIsSensitiveOperator(negated bool) QueryOperator {
	return &BoolFilterOperator{
		negated: negated,
		column:  "status.sensitive",
	}
}

func NewIsLocalOperator(negated bool) QueryOperator {
	return &BoolFilterOperator{
		negated: negated,
		column:  "status.local",
	}
}

func NewIsFederatedOperator(negated bool) QueryOperator {
	return &BoolFilterOperator{
		negated: negated,
		column:  "status.federated",
	}
}

// endregion

// region ValueFilterOperator

type ValueFilterOperator struct {
	negated bool
	column  bun.Ident
	value   any
}

func (o *ValueFilterOperator) Modify(dialect schema.Dialect, q *bun.SelectQuery) *bun.SelectQuery {
	query := "? = ?"
	if o.negated {
		query = "? != ?"
	}
	return q.Where(query, o.column, o.value)
}

func NewFromAccountOperator(negated bool, accountID string) QueryOperator {
	return &ValueFilterOperator{
		negated: negated,
		column:  "status.account_id",
		value:   accountID,
	}
}

func NewToAccountOperator(negated bool, accountID string) QueryOperator {
	return &ValueFilterOperator{
		negated: negated,
		column:  "status.in_reply_to_account_id",
		value:   accountID,
	}
}

func NewIsVisibilityOperator(negated bool, visibility Visibility) QueryOperator {
	return &ValueFilterOperator{
		negated: negated,
		column:  "status.visibility",
		value:   visibility,
	}
}

func NewIsActivityTypeOperator(negated bool, asType string) QueryOperator {
	return &ValueFilterOperator{
		negated: negated,
		column:  "status.activity_streams_type",
		value:   asType,
	}
}

func NewHasContentOperator(negated bool) QueryOperator {
	return &ValueFilterOperator{
		negated: !negated, // Reverse sense from most other operators.
		column:  "status.content",
		value:   "",
	}
}

func NewHasContentWarningOperator(negated bool) QueryOperator {
	return &ValueFilterOperator{
		negated: !negated, // Reverse sense from most other operators.
		column:  "status.content_warning",
		value:   "",
	}
}

// endregion

// region EmptyArrayFilterOperator

type NonEmptyArrayFilterOperator struct {
	negated bool
	column  bun.Ident
}

func (o *NonEmptyArrayFilterOperator) Modify(dialect schema.Dialect, q *bun.SelectQuery) *bun.SelectQuery {
	var query string
	switch dialect.Name() {
	case bundialect.PG:
		query = "COALESCE(CARDINALITY(?), 0)"
	case bundialect.SQLite:
		query = "COALESCE(JSON_ARRAY_LENGTH(?), 0)"
	default:
		panic("db conn was neither pg not sqlite")
	}

	if o.negated {
		query += " = 0"
	} else {
		query += " > 0"
	}

	return q.Where(query, o.column)
}

func NewHasAttachmentOperator(negated bool) QueryOperator {
	return &NonEmptyArrayFilterOperator{
		negated: negated,
		column:  "status.attachments",
	}
}

func NewHasTagOperator(negated bool) QueryOperator {
	return &NonEmptyArrayFilterOperator{
		negated: negated,
		column:  "status.tags",
	}
}

func NewHasMentionOperator(negated bool) QueryOperator {
	return &NonEmptyArrayFilterOperator{
		negated: negated,
		column:  "status.mentions",
	}
}

func NewHasEmojiOperator(negated bool) QueryOperator {
	return &NonEmptyArrayFilterOperator{
		negated: negated,
		column:  "status.emojis",
	}
}

func NewHasEditOperator(negated bool) QueryOperator {
	return &NonEmptyArrayFilterOperator{
		negated: negated,
		column:  "status.edits",
	}
}

// endregion
