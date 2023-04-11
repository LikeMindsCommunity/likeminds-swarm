package utils

import (
	"fmt"
	"log"
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

var FeedImageAndLinkMediaVersions = map[string]int{
	PlatformAndroid:     UnreleasedMinVersion,
	PlatformWeb:         UnreleasedMinVersion,
	PlatformIoS:         UnreleasedMinVersion,
	PlatformFlutter:     4,
	PlatformReactNative: UnreleasedMinVersion,
}

var EditFeedEntityVersions = map[string]int{
	PlatformAndroid:     UnreleasedMaxVersion,
	PlatformWeb:         UnreleasedMaxVersion,
	PlatformIoS:         UnreleasedMaxVersion,
	PlatformFlutter:     UnreleasedMaxVersion,
	PlatformReactNative: UnreleasedMaxVersion,
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
		log.Printf("CheckVersion() - setting default version code as header is missing in request")
	}

	if platformCode == "" {
		platformCode = PlatformWeb
		log.Printf("CheckVersion() - setting default platform code as header is missing in request")
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
		log.Printf("CheckVersion() - setting default version code as header is missing in request")
	}

	if platformCode == "" {
		platformCode = PlatformWeb
		log.Printf("CheckVersion() - setting default platform code as header is missing in request")
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
