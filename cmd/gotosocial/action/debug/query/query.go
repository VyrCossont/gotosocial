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

package query

import (
	"context"
	"fmt"
	"os"
	"strings"

	"code.superseriousbusiness.org/gotosocial/cmd/gotosocial/action"
	"code.superseriousbusiness.org/gotosocial/internal/db/bundb"
	statusfilter "code.superseriousbusiness.org/gotosocial/internal/filter/status"
	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"code.superseriousbusiness.org/gotosocial/internal/log"
	searchembedding "code.superseriousbusiness.org/gotosocial/internal/processing/search/embedding"
	"code.superseriousbusiness.org/gotosocial/internal/state"
)

// TODO: (Vyr) here thru the defer below was copied from admin commands

var (
	// check function conformance
	_ action.GTSAction = Query
)

func initState(ctx context.Context) (*state.State, error) {
	var state state.State
	state.Caches.Init()
	if err := state.Caches.Start(); err != nil {
		return nil, fmt.Errorf("error starting caches: %w", err)
	}

	// Only set state DB connection.
	// Don't need Actions or Workers for this (yet).
	dbConn, err := bundb.NewBunDBService(ctx, &state)
	if err != nil {
		return nil, fmt.Errorf("error creating dbConn: %w", err)
	}
	state.DB = dbConn

	return &state, nil
}

func stopState(state *state.State) error {
	err := state.DB.Close()
	state.Caches.Stop()
	return err
}

func Query(ctx context.Context) error {
	state, err := initState(ctx)
	if err != nil {
		return err
	}

	defer func() {
		// Ensure state gets stopped on return.
		if err := stopState(state); err != nil {
			log.Error(ctx, err)
		}
	}()

	embedder, err := searchembedding.FromConfig()
	if err != nil {
		return err
	}

	isQuery := false
	for i, arg := range os.Args {
		if !isQuery {
			if arg == "query" && i > 0 && os.Args[i-1] == "debug" {
				isQuery = true
			}
			continue
		}

		println("=== " + arg + " ===")
		println()

		queryEmbedding, err := embedder.CreateQueryEmbedding(ctx, arg)
		if err != nil {
			return err
		}

		statuses, err := state.DB.StatusEmbeddingBestMatches(ctx, queryEmbedding, 40)
		if err != nil {
			return err
		}

		for _, status := range statuses {
			if status.Visibility != gtsmodel.VisibilityPublic {
				continue
			}
			println(status.URL)
			if status.ContentWarning != "" {
				println("CW: " + status.ContentWarning)
			}
			println(strings.Join(statusfilter.GetFilterableFields(status), "\n"))
			println()
		}
		println()
	}

	return nil
}
