package com.travelgd.dto;

import jakarta.validation.constraints.NotBlank;

public record BlogDraftRequest(@NotBlank String topic) {}
