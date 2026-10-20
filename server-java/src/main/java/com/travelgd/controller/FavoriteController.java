package com.travelgd.controller;

import com.travelgd.common.ApiResponse;
import com.travelgd.dto.FavoriteAddRequest;
import com.travelgd.model.Place;
import com.travelgd.security.CurrentUser;
import com.travelgd.service.FavoriteService;
import jakarta.validation.Valid;
import org.springframework.http.HttpStatus;
import org.springframework.web.bind.annotation.DeleteMapping;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.PathVariable;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.RequestBody;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.ResponseStatus;
import org.springframework.web.bind.annotation.RestController;

import java.util.List;

@RestController
@RequestMapping("/api/favorites")
public class FavoriteController {

    private final FavoriteService favoriteService;

    public FavoriteController(FavoriteService favoriteService) {
        this.favoriteService = favoriteService;
    }

    @GetMapping
    public ApiResponse<List<Place>> list(@CurrentUser Long userId) {
        return ApiResponse.ok(favoriteService.listFavorites(userId));
    }

    @PostMapping
    @ResponseStatus(HttpStatus.CREATED)
    public ApiResponse<Void> add(@CurrentUser Long userId, @Valid @RequestBody FavoriteAddRequest req) {
        favoriteService.addFavorite(userId, req.placeId());
        return ApiResponse.ok();
    }

    @DeleteMapping("/{placeId}")
    public ApiResponse<Void> remove(@CurrentUser Long userId, @PathVariable Long placeId) {
        favoriteService.removeFavorite(userId, placeId);
        return ApiResponse.ok();
    }
}
