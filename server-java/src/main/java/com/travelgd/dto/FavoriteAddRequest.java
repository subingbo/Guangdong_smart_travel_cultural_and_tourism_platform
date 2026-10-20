package com.travelgd.dto;

import jakarta.validation.constraints.NotNull;

public record FavoriteAddRequest(@NotNull Long placeId) {}
