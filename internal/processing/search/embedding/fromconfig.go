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

import (
	"net/netip"
	"net/url"

	"code.superseriousbusiness.org/gotosocial/internal/config"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"code.superseriousbusiness.org/gotosocial/internal/httpclient"
)

// FromConfig creates a no-op, API-backed, or internal text embedder based on GTS config.
func FromConfig() (Embedder, error) {
	embeddingBackend := config.GetSearchEmbeddingBackend()
	switch embeddingBackend {
	case EmbedderTypeAPI:
		apiBaseURL := config.GetSearchEmbeddingAPIBaseURL()
		if apiBaseURL == "" {
			return nil, gtserror.Newf("search embedding backend is %s but API base URL has not been set", embeddingBackend)
		}
		// This can be empty if no authentication is required (self-hosted Ollama etc.)
		apiKey := config.GetSearchEmbeddingAPIKey()
		model := config.GetSearchEmbeddingModel()
		if model == "" {
			return nil, gtserror.Newf("search embedding backend is %s but model name has not been set", embeddingBackend)
		}
		vectorSize := config.GetSearchEmbeddingVectorSize()
		if vectorSize == 0 {
			return nil, gtserror.Newf("search embedding backend is %s but vector size has not been set", embeddingBackend)
		}
		distanceMetric := DistanceMetric(config.GetSearchEmbeddingDistanceMetric())
		switch distanceMetric {
		case DistanceMetricCosine, DistanceMetricL1, DistanceMetricL2:
			break
		default:
			return nil, gtserror.Newf("search embedding backend is %s but distance metric has not been set or is an unknown value", embeddingBackend)
		}
		// One or both of these prompts may be empty, depending on the model.
		documentPrompt := config.GetSearchEmbeddingPromptDocument()
		queryPrompt := config.GetSearchEmbeddingPromptQuery()
		maxChars := config.GetSearchEmbeddingMaxChars()
		if maxChars == 0 {
			return nil, gtserror.Newf("search embedding backend is %s but max input length has not been set", embeddingBackend)
		}

		// Same timeout and TLS bypass settings GTS uses for other outbound HTTP calls.
		// However, IP ranges are different.
		// TODO: (Vyr) we should have a way to allow connections to localhost and allow cleartext HTTP only for embedding providers.
		// 	We'll have the same problem when we add webhooks.
		//  For now, if the URL scheme is unencrypted HTTP, we always require the embedding server to be on localhost.
		//  In the future, service providers like this should have separately configurable allowed/denied IP ranges from the normal GTS HTTP client.
		clientConfig := httpclient.Config{
			Timeout:               config.GetHTTPClientTimeout(),
			TLSInsecureSkipVerify: config.GetHTTPClientTLSInsecureSkipVerify(),
		}
		parsedURL, err := url.Parse(apiBaseURL)
		if err != nil {
			return nil, gtserror.Newf("search embedding backend is %s but couldn't parse API base URL %s: %w", embeddingBackend, apiBaseURL, err)
		}
		switch parsedURL.Scheme {
		case "https":
			// Allow connections to any HTTPS server, including private, unroutable, localhost, etc. ranges.
			clientConfig.AllowRanges = []netip.Prefix{
				netip.PrefixFrom(netip.IPv4Unspecified(), 0),
				netip.PrefixFrom(netip.IPv6Unspecified(), 0),
			}
			clientConfig.BlockRanges = nil

		case "http":
			// Allow connections only to localhost.
			clientConfig.AllowRanges = []netip.Prefix{
				// All 127.x.x.x IPv4 addresses are considered loopback addresses.
				netip.PrefixFrom(netip.AddrFrom4([4]byte{127, 0, 0, 0}), 8),
				// IPv6 has a single loopback address.
				netip.PrefixFrom(netip.IPv6Loopback(), 128),
			}
			// Block the entire IPv4 and IPv6 address spaces. (Blocks are applied after allows.)
			clientConfig.BlockRanges = []netip.Prefix{
				netip.PrefixFrom(netip.IPv4Unspecified(), 0),
				netip.PrefixFrom(netip.IPv6Unspecified(), 0),
			}

		default:
			return nil, gtserror.Newf("search embedding backend is %s but API base URL %s has an unknown or unsupported scheme: %s", embeddingBackend, apiBaseURL, parsedURL.Scheme)
		}
		return NewEmbedder(
			httpclient.New(clientConfig),
			apiBaseURL,
			apiKey,
			model,
			vectorSize,
			documentPrompt,
			queryPrompt,
			maxChars,
		), nil

	case EmbedderTypeNoop:
		return NewNoopEmbedder(), nil

	default:
		// TODO: (Vyr) includes "internal" for now, which isn't implemented yet.
		return nil, gtserror.Newf("unknown or unimplemented search embedding backend: %s", embeddingBackend)
	}
}
