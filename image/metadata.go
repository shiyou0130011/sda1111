package image

import (
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"hash/crc32"
	"io"
	"strconv"
	"strings"

	imageerrors "github.com/shiyou0130011/sda1111/errors"
)

const (
	maxImageSize    = 64 << 20
	maxMetadataSize = 1 << 20
)

var pngSignature = []byte{137, 80, 78, 71, 13, 10, 26, 10}

// GetImageMetadata reads a PNG image and returns its prompt, seed, and model.
// The reader can be an open local file or a remote response body.
func GetImageMetadata(reader io.Reader) (prompt string, seed int64, model string, err error) {
	if reader == nil {
		return "", 0, "", imageerrors.ErrNilImageReader
	}

	data, err := io.ReadAll(io.LimitReader(reader, maxImageSize+1))
	if err != nil {
		return "", 0, "", imageerrors.WrapImageRead(err)
	}
	if len(data) > maxImageSize {
		return "", 0, "", imageerrors.ImageTooLarge(len(data), maxImageSize)
	}

	parameters, err := readPNGParameters(data)
	if err != nil {
		return "", 0, "", err
	}
	return parseParameters(parameters)
}

func readPNGParameters(data []byte) (string, error) {
	if len(data) < len(pngSignature) || !bytes.Equal(data[:len(pngSignature)], pngSignature) {
		return "", imageerrors.ErrNotPNG
	}

	offset := len(pngSignature)
	var parameters string
	foundParameters := false
	foundEnd := false

	for offset < len(data) {
		if len(data)-offset < 12 {
			return "", imageerrors.ErrTruncatedPNGChunk
		}

		length := uint64(binary.BigEndian.Uint32(data[offset : offset+4]))
		chunkEnd := uint64(offset) + 12 + length
		if chunkEnd > uint64(len(data)) {
			return "", imageerrors.ErrTruncatedPNGChunkData
		}

		chunkType := data[offset+4 : offset+8]
		chunkData := data[offset+8 : offset+8+int(length)]
		checksum := binary.BigEndian.Uint32(data[offset+8+int(length) : int(chunkEnd)])
		if crc32.ChecksumIEEE(data[offset+4:offset+8+int(length)]) != checksum {
			return "", imageerrors.InvalidPNGChecksum(chunkType)
		}

		if !foundParameters && (bytes.Equal(chunkType, []byte("tEXt")) ||
			bytes.Equal(chunkType, []byte("zTXt")) ||
			bytes.Equal(chunkType, []byte("iTXt"))) {
			keyword, text, err := decodeTextChunk(chunkType, chunkData)
			if err != nil {
				return "", err
			}
			if keyword == "parameters" {
				parameters = text
				foundParameters = true
			}
		}

		offset = int(chunkEnd)
		if bytes.Equal(chunkType, []byte("IEND")) {
			foundEnd = true
			break
		}
	}

	if !foundEnd {
		return "", imageerrors.ErrPNGMissingIEND
	}
	if !foundParameters {
		return "", imageerrors.ErrPNGParametersMissing
	}
	return parameters, nil
}

func decodeTextChunk(chunkType, data []byte) (string, string, error) {
	separator := bytes.IndexByte(data, 0)
	if separator < 0 {
		return "", "", imageerrors.MalformedPNGTextChunk(chunkType)
	}
	keyword := string(data[:separator])
	textData := data[separator+1:]

	switch string(chunkType) {
	case "tEXt":
		return keyword, string(textData), nil
	case "zTXt":
		if len(textData) < 2 || textData[0] != 0 {
			return "", "", imageerrors.ErrUnsupportedZTXtCompression
		}
		text, err := decompressMetadata(textData[1:])
		return keyword, text, err
	case "iTXt":
		if len(textData) < 2 {
			return "", "", imageerrors.ErrMalformedITXtChunk
		}
		compressed := textData[0] == 1
		if textData[0] > 1 || textData[1] != 0 {
			return "", "", imageerrors.ErrUnsupportedITXtCompression
		}
		rest := textData[2:]
		for i := 0; i < 2; i++ {
			separator := bytes.IndexByte(rest, 0)
			if separator < 0 {
				return "", "", imageerrors.ErrMalformedITXtChunk
			}
			rest = rest[separator+1:]
		}
		if compressed {
			text, err := decompressMetadata(rest)
			return keyword, text, err
		}
		if len(rest) > maxMetadataSize {
			return "", "", imageerrors.ErrPNGMetadataTooLarge
		}
		return keyword, string(rest), nil
	default:
		return "", "", imageerrors.UnsupportedPNGTextChunk(chunkType)
	}
}

func decompressMetadata(data []byte) (string, error) {
	reader, err := zlib.NewReader(bytes.NewReader(data))
	if err != nil {
		return "", imageerrors.WrapPNGMetadataDecompression(err)
	}
	defer reader.Close()

	text, err := io.ReadAll(io.LimitReader(reader, maxMetadataSize+1))
	if err != nil {
		return "", imageerrors.WrapPNGMetadataDecompression(err)
	}
	if len(text) > maxMetadataSize {
		return "", imageerrors.ErrPNGMetadataTooLarge
	}
	return string(text), nil
}

func parseParameters(parameters string) (string, int64, string, error) {
	promptEnd := len(parameters)
	for _, marker := range []string{"\nNegative prompt:", "\nSteps:"} {
		if index := strings.Index(parameters, marker); index >= 0 && index < promptEnd {
			promptEnd = index
		}
	}
	prompt := strings.TrimSpace(parameters[:promptEnd])

	var seed int64
	var model string
	foundSeed, foundModel := false, false
	for _, field := range strings.Split(parameters, ",") {
		key, value, ok := strings.Cut(strings.TrimSpace(field), ":")
		if !ok {
			continue
		}
		switch strings.TrimSpace(key) {
		case "Seed":
			parsedSeed, parseErr := strconv.ParseInt(strings.TrimSpace(value), 10, 64)
			if parseErr != nil {
				return "", 0, "", imageerrors.InvalidPNGSeed(value, parseErr)
			}
			seed = parsedSeed
			foundSeed = true
		case "Model":
			model = strings.TrimSpace(value)
			foundModel = model != ""
		}
	}

	if prompt == "" {
		return "", 0, "", imageerrors.ErrPNGPromptMissing
	}
	if !foundSeed {
		return "", 0, "", imageerrors.ErrPNGSeedMissing
	}
	if !foundModel {
		return "", 0, "", imageerrors.ErrPNGModelMissing
	}
	return prompt, seed, model, nil
}
