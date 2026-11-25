package sda1111

import (
	"bytes"
	"encoding/binary"
	stderrors "errors"
	"hash/crc32"
	"image"
	"image/png"
	"testing"

	imageerrors "github.com/shiyou0130011/sda1111/errors"
)

func TestGetImageMetadata(t *testing.T) {
	const prompt = "1 girl, black hair, long hair, red eyes, bangs, blunt_bangs, low twintails, big breasts, smile, kimono,fox mask,temple festival,"
	const parameters = prompt + "\nNegative prompt: low quality\nSteps: 30, Sampler: Euler a, Seed: 1085462827, Size: 1536x1024, Model: HALO"

	reader := bytes.NewReader(pngWithParameters(t, parameters))
	gotPrompt, seed, model, err := GetImageMetadata(reader)
	if err != nil {
		t.Fatalf("GetImageMetadata() error = %v", err)
	}
	if gotPrompt != prompt {
		t.Errorf("prompt = %q, want %q", gotPrompt, prompt)
	}
	if seed != 1085462827 {
		t.Errorf("seed = %d, want 1085462827", seed)
	}
	if model != "HALO" {
		t.Errorf("model = %q, want HALO", model)
	}
}

func TestGetImageMetadataRejectsMissingFields(t *testing.T) {
	reader := bytes.NewReader(pngWithParameters(t, "a prompt\nSteps: 30, Sampler: Euler a"))
	_, _, _, err := GetImageMetadata(reader)
	if err == nil {
		t.Fatal("GetImageMetadata() error = nil, want missing metadata error")
	}
}

func TestGetImageMetadataRejectsNilReader(t *testing.T) {
	_, _, _, err := GetImageMetadata(nil)
	if !stderrors.Is(err, imageerrors.ErrNilImageReader) {
		t.Fatalf("GetImageMetadata() error = %v, want %v", err, imageerrors.ErrNilImageReader)
	}
}

func pngWithParameters(t *testing.T, parameters string) []byte {
	t.Helper()

	var encoded bytes.Buffer
	if err := png.Encode(&encoded, image.NewRGBA(image.Rect(0, 0, 1, 1))); err != nil {
		t.Fatalf("encode PNG: %v", err)
	}
	data := encoded.Bytes()
	textChunk := makePNGChunk("tEXt", append([]byte("parameters\x00"), []byte(parameters)...))
	result := make([]byte, 0, len(data)+len(textChunk))
	result = append(result, data[:33]...)
	result = append(result, textChunk...)
	result = append(result, data[33:]...)
	return result
}

func makePNGChunk(kind string, data []byte) []byte {
	chunk := make([]byte, 12+len(data))
	binary.BigEndian.PutUint32(chunk[:4], uint32(len(data)))
	copy(chunk[4:8], kind)
	copy(chunk[8:], data)
	checksum := crc32.ChecksumIEEE(chunk[4 : 8+len(data)])
	binary.BigEndian.PutUint32(chunk[8+len(data):], checksum)
	return chunk
}
