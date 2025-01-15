// Copyright 2024 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package genai

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

// ReplayAPIClient is a client that reads responses from a replay session file.
type replayAPIClient struct {
	ReplayFile              *replayFile
	ReplaysDirectory        string
	currentInteractionIndex int
	t                       *testing.T
	server                  *httptest.Server
}

// NewReplayAPIClient creates a new ReplayAPIClient from a replay session file.
func newReplayAPIClient(t *testing.T) *replayAPIClient {
	t.Helper()
	// The replay files are expected to be in a directory specified by the environment variable
	// GOOGLE_GENAI_REPLAYS_DIRECTORY.
	replaysDirectory := os.Getenv("GOOGLE_GENAI_REPLAYS_DIRECTORY")
	rac := &replayAPIClient{
		ReplayFile:              nil,
		ReplaysDirectory:        replaysDirectory,
		currentInteractionIndex: 0,
		t:                       t,
	}
	rac.server = httptest.NewServer(rac)
	rac.t.Cleanup(func() {
		rac.server.Close()
	})
	return rac
}

// GetBaseURL returns the URL of the mocked HTTP server.
func (rac *replayAPIClient) GetBaseURL() string {
	return rac.server.URL
}

// CreateClient creates a new HTTP client that uses the replay session file.
func (rac *replayAPIClient) CreateClient(ctx context.Context) (*http.Client, error) {
	return rac.server.Client(), nil
}

// LoadReplay populates a replay session from a file based on the provided path.
func (rac *replayAPIClient) LoadReplay(replayFilePath string) {
	rac.t.Helper()
	fullReplaysPath := replayFilePath
	if rac.ReplaysDirectory != "" {
		fullReplaysPath = filepath.Join(rac.ReplaysDirectory, replayFilePath)
	}
	ReplayFile, err := readReplayFile(fullReplaysPath)
	if err != nil {
		rac.t.Errorf("error loading replay file, %v", err)
	}
	rac.ReplayFile = ReplayFile
}

func readReplayFile(path string) (*replayFile, error) {
	dat, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var replayFile replayFile
	if err := unmarshal(dat, &replayFile); err != nil {
		return nil, err
	}
	return &replayFile, nil
}

// LatestInteraction returns the interaction that was returned by the last call to ServeHTTP.
func (rac *replayAPIClient) LatestInteraction() *replayInteraction {
	rac.t.Helper()
	if rac.currentInteractionIndex == 0 {
		rac.t.Fatalf("no interactions has been made in replay session so far")
	}
	return rac.ReplayFile.Interactions[rac.currentInteractionIndex-1]
}

// ServeHTTP mocks serving HTTP requests.
func (rac *replayAPIClient) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	rac.t.Helper()
	if rac.ReplayFile == nil {
		rac.t.Fatalf("no replay file loaded")
	}
	if rac.currentInteractionIndex >= len(rac.ReplayFile.Interactions) {
		rac.t.Fatalf("no more interactions in replay session")
	}
	interaction := rac.ReplayFile.Interactions[rac.currentInteractionIndex]

	rac.assertRequest(req, interaction.Request)
	rac.currentInteractionIndex++
	var bodySegments []string
	for i := 0; i < len(interaction.Response.BodySegments); i++ {
		responseBodySegment, err := json.Marshal(interaction.Response.BodySegments[i])
		if err != nil {
			rac.t.Errorf("error marshalling responseBodySegment [%s], err: %+v", rac.ReplayFile.ReplayID, err)
		}
		bodySegments = append(bodySegments, string(responseBodySegment))
	}
	if interaction.Response.StatusCode != 0 {
		w.WriteHeader(int(interaction.Response.StatusCode))
	} else {
		w.WriteHeader(http.StatusOK)
	}
	w.Write([]byte(strings.Join(bodySegments, "\n")))
}

// `re“ is a regex that matches the fields that should be ignored in a given cmp.Path.
var re = regexp.MustCompile(strings.Join([]string{
	"map\\[string\\]any", // map[string]any
	"\\[\\]any",          // []any
	"\\d",                // [0..9]
	"\\[",                // [
	"\\]",                // ]
	"\\(",                // (
	"\\)",                // )
	"\"",                 // "
	"\\*",                // *
	"\\.",                // .
}, "|"))

