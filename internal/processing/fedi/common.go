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

package fedi

import (
	"context"
	"errors"
	"net/url"

	"code.superseriousbusiness.org/gotosocial/internal/db"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
)

type commonAuth struct {
	handshakingURI *url.URL          // Set to requestingAcct's URI if we're currently handshaking them.
	requester      *gtsmodel.Account // Remote account making request to this instance.
	receiver       *gtsmodel.Account // Local account receiving the request.
}

// authenticate is a util function for authenticating a signed GET
// request to one of the AP/fedi resources handled in this package.
func (p *Processor) authenticate(ctx context.Context, requestedUser string) (*commonAuth, gtserror.WithCode) {
	// Get the requested local account
	// with given username from database.
	receiver, err := p.state.DB.GetAccountByUsernameDomain(ctx, requestedUser, "")
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		err = gtserror.Newf("db error getting account %s: %w", requestedUser, err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	if receiver == nil {
		err := gtserror.Newf("account %s not found in the db", requestedUser)
		return nil, gtserror.NewErrorNotFound(err)
	}

	// Ensure request signed, and use signature URI to
	// get requesting account, dereferencing if necessary.
	pubKeyAuth, errWithCode := p.federator.AuthenticateFederatedRequest(ctx, requestedUser)
	if errWithCode != nil {
		return nil, errWithCode
	}

	if pubKeyAuth.Handshaking {
		// We're still handshaking so we
		// don't know the requester yet.
		return &commonAuth{
			handshakingURI: pubKeyAuth.OwnerURI,
			receiver:       receiver,
		}, nil
	}

	// Get requester from auth.
	requester := pubKeyAuth.Owner

	// Check if requester is suspended.
	switch {
	case !requester.IsSuspended():
		// No problem.

	case requester.DeletedSelf():
		// Requester deleted their own account.
		// Why are they now requesting something?
		err := gtserror.Newf("requester %s self-deleted", requester.UsernameDomain())
		return nil, gtserror.NewErrorUnauthorized(err)

	default:
		// Admin from our instance likely suspended account.
		err := gtserror.Newf("requester %s is suspended", requester.UsernameDomain())
		return nil, gtserror.NewErrorForbidden(err)
	}

	// Ensure receiver does not block requester.
	blocked, err := p.state.DB.IsBlocked(ctx, receiver.ID, requester.ID)
	if err != nil {
		err := gtserror.Newf("db error checking block: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	if blocked {
		var text = requestedUser + " blocks " + requester.Username
		return nil, gtserror.NewErrorForbidden(errors.New(text))
	}

	return &commonAuth{
		requester: requester,
		receiver:  receiver,
	}, nil
}
