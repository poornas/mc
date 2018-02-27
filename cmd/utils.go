/*
 * Minio Client (C) 2015, 2016, 2017 Minio, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package cmd

import (
	"crypto/md5"
	"crypto/tls"
	"encoding/base64"
	"io"
	"math/rand"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/minio/minio-go"

	"github.com/minio/mc/pkg/console"
	"github.com/minio/mc/pkg/probe"
)

func isErrIgnored(err *probe.Error) (ignored bool) {
	// For all non critical errors we can continue for the remaining files.
	switch err.ToGoError().(type) {
	// Handle these specifically for filesystem related errors.
	case BrokenSymlink, TooManyLevelsSymlink, PathNotFound, PathInsufficientPermission:
		ignored = true
	// Handle these specifically for object storage related errors.
	case BucketNameEmpty, ObjectMissing, ObjectAlreadyExists:
		ignored = true
	case ObjectAlreadyExistsAsDirectory, BucketDoesNotExist, BucketInvalid, ObjectOnGlacier:
		ignored = true
	default:
		ignored = false
	}
	return ignored
}

// UTCNow - returns current UTC time.
func UTCNow() time.Time {
	return time.Now().UTC()
}

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

// newRandomID generates a random id of regular lower case and uppercase english characters.
func newRandomID(n int) string {
	rand.Seed(UTCNow().UnixNano())
	sid := make([]rune, n)
	for i := range sid {
		sid[i] = letters[rand.Intn(len(letters))]
	}
	return string(sid)
}

// dumpTlsCertificates prints some fields of the certificates received from the server.
// Fields will be inspected by the user, so they must be conscise and useful
func dumpTLSCertificates(t *tls.ConnectionState) {
	for _, cert := range t.PeerCertificates {
		console.Debugln("TLS Certificate found: ")
		if len(cert.Issuer.Country) > 0 {
			console.Debugln(" >> Country: " + cert.Issuer.Country[0])
		}
		if len(cert.Issuer.Organization) > 0 {
			console.Debugln(" >> Organization: " + cert.Issuer.Organization[0])
		}
		console.Debugln(" >> Expires: " + cert.NotAfter.String())
	}
}

// isStdIO checks if the input parameter is one of the standard input/output streams
func isStdIO(reader io.Reader) bool {
	return reader == os.Stdin || reader == os.Stdout || reader == os.Stderr
}

// splitStr splits a string into n parts, empty strings are added
// if we are not able to reach n elements
func splitStr(path, sep string, n int) []string {
	splits := strings.SplitN(path, sep, n)
	// Add empty strings if we found elements less than nr
	for i := n - len(splits); i > 0; i-- {
		splits = append(splits, "")
	}
	return splits
}

// newS3Config simply creates a new Config struct using the passed
// parameters.
func newS3Config(urlStr string, hostCfg *hostConfigV9) *Config {
	// We have a valid alias and hostConfig. We populate the
	// credentials from the match found in the config file.
	s3Config := new(Config)

	s3Config.AppName = "mc"
	s3Config.AppVersion = Version
	s3Config.AppComments = []string{os.Args[0], runtime.GOOS, runtime.GOARCH}
	s3Config.Debug = globalDebug
	s3Config.Insecure = globalInsecure

	s3Config.HostURL = urlStr
	if hostCfg != nil {
		s3Config.AccessKey = hostCfg.AccessKey
		s3Config.SecretKey = hostCfg.SecretKey
		s3Config.Signature = hostCfg.API
	}
	s3Config.Lookup = getLookupType(hostCfg.Lookup)
	return s3Config
}

// lineTrunc - truncates a string to the given maximum length by
// adding ellipsis in the middle
func lineTrunc(content string, maxLen int) string {
	runes := []rune(content)
	rlen := len(runes)
	if rlen <= maxLen {
		return content
	}
	halfLen := maxLen / 2
	fstPart := string(runes[0:halfLen])
	sndPart := string(runes[rlen-halfLen:])
	return fstPart + "…" + sndPart
}

// isOlder returns true if the passed object is older than olderRef
func isOlder(c *clientContent, olderRef int) bool {
	objectAge := UTCNow().Sub(c.Time)
	return objectAge < (time.Duration(olderRef) * Day)
}

// isNewer returns true if the passed object is newer than newerRef
func isNewer(c *clientContent, newerRef int) bool {
	objectAge := UTCNow().Sub(c.Time)
	return objectAge > (time.Duration(newerRef) * Day)
}

// getLookupType returns the minio.BucketLookupType for lookup
// option entered on the command line
func getLookupType(l string) minio.BucketLookupType {
	l = strings.ToLower(l)
	switch l {
	case "dns":
		return minio.BucketLookupDNS
	case "path":
		return minio.BucketLookupPath
	}
	return minio.BucketLookupAuto
}

// sumMD5Base64 calculate md5sum for an input byte array, returns base64 encoded.
func sumMD5Base64(data []byte) string {
	hash := md5.New()
	hash.Write(data)
	return base64.StdEncoding.EncodeToString(hash.Sum(nil))
}
