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

// DistanceMetric is a vector distance metric used for text embedding queries.
// TODO: (Vyr) there are a lot of caveats and options for vector indexing.
// See: <https://github.com/pgvector/pgvector?tab=readme-ov-file#filtering>
type DistanceMetric string

// PGOperator returns the distance operator for a pgvector query.
func (dm DistanceMetric) PGOperator() string {
	switch dm {
	case DistanceMetricCosine:
		return "<=>"
	case DistanceMetricL1:
		return "<+>"
	case DistanceMetricL2:
		return "<->"
	default:
		return ""
	}
}

// PGIndex returns the vector operations type for a pgvector index.
func (dm DistanceMetric) PGIndex() string {
	switch dm {
	case DistanceMetricCosine:
		return "vector_cosine_ops"
	case DistanceMetricL1:
		return "vector_l1_ops"
	case DistanceMetricL2:
		return "vector_l2_ops"
	default:
		return ""
	}
}

const (
	DistanceMetricCosine DistanceMetric = "cosine"
	DistanceMetricL1     DistanceMetric = "l1"
	DistanceMetricL2     DistanceMetric = "l2"
)
