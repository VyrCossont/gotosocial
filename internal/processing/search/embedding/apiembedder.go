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
	"context"
	"strings"

	statusfilter "code.superseriousbusiness.org/gotosocial/internal/filter/status"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"code.superseriousbusiness.org/gotosocial/internal/httpclient"
	"github.com/openai/openai-go/v3/option"
	"github.com/openai/openai-go/v3/packages/param"

	"github.com/openai/openai-go/v3"
)

const EmbedderTypeAPI string = "api"

// apiEmbedder is an OpenAI-API-backed embedder,
// Note that this does *not* mean OpenAI is required as a backend.
// Several projects (including Ollama) implement a subset of their API,
// and we only need the one API method for creating a text embedding.
type apiEmbedder struct {
	apiClient      openai.Client
	model          string
	vectorSize     int
	documentPrompt string
	queryPrompt    string
	maxChars       int
}

// NewEmbedder creates an API-backed embedder using a GTS HTTP client.
func NewEmbedder(httpClient *httpclient.Client, apiBaseURL string, apiKey string, model string, vectorSize int, documentPrompt string, queryPrompt string, maxChars int) Embedder {
	// TODO: (Vyr) there are additional client options (timeouts, etc.) that could be exposed in the GTS config file.
	return &apiEmbedder{
		apiClient: openai.NewClient(
			option.WithBaseURL(apiBaseURL),
			option.WithAPIKey(apiKey),
			option.WithHTTPClient(httpClient),
		),
		model:          model,
		vectorSize:     vectorSize,
		documentPrompt: documentPrompt,
		queryPrompt:    queryPrompt,
		maxChars:       maxChars,
	}
}

// numStatusIDsInLogs is the number of status IDs to show in batch IDs in log messages.
const numStatusIDsInLogs int = 3

// formatStatusBatchID returns a string containing the first few IDs in a batch of statuses.
func formatStatusBatchID(statuses []*gtsmodel.Status) string {
	statusIDs := make([]string, 0, numStatusIDsInLogs+1)
	for i := 0; i < min(len(statuses), numStatusIDsInLogs); i++ {
		statusIDs = append(statusIDs, statuses[i].ID)
	}
	statusBatchID := strings.Join(statusIDs, ", ")
	if len(statuses) > numStatusIDsInLogs {
		statusBatchID += "…"
	}
	return "(" + statusBatchID + ")"
}

func (e *apiEmbedder) CreateStatusEmbeddings(ctx context.Context, statuses []*gtsmodel.Status) ([]*gtsmodel.StatusEmbedding, error) {
	// Used in log messages.
	statusBatchID := formatStatusBatchID(statuses)

	// Collect statuses with non-empty text.
	statusIDs := make([]string, 0, len(statuses))
	statusTexts := make([]string, 0, len(statuses))
	for _, status := range statuses {
		statusText := strings.Join(statusfilter.GetFilterableFields(status), "\n")
		if strings.TrimSpace(statusText) == "" {
			continue
		}
		if e.maxChars > 0 && len(statusText) > e.maxChars {
			continue
		}
		statusIDs = append(statusIDs, status.ID)
		statusTexts = append(statusTexts, e.documentPrompt+statusText)
	}

	if len(statusIDs) == 0 {
		return nil, nil
	}

	// Request embeddings for the batch of statuses.
	statusEmbeddings := make([]*gtsmodel.StatusEmbedding, 0, len(statusIDs))
	resp, err := e.apiClient.Embeddings.New(ctx, openai.EmbeddingNewParams{
		Input: openai.EmbeddingNewParamsInputUnion{
			OfArrayOfStrings: statusTexts,
		},
		Model:          e.model,
		Dimensions:     param.NewOpt(int64(e.vectorSize)),
		EncodingFormat: "float",
	})
	if err != nil {
		err := gtserror.Newf("error getting text embedding for status batch %s: %w", statusBatchID, err)
		return nil, err
	}
	if len(resp.Data) != len(statusIDs) {
		err := gtserror.Newf("wrong number of embeddings returned for status batch %s: expected %d, got %d", statusBatchID, len(statusIDs), len(resp.Data))
		return nil, err
	}
	for _, data := range resp.Data {
		embedding := data.Embedding
		if len(embedding) != e.vectorSize {
			err := gtserror.Newf("wrong vector size for embedding returned for status batch %s: expected %d, got %d", statusBatchID, e.vectorSize, len(embedding))
			return nil, err
		}
		statusEmbeddings = append(statusEmbeddings, &gtsmodel.StatusEmbedding{
			StatusID:  statusIDs[data.Index],
			Embedding: embedding,
		})
	}

	return statusEmbeddings, nil
}

func (e *apiEmbedder) CreateQueryEmbedding(ctx context.Context, query string) (*gtsmodel.QueryEmbedding, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return nil, ErrQueryEmpty
	}
	if e.maxChars > 0 && len(query) > e.maxChars {
		return nil, ErrQueryTooLong
	}

	resp, err := e.apiClient.Embeddings.New(ctx, openai.EmbeddingNewParams{
		Input: openai.EmbeddingNewParamsInputUnion{
			OfString: param.NewOpt(e.queryPrompt + query),
		},
		Model:          e.model,
		Dimensions:     param.NewOpt(int64(e.vectorSize)),
		EncodingFormat: "float",
	})
	if err != nil {
		err := gtserror.Newf("error getting text embedding for search query: %w", err)
		return nil, err
	}
	if len(resp.Data) != 1 {
		err := gtserror.Newf("wrong number of embeddings returned for search query: expected 1, got %d", len(resp.Data))
		return nil, err
	}
	embedding := resp.Data[0].Embedding
	if len(embedding) != e.vectorSize {
		err := gtserror.Newf("wrong vector size for embedding returned for search query: expected %d, got %d", e.vectorSize, len(embedding))
		return nil, err
	}
	return &gtsmodel.QueryEmbedding{Embedding: embedding}, nil
}
