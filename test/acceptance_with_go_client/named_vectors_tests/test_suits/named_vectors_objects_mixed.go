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

package test_suits

import (
	"context"
	"fmt"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	wvt "github.com/weaviate/weaviate-go-client/v5/weaviate"
	"github.com/weaviate/weaviate-go-client/v5/weaviate/fault"
	"github.com/weaviate/weaviate/entities/models"
	"github.com/weaviate/weaviate/entities/schema"
)

func testCreateSchemaWithMixedVectorizers(host string) func(t *testing.T) {
	return func(t *testing.T) {
		ctx := context.Background()
		client, err := wvt.NewClient(wvt.Config{Scheme: "http", Host: host})
		require.Nil(t, err)

		cleanup := func() {
			err := client.Schema().AllDeleter().Do(context.Background())
			require.Nil(t, err)
		}

		t.Run("none vectorizer", func(t *testing.T) {
			cleanup()

			className := "BringYourOwnVector"
			none1 := "none1"
			none2 := "none2"
			mixedTargetVectors := []string{none1, none2, c11y, transformers_bq}
			vector1a := []float32{0.1, 0.2, 0.3}
			vector2a := []float32{-0.1001, 0.2002, -0.3003, -0.4, -0.5}
			vector1b := []float32{0.1111, 0.4, 0.3}
			vector2b := []float32{-0.11, 0.11111, -0.2222, -0.4, -0.5}
			class := &models.Class{
				Class: className,
				Properties: []*models.Property{
					{
						Name: "text", DataType: []string{schema.DataTypeText.String()},
					},
				},
				VectorConfig: map[string]models.VectorConfig{
					none1: {
						Vectorizer: map[string]interface{}{
							"none": nil,
						},
						VectorIndexType: "hnsw",
					},
					none2: {
						Vectorizer: map[string]interface{}{
							"none": nil,
						},
						VectorIndexType: "flat",
					},
					c11y: {
						Vectorizer: map[string]interface{}{
							text2vecContextionary: map[string]interface{}{
								"vectorizeClassName": false,
							},
						},
						VectorIndexType: "hnsw",
					},
					transformers_bq: {
						Vectorizer: map[string]interface{}{
							text2vecTransformers: map[string]interface{}{
								"vectorizeClassName": false,
							},
						},
						VectorIndexType:   "flat",
						VectorIndexConfig: bqFlatIndexConfig(),
					},
				},
			}

			t.Run("create schema", func(t *testing.T) {
				err := client.Schema().ClassCreator().WithClass(class).Do(ctx)
				require.NoError(t, err)

				cls, err := client.Schema().ClassGetter().WithClassName(className).Do(ctx)
				require.NoError(t, err)
				assert.Equal(t, class.Class, cls.Class)
				require.NotEmpty(t, cls.VectorConfig)
				require.Len(t, cls.VectorConfig, len(mixedTargetVectors))
				for _, targetVector := range mixedTargetVectors {
					require.NotEmpty(t, cls.VectorConfig[targetVector])
					assert.NotEmpty(t, cls.VectorConfig[targetVector].VectorIndexType)
					vectorizerConfig, ok := cls.VectorConfig[targetVector].Vectorizer.(map[string]interface{})
					require.True(t, ok)
					assert.Len(t, vectorizerConfig, 1)
				}
			})

			t.Run("add objects", func(t *testing.T) {
				objects := []struct {
					id      string
					text    string
					vectors models.Vectors
				}{
					{
						id:   id1,
						text: "bring your own first vector",
						vectors: models.Vectors{
							none1: vector1a,
							none2: vector2a,
						},
					},
					{
						id:   id2,
						text: "bring your own second vector",
						vectors: models.Vectors{
							none1: vector1b,
							none2: vector2b,
						},
					},
				}
				for _, tt := range objects {
					objWrapper, err := client.Data().Creator().
						WithClassName(className).
						WithID(tt.id).
						WithProperties(map[string]interface{}{
							"text": tt.text,
						}).
						WithVectors(tt.vectors).
						Do(ctx)
					require.NoError(t, err)
					require.NotNil(t, objWrapper)
					assert.Len(t, objWrapper.Object.Vectors, 4)

					objs, err := client.Data().ObjectsGetter().
						WithClassName(className).
						WithID(tt.id).
						WithVector().
						Do(ctx)
					require.NoError(t, err)
					require.Len(t, objs, 1)
					require.NotNil(t, objs[0])
					properties, ok := objs[0].Properties.(map[string]interface{})
					require.True(t, ok)
					assert.Equal(t, tt.text, properties["text"])
					assert.Nil(t, objs[0].Vector)
					assert.Len(t, objs[0].Vectors, len(mixedTargetVectors))
					for targetVector, vector := range tt.vectors {
						require.NotNil(t, objs[0].Vectors[targetVector])
						assert.Equal(t, vector, objs[0].Vectors[targetVector])
					}
				}
			})

			t.Run("update vectors", func(t *testing.T) {
				beforeUpdateVectors := getVectors(t, client, className, id1, mixedTargetVectors...)

				updatedVector1 := []float32{0.11111111111, 0.2222222222, 0.3333333333}
				updatedVector2 := []float32{0.1, 0.2, 0.3, 0.4, 0.5}
				updatedVectors := models.Vectors{
					none1: updatedVector1,
					none2: updatedVector2,
				}

				err := client.Data().Updater().
					WithClassName(className).
					WithID(id1).
					WithVectors(updatedVectors).
					Do(ctx)
				require.NoError(t, err)
				afterUpdateVectors := getVectors(t, client, className, id1, mixedTargetVectors...)
				for targetVector, vector := range updatedVectors {
					assert.NotEqual(t, beforeUpdateVectors[targetVector], afterUpdateVectors[targetVector])
					assert.Equal(t, vector, models.Vector(afterUpdateVectors[targetVector]))
				}
			})

			t.Run("update vectors with merge", func(t *testing.T) {
				beforeUpdateVectors := getVectors(t, client, className, id1, mixedTargetVectors...)

				updatedVector1 := []float32{0.00001, 0.0002, 0.00003}
				updatedVector2 := []float32{1.1, 1.2, 1.3, 1.4, 1.5}
				updatedVectors := models.Vectors{
					none1: updatedVector1,
					none2: updatedVector2,
				}

				err := client.Data().Updater().
					WithMerge().
					WithClassName(className).
					WithID(id1).
					WithProperties(map[string]interface{}{
						"text": "This change should change vector",
					}).
					WithVectors(updatedVectors).
					Do(ctx)
				require.NoError(t, err)
				afterUpdateVectors := getVectors(t, client, className, id1, mixedTargetVectors...)
				for _, targetVector := range mixedTargetVectors {
					assert.NotEqual(t, beforeUpdateVectors[targetVector], afterUpdateVectors[targetVector])
				}
				for targetVector, vector := range updatedVectors {
					assert.Equal(t, vector, models.Vector(afterUpdateVectors[targetVector]))
				}
			})

			t.Run("check BYOV vector names for existence when inserting", func(t *testing.T) {
				_, err = client.Data().Creator().
					WithClassName(className).
					WithID(id3).
					WithProperties(map[string]interface{}{
						"text": "Lorem ipsum dolor sit amet",
					}).
					WithVectors(models.Vectors{
						"non_existent_vector": []float32{0.1, 0.2, 0.3},
					}).
					Do(ctx)
				require.Error(t, err)
				var clientError *fault.WeaviateClientError
				require.True(t, errors.As(err, &clientError))
				require.Equal(t, 422, clientError.StatusCode)
				require.Contains(t, clientError.Msg, fmt.Sprintf("collection %s does not have configuration for vector non_existent_vector", className))
			})

			t.Run("check BYOV vector names for existence when updating", func(t *testing.T) {
				err = client.Data().Updater().
					WithClassName(className).
					WithID(id1).
					WithVectors(models.Vectors{
						"non_existent_vector": []float32{0.1, 0.2, 0.3},
					}).
					Do(ctx)
				require.Error(t, err)
				var clientError *fault.WeaviateClientError
				require.True(t, errors.As(err, &clientError))
				require.Equal(t, 422, clientError.StatusCode)
				require.Contains(t, clientError.Msg, fmt.Sprintf("collection %s does not have configuration for vector non_existent_vector", className))
			})

			t.Run("check BYOV vector names for existence when merge updating", func(t *testing.T) {
				err = client.Data().Updater().WithMerge().
					WithClassName(className).
					WithID(id1).
					WithVectors(models.Vectors{
						"non_existent_vector": []float32{0.1, 0.2, 0.3},
					}).
					Do(ctx)
				require.Error(t, err)
				var clientError *fault.WeaviateClientError
				require.True(t, errors.As(err, &clientError))
				require.Equal(t, 422, clientError.StatusCode)
				require.Contains(t, clientError.Msg, fmt.Sprintf("collection %s does not have configuration for vector non_existent_vector", className))
			})
		})
	}
}
