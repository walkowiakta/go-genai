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

const (
	roleUser = "user"
)

// Contents is satisfied by [Text], [Texts], [PartSlice], [ContentSlice]
//
// Examples:
//
//	Text("Hello")
//	Texts{"Hello", "World"}
//	PartSlice{Text("Hello"), Text("World"), FileData{FileURI: "https://.../bg.jpg", MIMEType: "image/jpeg"}}
//	ContentSlice{Content{Parts: Text("Hello"), Role: roleUser}, Content{Parts: Text("World"), Role: roleUser}}
type Contents interface {
	ToContents() []*Content
}

// part is satisfied by [Text], [Texts], [Part],
// [VideoMetadata], [CodeExecutionResult],
// [ExecutableCode], [FileData], [FunctionCall],
// [FunctionResponse], [InlineData]
type part interface {
	toPart() *Part
}

// Text is a string.
//
// Example usage:
//
//	client.Models.GenerateContent(ctx, "gemini-1.5-flash", genai.Text("Hello"), nil)
type Text string

// ToContents satisfies [Contents]
func (t Text) ToContents() []*Content {
	return []*Content{t.ToContent()}
}

// ToContent satisfies [Contents]
func (t Text) ToContent() *Content {
	return &Content{
		Parts: []*Part{t.toPart()},
		Role:  roleUser,
	}
}

// toPart satisfies [part]
func (t Text) toPart() *Part {
	return &Part{Text: string(t)}
}

// Texts is a list of string and satisfies [Contents] interface.
//
// Example usage:
//
//	client.Models.GenerateContent(ctx, "gemini-1.5-flash", genai.Texts{"Hello", "World"}, nil)
type Texts []string

// ToContents satisfies [Contents]
func (t Texts) ToContents() []*Content {
	var parts []*Part
	for _, text := range t {
		parts = append(parts, Text(text).toPart())
	}
	return []*Content{{
		Parts: parts,
		Role:  roleUser,
	}}
}

// ContentSlice is a list of [Content] struct.
//
// Example usage:
//
//	client.Models.GenerateContent(ctx, "gemini-1.5-flash", genai.ContentSlice{
//	  genai.Content{Parts: genai.Text("Hello"), Role: "user"},
//	  genai.Content{Parts: genai.Text("World"), Role: "user"},
//	}, nil)
type ContentSlice []*Content

// ToContents satisfies [Contents]
func (t ContentSlice) ToContents() []*Content {
	return t
}

// PartSlice is a single Content with multiple part data.
//
// Example usage:
//
//	client.Models.GenerateContent(ctx, "gemini-1.5-flash", genai.PartSlice{
//	  genai.Text("Hello"),
//	  genai.Text("World"),
//	  genai.FileData{FileURI: "https://.../bg.jpg", MIMEType: "image/jpeg"},
//	}, nil)
type PartSlice []part

// Satisfy [part]
func (t PartSlice) toPart() []*Part {
	var parts []*Part
	for _, part := range t {
		parts = append(parts, part.toPart())
	}
	return parts
}

// ToContents satisfies [Contents]
func (t PartSlice) ToContents() []*Content {
	return []*Content{{
		Parts: t.toPart(),
		Role:  roleUser,
	}}
}

// toPart satisfies [part]
func (p Part) toPart() *Part {
	return &p
}

// toPart satisfies [part]
func (p FileData) toPart() *Part {
	return &Part{FileData: &p}
}

// InlineData is an alias for Blob.
type InlineData = Blob

// Satisfy [part]
func (p Blob) toBlob() *Blob {
	return &Blob{Data: p.Data, MIMEType: p.MIMEType}
}

// Satisfy [part]
func (p Blob) toPart() *Part {
	return &Part{InlineData: p.toBlob()}
}

// Satisfy [part]
func (p FunctionResponse) toPart() *Part {
	return &Part{FunctionResponse: &p}
}

// Satisfy [part]
func (p FunctionCall) toPart() *Part {
	return &Part{FunctionCall: &p}
}

// Satisfy [part]
func (p ExecutableCode) toPart() *Part {
	return &Part{ExecutableCode: &p}
}

// Satisfy [part]
func (p CodeExecutionResult) toPart() *Part {
	return &Part{CodeExecutionResult: &p}
}

// Satisfy [part]
func (p VideoMetadata) toPart() *Part {
	return &Part{VideoMetadata: &p}
}
