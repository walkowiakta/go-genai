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

// Package main contains the sample code for the GenerateContent API.
package main

/*
# For Vertex AI API
export GOOGLE_GENAI_USE_VERTEXAI=true
export GOOGLE_CLOUD_PROJECT=cloud-llm-preview1
export GOOGLE_CLOUD_LOCATION=us-central1

# For Gemini AI API
export GOOGLE_GENAI_USE_VERTEXAI=false
export GOOGLE_API_KEY={YOUR_API_KEY}

go run samples/generate_content.go --model=gemini-1.5-pro-002
*/

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"

	genai "google.golang.org/genai"
)

var model = flag.String("model", "gemini-1.5-pro-002", "the model name, e.g. gemini-1.5-pro-002")

func generateContent(ctx context.Context) {
	client, err := genai.NewClient(ctx, nil)
	if err != nil {
		log.Fatal(err)
	}
	if client.ClientConfig.Backend == genai.BackendVertexAI {
		fmt.Println("Calling VertexAI.GenerateContent API...")
	} else {
		fmt.Println("Calling GeminiAI.GenerateContent API...")
	}
	// No configs are being used in this sample, explicitly set it to nil for clarity.
	var config *genai.GenerateContentConfig = nil
	// Call the GenerateContent method.
	result, err := client.Models.GenerateContent(ctx, *model, genai.Text("What is your name?"), config)
	if err != nil {
		log.Fatal(err)
	}
	// Marshal the result to JSON and pretty-print it to a byte array.
	response, err := json.MarshalIndent(*result, "", "  ")
	if err != nil {
		log.Fatal(err)
	}
	// Log the output.
	fmt.Println(string(response))
}

func main() {
	ctx := context.Background()
	flag.Parse()
	generateContent(ctx)
}
