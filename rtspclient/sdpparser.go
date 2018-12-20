package rtspclient

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/Nodeboy2/rtspclient/sdp"
)

type MediaSubsession struct {
	MediumName            string // video/audio
	PayloadFormat         int
	CodecName             string
	RtpTimestampFrequency int
	Channels              int
	TrackURL              string
	VideoFramerate        int // "a=framerate: <fps>" or "a=x-framerate: <fps>"
	VideoWidth            int // "a=x-dimensions:<width>,<height>"
	VideoHeight           int
	Fmtp                  map[string]string
}

type SDPInfo struct {
	Medias []MediaSubsession
}

func parsingSDPLine(strSdp string) (*sdp.Message, error) {
	var sdpSession sdp.Session
	var err error
	sdpSession, err = sdp.DecodeSession([]byte(strSdp), sdpSession)
	if nil != err {
		log.Fatal("sdp decode session error: ", err)
		return nil, err
	}
	sdpDecoder := sdp.NewDecoder(sdpSession)
	sdpMessage := &sdp.Message{}
	err = sdpDecoder.Decode(sdpMessage)
	if nil != err {
		log.Fatal("sdp decode error: ", err)
		return nil, err
	}
	return sdpMessage, nil
}

func parsingSDP(sdp string) *SDPInfo {

	sdpMessage, errorInfo := parsingSDPLine(sdp)
	if nil != errorInfo {
		return nil
	}

	if 0 >= len(sdpMessage.Medias) {
		return nil
	}

	sdpInfo := &SDPInfo{Medias: make([]MediaSubsession, len(sdpMessage.Medias))}

	for index, meidaInfo := range sdpMessage.Medias {
		sdpInfo.Medias[index].MediumName = meidaInfo.Description.Type
		sdpInfo.Medias[index].PayloadFormat, _ = strconv.Atoi(meidaInfo.Description.Format)
		codecName, _, timestampFrequency, channels := getPayloadInfo(sdpInfo.Medias[index].PayloadFormat)
		if "" == codecName {
			sdpRtpmap := meidaInfo.Attributes.Value("rtpmap")
			if "" != sdpRtpmap {
				codecName, timestampFrequency, channels = getPayloadInfoForRtpmap(sdpRtpmap)
			}
		}

		sdpInfo.Medias[index].CodecName = codecName
		sdpInfo.Medias[index].RtpTimestampFrequency = timestampFrequency
		sdpInfo.Medias[index].Channels = channels

		sdpInfo.Medias[index].TrackURL = meidaInfo.Attributes.Value("control")

		sdpXDimensions := meidaInfo.Attributes.Value("x-dimensions")
		if "" != sdpXDimensions {
			sdpInfo.Medias[index].VideoWidth, sdpInfo.Medias[index].VideoHeight = getSizeForXDimensions(sdpXDimensions)
		}

		var sdpFramerate string
		sdpFramerate = meidaInfo.Attributes.Value("framerate")
		if "" == sdpFramerate {
			sdpFramerate = meidaInfo.Attributes.Value("x-framerate")
		}
		if "" != sdpFramerate {
			sdpInfo.Medias[index].VideoFramerate = getFramerateForFramerate(sdpFramerate)
		}

		sdpFmtp := meidaInfo.Attributes.Value("fmtp")
		if "" != sdpFmtp {
			sdpInfo.Medias[index].Fmtp = getFmtParame(sdpFmtp)
			log.Print(sdpFmtp)
		}

	}

	return sdpInfo
}

func getFramerateForFramerate(sdpFramerate string) int {
	// Check for a "a=framerate: <fps>" or "a=x-framerate: <fps>" line
	framerate, _ := strconv.Atoi(sdpFramerate)
	return framerate
}

func getFmtParame(sdpFmtp string) map[string]string {
	fmtParame := make(map[string]string)

	fieldsFmtp := strings.Split(sdpFmtp, " ")
	for _, field := range fieldsFmtp {
		fields := strings.Split(field, " ")
		for _, fieldTemp := range fields {
			fieldTemp = strings.Replace(fieldTemp, ";", "", -1)
			index := strings.Index(fieldTemp, "=")
			if index == -1 {
				fmtParame[fieldTemp] = ""
			} else {
				fmtParame[fieldTemp[:index]] = field[index+1:]
			}
		}
	}
	return fmtParame
}

func getSizeForXDimensions(sdpXDimensions string) (int, int) {
	// Check for a "a=x-dimensions:<width>,<height>" line:
	sdpXDimensions = strings.Replace(sdpXDimensions, ",", " ", -1)

	width := 0
	height := 0
	fmt.Sscanf(sdpXDimensions, "%d %d", &width, &height)
	return width, height
}

func getPayloadInfoForRtpmap(sdpRtpmap string) (string, int, int) {
	// Check for a "<fmt> <codec>/<freq>" line:
	// (Also check without the "/<freq>"; RealNetworks omits this)
	// Also check for a trailing "/<numChannels>".
	sdpRtpmap = strings.Replace(sdpRtpmap, "/", " ", -1)
	var payloadFormat int
	var codecName string
	timestampFrequency := -1
	channels := -1
	fmt.Sscanf(sdpRtpmap, "%d %s %d %d", &payloadFormat, &codecName, &timestampFrequency, &channels)
	return codecName, timestampFrequency, channels
}

// return encodingName, audio/video/data clock rate channels
func getPayloadInfo(payloadFormat int) (string, string, int, int) {
	switch payloadFormat {
	case 0:
		return "PCMU", "audio", 8000, 1
	case 3:
		return "GSM", "audio", 8000, 1
	case 4:
		return "G723", "audio", 8000, 1
	case 5:
		return "DVI4", "audio", 8000, 1
	case 6:
		return "DVI4", "audio", 16000, 1
	case 7:
		return "LPC", "audio", 8000, 1
	case 8:
		return "PCMA", "audio", 8000, 1
	case 9:
		return "G722", "audio", 8000, 1
	case 10:
		return "L16", "audio", 44100, 2
	case 11:
		return "L16", "audio", 44100, 1
	case 12:
		return "QCELP", "audio", 8000, 1
	case 13:
		return "CN", "audio", 8000, 1
	case 14:
		return "MPA", "audio", 90000, -1
	case 15:
		return "G728", "audio", 8000, 1
	case 16:
		return "DVI4", "audio", 11025, 1
	case 17:
		return "DVI4", "audio", 22050, 1
	case 18:
		return "G729", "audio", 8000, 1
	case 25:
		return "CelB", "video", 90000, -1
	case 26:
		return "JPEG", "video", 90000, -1
	case 28:
		return "nv", "video", 90000, -1
	case 31:
		return "H261", "video", 90000, -1
	case 32:
		return "MPV", "video", 90000, -1
	case 33:
		return "MP2T", "data", 90000, -1
	case 34:
		return "H263", "video", 90000, -1
	}
	return "", "", 0, 0
}
