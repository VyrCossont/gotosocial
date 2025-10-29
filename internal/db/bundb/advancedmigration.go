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
	"bytes"
	"context"
	"encoding/hex"

	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"code.superseriousbusiness.org/gotosocial/internal/state"
	"github.com/uptrace/bun"
)

type advancedMigrationDB struct {
	db    *bun.DB
	state *state.State
}

func (a *advancedMigrationDB) GetAdvancedMigration(ctx context.Context, id string) (*gtsmodel.AdvancedMigration, error) {
	var advancedMigration gtsmodel.AdvancedMigration
	err := a.db.NewSelect().
		Model(&advancedMigration).
		Where("? = ?", bun.Ident("id"), id).
		Limit(1).
		Scan(ctx)
	if err != nil {
		return nil, err
	}
	// TODO: (Vyr) seems to affect all advanced migrations with a PG database. What's going on here?
	// 	See: <https://www.postgresql.org/docs/current/datatype-binary.html#DATATYPE-BINARY-BYTEA-HEX-FORMAT>
	if base16Bytes, isPGByteA := bytes.CutPrefix(advancedMigration.StateJSON, []byte{'\\', 'x'}); isPGByteA {
		advancedMigration.StateJSON = make([]byte, hex.DecodedLen(len(base16Bytes)))
		_, err := hex.Decode(advancedMigration.StateJSON, base16Bytes)
		if err != nil {
			return nil, gtserror.New("detected PG bytea format for StateJSON but couldn't decode it")
		}
	}
	return &advancedMigration, nil
}

func (a *advancedMigrationDB) PutAdvancedMigration(ctx context.Context, advancedMigration *gtsmodel.AdvancedMigration) error {
	_, err := NewUpsert(a.db).
		Model(advancedMigration).
		Constraint("id").
		Exec(ctx)
	return err
}
