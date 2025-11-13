package service

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"developer-portal-backend/internal/logger"

	"github.com/gin-gonic/gin"
)

// ChatInferenceStream handles streaming chat inference using Server-Sent Events
func (s *AICoreService) ChatInferenceStream(c *gin.Context, req *AICoreInferenceRequest, writer gin.ResponseWriter) error {
	// Get all deployments accessible to the user
	deploymentsResp, err := s.GetDeployments(c)
	if err != nil {
		return fmt.Errorf("failed to get deployments: %w", err)
	}

	// Find the deployment by ID across all teams
	var targetDeployment *AICoreDeployment
	var targetTeamName string

	for _, teamDeployments := range deploymentsResp.Deployments {
		for _, deployment := range teamDeployments.Deployments {
			if deployment.ID == req.DeploymentID {
				targetDeployment = &deployment
				targetTeamName = teamDeployments.Team
				break
			}
		}
		if targetDeployment != nil {
			break
		}
	}

	if targetDeployment == nil {
		return fmt.Errorf("deployment %s not found or user does not have access to it", req.DeploymentID)
	}

	if targetDeployment.DeploymentURL == "" {
		return fmt.Errorf("deployment URL not available for deployment %s", req.DeploymentID)
	}

	// Get credentials and token for the team that owns this deployment
	credentials, err := s.getCredentialsForTeam(targetTeamName)
	if err != nil {
		return fmt.Errorf("failed to get credentials for team %s: %w", targetTeamName, err)
	}

	accessToken, err := s.getAccessToken(credentials)
	if err != nil {
		return fmt.Errorf("failed to get access token: %w", err)
	}

	// Determine model type (same logic as ChatInference)
	isOrchestration := false
	isGPTModel := false
	isGeminiModel := false
	modelName := ""

	// Check if this is orchestration based on scenario ID
	if strings.Contains(strings.ToLower(targetDeployment.ScenarioID), "orchestration") {
		isOrchestration = true
	}

	// Extract model name and check model type
	if extractedName := extractModelNameFromDetails(targetDeployment.Details); extractedName != "" {
		modelName = extractedName
		lowerName := strings.ToLower(extractedName)
		if strings.Contains(lowerName, "gpt") || strings.Contains(lowerName, "o1") ||
			strings.Contains(lowerName, "o3") || strings.Contains(lowerName, "openai") {
			isGPTModel = true
		} else if strings.Contains(lowerName, "gemini") {
			isGeminiModel = true
		}
	}

	// Trim messages to fit within model context limits
	contextLimit := getModelContextLimit(modelName)
	req.Messages = trimMessagesToContextLimit(req.Messages, contextLimit)

	// Build inference payload (reuse logic from ChatInference)
	var inferencePayload map[string]interface{}
	var inferenceURL string

	// Helper function to extract text content from message
	getMessageText := func(msg AICoreInferenceMessage) string {
		if str, ok := msg.Content.(string); ok {
			return str
		}
		if contentArr, ok := msg.Content.([]interface{}); ok {
			for _, part := range contentArr {
				if partMap, ok := part.(map[string]interface{}); ok {
					if partMap["type"] == "text" {
						if text, ok := partMap["text"].(string); ok {
							return text
						}
					}
				}
			}
		}
		return ""
	}

	if isGeminiModel {
		// Gemini models use /models/<model>:streamGenerateContent endpoint
		var parts []map[string]interface{}

		for _, msg := range req.Messages {
			if msg.Role == "system" {
				parts = append(parts, map[string]interface{}{
					"text": fmt.Sprintf("[System]: %s", getMessageText(msg)),
				})
				continue
			}

			// Handle multimodal content (text + images)
			if contentArr, ok := msg.Content.([]interface{}); ok {
				for _, part := range contentArr {
					if partMap, ok := part.(map[string]interface{}); ok {
						partType := partMap["type"].(string)
						if partType == "text" {
							parts = append(parts, map[string]interface{}{
								"text": partMap["text"],
							})
						} else if partType == "image_url" {
							if imageURL, ok := partMap["image_url"].(map[string]interface{}); ok {
								parts = append(parts, map[string]interface{}{
									"fileData": map[string]interface{}{
										"mimeType": "image/png",
										"fileUri":  imageURL["url"],
									},
								})
							}
						}
					}
				}
			} else {
				parts = append(parts, map[string]interface{}{
					"text": getMessageText(msg),
				})
			}
		}

		inferencePayload = map[string]interface{}{
			"contents": map[string]interface{}{
				"role":  "user",
				"parts": parts,
			},
		}

		if req.MaxTokens > 0 || req.Temperature > 0 {
			generationConfig := make(map[string]interface{})
			if req.MaxTokens > 0 {
				generationConfig["maxOutputTokens"] = req.MaxTokens
			}
			if req.Temperature > 0 {
				generationConfig["temperature"] = req.Temperature
			}
			inferencePayload["generation_config"] = generationConfig
		}

		inferenceURL = fmt.Sprintf("%s/models/%s:streamGenerateContent", targetDeployment.DeploymentURL, modelName)
	} else if isOrchestration {
		// Orchestration models use orchestration config
		if modelName == "" {
			modelName = "gpt-4o-mini"
		}

		templateMessages := make([]map[string]interface{}, 0)
		for _, msg := range req.Messages {
			templateMessages = append(templateMessages, map[string]interface{}{
				"role":    msg.Role,
				"content": msg.Content,
			})
		}

		modelParams := map[string]interface{}{
			"frequency_penalty": 0,
			"presence_penalty":  0,
		}

		if req.MaxTokens > 0 {
			modelParams["max_tokens"] = req.MaxTokens
		} else {
			modelParams["max_tokens"] = 1000
		}

		if req.Temperature > 0 {
			modelParams["temperature"] = req.Temperature
		} else {
			modelParams["temperature"] = 0.7
		}

		inferencePayload = map[string]interface{}{
			"orchestration_config": map[string]interface{}{
				"module_configurations": map[string]interface{}{
					"templating_module_config": map[string]interface{}{
						"template": templateMessages,
					},
					"llm_module_config": map[string]interface{}{
						"model_name":    modelName,
						"model_params":  modelParams,
						"model_version": "latest",
					},
				},
			},
			"input_params": map[string]interface{}{},
			"stream":       true, // Enable streaming
		}

		inferenceURL = fmt.Sprintf("%s/completion", targetDeployment.DeploymentURL)
	} else if isGPTModel {
		// Build messages array
		messages := make([]map[string]interface{}, 0)
		for _, msg := range req.Messages {
			message := map[string]interface{}{
				"role": msg.Role,
			}

			if contentArr, ok := msg.Content.([]interface{}); ok {
				message["content"] = contentArr
			} else {
				message["content"] = msg.Content
			}

			messages = append(messages, message)
		}

		inferencePayload = map[string]interface{}{
			"messages": messages,
			"stream":   true, // Enable streaming
		}

		apiVersion := getGPTAPIVersion(modelName)
		isReasoningModel := strings.Contains(strings.ToLower(modelName), "o1") ||
			strings.Contains(strings.ToLower(modelName), "o3-mini") ||
			strings.Contains(strings.ToLower(modelName), "gpt-5")

		if !isReasoningModel {
			if req.MaxTokens > 0 {
				inferencePayload["max_tokens"] = req.MaxTokens
			} else {
				inferencePayload["max_tokens"] = 1000
			}
			if req.Temperature > 0 {
				inferencePayload["temperature"] = req.Temperature
			} else {
				inferencePayload["temperature"] = 0.7
			}
			if req.TopP > 0 {
				inferencePayload["top_p"] = req.TopP
			}
		}

		inferenceURL = fmt.Sprintf("%s/chat/completions?api-version=%s", targetDeployment.DeploymentURL, apiVersion)
	} else {
		// Anthropic Claude models (default if not GPT, Gemini, or Orchestration)
		var systemPrompt string
		var userMessages []map[string]string

		for _, msg := range req.Messages {
			if msg.Role == "system" {
				systemPrompt = getMessageText(msg)
			} else {
				userMessages = append(userMessages, map[string]string{
					"role":    msg.Role,
					"content": getMessageText(msg),
				})
			}
		}

		inferencePayload = map[string]interface{}{
			"anthropic_version": "bedrock-2023-05-31",
			"messages":          userMessages,
		}

		if systemPrompt != "" {
			inferencePayload["system"] = systemPrompt
		}

		if req.MaxTokens > 0 {
			inferencePayload["max_tokens"] = req.MaxTokens
		} else {
			inferencePayload["max_tokens"] = 1000
		}
		if req.Temperature > 0 {
			inferencePayload["temperature"] = req.Temperature
		} else {
			inferencePayload["temperature"] = 0.7
		}
		if req.TopP > 0 {
			inferencePayload["top_p"] = req.TopP
		}

		// SAP AI Core Claude streaming uses invoke-with-response-stream endpoint
		inferenceURL = fmt.Sprintf("%s/invoke-with-response-stream", targetDeployment.DeploymentURL)
	}

	// Make the streaming request
	resp, err := s.makeAICoreRequest("POST", inferenceURL, accessToken, credentials.ResourceGroup, inferencePayload)
	if err != nil {
		return fmt.Errorf("failed to make inference request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("inference request failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Stream the response using SSE
	flusher, ok := writer.(http.Flusher)
	if !ok {
		return fmt.Errorf("streaming not supported")
	}

	log := logger.FromGinContext(c)

	// Read the streaming response line by line
	reader := bufio.NewReader(resp.Body)
	for {
		line, err := reader.ReadBytes('\n')
		if err != nil {
			if err == io.EOF {
				// Send a final [DONE] message
				writer.Write([]byte("data: [DONE]\n\n"))
				flusher.Flush()
				break
			}
			log.Errorf("Error reading stream: %v", err)
			return fmt.Errorf("error reading stream: %w", err)
		}

		line = bytes.TrimSpace(line)
		if len(line) == 0 {
			continue
		}

		// SSE format: "data: {...}"
		if bytes.HasPrefix(line, []byte("data: ")) {
			data := bytes.TrimPrefix(line, []byte("data: "))

			// Check for [DONE] marker
			if bytes.Equal(data, []byte("[DONE]")) {
				writer.Write([]byte("data: [DONE]\n\n"))
				flusher.Flush()
				break
			}

			// Parse the chunk
			var chunk map[string]interface{}
			if err := json.Unmarshal(data, &chunk); err != nil {
				log.Warnf("Failed to parse chunk: %v", err)
				continue
			}

			// Convert Gemini format to OpenAI format if needed
			if isGeminiModel {
				// Gemini streaming response format:
				// {"candidates": [{"content": {"parts": [{"text": "..."}]}}]}
				if candidates, ok := chunk["candidates"].([]interface{}); ok && len(candidates) > 0 {
					if candidate, ok := candidates[0].(map[string]interface{}); ok {
						if content, ok := candidate["content"].(map[string]interface{}); ok {
							if parts, ok := content["parts"].([]interface{}); ok && len(parts) > 0 {
								if part, ok := parts[0].(map[string]interface{}); ok {
									if text, ok := part["text"].(string); ok {
										// Convert to OpenAI streaming format
										openAIChunk := map[string]interface{}{
											"id":      fmt.Sprintf("gemini-%d", time.Now().UnixNano()),
											"object":  "chat.completion.chunk",
											"created": time.Now().Unix(),
											"model":   modelName,
											"choices": []map[string]interface{}{
												{
													"index": 0,
													"delta": map[string]interface{}{
														"content": text,
													},
													"finish_reason": nil,
												},
											},
										}

										// Check for finish reason
										if finishReason, ok := candidate["finishReason"].(string); ok && finishReason != "" {
											openAIChunk["choices"].([]map[string]interface{})[0]["finish_reason"] = strings.ToLower(finishReason)
										}

										convertedData, _ := json.Marshal(openAIChunk)
										writer.Write([]byte(fmt.Sprintf("data: %s\n\n", convertedData)))
										flusher.Flush()
										continue
									}
								}
							}
						}
					}
				}
			}

			// Forward the chunk as-is (OpenAI/Anthropic/Orchestration format)
			writer.Write([]byte(fmt.Sprintf("data: %s\n\n", data)))
			flusher.Flush()
		}
	}

	return nil
}
