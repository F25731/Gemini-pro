package app

import (
	"fmt"
	"strings"
)

type MediaKind string

const (
	MediaImage MediaKind = "image"
)

type ModelSpec struct {
	ID                 string    `json:"id"`
	Name               string    `json:"name"`
	Family             string    `json:"family"`
	Media              MediaKind `json:"media"`
	Resolution         string    `json:"resolution"`
	QualityTiers       []string  `json:"qualityTiers,omitempty"`
	AspectRatios       []string  `json:"aspectRatios,omitempty"`
	DefaultAspectRatio  string    `json:"defaultAspectRatio,omitempty"`
	Capabilities        []string  `json:"capabilities,omitempty"`
	TextEndpoint        string    `json:"textEndpoint"`
	ImageEndpoint       string    `json:"imageEndpoint,omitempty"`
	MinImageInputs      int       `json:"minImageInputs,omitempty"`
	MaxImageInputs      int       `json:"maxImageInputs,omitempty"`
	MaxFileSizeMB       int       `json:"maxFileSizeMb,omitempty"`
	Notes              []string  `json:"notes,omitempty"`
}

func allModelSpecs() []ModelSpec {
	specs := []ModelSpec{}
	imageRatios := []string{"auto", "1:1", "16:9", "9:16", "4:3", "3:4", "3:2", "2:3", "5:4", "4:5", "21:9"}
	banana2Ratios := append(append([]string{}, imageRatios...), "1:4", "4:1", "1:8", "8:1")

	for _, resolution := range []string{"1k", "2k", "4k"} {
		specs = append(specs, ModelSpec{
			ID:             "banana-pro-" + resolution,
			Name:           "Banana Pro " + resolution,
			Family:         "banana-pro",
			Media:          MediaImage,
			Resolution:     resolution,
			QualityTiers:   []string{"1k", "2k", "4k"},
			AspectRatios:   imageRatios,
			Capabilities:   []string{"text-to-image", "image-to-image"},
			TextEndpoint:   "/v1/banana_pro/text-to-image",
			ImageEndpoint:  "/v1/banana_pro/image-to-image",
			MinImageInputs: 1,
			MaxImageInputs: 10,
			MaxFileSizeMB:  10,
			Notes:          []string{"Single public model per resolution; requests with reference images use image-to-image."},
		})
	}

	for _, resolution := range []string{"512", "1k", "2k", "4k"} {
		specs = append(specs, ModelSpec{
			ID:             "banana2-" + resolution,
			Name:           "Banana2 " + resolution,
			Family:         "banana2",
			Media:          MediaImage,
			Resolution:     resolution,
			QualityTiers:   []string{"512", "1k", "2k", "4k"},
			AspectRatios:   banana2Ratios,
			Capabilities:   []string{"text-to-image", "image-to-image"},
			TextEndpoint:   "/v1/banana2/text-to-image",
			ImageEndpoint:  "/v1/banana2/image-to-image",
			MinImageInputs: 1,
			MaxImageInputs: 10,
			MaxFileSizeMB:  30,
			Notes:          []string{"Banana2 supports 512; 512px is normalized to 512 upstream."},
		})
	}

	return specs
}

func modelIDs() []string {
	specs := allModelSpecs()
	ids := make([]string, 0, len(specs))
	for _, spec := range specs {
		ids = append(ids, spec.ID)
	}
	return ids
}

func modelSpecByID(model string) (ModelSpec, error) {
	value := strings.ToLower(strings.TrimSpace(model))
	for _, spec := range allModelSpecs() {
		if value == strings.ToLower(spec.ID) {
			return spec, nil
		}
	}
	return ModelSpec{}, fmt.Errorf("unsupported model %q", model)
}

func publicModelSpecs() []ModelSpec {
	return allModelSpecs()
}
