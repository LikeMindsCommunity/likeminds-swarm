package externalHelpers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/nateshr/likeminds-swarm/internal/services/logging"
)

// InvalidateKettleCache | deletes the cache for the given keyPattern in kettle service
func InvalidateKettleCache(keyPattern string) {

	headers := gin.H{
		"x-platform-type": SwarmServiceHeader,
	}

	requestBody := map[string]interface{}{
		"key_pattern": keyPattern,
	}

	// Send request to disable webhook
	respBytes, statusCode, err := GetRequestResponse(KettleService, KettleCacheDeleteEndpoint, DELETERequest, headers, nil, requestBody)
	if err != nil || statusCode != http.StatusOK {
		logging.Error("Error Deleting cache for key: ", keyPattern, " Response: ", string(respBytes), " Error: ", err)
		return
	}

	logging.Info("Kettle Cache deleted successfully for key: ", keyPattern)
}
