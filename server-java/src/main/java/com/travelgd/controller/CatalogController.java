package com.travelgd.controller;

import com.travelgd.common.ApiResponse;
import com.travelgd.model.Place;
import com.travelgd.model.ScenicImage;
import com.travelgd.service.CatalogService;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.PathVariable;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RequestParam;
import org.springframework.web.bind.annotation.RestController;

import java.util.List;

@RestController
@RequestMapping("/api")
public class CatalogController {

    private final CatalogService catalogService;

    public CatalogController(CatalogService catalogService) {
        this.catalogService = catalogService;
    }

    @GetMapping("/scenics")
    public ApiResponse<List<ScenicImage>> scenics() {
        return ApiResponse.ok(catalogService.listScenics());
    }

    @GetMapping("/scenics/{id}")
    public ApiResponse<ScenicImage> scenic(@PathVariable Long id) {
        return ApiResponse.ok(catalogService.getScenic(id));
    }

    @GetMapping("/places")
    public ApiResponse<List<Place>> places(@RequestParam(required = false) String tag) {
        return ApiResponse.ok(catalogService.listPlaces(tag));
    }

    @GetMapping("/places/{id}")
    public ApiResponse<Place> place(@PathVariable Long id) {
        return ApiResponse.ok(catalogService.getPlace(id));
    }
}
