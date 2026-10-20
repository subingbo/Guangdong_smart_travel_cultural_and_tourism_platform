package com.travelgd.controller;

import com.travelgd.common.ApiResponse;
import com.travelgd.dto.GuideChatRequest;
import com.travelgd.model.GuideSession;
import com.travelgd.security.CurrentUser;
import com.travelgd.service.GuideService;
import jakarta.validation.Valid;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.PathVariable;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.RequestBody;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;

import java.util.Map;

@RestController
@RequestMapping("/api/guide")
public class GuideController {

    private final GuideService guideService;

    public GuideController(GuideService guideService) {
        this.guideService = guideService;
    }

    @PostMapping("/chat")
    public ApiResponse<Map<String, Object>> chat(@CurrentUser Long userId,
                                                 @Valid @RequestBody GuideChatRequest req) {
        return ApiResponse.ok(guideService.chat(userId, req));
    }

    @GetMapping("/sessions/{id}")
    public ApiResponse<GuideSession> session(@CurrentUser Long userId, @PathVariable Long id) {
        return ApiResponse.ok(guideService.getSession(userId, id));
    }
}
