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
	"errors"
	"reflect"
	"testing"
)

func createGenerateContentResponse(candidates []*Candidate) *GenerateContentResponse {
	return &GenerateContentResponse{
		Candidates: candidates,
	}
}

func TestText(t *testing.T) {
	tests := []struct {
		name          string
		response      *GenerateContentResponse
		expectedText  string
		expectedError error
	}{
		{
			name:         "Empty Candidates",
			response:     createGenerateContentResponse([]*Candidate{}),
			expectedText: "",
		},
		{
			name: "Multiple Candidates",
			response: createGenerateContentResponse([]*Candidate{
				{Content: &Content{Parts: []*Part{{Text: "text1", Thought: false}}}},
				{Content: &Content{Parts: []*Part{{Text: "text2", Thought: false}}}},
			}),
			expectedText: "text1",
		},
		{
			name: "Empty Parts",
			response: createGenerateContentResponse([]*Candidate{
				{Content: &Content{Parts: []*Part{}}},
			}),
			expectedText: "",
		},
		{
			name: "Part With Text",
			response: createGenerateContentResponse([]*Candidate{
				{Content: &Content{Parts: []*Part{{Text: "text", Thought: false}}}},
			}),
			expectedText: "text",
		},
		{
			name: "Multiple Parts With Text",
			response: createGenerateContentResponse([]*Candidate{
				{Content: &Content{Parts: []*Part{
					{Text: "text1", Thought: false},
					{Text: "text2", Thought: false},
				}}},
			}),
			expectedText: "text1text2",
		},
		{
			name: "Multiple Parts With Thought",
			response: createGenerateContentResponse([]*Candidate{
				{Content: &Content{Parts: []*Part{
					{Text: "text1", Thought: false},
					{Text: "text2", Thought: true},
				}}},
			}),
			expectedText: "text1",
		},
		{
			name: "Part With InlineData",
			response: createGenerateContentResponse([]*Candidate{
				{Content: &Content{Parts: []*Part{
					{Text: "text1", Thought: false},
					{InlineData: &Blob{}},
				}}},
			}),
			expectedError: errors.New("GenerateContentResponse.Text only supports text parts, but got InlineData"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tt.response.Text()

			if tt.expectedError != nil {
				if err == nil || err.Error() != tt.expectedError.Error() {
					t.Fatalf("expected error %v, got %v", tt.expectedError, err)
				}
			} else {
				if err != nil {
					t.Fatalf("expected no error, got %v", err)
				}
				if result != tt.expectedText {
					t.Fatalf("expected text %v, got %v", tt.expectedText, result)
				}
			}
		})
	}
}

func TestFunctionCalls(t *testing.T) {
	tests := []struct {
		name                  string
		response              *GenerateContentResponse
		expectedFunctionCalls []*FunctionCall
	}{
		{
			name:                  "Empty Candidates",
			response:              createGenerateContentResponse([]*Candidate{}),
			expectedFunctionCalls: nil,
		},
		{
			name: "Multiple Candidates",
			response: createGenerateContentResponse([]*Candidate{
				{Content: &Content{Parts: []*Part{{FunctionCall: &FunctionCall{Name: "funcCall1", Args: map[string]any{"key1": "val1"}}}}}},
				{Content: &Content{Parts: []*Part{{FunctionCall: &FunctionCall{Name: "funcCall2", Args: map[string]any{"key2": "val2"}}}}}},
			}),
			expectedFunctionCalls: []*FunctionCall{
				{Name: "funcCall1", Args: map[string]any{"key1": "val1"}},
			},
		},
		{
			name: "Empty Parts",
			response: createGenerateContentResponse([]*Candidate{
				{Content: &Content{Parts: []*Part{}}},
			}),
			expectedFunctionCalls: nil,
		},
		{
			name: "Part With FunctionCall",
			response: createGenerateContentResponse([]*Candidate{
				{Content: &Content{Parts: []*Part{{FunctionCall: &FunctionCall{Name: "funcCall1", Args: map[string]any{"key1": "val1"}}}}}},
			}),
			expectedFunctionCalls: []*FunctionCall{
				{Name: "funcCall1", Args: map[string]any{"key1": "val1"}},
			},
		},
		{
			name: "Multiple Parts With FunctionCall",
			response: createGenerateContentResponse([]*Candidate{
				{Content: &Content{Parts: []*Part{
					{FunctionCall: &FunctionCall{Name: "funcCall1", Args: map[string]any{"key1": "val1"}}},
					{FunctionCall: &FunctionCall{Name: "funcCall2", Args: map[string]any{"key2": "val2"}}},
				}}},
			}),
			expectedFunctionCalls: []*FunctionCall{
				{Name: "funcCall1", Args: map[string]any{"key1": "val1"}},
				{Name: "funcCall2", Args: map[string]any{"key2": "val2"}},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.response.FunctionCalls()

			if !reflect.DeepEqual(result, tt.expectedFunctionCalls) {
				t.Fatalf("expected function calls %v, got %v", tt.expectedFunctionCalls, result)
			}
		})
	}
}

