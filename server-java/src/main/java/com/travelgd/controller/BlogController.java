package com.travelgd.controller;

import com.travelgd.common.ApiResponse;
import com.travelgd.dto.BlogCreateRequest;
import com.travelgd.dto.BlogDraftRequest;
import com.travelgd.model.TravelBlog;
import com.travelgd.security.CurrentUser;
import com.travelgd.security.JwtService;
import com.travelgd.service.BlogService;
import jakarta.validation.Valid;
import org.springframework.http.HttpStatus;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.PathVariable;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.RequestBody;
import org.springframework.web.bind.annotation.RequestHeader;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.ResponseStatus;
import org.springframework.web.bind.annotation.RestController;

import java.util.List;
import java.util.Map;

/**
 * Feed and detail are public: the viewer id is OPTIONAL and anonymous visitors
 * simply get likedByMe=false. Publish/draft require a token via @CurrentUser.
 */
@RestController
@RequestMapping("/api/blogs")
public class BlogController {

    private final BlogService blogService;
    private final JwtService jwtService;

    public BlogController(BlogService blogService, JwtService jwtService) {
        this.blogService = blogService;
        this.jwtService = jwtService;
    }

    @GetMapping
    public ApiResponse<List<TravelBlog>> feed(
            @RequestHeader(value = "Authorization", required = false) String auth) {
        return ApiResponse.ok(blogService.listFeed(viewerId(auth)));
    }

    @GetMapping("/{id}")
    public ApiResponse<TravelBlog> detail(@PathVariable Long id,
            @RequestHeader(value = "Authorization", required = false) String auth) {
        return ApiResponse.ok(blogService.getBlog(id, viewerId(auth)));
    }

    @PostMapping
    @ResponseStatus(HttpStatus.CREATED)
    public ApiResponse<TravelBlog> publish(@CurrentUser Long userId,
                                           @Valid @RequestBody BlogCreateRequest req) {
        return ApiResponse.ok(blogService.publish(userId, req));
    }

    @PostMapping("/draft")
    public ApiResponse<Map<String, String>> draft(@CurrentUser Long userId,
                                                  @Valid @RequestBody BlogDraftRequest req) {
        return ApiResponse.ok(blogService.draft(userId, req.topic()));
    }

    @PostMapping("/{id}/like")
    public ApiResponse<Map<String, Integer>> like(@PathVariable Long id) {
        return ApiResponse.ok(Map.of("likes", blogService.like(id)));
    }

    private Long viewerId(String auth) {
        if (auth != null && auth.startsWith("Bearer ")) {
            return jwtService.parseUserId(auth.substring(7));
        }
        return null;
    }
}
