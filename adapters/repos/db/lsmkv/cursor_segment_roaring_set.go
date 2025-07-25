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

package lsmkv

import (
	"github.com/weaviate/weaviate/adapters/repos/db/lsmkv/segmentindex"
	"github.com/weaviate/weaviate/adapters/repos/db/roaringset"
)

func (s *segment) newRoaringSetCursor() *roaringset.SegmentCursor {
	return roaringset.NewSegmentCursor(s.contents[s.dataStartPos:s.dataEndPos],
		&roaringSetSeeker{s.index})
}

func (sg *SegmentGroup) newRoaringSetCursors() ([]roaringset.InnerCursor, func()) {
	segments, release := sg.getAndLockSegments()

	out := make([]roaringset.InnerCursor, len(segments))

	for i, segment := range segments {
		out[i] = segment.newRoaringSetCursor()
	}

	return out, release
}

// diskIndex returns node's Start and End offsets
// taking into account HeaderSize. SegmentCursor of RoaringSet
// accepts only payload part of underlying segment content, therefore
// offsets should be adjusted and reduced by HeaderSize
type roaringSetSeeker struct {
	diskIndex diskIndex
}

func (s *roaringSetSeeker) Seek(key []byte) (segmentindex.Node, error) {
	node, err := s.diskIndex.Seek(key)
	if err != nil {
		return segmentindex.Node{}, err
	}
	return segmentindex.Node{
		Key:   node.Key,
		Start: node.Start - segmentindex.HeaderSize,
		End:   node.End - segmentindex.HeaderSize,
	}, nil
}
