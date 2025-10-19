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
	"code.superseriousbusiness.org/gotosocial/internal/config"
	"code.superseriousbusiness.org/gotosocial/internal/db"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
)

// EmojiGet handles the GET for an emoji originating from this instance.
func (p *Processor) EmojiGet(ctx context.Context, emojiID string) (any, gtserror.WithCode) {
	// Authenticate incoming request.
	//
	// Pass hostname string to this function to indicate
	// it's the instance account being requested, as
	// emojis are always owned by the instance account.
	auth, errWithCode := p.authenticate(ctx, config.GetHost())
	if errWithCode != nil {
		return nil, errWithCode
	}

	if auth.handshakingURI != nil {
		// We're currently handshaking, which means
		// we don't know this account yet. This should
		// be a very rare race condition.
		err := gtserror.Newf("network race handshaking %s", auth.handshakingURI)
		return nil, gtserror.NewErrorInternalError(err)
	}

	// Get the requested emoji.
	emoji, err := p.state.DB.GetEmojiByID(ctx, emojiID)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		err := gtserror.Newf("db error getting emoji %s: %w", emojiID, err)
		return nil, gtserror.NewErrorNotFound(err)
	}

	if emoji == nil {
		err := gtserror.Newf("emoji %s not found in the db", emojiID)
		return nil, gtserror.NewErrorNotFound(err)
	}

	// Only serve *our*
	// emojis on this path.
	if !emoji.IsLocal() {
		err := gtserror.Newf("emoji %s doesn't belong to this instance (domain is %s)", emojiID, emoji.Domain)
		return nil, gtserror.NewErrorNotFound(err)
	}

	// Don't serve emojis that have
	// been disabled by an admin.
	if *emoji.Disabled {
		err := gtserror.Newf("emoji with id %s has been disabled by an admin", emojiID)
		return nil, gtserror.NewErrorNotFound(err)
	}

	apEmoji, err := p.converter.EmojiToAS(ctx, emoji)
	if err != nil {
		err := gtserror.Newf("error converting emoji %s to ap: %s", emojiID, err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	data, err := ap.Serialize(apEmoji)
	if err != nil {
		err := gtserror.Newf("error serializing emoji %s: %w", emojiID, err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	return data, nil
}
