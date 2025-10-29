// GoToSocial
// Copyright (C) GoToSocial Authors admin@gotosocial.org
// SPDX-License-Identifier: AGPL-3.0-or-later
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package embedding

import (
	"context"
	"errors"

	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
)

// Embedder can generate a text embedding vector for a status.
type Embedder interface {
	// CreateStatusEmbeddings takes a batch of statuses and tries to return an embedding vector for them.
	// The output is not necessarily in the same order.
	// In the event an embedding cannot be created, it is allowed to return fewer output status embeddings than input statuses.
	CreateStatusEmbeddings(ctx context.Context, statuses []*gtsmodel.Status) ([]*gtsmodel.StatusEmbedding, error)

	// CreateQueryEmbedding tries to return a query embedding for a given search query.
	// If the query is empty, it will return the ErrQueryEmpty error.
	// If the query is too long, it will return the ErrQueryTooLong error.
	CreateQueryEmbedding(ctx context.Context, query string) (*gtsmodel.QueryEmbedding, error)
}

// ErrNotEnabled is a special error indicating that text embedding search is not enabled,
// so that status handling paths can bypass any further processing if they see this.
var ErrNotEnabled = errors.New("text embedding search is not enabled")

// ErrQueryEmpty indicates that a search query was empty.
var ErrQueryEmpty = errors.New("search query is empty")

// ErrQueryTooLong indicates that a search query was too long to create an embedding from.
var ErrQueryTooLong = errors.New("search query is too long")
