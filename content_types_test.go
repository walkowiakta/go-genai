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
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestUnionTypes(t *testing.T) {
	mapValue := map[string]any{"key": "value"}
	t.Run("TextToContents", func(t *testing.T) {
		text := Text("Hello")
		expected := []*Content{{
			Parts: []*Part{{Text: "Hello"}},
			Role:  roleUser,
		}}
		got := text.ToContents()
		if diff := cmp.Diff(got, expected); diff != "" {
			t.Errorf("ToContents() mismatch (-want +got):\n%s", diff)
		}
	})

	t.Run("TextsToContents", func(t *testing.T) {
		texts := Texts{"Hello", "World"}
		expected := []*Content{{
			Parts: []*Part{{Text: "Hello"}, {Text: "World"}},
			Role:  roleUser,
		}}

		got := texts.ToContents()
		if diff := cmp.Diff(got, expected); diff != "" {
			t.Errorf("ToContents() mismatch (-want +got):\n%s", diff)
		}
	})

	t.Run("ContentsToContents", func(t *testing.T) {
		contents := ContentSlice{
			{
				Parts: []*Part{{Text: "Hello"}},
				Role:  roleUser,
			},
			{
				Parts: []*Part{{Text: "World"}},
				Role:  roleUser,
			},
		}

		expected := []*Content{
			{
				Parts: []*Part{{Text: "Hello"}},
				Role:  roleUser,
			},
			{
				Parts: []*Part{{Text: "World"}},
				Role:  roleUser,
			},
		}
		got := contents.ToContents()

		if diff := cmp.Diff(got, expected); diff != "" {
			t.Errorf("ToContents() mismatch (-want +got):\n%s", diff)
		}
	})

	t.Run("PartsToPart", func(t *testing.T) {
		parts := PartSlice{
			Text("Hello"),
			FileData{FileURI: "gs://generativeai-downloads/images/scones.jpg", MIMEType: "image/jpeg"},
		}
		expected := []*Part{
			{Text: "Hello"},
			{FileData: &FileData{FileURI: "gs://generativeai-downloads/images/scones.jpg", MIMEType: "image/jpeg"}},
		}

		got := parts.toPart()

		if diff := cmp.Diff(got, expected, cmp.AllowUnexported(FileData{})); diff != "" {
			t.Errorf("toPart() mismatch (-want +got):\n%s", diff)
		}
	})

	t.Run("PartsToContents", func(t *testing.T) {
		parts := PartSlice{Text("one"), Text("two")}
		expected := []*Content{{
			Parts: []*Part{{Text: "one"}, {Text: "two"}},
			Role:  roleUser,
		}}
		got := parts.ToContents()

		if diff := cmp.Diff(got, expected); diff != "" {
			t.Errorf("ToContents() mismatch (-want +got):\n%s", diff)
		}
	})

	t.Run("PartToPart", func(t *testing.T) {
		expected := &Part{Text: "hello"}
		got := Part{Text: "hello"}.toPart()
		if diff := cmp.Diff(got, expected); diff != "" {
			t.Errorf("toPart() mismatch (-want +got):\n%s", diff)
		}
	})

	t.Run("FileDataPart", func(t *testing.T) {
		fileData := FileData{FileURI: "file123"}
		expected := &Part{FileData: &FileData{FileURI: "file123"}}
		got := fileData.toPart()
		if diff := cmp.Diff(got, expected); diff != "" {
			t.Errorf("toPart() mismatch (-want +got):\n%s", diff)
		}
	})

	t.Run("InlinePart", func(t *testing.T) {
		blob := InlineData{MIMEType: "text/plain", Data: []byte("hello")}
		expected := &Part{InlineData: &Blob{MIMEType: "text/plain", Data: []byte("hello")}}
		got := blob.toPart()
		if diff := cmp.Diff(got, expected, cmp.AllowUnexported(Blob{})); diff != "" {
			t.Errorf("toPart() mismatch (-want +got):\n%s", diff)
		}
	})

	t.Run("FunctionCallPart", func(t *testing.T) {
		functionCall := FunctionCall{Name: "test_function", Args: mapValue}
		expected := &Part{FunctionCall: &FunctionCall{Name: "test_function", Args: mapValue}}
		got := functionCall.toPart()
		if diff := cmp.Diff(got, expected); diff != "" {
			t.Errorf("toPart() mismatch (-want +got):\n%s", diff)
		}
	})

	t.Run("FunctionResponsePart", func(t *testing.T) {
		functionResponse := FunctionResponse{Name: "test_function", Response: mapValue}
		expected := &Part{FunctionResponse: &FunctionResponse{Name: "test_function", Response: mapValue}}
		got := functionResponse.toPart()
		if diff := cmp.Diff(got, expected); diff != "" {
			t.Errorf("toPart() mismatch (-want +got):\n%s", diff)
		}
	})

	t.Run("ExecutableCodePart", func(t *testing.T) {
		executableCode := ExecutableCode{Code: "print('hello')"}
		expected := &Part{ExecutableCode: &ExecutableCode{Code: "print('hello')"}}
		got := executableCode.toPart()
		if diff := cmp.Diff(got, expected); diff != "" {
			t.Errorf("toPart() mismatch (-want +got):\n%s", diff)
		}
	})

	t.Run("CodeExecutionResultPart", func(t *testing.T) {
		codeExecutionResult := CodeExecutionResult{Output: "print('hello')", Outcome: "hello"}
		expected := &Part{CodeExecutionResult: &CodeExecutionResult{Output: "print('hello')", Outcome: "hello"}}
		got := codeExecutionResult.toPart()
		if diff := cmp.Diff(got, expected); diff != "" {
			t.Errorf("toPart() mismatch (-want +got):\n%s", diff)
		}
	})

	t.Run("VideoMetadataPart", func(t *testing.T) {
		videoMetadata := VideoMetadata{StartOffset: "0s", EndOffset: "10s"}
		expected := &Part{VideoMetadata: &VideoMetadata{StartOffset: "0s", EndOffset: "10s"}}
		got := videoMetadata.toPart()
		if diff := cmp.Diff(got, expected); diff != "" {
			t.Errorf("toPart() mismatch (-want +got):\n%s", diff)
		}
	})
}
