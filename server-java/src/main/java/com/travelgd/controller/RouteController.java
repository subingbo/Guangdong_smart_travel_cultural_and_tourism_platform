package com.travelgd.controller;

import com.travelgd.common.ApiResponse;
import com.travelgd.dto.RouteGenRequest;
import com.travelgd.model.Route;
import com.travelgd.security.CurrentUser;
import com.travelgd.service.RouteService;
import jakarta.validation.Valid;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.PathVariable;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.RequestBody;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;

import java.util.List;

@RestController
@RequestMapping("/api/routes")
public class RouteController {

    private final RouteService routeService;

    public RouteController(RouteService routeService) {
        this.routeService = routeService;
    }

    @PostMapping("/generate")
    public ApiResponse<Route> generate(@CurrentUser Long userId, @Valid @RequestBody RouteGenRequest req) {
        return ApiResponse.ok(routeService.generate(userId, req));
    }

    @GetMapping
    public ApiResponse<List<Route>> list(@CurrentUser Long userId) {
        return ApiResponse.ok(routeService.listMine(userId));
    }

    @GetMapping("/{id}")
    public ApiResponse<Route> detail(@CurrentUser Long userId, @PathVariable Long id) {
        return ApiResponse.ok(routeService.getMine(userId, id));
    }
}
