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

	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
)

const EmbedderTypeNoop string = ""

// noopEmbedder does nothing, always returning the ErrNotEnabled error.
type noopEmbedder struct{}

// NewNoopEmbedder returns an embedder that does nothing.
func NewNoopEmbedder() Embedder {
	return &noopEmbedder{}
}

func (e *noopEmbedder) CreateStatusEmbeddings(ctx context.Context, statuses []*gtsmodel.Status) ([]*gtsmodel.StatusEmbedding, error) {
	return nil, ErrNotEnabled
}

func (e *noopEmbedder) CreateQueryEmbedding(ctx context.Context, query string) (*gtsmodel.QueryEmbedding, error) {
	return nil, ErrNotEnabled
}
