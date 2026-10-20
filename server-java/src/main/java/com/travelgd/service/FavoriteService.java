package com.travelgd.service;

import com.travelgd.common.ApiException;
import com.travelgd.mapper.FavoriteMapper;
import com.travelgd.model.Place;
import org.springframework.stereotype.Service;

import java.util.List;

@Service
public class FavoriteService {

    private final FavoriteMapper favoriteMapper;
    private final CatalogService catalogService;

    public FavoriteService(FavoriteMapper favoriteMapper, CatalogService catalogService) {
        this.favoriteMapper = favoriteMapper;
        this.catalogService = catalogService;
    }

    public List<Place> listFavorites(Long userId) {
        return favoriteMapper.findPlacesByUserId(userId);
    }

    public void addFavorite(Long userId, Long placeId) {
        catalogService.getPlace(placeId); // 404 when place does not exist
        if (favoriteMapper.exists(userId, placeId)) {
            throw ApiException.conflict("Place already favorited");
        }
        favoriteMapper.insert(userId, placeId);
    }

    public void removeFavorite(Long userId, Long placeId) {
        favoriteMapper.delete(userId, placeId);
    }
}
