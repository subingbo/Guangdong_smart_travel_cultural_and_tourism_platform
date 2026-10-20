package com.travelgd.dto;

import com.travelgd.dto.UserDto;

public record AuthResponse(String token, UserDto user) {}
