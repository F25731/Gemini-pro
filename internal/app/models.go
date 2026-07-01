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
	ID                string    `json:"id"`
	Name              string    `json:"name"`
	Family            string    `json:"family"`
	Media             MediaKind `json:"media"`
	Resolution        string    `json:"resolution"`
	DefaultAspectRatio string    `json:"defaultAspectRatio,omitempty"`
	DefaultDuration    string    `json:"defaultDuration,omitempty"`
	TextEndpoint       string    `json:"textEndpoint"`
	ImageEndpoint      string    `json:"imageEndpoint,omitempty"`
	StartEndEndpoint   string    `json:"startEndEndpoint,omitempty"`
	MaxImageInputs     int       `json:"maxImageInputs,omitempty"`
}

func allModelSpecs() []ModelSpec {
	specs := []ModelSpec{}
	for _, resolution := range []string{"1k", "2k", "4k"} {
		specs = append(specs, ModelSpec{
			ID:             "banana-pro-" + resolution,
			Name:           "Banana Pro " + resolution,
			Family:         "banana-pro",
			Media:          MediaImage,
			Resolution:     resolution,
			TextEndpoint:   "/v1/banana_pro/text-to-image",
			ImageEndpoint:  "/v1/banana_pro/image-to-image",
			MaxImageInputs: 10,
		})
	}
	for _, resolution := range []string{"512", "1k", "2k", "4k"} {
		specs = append(specs, ModelSpec{
			ID:             "banana2-" + resolution,
			Name:           "Banana2 " + resolution,
			Family:         "banana2",
			Media:          MediaImage,
			Resolution:     resolution,
			TextEndpoint:   "/v1/banana2/text-to-image",
			ImageEndpoint:  "/v1/banana2/image-to-image",
			MaxImageInputs: 10,
		})
	}
	for _, family := range []struct {
		prefix  string
		label   string
		startEnd bool
	}{
		{"veo31-pro", "Veo3.1 Pro", true},
		{"veo31-fast", "Veo3.1 Fast", false},
	} {
		for _, resolution := range []string{"720p", "1080p", "4k"} {
			spec := ModelSpec{
				ID:                 family.prefix + "-" + resolution,
				Name:               family.label + " " + resolution,
				Family:             family.prefix,
				Media:              MediaVideo,
				Resolution:         resolution,
				DefaultAspectRatio:  "16:9",
				DefaultDuration:     "8",
				TextEndpoint:        "/v1/" + family.prefix + "/text-to-video",
				ImageEndpoint:       "/v1/" + family.prefix + "/image-to-video",
				MaxImageInputs:      1,
			}
			if family.startEnd {
				spec.StartEndEndpoint = "/v1/" + family.prefix + "/start-end-to-video"
				spec.MaxImageInputs = 2
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
