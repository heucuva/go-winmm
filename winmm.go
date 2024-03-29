//go:build windows
// +build windows

package winmm

import (
	"errors"
	"fmt"
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	// ErrWinMM is for tagging Windows Multimedia errors appropriately
	ErrWinMM = errors.New("winmm error")
)

var (
	winmmDll = windows.NewLazySystemDLL("winmm.dll")

	waveOutOpen            = winmmDll.NewProc("waveOutOpen")
	waveOutPrepareHeader   = winmmDll.NewProc("waveOutPrepareHeader")
	waveOutWrite           = winmmDll.NewProc("waveOutWrite")
	waveOutUnprepareHeader = winmmDll.NewProc("waveOutUnprepareHeader")
	waveOutClose           = winmmDll.NewProc("waveOutClose")
)

// WaveOutData is a structure holding the header and the go version of the data
// sent out to the sound device (for garbage collection reasons)
type WaveOutData struct {
	hdr  WAVEHDR
	data []uint8
}

// WaveOut is a sound device for the windows multimedia system
type WaveOut struct {
	handle    HWAVEOUT
	buffers   [3]WaveOutData
	available chan *WaveOutData
}

// New creates a new WaveOut device based on the parameters provided
func New(channels int, samplesPerSec int, bitsPerSample int) (*WaveOut, error) {
	w := WaveOut{}
	w.available = make(chan *WaveOutData, len(w.buffers))
	// make a circular buffer out of the headers
	for i := 0; i < len(w.buffers); i++ {
		var next *WaveOutData
		if i < len(w.buffers)-1 {
			next = &w.buffers[i+1]
		} else {
			next = &w.buffers[0]
		}
		w.buffers[i].hdr.LpNext = uintptr(unsafe.Pointer(&next.hdr))
		w.available <- &w.buffers[i]
	}

	wfx := WAVEFORMATEX{
		WFormatTag:     WAVE_FORMAT_PCM,
		NChannels:      uint16(channels),
		NSamplesPerSec: uint32(samplesPerSec),
		WBitsPerSample: uint16(bitsPerSample),
	}
	wfx.CbSize = uint16(unsafe.Sizeof(wfx))
	wfx.NBlockAlign = uint16(channels * bitsPerSample / 8)
	wfx.NAvgBytesPerSec = wfx.NSamplesPerSec * uint32(wfx.NBlockAlign)

	result, _, _ := waveOutOpen.Call(
		uintptr(unsafe.Pointer(&w.handle)), // phwo
		uintptr(WAVE_MAPPER),               // uDeviceID = WAVE_MAPPER
		uintptr(unsafe.Pointer(&wfx)),      // pwfx
		uintptr(0),                         // dwCallback
		uintptr(0),                         // dwInstance
		uintptr(0))                         // fdwOpen
	if result != 0 { // MMSYSERR_NOERROR
		return nil, fmt.Errorf("%w: waveOutOpen returned %d", ErrWinMM, result)
	}

	return &w, nil
}

// Write prepares a byte array for output to the WaveOut device
func (w *WaveOut) Write(data []byte) *WaveOutData {
	// pull a buffer
	wave := <-w.available

	wave.data = data
	wave.hdr.LpData = uintptr(unsafe.Pointer(&wave.data[0]))
	wave.hdr.DwBufferLength = uint32(len(wave.data))

	_, _, _ = waveOutPrepareHeader.Call(
		uintptr(w.handle),                  // hwo
		uintptr(unsafe.Pointer(&wave.hdr)), // pwh
		uintptr(unsafe.Sizeof(wave.hdr)))   // cbwh

	_, _, _ = waveOutWrite.Call(
		uintptr(w.handle),                  // hwo
		uintptr(unsafe.Pointer(&wave.hdr)), // pwh
		uintptr(unsafe.Sizeof(wave.hdr)))   // cbwh

	return wave
}

// IsHeaderFinished determines if a wave output buffer has finished playing
// and will readd it to the available buffer queue when it is
func (w *WaveOut) IsHeaderFinished(hdr *WaveOutData) bool {
	result, _, _ := waveOutUnprepareHeader.Call(
		uintptr(w.handle),                 // hwo
		uintptr(unsafe.Pointer(&hdr.hdr)), // pwh
		uintptr(unsafe.Sizeof(hdr.hdr)))   // cbwh
	if result == WAVERR_STILLPLAYING {
		return false
	}

	// put it back!
	w.available <- hdr
	return true
}

// Close terminates a WaveOut device
func (w *WaveOut) Close() {
	if w.handle != 0 {
		var h uintptr
		h, w.handle = uintptr(w.handle), 0
		_, _, _ = waveOutClose.Call(h)
	}
	close(w.available)
}
