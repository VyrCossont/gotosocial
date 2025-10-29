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

package migrations

import (
	"context"

	"code.superseriousbusiness.org/gotosocial/internal/config"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"code.superseriousbusiness.org/gotosocial/internal/log"
	searchembedding "code.superseriousbusiness.org/gotosocial/internal/processing/search/embedding"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect"
)

func init() {
	up := func(ctx context.Context, db *bun.DB) error {
		return db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
			// Add a new table for status text embeddings.
			// The same schema works for SQLite and PG because SQLite mostly ignores column types.
			if _, err := tx.
				NewCreateTable().
				Model((*gtsmodel.StatusEmbedding)(nil)).
				IfNotExists().
				Exec(ctx); // nocollapse
			err != nil {
				return err
			}

			switch tx.Dialect().Name() {
			case dialect.SQLite:
				// Vector-specific operations aren't implemented for SQLite.
				// TODO: (Vyr) Perhaps they could be with <https://github.com/asg017/sqlite-vec>?
				// Meanwhile, the SQLite version of this table will only be useful for tests.
				log.Warn(ctx, "Text embedding search isn't fully implemented for SQLite")
				return nil

			case dialect.PG:
				// Register the PGVector extension.
				// This will fail if PGVector isn't present.
				// This can also fail if the DB user isn't a superuser for the GTS database,
				// in which case the admin will have to create the extension by hand.
				if _, err := tx.
					NewRaw("CREATE EXTENSION IF NOT EXISTS vector").
					Exec(ctx); // nocollapse
				err != nil {
					return err
				}

				vectorSize := config.GetSearchEmbeddingVectorSize()
				if vectorSize == 0 {
					return gtserror.New("search embedding vector size has not been set")
				}

				distanceMetric := searchembedding.DistanceMetric(config.GetSearchEmbeddingDistanceMetric())
				switch distanceMetric {
				case searchembedding.DistanceMetricCosine, searchembedding.DistanceMetricL1, searchembedding.DistanceMetricL2:
					break
				default:
					return gtserror.Newf("search embedding distance metric has not been set or is an unknown value")
				}

				// Add an approximate search index.
				// Uses HNSW since that doesn't require any data and is faster than IVFFlat.
				// Avoids hardcoding the embedding vector size from the GTS config into the StatusEmbedding struct,
				// but we do need to use it here to set up the index. Not known if this has a performance impact.
				// See: <https://github.com/pgvector/pgvector?tab=readme-ov-file#hnsw>
				// See: <https://github.com/pgvector/pgvector?tab=readme-ov-file#ivfflat>
				// See: <https://github.com/pgvector/pgvector?tab=readme-ov-file#can-i-store-vectors-with-different-dimensions-in-the-same-column>
				// TODO: (Vyr) this kind of config-sensitive migration seems like it needs its own admin tool,
				// 	since changing the model, vector size, or distance metric after this runs renders the existing data and index useless.
				if _, err := tx.
					NewCreateIndex().
					Model((*gtsmodel.StatusEmbedding)(nil)).
					Index("status_embeddings_embedding_approx_idx").
					Using("hnsw").
					ColumnExpr("(embedding::vector(?)) ?", vectorSize, bun.Safe(distanceMetric.PGIndex())).
					IfNotExists().
					Exec(ctx); // nocollapse
				err != nil {
					return err
				}

			default:
				panic("unsupported db type")
			}

			return nil
		})
	}

	down := func(ctx context.Context, db *bun.DB) error {
		return nil
	}

	if err := Migrations.Register(up, down); err != nil {
		panic(err)
	}
}
