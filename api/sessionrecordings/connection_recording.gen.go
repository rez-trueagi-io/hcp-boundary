// Code generated by "make api"; DO NOT EDIT.
// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package sessionrecordings

import (
	"time"

	"github.com/hashicorp/boundary/api"
)

type ConnectionRecording struct {
	Id                string              `json:"id,omitempty"`
	ConnectionId      string              `json:"connection_id,omitempty"`
	BytesUp           uint64              `json:"bytes_up,string,omitempty"`
	BytesDown         uint64              `json:"bytes_down,string,omitempty"`
	StartTime         time.Time           `json:"start_time,omitempty"`
	EndTime           time.Time           `json:"end_time,omitempty"`
	Duration          api.Duration        `json:"duration,omitempty"`
	MimeTypes         []string            `json:"mime_types,omitempty"`
	ChannelRecordings []*ChannelRecording `json:"channel_recordings,omitempty"`
}