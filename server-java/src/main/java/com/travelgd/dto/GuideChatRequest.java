package com.travelgd.dto;

import jakarta.validation.constraints.NotBlank;

public record GuideChatRequest(Long sessionId, @NotBlank String question) {}
