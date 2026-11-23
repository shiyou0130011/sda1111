package errors

import (
	stderrors "errors"
	"fmt"
)

var (
	ErrNilImageReader             = stderrors.New("image reader is nil")
	ErrImageTooLarge              = stderrors.New("image exceeds maximum size")
	ErrNotPNG                     = stderrors.New("image is not a PNG")
	ErrTruncatedPNGChunk          = stderrors.New("truncated PNG chunk")
	ErrTruncatedPNGChunkData      = stderrors.New("truncated PNG chunk data")
	ErrInvalidPNGChecksum         = stderrors.New("invalid PNG checksum for chunk")
	ErrPNGMissingIEND             = stderrors.New("PNG is missing its IEND chunk")
	ErrPNGParametersMissing       = stderrors.New("PNG does not contain parameters metadata")
	ErrMalformedPNGTextChunk      = stderrors.New("malformed PNG text chunk")
	ErrUnsupportedZTXtCompression = stderrors.New("unsupported or malformed zTXt compression")
	ErrMalformedITXtChunk         = stderrors.New("malformed iTXt PNG text chunk")
	ErrUnsupportedITXtCompression = stderrors.New("unsupported iTXt compression")
	ErrPNGMetadataTooLarge        = stderrors.New("PNG metadata exceeds maximum size")
	ErrUnsupportedPNGTextChunk    = stderrors.New("unsupported PNG text chunk")
	ErrPNGMetadataDecompression   = stderrors.New("decompress PNG metadata")
	ErrInvalidPNGSeed             = stderrors.New("invalid Seed metadata")
	ErrPNGPromptMissing           = stderrors.New("PNG parameters metadata does not contain a prompt")
	ErrPNGSeedMissing             = stderrors.New("PNG parameters metadata does not contain a Seed")
	ErrPNGModelMissing            = stderrors.New("PNG parameters metadata does not contain a Model")
	ErrImageRead                  = stderrors.New("read image")
)

func WrapImageRead(err error) error {
	return fmt.Errorf("%w: %w", ErrImageRead, err)
}

func ImageTooLarge(size, limit int) error {
	return fmt.Errorf("%w: got %d bytes, limit %d", ErrImageTooLarge, size, limit)
}

func InvalidPNGChecksum(chunkType []byte) error {
	return fmt.Errorf("%w %q", ErrInvalidPNGChecksum, chunkType)
}

func MalformedPNGTextChunk(chunkType []byte) error {
	return fmt.Errorf("%w %q", ErrMalformedPNGTextChunk, chunkType)
}

func UnsupportedPNGTextChunk(chunkType []byte) error {
	return fmt.Errorf("%w %q", ErrUnsupportedPNGTextChunk, chunkType)
}

func WrapPNGMetadataDecompression(err error) error {
	return fmt.Errorf("%w: %w", ErrPNGMetadataDecompression, err)
}

func InvalidPNGSeed(value string, err error) error {
	return fmt.Errorf("%w %q: %w", ErrInvalidPNGSeed, value, err)
}
