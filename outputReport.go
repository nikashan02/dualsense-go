// References C++ structures defined at https://controllers.fandom.com/wiki/Sony_DualSense#Output_Reports

package dualsense

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

type MuteLightMode uint8

const (
	MuteLightModeOff MuteLightMode = iota
	MuteLightModeOn
	MuteLightModeBreathing
	MuteLightModeDoNothing
	MuteLightModeNoAction4
	MuteLightModeNoAction5
	MuteLightModeNoAction6
	MuteLightModeNoAction7
)

type LightFadeAnimation uint8

const (
	LightFadeAnimationNothing LightFadeAnimation = iota
	LightFadeAnimationFadeIn
	LightFadeAnimationFadeOut
)

type LightBrightness uint8

const (
	LightBrightnessBright LightBrightness = iota
	LightBrightnessMid
	LightBrightnessDim
	LightBrightnessNoAction3
	LightBrightnessNoAction4
	LightBrightnessNoAction5
	LightBrightnessNoAction6
	LightBrightnessNoAction7
)

type packedSetStateData struct {
	SetFlags0            uint8 // Contains EnableRumbleEmulation, UseRumbleNotHaptics, AllowRightTriggerFFB, AllowLeftTriggerFFB, AllowHeadphoneVolume, AllowSpeakerVolume, AllowMicVolume, AllowAudioControl
	SetFlags1            uint8 // Contains AllowMuteLight, AllowAudioMute, AllowLedColor, ResetLights, AllowPlayerIndicators, AllowHapticLowPassFilter, AllowMotorPowerLevel, AllowAudioControl2
	RumbleEmulationRight uint8
	RumbleEmulationLeft  uint8
	VolumeHeadphones     uint8
	VolumeSpeaker        uint8
	VolumeMic            uint8
	AudioControl         uint8 // Contains MicSelect, EchoCancelEnable, NoiseCancelEnable, OutputPathSelect, InputPathSelect
	MuteLight            MuteLightMode
	MuteControl          uint8 // Contains TouchPowerSave, MotionPowerSave, HapticPowerSave, AudioPowerSave, MicMute, SpeakerMute, HeadphoneMute, HapticMute
	RightTriggerFFB      [11]uint8
	LeftTriggerFFB       [11]uint8
	HostTimestamp        uint32
	MotorPowerLevel      uint8 // Contains TriggerMotorPowerReduction, RumbleMotorPowerReduction
	AudioControl2        uint8 // Contains SpeakerCompPreGain, BeamformingEnable, UnkAudioControl2
	SetFlags38           uint8 // Contains AllowLightBrightnessChange, AllowColorLightFadeAnimation, EnableImprovedRumbleEmulation, UNKBITC
	SetFlags39           uint8 // Contains HapticLowPassFilter, UNKBIT
	UNKBYTE              uint8
	LightFadeAnimation   LightFadeAnimation
	LightBrightness      LightBrightness
	PlayerIndicators     uint8 // Contains PlayerLight1, PlayerLight2, PlayerLight3, PlayerLight4, PlayerLight5, PlayerLightFade, PlayerLightUNK
	LedRed               uint8
	LedGreen             uint8
	LedBlue              uint8
}

type packedUSBReportOut struct {
	ReportID        uint8
	USBSetStateDate packedSetStateData
}

type MicSelectType uint8

const (
	MicSelectAuto MicSelectType = iota
	MicSelectInternalOnly
	MicSelectExternalOnly
	MicSelectUnknown
)

