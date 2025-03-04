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

	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect"
)

func init() {
	up := func(ctx context.Context, db *bun.DB) error {
		return db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
			switch tx.Dialect().Name() {
			case dialect.SQLite:
				// SQLite's virtual tables are too messy for this.
				return nil

			case dialect.PG:
				// Add the full-text index.
				if _, err := tx.
					NewCreateIndex().
					Model((*gtsmodel.Status)(nil)).
					Index("statuses_full_text_english_idx").
					// TODO: (Vyr) this is uncooked and doesn't include media alt text or poll options but does include HTML
					ColumnExpr("(to_tsvector('english', coalesce(content_warning, '')) || to_tsvector('english', content))").
					// TODO: (Vyr) could further restrict it with `language = 'en' or starts_with(language, 'en-')` but many posts don't have language tags at all
					Where("boost_of_id is null").
					Using("gin").
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
