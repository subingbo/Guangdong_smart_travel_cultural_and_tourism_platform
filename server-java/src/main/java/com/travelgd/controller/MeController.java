package com.travelgd.controller;

import com.travelgd.common.ApiResponse;
import com.travelgd.dto.StatsDto;
import com.travelgd.dto.UserDto;
import com.travelgd.security.CurrentUser;
import com.travelgd.service.MeService;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.PutMapping;
import org.springframework.web.bind.annotation.RequestBody;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;

import java.util.Map;

@RestController
@RequestMapping("/api/me")
public class MeController {

    private final MeService meService;

    public MeController(MeService meService) {
        this.meService = meService;
    }

    @GetMapping
    public ApiResponse<UserDto> profile(@CurrentUser Long userId) {
        return ApiResponse.ok(meService.getProfile(userId));
    }

    @PutMapping
    public ApiResponse<UserDto> update(@CurrentUser Long userId,
                                       @RequestBody Map<String, String> body) {
        return ApiResponse.ok(meService.updateProfile(
                userId, body.get("nickname"), body.get("avatarUrl"), body.get("language")));
    }

    @GetMapping("/stats")
    public ApiResponse<StatsDto> stats(@CurrentUser Long userId) {
        return ApiResponse.ok(meService.getStats(userId));
    }
}
