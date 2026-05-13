package partitioning

import (
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data/batch"
	"github.com/multiversx/mx-chain-core-go/marshal"
)

// SimpleDataPacker can split a large slice of byte slices in chunks <= maxPacketSize
// If one element still exceeds maxPacketSize, it will be returned alone
// It does the marshaling of the resulted (smaller) slice of byte slices
// This is a simpler version of a data packer that does not marshall in a repetitive manner currentChunk slice
// as the SizeDataPacker does. This limitation is lighter in terms of CPU cycles and memory used but is not as precise
// as SizeDataPacker.
type SimpleDataPacker struct {
	marshalizer marshal.Marshalizer
}

// NewSimpleDataPacker creates a new SizeDataPacker instance
func NewSimpleDataPacker(marshalizer marshal.Marshalizer) (*SimpleDataPacker, error) {
	if marshalizer == nil || marshalizer.IsInterfaceNil() {
		return nil, core.ErrNilMarshalizer
	}

	return &SimpleDataPacker{
		marshalizer: marshalizer,
	}, nil
}

// PackDataInChunks packs the provided data into smaller chunks
// limit is expressed in bytes
func (sdp *SimpleDataPacker) PackDataInChunks(data [][]byte, limit int) ([][]byte, error) {
	if limit < minimumMaxPacketSizeInBytes {
		return nil, core.ErrInvalidValue
	}
	if data == nil {
		return nil, core.ErrNilInputData
	}

	returningBuff := make([][]byte, 0)

	currentChunk := make([][]byte, 0)
	lenChunk := 0
	for _, element := range data {
		isBuffToLarge := lenChunk+len(element) >= limit
		chunkNotEmpty := len(currentChunk) > 0
		if isBuffToLarge && chunkNotEmpty {
			// ISSUE-043: surface marshal errors on intermediate chunks
			// rather than silently appending a zero-length / partial
			// `marshaledChunk` to the output. The final-flush path below
			// already returns the error correctly; making the intermediate
			// path symmetric closes a P2P-data-integrity silent-loss bug
			// where a marshal failure mid-packing would produce a buffer
			// list whose i-th entry is empty bytes and whose subsequent
			// entries are correct — undetectable by the caller.
			marshaledChunk, err := sdp.marshalizer.Marshal(&batch.Batch{Data: currentChunk})
			if err != nil {
				return nil, err
			}
			returningBuff = append(returningBuff, marshaledChunk)
			currentChunk = make([][]byte, 0)
			lenChunk = 0
		}

		currentChunk = append(currentChunk, element)
		lenChunk += len(element)
	}

	if len(currentChunk) > 0 {
		marshaledElements, err := sdp.marshalizer.Marshal(&batch.Batch{Data: currentChunk})
		if err != nil {
			return nil, err
		}
		returningBuff = append(returningBuff, marshaledElements)
	}

	return returningBuff, nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (sdp *SimpleDataPacker) IsInterfaceNil() bool {
	return sdp == nil
}
