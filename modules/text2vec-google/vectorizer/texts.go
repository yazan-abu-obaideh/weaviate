//                           _       _
// __      _____  __ ___   ___  __ _| |_ ___
// \ \ /\ / / _ \/ _` \ \ / / |/ _` | __/ _ \
//  \ V  V /  __/ (_| |\ V /| | (_| | ||  __/
//   \_/\_/ \___|\__,_| \_/ |_|\__,_|\__\___|
//
//  Copyright © 2016 - 2025 Weaviate B.V. All rights reserved.
//
//  CONTACT: hello@weaviate.io
//

package vectorizer

import (
	"context"

	"github.com/pkg/errors"
	"github.com/weaviate/weaviate/entities/moduletools"
	"github.com/weaviate/weaviate/modules/text2vec-google/ent"
	libvectorizer "github.com/weaviate/weaviate/usecases/vectorizer"
)

func (v *Vectorizer) Texts(ctx context.Context, inputs []string,
	cfg moduletools.ClassConfig,
) ([]float32, error) {
	settings := NewClassSettings(cfg)
	res, err := v.client.VectorizeQuery(ctx, inputs, ent.VectorizationConfig{
		ApiEndpoint: settings.ApiEndpoint(),
		ProjectID:   settings.ProjectID(),
		Model:       settings.ModelID(),
	})
	if err != nil {
		return nil, errors.Wrap(err, "remote client vectorize")
	}
	return libvectorizer.CombineVectors(res.Vectors), nil
}
