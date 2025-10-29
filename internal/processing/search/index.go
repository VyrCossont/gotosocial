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
	"errors"

	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	searchembedding "code.superseriousbusiness.org/gotosocial/internal/processing/search/embedding"
)

// Index indexes a status for search.
// Currently, this means creating a text embedding from its contents.
func (p *Processor) Index(ctx context.Context, status *gtsmodel.Status) error {
	statusEmbeddings, err := p.embedder.CreateStatusEmbeddings(ctx, []*gtsmodel.Status{status})
	if errors.Is(err, searchembedding.ErrNotEnabled) {
		return nil
	}
	if err != nil {
		return err
	}
	if len(statusEmbeddings) == 0 {
		// This may happen if the status's text was too long or too short to index.
		return nil
	}

	return p.state.DB.UpsertStatusEmbeddings(ctx, statusEmbeddings)
}

// Deindex removes a status from search indexes given its ID.
func (p *Processor) Deindex(ctx context.Context, statusID string) error {
	_, err := p.state.DB.DeleteStatusEmbeddings(ctx, []string{statusID})
	return err
}
