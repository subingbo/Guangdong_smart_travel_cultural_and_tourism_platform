package com.travelgd.service;

import com.travelgd.common.ApiException;
import com.travelgd.mapper.PlaceMapper;
import com.travelgd.mapper.ScenicImageMapper;
import com.travelgd.model.Place;
import com.travelgd.model.ScenicImage;
import org.springframework.stereotype.Service;

import java.util.List;

@Service
public class CatalogService {

    private final ScenicImageMapper scenicImageMapper;
    private final PlaceMapper placeMapper;

    public CatalogService(ScenicImageMapper scenicImageMapper, PlaceMapper placeMapper) {
        this.scenicImageMapper = scenicImageMapper;
        this.placeMapper = placeMapper;
    }

    public List<ScenicImage> listScenics() {
        return scenicImageMapper.findAll();
    }

    public ScenicImage getScenic(Long id) {
        ScenicImage image = scenicImageMapper.findById(id);
        if (image == null) throw ApiException.notFound("Scenic image not found");
        return image;
    }

    public List<Place> listPlaces(String tag) {
        return placeMapper.findAll(tag == null || tag.isBlank() ? null : tag);
    }

    public Place getPlace(Long id) {
        Place place = placeMapper.findById(id);
        if (place == null) throw ApiException.notFound("Place not found");
        return place;
    }
}
