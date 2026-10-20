package com.travelgd.service.ai;

import com.fasterxml.jackson.databind.JsonNode;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.travelgd.common.ApiException;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.stereotype.Component;

import java.net.URI;
import java.net.http.HttpClient;
import java.net.http.HttpRequest;
import java.net.http.HttpResponse;
import java.time.Duration;
import java.util.List;
import java.util.Map;

/**
 * Thin client for the Alibaba DashScope (Qwen) text-generation API.
 * Mirrors the request shape used by the original cloud function aiService.
 */
@Component
public class QwenClient {

    private final HttpClient http = HttpClient.newBuilder()
            .connectTimeout(Duration.ofSeconds(10))
            .build();
    private final ObjectMapper mapper = new ObjectMapper();

    private final String apiKey;
    private final String chatUrl;

    public QwenClient(@Value("${app.ai.dashscope-api-key}") String apiKey,
                      @Value("${app.ai.chat-url}") String chatUrl) {
        this.apiKey = apiKey;
        this.chatUrl = chatUrl;
    }

    public boolean isEnabled() {
        return apiKey != null && !apiKey.isBlank();
    }

    /** Sends a chat completion request and returns the assistant message text. */
    public String chat(String model, List<Map<String, String>> messages) {
        if (!isEnabled()) {
            throw ApiException.badGateway("AI provider not configured (DASHSCOPE_API_KEY missing)");
        }
        try {
            var body = Map.of(
                    "model", model,
                    "input", Map.of("messages", messages),
                    "parameters", Map.of("result_format", "message"));
            HttpRequest request = HttpRequest.newBuilder()
                    .uri(URI.create(chatUrl))
                    .header("Authorization", "Bearer " + apiKey)
                    .header("Content-Type", "application/json")
                    .POST(HttpRequest.BodyPublishers.ofString(mapper.writeValueAsString(body)))
                    .timeout(Duration.ofSeconds(60))
                    .build();
            HttpResponse<String> response = http.send(request, HttpResponse.BodyHandlers.ofString());
            if (response.statusCode() != 200) {
                throw ApiException.badGateway("AI provider error: HTTP " + response.statusCode());
            }
            JsonNode root = mapper.readTree(response.body());
            JsonNode content = root.path("output").path("choices").get(0).path("message").path("content");
            if (content.isMissingNode()) {
                throw ApiException.badGateway("Unexpected AI response shape");
            }
            return content.asText();
        } catch (ApiException e) {
            throw e;
        } catch (Exception e) {
            throw ApiException.badGateway("AI provider unreachable: " + e.getMessage());
        }
    }

    /** Strips ```json fences the model sometimes adds before JSON.parse. */
    public static String stripFences(String text) {
        return text == null ? "" : text.replace("```json", "").replace("```", "").trim();
    }
}
