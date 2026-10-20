package com.travelgd.dto;

import jakarta.validation.constraints.NotBlank;

import java.util.List;

public record BlogCreateRequest(@NotBlank String title, @NotBlank String content, List<String> images) {}
