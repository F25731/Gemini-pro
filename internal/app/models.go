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
	QualityTiers      []string  `json:"qualityTiers,omitempty"`
	AspectRatios      []string  `json:"aspectRatios,omitempty"`
	DefaultAspectRatio string    `json:"defaultAspectRatio,omitempty"`
	DefaultDuration    string    `json:"defaultDuration,omitempty"`
	DurationOptions    []string  `json:"durationOptions,omitempty"`
	Capabilities       []string  `json:"capabilities,omitempty"`
	TextEndpoint       string    `json:"textEndpoint"`
	ImageEndpoint      string    `json:"imageEndpoint,omitempty"`
	StartEndEndpoint   string    `json:"startEndEndpoint,omitempty"`
	MinImageInputs     int       `json:"minImageInputs,omitempty"`
	MaxImageInputs     int       `json:"maxImageInputs,omitempty"`
	MaxFileSizeMB      int       `json:"maxFileSizeMb,omitempty"`
	Notes             []string  `json:"notes,omitempty"`
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
			Capabilities:   []string{"文生图", "图生图"},
			TextEndpoint:   "/v1/banana_pro/text-to-image",
			ImageEndpoint:  "/v1/banana_pro/image-to-image",
			MinImageInputs: 1,
			MaxImageInputs: 10,
			MaxFileSizeMB:  10,
			Notes:          []string{"不区分文生图/图生图模型，请求带参考图时自动走图生图。"},
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
			Capabilities:   []string{"文生图", "图生图"},
			TextEndpoint:   "/v1/banana2/text-to-image",
			ImageEndpoint:  "/v1/banana2/image-to-image",
			MinImageInputs: 1,
			MaxImageInputs: 10,
			MaxFileSizeMB:  30,
			Notes:          []string{"Banana2 支持 512 档；512px 会归一化为 512。"},
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
				QualityTiers:       []string{"720p", "1080p", "4k"},
				AspectRatios:       videoRatios,
				DefaultAspectRatio:  "16:9",
				DefaultDuration:     "8",
				DurationOptions:    []string{"8"},
				Capabilities:       []string{"文生视频", "图生视频"},
				TextEndpoint:        "/v1/" + family.prefix + "/text-to-video",
				ImageEndpoint:       "/v1/" + family.prefix + "/image-to-video",
				MinImageInputs:     1,
				MaxImageInputs:      1,
				MaxFileSizeMB:      10,
				Notes:              []string{"视频模型按 720p / 1080p / 4k 分档，不区分文生视频和图生视频。"},
			}
			if family.startEnd {
				spec.StartEndEndpoint = "/v1/" + family.prefix + "/start-end-to-video"
				spec.Capabilities = append(spec.Capabilities, "首尾帧视频")
				spec.MaxImageInputs = 2
				spec.Notes = append(spec.Notes, "传 2 张参考图时自动走首尾帧接口。")
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
