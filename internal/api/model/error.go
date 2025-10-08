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

package model

// Error is a generic public-facing API error.
//
// swagger:model error
type Error struct {
	// Error is usually a human-readable description of the error.
	//
	// For OAuth-related errors, it will be one of the codes listed in
	// <https://datatracker.ietf.org/doc/html/rfc6749#section-5.2>.
	Error string `json:"error"`

	// ErrorDescription is only used for OAuth errors, and is a human-readable description of the error.
	ErrorDescription string `json:"error_description,omitempty"`
}
