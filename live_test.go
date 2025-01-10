package genai

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/gorilla/websocket"
)

func TestLiveConnect(t *testing.T) {
	ctx := context.Background()
	const model = "test-model"

	tests := []struct {
		desc        string
		backend     Backend
		apiKey      string
		config      *LiveConnectConfig
		requestBody string
		wantErr     bool
	}{
		{
			desc:        "successful connection mldev",
			backend:     BackendGoogleAI,
			apiKey:      "test-api-key",
			requestBody: `{"setup":{"model":"test-model"}}`,
			wantErr:     false,
		},
		{
			desc:    "successful connection with config mldev",
			backend: BackendGoogleAI,
			apiKey:  "test-api-key",
			config: &LiveConnectConfig{
				GenerationConfig:  &GenerationConfig{Temperature: Ptr(0.5)},
				SystemInstruction: &Content{Parts: []*Part{{Text: "test instruction"}}},
				Tools:             []*Tool{{GoogleSearch: &GoogleSearch{}}},
			},
			requestBody: `{"setup":{"generationConfig":{"temperature":0.5},"model":"test-model","systemInstruction":{"parts":[{"text":"test instruction"}]},"tools":[{"googleSearch":{}}]}}`,
			wantErr:     false,
		},
		// TODO(b/365983028): Add Vertex AI tests
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			var upgrader = websocket.Upgrader{}
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				conn, _ := upgrader.Upgrade(w, r, nil)
				defer conn.Close()

				mt, message, err := conn.ReadMessage()
				if err != nil {
					if tt.wantErr {
						return
					}
					t.Fatalf("ReadMessage: %v", err)
				}
				if diff := cmp.Diff(string(message), tt.requestBody); diff != "" {
					t.Errorf("Request message mismatch (-want +got):\n%s", diff)
				}

				response := &LiveServerMessage{}
				if err := json.Unmarshal([]byte(`{"setupComplete":{}}`), response); err != nil {
					t.Fatalf("Unmarshal: %v", err)
				}
				responseBytes, err := json.Marshal(response)
				if err != nil {
					t.Fatalf("Marshal: %v", err)
				}

				conn.WriteMessage(mt, responseBytes)
			}))
			defer ts.Close()

			if tt.backend == BackendVertexAI {
				return
			}
			client, err := NewClient(ctx, &ClientConfig{
				baseURL:    strings.Replace(ts.URL, "http", "ws", 1),
				HTTPClient: ts.Client(),
				Backend:    tt.backend,
				APIKey:     tt.apiKey,
			})
			if err != nil {
				t.Fatalf("NewClient failed: %v", err)
			}
			session, err := client.Live.Connect(model, tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("Connect() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				// Validate the session setup response if connection is successful
				message, err := session.Receive()

				if err != nil {
					t.Errorf("Receive() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if diff := cmp.Diff(message.SetupComplete, &LiveServerSetupComplete{}); diff != "" {
					t.Errorf("session setup mismatch (-want +got):\n%s", diff)
				}

			}
		})
	}
}

func TestLiveSendAndReceive(t *testing.T) {
	ctx := context.Background()
	ts := setupTestWebsocketServer(t, []string{`{"setup":{"model":"test-model"}}`, `{"clientContent":{"turns":[{"parts":[{"text":"client test message"}],"role":"user"}]}}`}, []string{`"setupComplete":{}`, `{"serverContent":{"modelTurn":{"parts":[{"text":"server test message"}],"role":"user"}}}`})

	defer ts.Close()
	client := fakeLiveClient(ctx, t, ts)
	session, err := client.Live.Connect("test-model", &LiveConnectConfig{})
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}
	defer session.Close()
	// Discard the initial setup message.
	_, _ = session.Receive()

	// Construct a test message
	clientMessage := &LiveClientMessage{
		ClientContent: &LiveClientContent{Turns: Text("client test message")},
	}

	// Test sending the message
	err = session.Send(clientMessage)
	if err != nil {
		t.Errorf("Send failed : %v", err)
	}

	// Construct the expected response
	serverMessage := &LiveServerMessage{ServerContent: &LiveServerContent{ModelTurn: Text("server test message")[0]}}
	// Test receiving the response
	gotMessage, err := session.Receive()
	if err != nil {
		t.Errorf("Receive failed: %v", err)
	}
	if diff := cmp.Diff(gotMessage, serverMessage); diff != "" {
		t.Errorf("Response message mismatch (-want +got):\n%s", diff)
	}
}

// Helper function to set up a test websocket server.
func setupTestWebsocketServer(t *testing.T, wantRequestBodySlice []string, responseBodySlice []string) *httptest.Server {
	t.Helper()

	var upgrader = websocket.Upgrader{}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, _ := upgrader.Upgrade(w, r, nil)
		defer conn.Close()

		index := 0

		for {
			mt, message, err := conn.ReadMessage()
			if err != nil {
				t.Logf("read error: %v", err)
				break
			}
			var clientMessage = &LiveClientMessage{}
			if err := json.Unmarshal(message, clientMessage); err != nil {
				t.Fatalf("unmarshal error: %v", err)
			}
			if diff := cmp.Diff(string(message), wantRequestBodySlice[index]); diff != "" {
				t.Errorf("Request message mismatch (-want +got):\n%s", diff)
			}
			err = conn.WriteMessage(mt, []byte(responseBodySlice[index]))
			index++
			if err != nil {
				t.Logf("write error: %v", err)
				break
			}
		}
	}))

	return ts
}

// Helper function to create a fake client for testing.
func fakeLiveClient(ctx context.Context, t *testing.T, server *httptest.Server) *Client {
	t.Helper()
	client, err := NewClient(ctx, &ClientConfig{
		baseURL:    strings.Replace(server.URL, "http", "ws", 1),
		HTTPClient: server.Client(),
		Backend:    BackendGoogleAI,
		APIKey:     "test-api-key",
	})
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}
	return client
}
