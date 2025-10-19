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
	"fmt"

	"code.superseriousbusiness.org/gotosocial/internal/ap"
	"code.superseriousbusiness.org/gotosocial/internal/db"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
)

// AuthorizationGet handles the getting of a fedi/activitypub
// representation of a local interaction authorization.
//
// It performs appropriate authentication before
// returning a JSON serializable interface.
func (p *Processor) AuthorizationGet(
	ctx context.Context,
	requestedUser string,
	intReqID string,
) (any, gtserror.WithCode) {
	// Ensure valid request, intReq exists, etc.
	intReq, errWithCode := p.validateAuthGetRequest(ctx, requestedUser, intReqID)
	if errWithCode != nil {
		return nil, errWithCode
	}

	// Convert + serialize the Authorization.
	authorization, err := p.converter.InteractionReqToASAuthorization(ctx, intReq)
	if err != nil {
		err := gtserror.Newf("error converting to authorization: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	data, err := ap.Serialize(authorization)
	if err != nil {
		err := gtserror.Newf("error serializing accept: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	return data, nil
}

// AcceptGet handles the getting of a fedi/activitypub
// representation of a local interaction acceptance.
//
// It performs appropriate authentication before
// returning a JSON serializable interface.
func (p *Processor) AcceptGet(
	ctx context.Context,
	requestedUser string,
	intReqID string,
) (any, gtserror.WithCode) {
	// Ensure valid request, intReq exists, etc.
	intReq, errWithCode := p.validateAuthGetRequest(ctx, requestedUser, intReqID)
	if errWithCode != nil {
		return nil, errWithCode
	}

	// Convert + serialize the Accept.
	accept, err := p.converter.InteractionReqToASAccept(ctx, intReq)
	if err != nil {
		err := gtserror.Newf("error converting to accept: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	data, err := ap.Serialize(accept)
	if err != nil {
		err := gtserror.Newf("error serializing accept: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	return data, nil
}

// validateAuthGetRequest is a shortcut function
// for returning an accepted interaction request
// targeting `requestedUser`.
func (p *Processor) validateAuthGetRequest(
	ctx context.Context,
	requestedUser string,
	intReqID string,
) (*gtsmodel.InteractionRequest, gtserror.WithCode) {
	// Authenticate incoming request, getting related accounts.
	auth, errWithCode := p.authenticate(ctx, requestedUser)
	if errWithCode != nil {
		return nil, errWithCode
	}

	if auth.handshakingURI != nil {
		// We're currently handshaking, which means we don't know
		// this account yet. This should be a very rare race condition.
		err := gtserror.Newf("network race handshaking %s", auth.handshakingURI)
		return nil, gtserror.NewErrorInternalError(err)
	}

	// Fetch interaction request with the given ID.
	req, err := p.state.DB.GetInteractionRequestByID(ctx, intReqID)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		err := gtserror.Newf("db error getting interaction request %s: %w", intReqID, err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	// Ensure that this is an existing
	// and *accepted* interaction request.
	if req == nil || !req.IsAccepted() {
		const text = "interaction request not found"
		return nil, gtserror.NewErrorNotFound(errors.New(text))
	}

	// Ensure interaction request was accepted
	// by the account in the request path.
	if req.TargetAccountID != auth.receiver.ID {
		text := fmt.Sprintf(
			"account %s is not targeted by interaction request %s and therefore can't accept it",
			requestedUser, intReqID,
		)
		return nil, gtserror.NewErrorNotFound(errors.New(text))
	}

	// All fine.
	return req, nil
}