func TestNewPartFromURI(t *testing.T) {
	fileURI := "http://example.com/video.mp4"
	mimeType := "video/mp4"
	expected := &Part{
		FileData: &FileData{
			FileURI:  fileURI,
			MIMEType: mimeType,
		},
	}

	result := NewPartFromURI(fileURI, mimeType)
	if !reflect.DeepEqual(result, expected) {
		t.Fatalf("expected %v, got %v", expected, result)
	}
}

func TestNewPartFromText(t *testing.T) {
	text := "Hello, world!"
	expected := &Part{
		Text: text,
	}

	result := NewPartFromText(text)
	if !reflect.DeepEqual(result, expected) {
		t.Fatalf("expected %v, got %v", expected, result)
	}
}

func TestNewPartFromBytes(t *testing.T) {
	data := []byte{0x01, 0x02, 0x03}
	mimeType := "application/octet-stream"
	expected := &Part{
		InlineData: &Blob{
			Data:     data,
			MIMEType: mimeType,
		},
	}

	result := NewPartFromBytes(data, mimeType)
	if !reflect.DeepEqual(result, expected) {
		t.Fatalf("expected %v, got %v", expected, result)
	}
}

func TestNewPartFromFunctionCall(t *testing.T) {
	funcName := "myFunction"
	args := map[string]any{"arg1": "value1"}
	expected := &Part{
		FunctionCall: &FunctionCall{
			Name: "myFunction",
			Args: map[string]any{"arg1": "value1"},
		},
	}

	result := NewPartFromFunctionCall(funcName, args)
	if !reflect.DeepEqual(result, expected) {
		t.Fatalf("expected %v, got %v", expected, result)
	}
}

func TestNewPartFromFunctionResponse(t *testing.T) {
	funcName := "myFunction"
	response := map[string]any{"result": "success"}
	expected := &Part{
		FunctionResponse: &FunctionResponse{
			Name:     "myFunction",
			Response: map[string]any{"result": "success"},
		},
	}

	result := NewPartFromFunctionResponse(funcName, response)
	if !reflect.DeepEqual(result, expected) {
		t.Fatalf("expected %v, got %v", expected, result)
	}
}

func TestNewPartFromVideoMetadata(t *testing.T) {
	endOffset := "00:01:00"
	startOffset := "00:00:00"
	expected := &Part{
		VideoMetadata: &VideoMetadata{
			EndOffset:   endOffset,
			StartOffset: startOffset,
		},
	}

	result := NewPartFromVideoMetadata(endOffset, startOffset)
	if !reflect.DeepEqual(result, expected) {
		t.Fatalf("expected %v, got %v", expected, result)
	}
}

func TestNewPartFromExecutableCode(t *testing.T) {
	code := "print('Hello, world!')"
	language := LanguagePython
	expected := &Part{
		ExecutableCode: &ExecutableCode{
			Code:     code,
			Language: language,
		},
	}

	result := NewPartFromExecutableCode(code, language)
	if !reflect.DeepEqual(result, expected) {
		t.Fatalf("expected %v, got %v", expected, result)
	}
}

func TestNewPartFromCodeExecutionResult(t *testing.T) {
	outcome := OutcomeOK
	output := "Execution output"
	expected := &Part{
		CodeExecutionResult: &CodeExecutionResult{
			Outcome: outcome,
			Output:  output,
		},
	}

	result := NewPartFromCodeExecutionResult(outcome, output)
	if !reflect.DeepEqual(result, expected) {
		t.Fatalf("expected %v, got %v", expected, result)
	}
}
