package com.travelgd.dto;

import jakarta.validation.constraints.Max;
import jakarta.validation.constraints.Min;
import jakarta.validation.constraints.NotNull;

import java.util.List;

public record RouteGenRequest(
        @NotNull @Min(1) @Max(7) Integer days,
        List<String> interests,
        String pace,
        String mode) {}
