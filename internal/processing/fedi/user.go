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

	"code.superseriousbusiness.org/gotosocial/internal/ap"
	"code.superseriousbusiness.org/gotosocial/internal/db"
	"code.superseriousbusiness.org/gotosocial/internal/gtscontext"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
)

// UserGet handles getting an AP representation of an account.
// It does auth before returning a JSON serializable interface to the caller.
func (p *Processor) UserGet(
	ctx context.Context,
	requestedUser string,
) (any, gtserror.WithCode) {
	// Authenticate incoming request, getting related accounts.
	//
	// We may currently be handshaking with the remote account
	// making the request. Unlike with other fedi endpoints,
	// don't bother checking this; if we're still handshaking
	// just serve the AP representation of our account anyway.
	//
	// This ensures that we don't get stuck in a loop with another
	// GtS instance, where each instance is trying repeatedly to
	// dereference the other account that's making the request
	// before it will reveal its own account.
	//
	// Instead, we end up in an 'I'll show you mine if you show me
	// yours' situation, where we sort of agree to reveal each
	// other's profiles at the same time.
	auth, errWithCode := p.authenticate(ctx, requestedUser)
	if errWithCode != nil {
		return nil, errWithCode
	}

	// Generate the proper AP representation.
	accountable, err := p.converter.AccountToAS(ctx, auth.receiver)
	if err != nil {
		err := gtserror.Newf("error converting to accountable: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	data, err := ap.Serialize(accountable)
	if err != nil {
		err := gtserror.Newf("error serializing accountable: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	return data, nil
}

// UserGetMinimal returns a minimal AP representation
// of the requested account, containing just the public
// key, without doing authentication.
func (p *Processor) UserGetMinimal(
	ctx context.Context,
	requestedUser string,
) (any, gtserror.WithCode) {
	acct, err := p.state.DB.GetAccountByUsernameDomain(
		gtscontext.SetBarebones(ctx),
		requestedUser, "",
	)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		err := gtserror.Newf("db error getting account %s: %w", requestedUser, err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	if acct == nil {
		err := gtserror.Newf("account %s not found in the db", requestedUser)
		return nil, gtserror.NewErrorNotFound(err)
	}

	// Generate minimal AP representation.
	accountable, err := p.converter.AccountToASMinimal(ctx, acct)
	if err != nil {
		err := gtserror.Newf("error converting to accountable: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	data, err := ap.Serialize(accountable)
	if err != nil {
		err := gtserror.Newf("error serializing accountable: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	return data, nil
}
