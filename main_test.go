package genai

import (
	"flag"
	"fmt"
	"os"
	"regexp"
	"testing"
)

const (
	apiMode     = "api"
	replayMode  = "replay"
	requestMode = "request"
)

// TODO(b/382773687): Enable the TestModelsGenerateContentStream tests.
var (
	disabledTestsForAll = []string{
		// TODO(b/380108306): bytes related tests are not supported in replay tests.
		"vertex/models/generate_content_part/test_image_base64",
		"mldev/models/generate_content_part/test_image_base64",
		// TODO(b/383753309): Refactor replay test to use url safe base64.
		"mldev/models/generate_image/test_all_mldev_config_parameters",
		"mldev/models/generate_image/test_all_vertexai_config_parameters",
		"mldev/models/generate_image/test_simple_prompt",
		"vertex/models/generate_image/test_all_mldev_config_parameters",
		"vertex/models/generate_image/test_all_vertexai_config_parameters",
		"vertex/models/generate_image/test_simple_prompt",
	}
	disabledTestsByMode = map[string][]string{
		apiMode: []string{
			"TestTable/",
			"TestModelsGenerateContentStream/mldev/models/generate_content/test_simple_shared_generation_config",
			"TestModelsGenerateContentStream/mldev/models/generate_content_cached_content/",
			"TestModelsGenerateContentStream/mldev/models/generate_content_part/",
			"TestModelsGenerateContentStream/vertex/models/generate_content/test_2_candidates_gemini_1_5_flash",
			"TestModelsGenerateContentStream/vertex/models/generate_content/test_llama",
			"TestModelsGenerateContentStream/vertex/models/generate_content/test_simple_shared_generation_config",
			"TestModelsGenerateContentStream/vertex/models/generate_content_cached_content",
			"TestModelsGenerateContentStream/vertex/models/generate_content_part/test_video_gcs_file_uri",
			"TestModelsGenerateContentStream/vertex/models/generate_content_tools/test_code_execution",
			"TestModelsGenerateContentStream/mldev/models/generate_content_model",
			"TestModelsGenerateContentAudio/",
		},
		replayMode: []string{
			"TestModelsGenerateContentStream/",
			// TODO(b/383351834): Enable the test after the bug is fixed.
			"TestTable/mldev/models/generate_content_config_zero_value",
		},
		requestMode: []string{
			"TestTable/",
			"TestModelsGenerateContentStream/",
		},
	}
	mode     = flag.String("mode", replayMode, "Test mode")
	backends = []struct {
		name    string
		Backend Backend
	}{
		{
			name:    "mldev",
			Backend: BackendGeminiAPI,
		},
		{
			name:    "vertex",
			Backend: BackendVertexAI,
		},
	}
)

func isDisabledTest(t *testing.T) bool {
	disabledTestPatterns := append(disabledTestsForAll, disabledTestsByMode[*mode]...)
	for _, p := range disabledTestPatterns {
		r := regexp.MustCompile(p)
		if r.MatchString(t.Name()) {
			return true
		}
	}
	return false
}

func TestMain(m *testing.M) {
	flag.Parse()
	fmt.Println("Running tests in", *mode)
	exitCode := m.Run()
	os.Exit(exitCode)
}
