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

package bundb

import (
	"context"

	"code.superseriousbusiness.org/gotosocial/internal/config"
	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"code.superseriousbusiness.org/gotosocial/internal/log"
	searchembedding "code.superseriousbusiness.org/gotosocial/internal/processing/search/embedding"
	"code.superseriousbusiness.org/gotosocial/internal/state"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect"
)

type statusEmbeddingDB struct {
	db    *bun.DB
	state *state.State
}

func (s *statusEmbeddingDB) UpsertStatusEmbeddings(ctx context.Context, statusEmbeddings []*gtsmodel.StatusEmbedding) error {
	_, err := NewUpsert(s.db).
		Model(&statusEmbeddings).
		Constraint("status_id").
		Column("embedding").
		Exec(ctx)
	return err
}

func (s *statusEmbeddingDB) DeleteStatusEmbeddings(ctx context.Context, statusIDs []string) (int64, error) {
	result, err := s.db.NewDelete().
		Model((*gtsmodel.StatusEmbedding)(nil)).
		Where("? IN (?)", bun.Ident("status_id"), bun.In(statusIDs)).
		Exec(ctx)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

func (s *statusEmbeddingDB) StatusEmbeddingBestMatches(ctx context.Context, queryEmbedding *gtsmodel.QueryEmbedding, count int) ([]*gtsmodel.Status, error) {
	if s.db.Dialect().Name() != dialect.PG {
		// TODO: (Vyr) investigate SQLite vector search extensions
		log.Warnf(ctx, "Text embedding vector search isn't supported for DB dialect %s", s.db.Dialect().Name())
		return nil, nil
	}
	distanceMetric := searchembedding.DistanceMetric(config.GetSearchEmbeddingDistanceMetric())

	// Search using PGVector with whatever distance metric is configured.
	// Should hit the index on the search embeddings table.
	vectorSize := config.GetSearchEmbeddingVectorSize()
	var statusIDs []string
	if _, err := s.db.NewSelect().
		Model((*gtsmodel.StatusEmbedding)(nil)).
		Column("status_id").
		OrderExpr("?::vector(?) ? ?", bun.Ident("embedding"), vectorSize, bun.Safe(distanceMetric.PGOperator()), queryEmbedding.Embedding).
		Limit(count).
		Exec(ctx, &statusIDs); // nocollapse
	err != nil {
		return nil, err
	}

	return s.state.DB.GetStatusesByIDs(ctx, statusIDs)
}
