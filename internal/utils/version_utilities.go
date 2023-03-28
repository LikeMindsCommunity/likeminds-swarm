package utils

import (
	"strconv"
)

const (
	PlatformAndroid     string = "an"
	PlatformWeb         string = "web"
	PlatformIoS         string = "ios"
	PlatformFlutter     string = "fl"
	PlatformReactNative string = "rn"
)

var UnreleasedMaxVersion int = 9999
var UnreleasedMinVersion int = -1

var FeedMediaVersions = map[string]int{
	PlatformAndroid:     UnreleasedMinVersion,
	PlatformWeb:         UnreleasedMinVersion,
	PlatformIoS:         UnreleasedMinVersion,
	PlatformFlutter:     4,
	PlatformReactNative: UnreleasedMinVersion,
}

// Exposed Method to check version code accessibility for a feature
func CheckVersion(featureVersionCode map[string]int, versionCode string, platformCode string) bool {
	/*
		returns True if,
		  versionCode >= featureVersionCode for the given platform
		returns False for all other cases
	*/

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