func ignoreFields(a any, ignoringPath string) cmp.Option {
	return cmp.FilterPath(func(p cmp.Path) bool {
		if reflect.TypeOf(a) != p[0].Type().Elem() {
			return false
		}
		pathArray := make([]string, 0, len(p))
		for _, pathStep := range p[1:] {
			step := re.ReplaceAllString(pathStep.String(), "")
			if len(step) == 0 {
				continue
			}
			pathArray = append(pathArray, step)
		}
		calculatedPath := strings.Join(pathArray, ".")
		return strings.EqualFold(calculatedPath, ignoringPath)
	}, cmp.Ignore())
}

func cmpOptionByPath(a any, FieldPath string, opt cmp.Option) cmp.Option {
	return cmp.FilterPath(func(p cmp.Path) bool {
		if reflect.TypeOf(a) != p[0].Type().Elem() {
			return false
		}
		pathArray := make([]string, 0, len(p))
		for _, pathStep := range p[1:] {
			step := re.ReplaceAllString(pathStep.String(), "")
			if len(step) == 0 {
				continue
			}
			pathArray = append(pathArray, step)
		}
		calculatedPath := strings.Join(pathArray, ".")
		return strings.EqualFold(calculatedPath, FieldPath)
	}, opt)
}

func (rac *replayAPIClient) assertRequest(sdkRequest *http.Request, want *replayRequest) {
	rac.t.Helper()
	sdkRequestBody, err := io.ReadAll(sdkRequest.Body)
	if err != nil {
		rac.t.Errorf("Error reading request body, err: %+v", err)
	}
	bodySegment := make(map[string]any)
	if err := json.Unmarshal(sdkRequestBody, &bodySegment); err != nil {
		rac.t.Errorf("Error unmarshalling body, err: %+v", err)
	}
	headers := make(map[string]string)
	for k, v := range sdkRequest.Header {
		headers[k] = strings.Join(v, ",")
	}
	got := &replayRequest{
		Method:       sdkRequest.Method,
		Url:          sdkRequest.URL.String(),
		Headers:      headers,
		BodySegments: []map[string]any{bodySegment},
	}
	// TODO(b/383753309): Maybe it's better to adjust the expected replayRequest when reading from the
	// replay file, instead of doing the adjustment logic in comparator.
	// Because when comparator returns false, the diff message still shows the content before
	// the adjustment.
	opts := cmp.Options{
		// Verifying that one url is the suffix of the other is enough for this validation.
		cmpOptionByPath(replayRequest{}, "Url", cmp.Comparer(func(x, y string) bool {
			vertexPrefix := regexp.MustCompile(`{VERTEX_URL_PREFIX}|{MLDEV_URL_PREFIX}`)
			x = vertexPrefix.ReplaceAllString(x, "")
			y = vertexPrefix.ReplaceAllString(y, "")
			return strings.HasSuffix(x, y) || strings.HasSuffix(y, x)
		})),
		cmpOptionByPath(replayRequest{}, "Headers", cmp.Comparer(func(x, y map[string]string) bool {
			requiredHeaders := []string{"user-agent", "x-goog-api-client", "content-type"}
			keyToLowerCase := func(m map[string]string) map[string]string {
				result := make(map[string]string)
				for k, v := range m {
					result[strings.ToLower(k)] = v
				}
				return result
			}
			x = keyToLowerCase(x)
			y = keyToLowerCase(y)
			for _, header := range requiredHeaders {
				if _, ok := x[header]; !ok {
					return false
				}
				if _, ok := y[header]; !ok {
					return false
				}
			}
			return true
		})),
		// Verifying that one cachedContent is the suffix of the other suffices for this validation.
		cmpOptionByPath(replayRequest{}, "BodySegments.cachedContent", cmp.Comparer(func(x, y any) bool {
			xStr := strings.ReplaceAll(x.(string), "{PROJECT_AND_LOCATION_PATH}", "")
			yStr := strings.ReplaceAll(y.(string), "{PROJECT_AND_LOCATION_PATH}", "")
			return strings.HasSuffix(xStr, yStr) || strings.HasSuffix(yStr, xStr)
		})),
		// Verifying generationConfig with default values logic.
		cmpOptionByPath(replayRequest{}, "BodySegments.generationConfig", cmp.Comparer(func(x, y any) bool {
			xStruct := &GenerateContentConfig{}
			yStruct := &GenerateContentConfig{}
			mapToStruct(x.(map[string]any), xStruct)
			mapToStruct(y.(map[string]any), yStruct)
			xStruct.setDefaults()
			yStruct.setDefaults()
			return reflect.DeepEqual(xStruct, yStruct)
		})),
		// Handles the lowercase/uppercase differences in the data, e.g. 'Method: "POST" vs "post"'.
		cmp.FilterPath(func(p cmp.Path) bool {
			return p.String() != "Url"
		}, cmp.Comparer(func(x, y string) bool {
			// There are cases where the replay file does not contain the new lines or the extra spaces in
			// the request while the actual request does. We remove the new lines and extra spaces to
			// avoid false negatives and only compare the actual data.
			newLineRegexp := regexp.MustCompile(`\r?\n`)
			x = newLineRegexp.ReplaceAllString(x, "")
			y = newLineRegexp.ReplaceAllString(y, "")
			multiSpaceRegexp := regexp.MustCompile(`\s+`)
			x = multiSpaceRegexp.ReplaceAllString(x, " ")
			y = multiSpaceRegexp.ReplaceAllString(y, " ")
			x = strings.TrimSpace(x)
			y = strings.TrimSpace(y)
			return strings.EqualFold(x, y)
		})),
	}
	if diff := cmp.Diff(got, want, opts); diff != "" {
		rac.t.Errorf("Requests had diffs (-got +want):\n%v", diff)
	}
}

