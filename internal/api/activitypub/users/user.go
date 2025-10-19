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

package users

import (
	"net/http"

	apiutil "code.superseriousbusiness.org/gotosocial/internal/api/util"
	"code.superseriousbusiness.org/gotosocial/internal/processing"
	"code.superseriousbusiness.org/gotosocial/internal/uris"
	"github.com/gin-gonic/gin"
)

const (
	OnlyOtherAccountsKey   = "only_other_accounts"
	BasePath               = "/:" + apiutil.UsernameKey
	InboxPath              = BasePath + "/" + uris.InboxPath
	OutboxPath             = BasePath + "/" + uris.OutboxPath
	FollowersPath          = BasePath + "/" + uris.FollowersPath
	FollowingPath          = BasePath + "/" + uris.FollowingPath
	FeaturedCollectionPath = BasePath + "/" + uris.CollectionsPath + "/" + uris.FeaturedPath
	StatusPath             = BasePath + "/" + uris.StatusesPath + "/:" + apiutil.IDKey
	StatusRepliesPath      = StatusPath + "/replies"
	AcceptPath             = BasePath + "/" + uris.AcceptsPath + "/:" + apiutil.IDKey
	AuthorizationsPath     = BasePath + "/" + uris.AuthorizationsPath + "/:" + apiutil.IDKey
)

type Module struct {
	processor *processing.Processor
}

func New(processor *processing.Processor) *Module {
	return &Module{
		processor: processor,
	}
}

func (m *Module) Route(attachHandler func(method string, path string, f ...gin.HandlerFunc) gin.IRoutes) {
	attachHandler(http.MethodGet, BasePath, m.UsersGETHandler)
	attachHandler(http.MethodPost, InboxPath, m.InboxPOSTHandler)
	attachHandler(http.MethodGet, FollowersPath, m.FollowersGETHandler)
	attachHandler(http.MethodGet, FollowingPath, m.FollowingGETHandler)
	attachHandler(http.MethodGet, FeaturedCollectionPath, m.FeaturedCollectionGETHandler)
	attachHandler(http.MethodGet, StatusPath, m.StatusGETHandler)
	attachHandler(http.MethodGet, StatusRepliesPath, m.StatusRepliesGETHandler)
	attachHandler(http.MethodGet, OutboxPath, m.OutboxGETHandler)
	attachHandler(http.MethodGet, AcceptPath, m.AcceptGETHandler)
	attachHandler(http.MethodGet, AuthorizationsPath, m.AuthorizationGETHandler)
}
