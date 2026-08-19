package externalHelpers

import (
	"net/http"

	"github.com/LikeMindsCommunity/likeminds-swarm/internal/services/logging"
	"github.com/gin-gonic/gin"
)

// InvalidateKettleCache | deletes the cache for the given keyPattern in kettle service
func InvalidateKettleCache(keyPatterns []string) {

	headers := gin.H{
		"x-platform-type": SwarmServiceHeader,
	}

	requestBody := map[string]interface{}{
		"key_patterns": keyPatterns,
	}

	// Send request to disable webhook
	respBytes, statusCode, err := GetRequestResponse(KettleService, KettleCacheDeleteEndpoint, DELETERequest, headers, nil, requestBody)
	if err != nil || statusCode != http.StatusOK {
		logging.Error("Error Deleting cache for key: ", keyPatterns, " Response: ", string(respBytes), " Error: ", err)
		return
	}

	logging.Info("Sent request for kettle cache deletion for keys: ", keyPatterns)
}
