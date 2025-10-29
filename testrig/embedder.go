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

package testrig

import (
	"context"

	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
)

// MockEmbedder is a mock of searchembedding.Embedder that stores any status or query passed to it,
// and always returns dummy embeddings with meaningless vectors. The zero value is usable.
type MockEmbedder struct {
	Statuses []*gtsmodel.Status
	Queries  []string
}

func (m *MockEmbedder) CreateStatusEmbeddings(ctx context.Context, statuses []*gtsmodel.Status) ([]*gtsmodel.StatusEmbedding, error) {
	m.Statuses = append(m.Statuses, statuses...)
	statusEmbeddings := make([]*gtsmodel.StatusEmbedding, 0, len(statuses))
	for _, status := range statuses {
		statusEmbeddings = append(statusEmbeddings, &gtsmodel.StatusEmbedding{
			StatusID:  status.ID,
			Embedding: []float64{0.0},
		})
	}
	return statusEmbeddings, nil
}

func (m *MockEmbedder) CreateQueryEmbedding(ctx context.Context, query string) (*gtsmodel.QueryEmbedding, error) {
	m.Queries = append(m.Queries, query)
	return &gtsmodel.QueryEmbedding{Embedding: []float64{0.0}}, nil
}
