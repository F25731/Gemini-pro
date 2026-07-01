package app

import (
	"fmt"
	"strings"
)

type MediaKind string

const (
	MediaImage MediaKind = "image"
	MediaVideo MediaKind = "video"
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
	DefaultDuration     string    `json:"defaultDuration,omitempty"`
	DurationOptions     []string  `json:"durationOptions,omitempty"`
	Capabilities        []string  `json:"capabilities,omitempty"`
	TextEndpoint        string    `json:"textEndpoint"`
	ImageEndpoint       string    `json:"imageEndpoint,omitempty"`
	StartEndEndpoint    string    `json:"startEndEndpoint,omitempty"`
	MinImageInputs      int       `json:"minImageInputs,omitempty"`
	MaxImageInputs      int       `json:"maxImageInputs,omitempty"`
	MaxFileSizeMB       int       `json:"maxFileSizeMb,omitempty"`
	Notes              []string  `json:"notes,omitempty"`
}

func allModelSpecs() []ModelSpec {
	specs := []ModelSpec{}
	imageRatios := []string{"auto", "1:1", "16:9", "9:16", "4:3", "3:4", "3:2", "2:3", "5:4", "4:5", "21:9"}
	banana2Ratios := append(append([]string{}, imageRatios...), "1:4", "4:1", "1:8", "8:1")
	videoRatios := []string{"16:9", "9:16"}

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

	for _, family := range []struct {
		modelPrefix    string
		endpointPrefix string
		label          string
		startEnd       bool
	}{
		{"veo31-pro", "veo3.1-pro", "Veo3.1 Pro", true},
		{"veo31-fast", "veo3.1-fast", "Veo3.1 Fast", false},
	} {
		for _, resolution := range []string{"720p", "1080p", "4k"} {
			spec := ModelSpec{
				ID:                 family.modelPrefix + "-" + resolution,
				Name:               family.label + " " + resolution,
				Family:             family.modelPrefix,
				Media:              MediaVideo,
				Resolution:         resolution,
				QualityTiers:       []string{"720p", "1080p", "4k"},
				AspectRatios:       videoRatios,
				DefaultAspectRatio:  "16:9",
				DefaultDuration:     "8",
				DurationOptions:    []string{"8"},
				Capabilities:       []string{"text-to-video", "image-to-video"},
				TextEndpoint:        "/v1/" + family.endpointPrefix + "/text-to-video",
				ImageEndpoint:       "/v1/" + family.endpointPrefix + "/image-to-video",
				MinImageInputs:     1,
				MaxImageInputs:     1,
				MaxFileSizeMB:      10,
				Notes:              []string{"Single public model per resolution; requests with reference images use image-to-video."},
			}
			if family.startEnd {
				spec.StartEndEndpoint = "/v1/" + family.endpointPrefix + "/start-end-to-video"
				spec.Capabilities = append(spec.Capabilities, "start-end-to-video")
				spec.MaxImageInputs = 2
				spec.Notes = append(spec.Notes, "Two reference images are routed to start-end-to-video.")
			}
			specs = append(specs, spec)
		}
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
