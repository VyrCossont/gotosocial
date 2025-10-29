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

package search

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"code.superseriousbusiness.org/gotosocial/internal/config"
	"code.superseriousbusiness.org/gotosocial/internal/db"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"code.superseriousbusiness.org/gotosocial/internal/id"
	"code.superseriousbusiness.org/gotosocial/internal/log"
	searchembedding "code.superseriousbusiness.org/gotosocial/internal/processing/search/embedding"
	"code.superseriousbusiness.org/gotosocial/internal/util"
)

const advancedMigrationID = "20251028213127_pg_add_status_embeddings"
const statusBatchSize = 100

type AdvancedMigrationState struct {
	MinID          string
	MaxIDInclusive string
}

// CreateEmbeddingsForStatuses creates text embeddings for all existing statuses.
func (p *Processor) CreateEmbeddingsForStatuses(ctx context.Context) error {
	backend := config.GetSearchEmbeddingBackend()
	if backend == searchembedding.EmbedderTypeNoop {
		log.Info(ctx, "text embedding search is not enabled. skipping advanced migration with ID %s", advancedMigrationID)
		return nil
	}
	dbtype := strings.ToLower(config.GetDbType())
	if dbtype != "postgres" {
		log.Info(ctx, "text embedding search is not implemented for %s databases. skipping advanced migration with ID %s", dbtype, advancedMigrationID)
		return nil
	}

	advancedMigration, err := p.state.DB.GetAdvancedMigration(ctx, advancedMigrationID)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		return gtserror.Newf("couldn't get advanced migration with ID %s: %w", advancedMigrationID, err)
	}
	state := AdvancedMigrationState{}
	if advancedMigration != nil {
		// There was a previous migration.
		if *advancedMigration.Finished {
			// This migration has already been run to completion; we don't need to run it again.
			return nil
		}
		// Otherwise, pick up where we left off.
		if err := json.Unmarshal(advancedMigration.StateJSON, &state); err != nil {
			// This should never happen.
			return gtserror.Newf("couldn't deserialize advanced migration state from JSON: %w", err)
		}
	} else {
		// Start at the beginning.
		state.MinID = id.Lowest

		// Find the max ID of all existing statuses.
		// This will be the last one we migrate;
		// newer ones will be handled by the normal conversation flow.
		state.MaxIDInclusive, err = p.state.DB.MaxAnyStatusID(ctx)
		if err != nil {
			return gtserror.Newf("couldn't get max status ID for migration: %w", err)
		}

		// Save a new advanced migration record.
		advancedMigration = &gtsmodel.AdvancedMigration{
			ID:       advancedMigrationID,
			Finished: util.Ptr(false),
		}
		if advancedMigration.StateJSON, err = json.Marshal(state); err != nil {
			// This should never happen.
			return gtserror.Newf("couldn't serialize advanced migration state to JSON: %w", err)
		}
		if err := p.state.DB.PutAdvancedMigration(ctx, advancedMigration); err != nil {
			return gtserror.Newf("couldn't save state for advanced migration with ID %s: %w", advancedMigrationID, err)
		}
	}

	log.Info(ctx, "creating text search embeddings for existing statuses…")

	// In batches, get all statuses up to and including the max ID,
	// and create embeddings for them.
	// TODO: (Vyr) this can be parallelized, or at least as parallelized as the embedding server we're talking to can handle.
	for {
		// Get status IDs for this batch.
		statusIDs, err := p.state.DB.GetAnyStatusIDsBatch(ctx, state.MinID, state.MaxIDInclusive, statusBatchSize)
		if err != nil {
			return gtserror.Newf("couldn't get status ID batch for migration: %w", err)
		}
		if len(statusIDs) == 0 {
			break
		}
		log.Infof(ctx, "creating embeddings for %d statuses starting after %s", len(statusIDs), state.MinID)

		// Load the batch by IDs.
		statuses, err := p.state.DB.GetStatusesByIDs(ctx, statusIDs)
		if err != nil {
			return gtserror.Newf("couldn't get statuses for migration: %w", err)
		}

		// Create embeddings for the batch.
		statusEmbeddings, err := p.embedder.CreateStatusEmbeddings(ctx, statuses)
		if err != nil {
			return gtserror.Newf("couldn't create embeddings for %d statuses starting after %s: %w", len(statusIDs), state.MinID, err)
		}
		err = p.state.DB.UpsertStatusEmbeddings(ctx, statusEmbeddings)
		if err != nil {
			return gtserror.Newf("couldn't upsert embeddings for %d statuses starting after %s: %w", len(statusIDs), state.MinID, err)
		}

		// Save the migration state with the new min ID.
		state.MinID = statusIDs[len(statusIDs)-1]
		if advancedMigration.StateJSON, err = json.Marshal(state); err != nil {
			// This should never happen.
			return gtserror.Newf("couldn't serialize advanced migration state to JSON: %w", err)
		}
		if err := p.state.DB.PutAdvancedMigration(ctx, advancedMigration); err != nil {
			return gtserror.Newf("couldn't save state for advanced migration with ID %s: %w", advancedMigrationID, err)
		}
	}

	// Mark the migration as finished.
	advancedMigration.Finished = util.Ptr(true)
	if err := p.state.DB.PutAdvancedMigration(ctx, advancedMigration); err != nil {
		return gtserror.Newf("couldn't save state for advanced migration with ID %s: %w", advancedMigrationID, err)
	}

	log.Info(ctx, "finished creating text search embeddings for existing statuses.")
	return nil
}
