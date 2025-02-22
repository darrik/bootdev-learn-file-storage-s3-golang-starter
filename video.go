package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"os/exec"
	"log"
)

type FFProbe struct {
	Streams []struct {
		Index              int    `json:"index,omitempty"`
		CodecName          string `json:"codec_name,omitempty"`
		CodecLongName      string `json:"codec_long_name,omitempty"`
		Profile            string `json:"profile,omitempty"`
		CodecType          string `json:"codec_type,omitempty"`
		CodecTagString     string `json:"codec_tag_string,omitempty"`
		CodecTag           string `json:"codec_tag,omitempty"`
		Width              int    `json:"width,omitempty"`
		Height             int    `json:"height,omitempty"`
		CodedWidth         int    `json:"coded_width,omitempty"`
		CodedHeight        int    `json:"coded_height,omitempty"`
		ClosedCaptions     int    `json:"closed_captions,omitempty"`
		FilmGrain          int    `json:"film_grain,omitempty"`
		HasBFrames         int    `json:"has_b_frames,omitempty"`
		SampleAspectRatio  string `json:"sample_aspect_ratio,omitempty"`
		DisplayAspectRatio string `json:"display_aspect_ratio,omitempty"`
		PixFmt             string `json:"pix_fmt,omitempty"`
		Level              int    `json:"level,omitempty"`
		ColorRange         string `json:"color_range,omitempty"`
		ColorSpace         string `json:"color_space,omitempty"`
		ColorTransfer      string `json:"color_transfer,omitempty"`
		ColorPrimaries     string `json:"color_primaries,omitempty"`
		ChromaLocation     string `json:"chroma_location,omitempty"`
		FieldOrder         string `json:"field_order,omitempty"`
		Refs               int    `json:"refs,omitempty"`
		IsAvc              string `json:"is_avc,omitempty"`
		NalLengthSize      string `json:"nal_length_size,omitempty"`
		ID                 string `json:"id,omitempty"`
		RFrameRate         string `json:"r_frame_rate,omitempty"`
		AvgFrameRate       string `json:"avg_frame_rate,omitempty"`
		TimeBase           string `json:"time_base,omitempty"`
		StartPts           int    `json:"start_pts,omitempty"`
		StartTime          string `json:"start_time,omitempty"`
		DurationTs         int    `json:"duration_ts,omitempty"`
		Duration           string `json:"duration,omitempty"`
		BitRate            string `json:"bit_rate,omitempty"`
		BitsPerRawSample   string `json:"bits_per_raw_sample,omitempty"`
		NbFrames           string `json:"nb_frames,omitempty"`
		ExtradataSize      int    `json:"extradata_size,omitempty"`
		Disposition        struct {
			Default         int `json:"default,omitempty"`
			Dub             int `json:"dub,omitempty"`
			Original        int `json:"original,omitempty"`
			Comment         int `json:"comment,omitempty"`
			Lyrics          int `json:"lyrics,omitempty"`
			Karaoke         int `json:"karaoke,omitempty"`
			Forced          int `json:"forced,omitempty"`
			HearingImpaired int `json:"hearing_impaired,omitempty"`
			VisualImpaired  int `json:"visual_impaired,omitempty"`
			CleanEffects    int `json:"clean_effects,omitempty"`
			AttachedPic     int `json:"attached_pic,omitempty"`
			TimedThumbnails int `json:"timed_thumbnails,omitempty"`
			NonDiegetic     int `json:"non_diegetic,omitempty"`
			Captions        int `json:"captions,omitempty"`
			Descriptions    int `json:"descriptions,omitempty"`
			Metadata        int `json:"metadata,omitempty"`
			Dependent       int `json:"dependent,omitempty"`
			StillImage      int `json:"still_image,omitempty"`
			Multilayer      int `json:"multilayer,omitempty"`
		} `json:"disposition,omitempty"`
		Tags struct {
			Language    string `json:"language,omitempty"`
			HandlerName string `json:"handler_name,omitempty"`
			VendorID    string `json:"vendor_id,omitempty"`
			Encoder     string `json:"encoder,omitempty"`
			Timecode    string `json:"timecode,omitempty"`
		} `json:"tags,omitempty"`
		SampleFmt      string `json:"sample_fmt,omitempty"`
		SampleRate     string `json:"sample_rate,omitempty"`
		Channels       int    `json:"channels,omitempty"`
		ChannelLayout  string `json:"channel_layout,omitempty"`
		BitsPerSample  int    `json:"bits_per_sample,omitempty"`
		InitialPadding int    `json:"initial_padding,omitempty"`
		Tags0          struct {
			Language    string `json:"language,omitempty"`
			HandlerName string `json:"handler_name,omitempty"`
			VendorID    string `json:"vendor_id,omitempty"`
		} `json:"tags,omitempty"`
		Tags1 struct {
			Language    string `json:"language,omitempty"`
			HandlerName string `json:"handler_name,omitempty"`
			Timecode    string `json:"timecode,omitempty"`
		} `json:"tags,omitempty"`
	} `json:"streams,omitempty"`
}

func getVideoAspectRatio(filePath string) (string, error) {
	var b bytes.Buffer

	cmd := exec.Command("ffprobe", "-v", "error", "-print_format", "json", "-show_streams", filePath)
	cmd.Stdout = &b
	err := cmd.Run()
	// log.Printf("kör %s\n", cmd.String())
	if err != nil {
		log.Fatalf("ffprobe failed: %s\n", err)
	}

	stream := FFProbe{}
	err = json.Unmarshal(b.Bytes(), &stream)
	if err != nil {
		log.Fatalf("%s\n",err)
	}

	as := stream.Streams[0].DisplayAspectRatio
	if as == "" {
		return as, errors.New("så jävla trött på golangs error hantering")
	}

	if as != "16:9" && as != "9:16" {
		as = "other"
	}

	return as, nil
}