type SetStateData struct {
	EnableRumbleEmulation         bool
	UseRumbleNotHaptics           bool
	AllowRightTriggerFFB          bool // Enable setting RightTriggerFFB
	AllowLeftTriggerFFB           bool // Enable setting LeftTriggerFFB
	AllowHeadphoneVolume          bool // Enable setting VolumeHeadphones
	AllowSpeakerVolume            bool // Enable setting VolumeSpeaker
	AllowMicVolume                bool // Enable setting VolumeMic
	AllowAudioControl             bool // Enable setting the "Audio Control" fields
	AllowMuteLight                bool // Enable setting MuteLight
	AllowAudioMute                bool // Enable setting the "Mute Control" fields
	AllowLedColor                 bool // Enable setting the "RGB LED" fields
	ResetLights                   bool
	AllowPlayerIndicators         bool // Enable setting the "Player Indicators" fields
	AllowHapticLowPassFilter      bool // Enable setting HapticLowPassFilter
	AllowMotorPowerLevel          bool // Enable setting the "Motor Power Level" fields
	AllowAudioControl2            bool // Enable setting the "Audio Control 2" fields
	RumbleEmulationRight          uint8
	RumbleEmulationLeft           uint8
	VolumeHeadphones              uint8
	VolumeSpeaker                 uint8
	VolumeMic                     uint8
	MicSelect                     MicSelectType // Audio Control
	EchoCancelEnable              bool          // Audio Control
	NoiseCancelEnable             bool          // Audio Control
	OutputPathSelect              uint8         // Audio Control: 0 L_R_X, 1 L_L_X, 2 L_L_R, 3 X_X_R
	InputPathSelect               uint8         // Audio Control: 0 CHAT_ASR, 1 CHAT_CHAT, 2 ASR_ASR, 3 invalid
	MuteLight                     MuteLightMode
	TouchPowerSave                bool      // Mute Control
	MotionPowerSave               bool      // Mute Control
	HapticPowerSave               bool      // Mute Control
	AudioPowerSave                bool      // Mute Control
	MicMute                       bool      // Mute Control
	SpeakerMute                   bool      // Mute Control
	HeadphoneMute                 bool      // Mute Control
	HapticMute                    bool      // Mute Control
	RightTriggerFFB               [11]uint8 // Use GenerateTriggerFFBParams
	LeftTriggerFFB                [11]uint8 // Use GenerateTriggerFFBParams
	HostTimestamp                 uint32
	TriggerMotorPowerReduction    uint8 // Motor Power Level
	RumbleMotorPowerReduction     uint8 // Motor Power Level
	SpeakerCompPreGain            uint8 // Audio Control 2
	BeamformingEnable             bool  // Audio Control 2
	AllowLightBrightnessChange    bool  // Allow setting LightBrightness
	AllowColorLightFadeAnimation  bool  // Allow setting LightFadeAnimation
	EnableImprovedRumbleEmulation bool  // Use instead of EnableRumbleEmulation
	HapticLowPassFilter           bool
	LightFadeAnimation            LightFadeAnimation
	LightBrightness               LightBrightness
	PlayerLight1                  bool  // Player Indicators
	PlayerLight2                  bool  // Player Indicators
	PlayerLight3                  bool  // Player Indicators
	PlayerLight4                  bool  // Player Indicators
	PlayerLight5                  bool  // Player Indicators
	PlayerLightFade               bool  // Player Indicators
	LedRed                        uint8 // RGB LED
	LedGreen                      uint8 // RGB LED
	LedBlue                       uint8 // RGB LED
}

type USBReportOut struct {
	ReportID        uint8
	USBSetStateDate SetStateData
}

func packBoolsToLittleEndianUint8(b [8]bool) uint8 {
	var packed uint8
	for _, v := range b {
		packed >>= 1
		if v {
			packed |= 0b10000000
		}
	}
	return packed
}

func packUSBReportOut(setStateData SetStateData) ([]byte, error) {
	setFlags0 := packBoolsToLittleEndianUint8([8]bool{
		setStateData.EnableRumbleEmulation,
		setStateData.UseRumbleNotHaptics,
		setStateData.AllowRightTriggerFFB,
		setStateData.AllowLeftTriggerFFB,
		setStateData.AllowHeadphoneVolume,
		setStateData.AllowSpeakerVolume,
		setStateData.AllowMicVolume,
		setStateData.AllowAudioControl,
	})

	setFlags1 := packBoolsToLittleEndianUint8([8]bool{
		setStateData.AllowMuteLight,
		setStateData.AllowAudioMute,
		setStateData.AllowLedColor,
		setStateData.ResetLights,
		setStateData.AllowPlayerIndicators,
		setStateData.AllowHapticLowPassFilter,
		setStateData.AllowMotorPowerLevel,
		setStateData.AllowAudioControl2,
	})

	audioControl := uint8(setStateData.MicSelect) << 6
	audioControl >>= 1
	if setStateData.EchoCancelEnable {
		audioControl |= 0b10000000
	}
	audioControl >>= 1
	if setStateData.NoiseCancelEnable {
		audioControl |= 0b10000000
	}
	audioControl >>= 2
	audioControl |= setStateData.OutputPathSelect << 6
	audioControl >>= 2
	audioControl |= setStateData.InputPathSelect << 6

	muteControl := packBoolsToLittleEndianUint8([8]bool{
		setStateData.TouchPowerSave,
		setStateData.MotionPowerSave,
		setStateData.HapticPowerSave,
		setStateData.AudioPowerSave,
		setStateData.MicMute,
		setStateData.SpeakerMute,
		setStateData.HeadphoneMute,
		setStateData.HapticMute,
	})

	motorPowerLevel := setStateData.TriggerMotorPowerReduction | (setStateData.RumbleMotorPowerReduction << 4)

	audioControl2 := setStateData.SpeakerCompPreGain << 5
	audioControl2 >>= 1
	if setStateData.BeamformingEnable {
		audioControl2 |= 0b10000000
	}
	audioControl2 >>= 4

	var setFlags38 uint8
	if setStateData.AllowLightBrightnessChange {
		setFlags38 |= 0b10000000
	}
	setFlags38 >>= 1
	if setStateData.AllowColorLightFadeAnimation {
		setFlags38 |= 0b10000000
	}
	setFlags38 >>= 1
	if setStateData.EnableImprovedRumbleEmulation {
		setFlags38 |= 0b10000000
	}
	setFlags38 >>= 5

	var setFlags39 uint8
	if setStateData.HapticLowPassFilter {
		setFlags39 |= 0b10000000
	}
	setFlags39 >>= 7

	playerIndicators := packBoolsToLittleEndianUint8([8]bool{
		setStateData.PlayerLight1,
		setStateData.PlayerLight2,
		setStateData.PlayerLight3,
		setStateData.PlayerLight4,
		setStateData.PlayerLight5,
		setStateData.PlayerLightFade,
		false,
		false,
	})

	var packedUSBReportOut = packedUSBReportOut{
		ReportID: 0x02,
		USBSetStateDate: packedSetStateData{
			SetFlags0:            setFlags0,
			SetFlags1:            setFlags1,
			RumbleEmulationRight: setStateData.RumbleEmulationRight,
			RumbleEmulationLeft:  setStateData.RumbleEmulationLeft,
			VolumeHeadphones:     setStateData.VolumeHeadphones,
			VolumeSpeaker:        setStateData.VolumeSpeaker,
			VolumeMic:            setStateData.VolumeMic,
			AudioControl:         audioControl,
			MuteLight:            setStateData.MuteLight,
			MuteControl:          muteControl,
			RightTriggerFFB:      setStateData.RightTriggerFFB,
			LeftTriggerFFB:       setStateData.LeftTriggerFFB,
			HostTimestamp:        setStateData.HostTimestamp,
			MotorPowerLevel:      motorPowerLevel,
			AudioControl2:        audioControl2,
			SetFlags38:           setFlags38,
			SetFlags39:           setFlags39,
			UNKBYTE:              0x00,
			LightFadeAnimation:   setStateData.LightFadeAnimation,
			LightBrightness:      setStateData.LightBrightness,
			PlayerIndicators:     playerIndicators,
			LedRed:               setStateData.LedRed,
			LedGreen:             setStateData.LedGreen,
			LedBlue:              setStateData.LedBlue,
		},
	}

	buffer := new(bytes.Buffer)
	err := binary.Write(buffer, binary.LittleEndian, packedUSBReportOut)
	if err != nil {
		return nil, fmt.Errorf("binary.Write: error trying to pack USBReportOut: %w", err)
	}
	return buffer.Bytes(), nil
}

