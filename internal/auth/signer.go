package auth

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strconv"
	"time"
)

// GenerateAuthToken creates the signature required for the Login endpoint.
func GenerateAuthToken(userNonce, userKey, integrationId string) string {
	// 1. Get current Unix timestamp (Seconds)
	// Ensure your system clock is synced! Time drift > 5-10 mins will cause 403s.
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)

	// 2. Construct the payload string: timestamp + userKey
	payload := timestamp + userKey

	// 3. Generate SHA-256 hash
	hash := sha256.Sum256([]byte(payload))
	hexEncodedHash := hex.EncodeToString(hash[:]) // Go outputs lowercase hex by default

	// 4. Construct final token string
	// Format: userNonce:timestamp:hexEncodedHash[:integrationIdentifier]
	
	baseToken := fmt.Sprintf("%s:%s:%s", userNonce, timestamp, hexEncodedHash)

	if integrationId != "" {
		baseToken = fmt.Sprintf("%s:%s", baseToken, integrationId)
	}

	// return strings.ToLower(baseToken) <--- REMOVED THIS LINE
	return baseToken
}
