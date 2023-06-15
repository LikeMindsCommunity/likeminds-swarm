package utils

import (
	"fmt"
	"strconv"
)

const (
	PlatformAndroid     string = "an"
	PlatformWeb         string = "web"
	PlatformIoS         string = "ios"
	PlatformFlutter     string = "fl"
	PlatformReactNative string = "rn"
	PlatformReactJS     string = "rt"
)

var UnreleasedMaxVersion int = 9999
var UnreleasedMinVersion int = -1

var FeedVideoAndDocumentMediaVersions = map[string]int{
	PlatformAndroid:     UnreleasedMinVersion,
	PlatformWeb:         UnreleasedMinVersion,
	PlatformIoS:         UnreleasedMinVersion,
	PlatformFlutter:     4,
	PlatformReactNative: UnreleasedMinVersion,
	PlatformReactJS:     UnreleasedMinVersion,
}

var FeedLinkMediaVersion = map[string]int{
	PlatformAndroid:     UnreleasedMinVersion,
	PlatformWeb:         UnreleasedMinVersion,
	PlatformIoS:         UnreleasedMinVersion,
	PlatformFlutter:     5,
	PlatformReactNative: UnreleasedMinVersion,
	PlatformReactJS:     UnreleasedMinVersion,
}

var EditFeedEntityVersions = map[string]int{
	PlatformAndroid:     2,
	PlatformWeb:         UnreleasedMaxVersion,
	PlatformIoS:         UnreleasedMaxVersion,
	PlatformFlutter:     5,
	PlatformReactNative: UnreleasedMaxVersion,
	PlatformReactJS:     UnreleasedMaxVersion,
}

// Exposed Method to check version code accessibility for a feature
func CheckVersion(featureVersionCode map[string]int, versionCode string, platformCode string) bool {
	/*
		returns True if,
		  versionCode >= featureVersionCode for the given platform
		returns False for all other cases
	*/

	if versionCode == "" {
		versionCode = fmt.Sprintf("%d", UnreleasedMinVersion)
	}

	if platformCode == "" {
		platformCode = PlatformWeb
	}

	var isVersionCheck bool = false

	featureVersionCodeForPlatform, ok := featureVersionCode[platformCode]

	if ok {
		versionCode, versionCodeConversionErr := strconv.Atoi(versionCode)

		if versionCodeConversionErr == nil {
			isVersionCheck = versionCode >= featureVersionCodeForPlatform
		}
	}

	return isVersionCheck
}

// Exposed Method to check version code accessibility for a feature
func CheckVersionInverted(featureVersionCode map[string]int, versionCode string, platformCode string) bool {
	/*
		returns True if,
		  versionCode < featureVersionCode for the given platform
		returns False for all other cases
	*/

	if versionCode == "" {
		versionCode = fmt.Sprintf("%d", UnreleasedMaxVersion)
	}

	if platformCode == "" {
		platformCode = PlatformWeb
	}

	var isVersionCheck bool = false

	featureVersionCodeForPlatform, ok := featureVersionCode[platformCode]

	if ok {
		versionCode, versionCodeConversionErr := strconv.Atoi(versionCode)

		if versionCodeConversionErr == nil {
			isVersionCheck = versionCode < featureVersionCodeForPlatform
		}
	}

	return isVersionCheck
}