type EffectType uint8

const (
	EffectTypeOff       = 0x05
	EffectTypeFeedback  = 0x21
	EffectTypeWeapon    = 0x25
	EffectTypeVibration = 0x26
)

func GenerateTriggerFFBParams(effectType EffectType, startPos, endPos, strength uint8) [11]uint8 {
	var params [11]uint8
	params[0] = uint8(effectType)
	params[1] = startPos
	params[2] = endPos
	params[3] = strength
	return params
}

var defaultSetStateData = SetStateData{
	EnableRumbleEmulation:         true,
	UseRumbleNotHaptics:           true,
	AllowRightTriggerFFB:          true,
	AllowLeftTriggerFFB:           true,
	AllowHeadphoneVolume:          true,
	AllowSpeakerVolume:            true,
	AllowMicVolume:                true,
	AllowAudioControl:             true,
	AllowMuteLight:                true,
	AllowAudioMute:                true,
	AllowLedColor:                 true,
	ResetLights:                   false,
	AllowPlayerIndicators:         true,
	AllowHapticLowPassFilter:      true,
	AllowMotorPowerLevel:          true,
	AllowAudioControl2:            true,
	RumbleEmulationRight:          0x00,
	RumbleEmulationLeft:           0x00,
	VolumeHeadphones:              0x00,
	VolumeSpeaker:                 0x00,
	VolumeMic:                     0x00,
	MicSelect:                     MicSelectAuto,
	EchoCancelEnable:              false,
	NoiseCancelEnable:             false,
	OutputPathSelect:              0x00,
	InputPathSelect:               0x00,
	MuteLight:                     MuteLightModeOff,
	TouchPowerSave:                false,
	MotionPowerSave:               false,
	HapticPowerSave:               false,
	AudioPowerSave:                false,
	MicMute:                       false,
	SpeakerMute:                   false,
	HeadphoneMute:                 false,
	HapticMute:                    false,
	RightTriggerFFB:               GenerateTriggerFFBParams(EffectTypeOff, 0x00, 0x00, 0x00),
	LeftTriggerFFB:                GenerateTriggerFFBParams(EffectTypeOff, 0x00, 0x00, 0x00),
	HostTimestamp:                 0x00,
	TriggerMotorPowerReduction:    0x00,
	RumbleMotorPowerReduction:     0x00,
	SpeakerCompPreGain:            0x00,
	BeamformingEnable:             false,
	AllowLightBrightnessChange:    false,
	AllowColorLightFadeAnimation:  false,
	EnableImprovedRumbleEmulation: false,
	HapticLowPassFilter:           true,
	LightFadeAnimation:            LightFadeAnimationFadeOut,
	LightBrightness:               LightBrightnessBright,
	PlayerLight1:                  false,
	PlayerLight2:                  false,
	PlayerLight3:                  false,
	PlayerLight4:                  false,
	PlayerLight5:                  false,
	PlayerLightFade:               false,
	LedRed:                        0xFF,
	LedGreen:                      0xFF,
	LedBlue:                       0xFF,
}
