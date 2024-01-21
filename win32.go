//go:build windows
// +build windows

package winmm

import (
	"golang.org/x/sys/windows"
)

const (
	// WAVERR_STILLPLAYING defines a value equal to the WAVERR_STILLPLAYING error code from winmm api
	WAVERR_STILLPLAYING = uintptr(33)

	// WAVE_FORMAT_PCM specifies PCM wave format
	WAVE_FORMAT_PCM = uint16(0x0001)

	// WAVE_MAPPER specifies the system default configured wave device
	WAVE_MAPPER = uint32(0xFFFFFFFF)
)

// HWAVEOUT is a handle for a WAVEOUT device
type HWAVEOUT windows.Handle

// WAVEHDR is a structure containing the details about a circular buffer of wave data
type WAVEHDR struct {
	LpData          uintptr
	DwBufferLength  uint32
	DwBytesRecorded uint32
	DwUser          uintptr
	DwFlags         uint32
	DwLoops         uint32
	LpNext          uintptr
	Reserved        uintptr
}

// WAVEFORMATEX is a structure containing data about a wave format
type WAVEFORMATEX struct {
	WFormatTag      uint16
	NChannels       uint16
	NSamplesPerSec  uint32
	NAvgBytesPerSec uint32
	NBlockAlign     uint16
	WBitsPerSample  uint16
	CbSize          uint16
}
