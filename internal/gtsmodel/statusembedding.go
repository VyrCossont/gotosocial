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

package gtsmodel

// StatusEmbedding stores the text embedding for a given status.
// Practical use of this table requires a DB extension for vector search;
// the details are different between PG and SQLite.
type StatusEmbedding struct {
	// StatusID is the originating status ID this was created from.
	StatusID string `bun:"type:CHAR(26),pk,nullzero,notnull"`
	// Embedding is the embedding vector of this status's text.
	// The dimension is not set here because it is a config file parameter.
	Embedding []float64 `bun:"type:vector,nullzero,notnull"`
}