// TODO(b/378168548): Decouple the custom JSON unmarshaller from ReplayAPIClient.
// Algorithm:
// 1. Unmarshal the JSON into a map[string]any.
// 2. Modify the map so the keys are all in camel case.
// 3. Marshal the modified map[string]any into JSON.
// 4. Unmarshal the modified JSON into the given type.
func unmarshal(data []byte, v any) error {
	// 1. Unmarshal the JSON into a map[string]any.
	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		return fmt.Errorf("ReplayAPIClient: error unmarshalling: %w", err)
	}
	// 2. Sanitize the map so it matches Go SDK replay tests.
	sanitizeReplayTestContent(m)
	// 3. Marshal the modified map[string]any into JSON.
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return fmt.Errorf("ReplayAPIClient: error marshalling the modified map: %w", err)
	}
	// 4. Unmarshal the modified JSON into the given type.
	return json.Unmarshal(data, v)
}

func sanitizeReplayTestContent(m any) {
	// TODO(b/380886719): Modify the replay parser to ignore empty values for `map[string]any` types.
	omitEmptyValues(m)
	// 2. Modify the map so the keys are all in camel case.
	convertKeysToCamelCase(m)
}

// omitEmptyValues recursively traverses the given value and if it is a `map[string]any` or
// `[]any`, it omits the empty values.
func omitEmptyValues(v any) {
	if v == nil {
		return
	}
	switch m := v.(type) {
	case map[string]any:
		for k, v := range m {
			// If the value is empty, delete the key from the map.
			if reflect.ValueOf(v).IsZero() {
				delete(m, k)
			} else {
				omitEmptyValues(v)
			}
		}
	case []any:
		for _, item := range m {
			omitEmptyValues(item)
		}
	}
}

// convertKeysToCamelCase recursively traverses the given value and if it is a `map[string]any`, it
// converts the keys to camel case.
func convertKeysToCamelCase(v any) {
	if v == nil {
		return
	}
	switch m := v.(type) {
	case map[string]any:
		for key, value := range m {
			camelCaseKey := toCamelCase(key)
			if camelCaseKey != key {
				m[camelCaseKey] = value
				delete(m, key)
			}
			convertKeysToCamelCase(value)
		}
	case []any:
		for _, item := range m {
			convertKeysToCamelCase(item)
		}
	}
}

// toCamelCase converts a string from snake case to camel case.
// Examples:
//
//	"foo" -> "foo"
//	"fooBar" -> "fooBar"
//	"foo_bar" -> "fooBar"
//	"foo_bar_baz" -> "fooBarBaz"
//	"foo-bar" -> "foo-bar"
func toCamelCase(s string) string {
	parts := strings.Split(s, "_")
	if len(parts) == 1 {
		// There is no underscore, so no need to modify the string.
		return s
	}
	// Skip the first word and convert the first letter of the remaining words to uppercase.
	for i, part := range parts[1:] {
		parts[i+1] = strings.ToUpper(part[:1]) + part[1:]
	}
	// Concat the parts back together to mak a camelCase string.
	return strings.Join(parts, "")
}
