package com.travelgd.service;

import com.fasterxml.jackson.databind.JsonNode;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.travelgd.common.ApiException;
import com.travelgd.dto.RouteGenRequest;
import com.travelgd.mapper.RouteMapper;
import com.travelgd.model.Route;
import com.travelgd.service.ai.QwenClient;
import com.travelgd.service.ai.TencentRouteClient;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.stereotype.Service;

import java.util.ArrayList;
import java.util.List;
import java.util.Map;

@Service
public class RouteService {

    private static final Logger log = LoggerFactory.getLogger(RouteService.class);
    private static final ObjectMapper MAPPER = new ObjectMapper();

    private final RouteMapper routeMapper;
    private final QwenClient qwen;
    private final TencentRouteClient routeClient;

    public RouteService(RouteMapper routeMapper, QwenClient qwen, TencentRouteClient routeClient) {
        this.routeMapper = routeMapper;
        this.qwen = qwen;
        this.routeClient = routeClient;
    }

    public Route generate(Long userId, RouteGenRequest req) {
        String pace = normalize(req.pace(), List.of("relaxed", "balanced", "packed"), "balanced");
        String mode = normalize(req.mode(), List.of("driving", "walking", "bicycling", "transit"), "driving");
        List<String> interests = req.interests() == null ? List.of("sightseeing") : req.interests();

        String systemPrompt = """
                You are an expert travel planner for Guangzhou. Generate a detailed %d-day itinerary \
                with a %s pace, focusing on these interests: %s.
                You MUST return a VALID JSON object matching the structure: \
                { "routeName": string, "duration": string, "totalDistance": string, \
                "waypoints": [{"day": number, "name": string, "nameEn": string, "latitude": number, \
                "longitude": number, "description": string, "stayTime": string}] }.
                Ensure coordinates are accurate for Guangzhou. Do NOT output markdown."""
                .formatted(req.days(), pace, String.join(", ", interests));

        String raw = qwen.chat("qwen-plus", List.of(
                Map.of("role", "system", "content", systemPrompt),
                Map.of("role", "user", "content", "Plan my trip in Guangzhou.")));

        Route route = parseRoute(raw, req, pace, mode);

        // Enrich with a real road polyline when the Tencent key is configured; otherwise straight lines.
        TencentRouteClient.RouteResult result = routeClient.fetchPolyline(route.getWaypoints(), mode);
        if (result != null && result.points().size() > 1) {
            route.setPolyline(result.points());
        } else {
            route.setPolyline(straightLines(route.getWaypoints()));
        }
        if (route.getTotalDistance() == null || route.getTotalDistance().isBlank()) {
            route.setTotalDistance(estimateDistance(route.getPolyline()));
        }

        route.setUserId(userId);
        routeMapper.insert(route);
        return routeMapper.findById(route.getId(), userId);
    }

    private Route parseRoute(String raw, RouteGenRequest req, String pace, String mode) {
        try {
            JsonNode node = MAPPER.readTree(QwenClient.stripFences(raw));
            Route route = new Route();
            route.setRouteName(node.path("routeName").asText("Guangzhou Trip"));
            route.setDurationDays(req.days());
            route.setTotalDistance(node.path("totalDistance").asText(""));
            route.setMode(mode);
            route.setPace(pace);
            route.setInterests(req.interests() == null ? List.of() : req.interests());

            List<Route.Waypoint> waypoints = new ArrayList<>();
            for (JsonNode w : node.path("waypoints")) {
                Route.Waypoint wp = new Route.Waypoint();
                wp.setName(w.path("name").asText(""));
                wp.setNameEn(w.path("nameEn").asText(w.path("name").asText("")));
                wp.setLat(w.path("latitude").asDouble());
                wp.setLng(w.path("longitude").asDouble());
                wp.setStayTime(w.path("stayTime").asText(""));
                wp.setDescription(w.path("description").asText(""));
                waypoints.add(wp);
            }
            if (waypoints.isEmpty()) throw ApiException.badGateway("AI returned no waypoints");
            route.setWaypoints(waypoints);
            return route;
        } catch (ApiException e) {
            throw e;
        } catch (Exception e) {
            log.warn("Failed to parse AI route response", e);
            throw ApiException.badGateway("Could not parse AI route response");
        }
    }

    private List<Route.Coordinate> straightLines(List<Route.Waypoint> waypoints) {
        List<Route.Coordinate> points = new ArrayList<>();
        for (Route.Waypoint wp : waypoints) {
            Route.Coordinate c = new Route.Coordinate();
            c.setLat(wp.getLat());
            c.setLng(wp.getLng());
            points.add(c);
        }
        return points;
    }

    private String estimateDistance(List<Route.Coordinate> polyline) {
        double total = 0;
        for (int i = 1; i < polyline.size(); i++) {
            total += haversine(polyline.get(i - 1), polyline.get(i));
        }
        return total >= 1000 ? String.format("%.1fkm", total / 1000) : Math.round(total) + "m";
    }

    private double haversine(Route.Coordinate a, Route.Coordinate b) {
        double R = 6371000;
        double dLat = Math.toRadians(b.getLat() - a.getLat());
        double dLng = Math.toRadians(b.getLng() - a.getLng());
        double h = Math.pow(Math.sin(dLat / 2), 2)
                + Math.cos(Math.toRadians(a.getLat())) * Math.cos(Math.toRadians(b.getLat()))
                * Math.pow(Math.sin(dLng / 2), 2);
        return 2 * R * Math.asin(Math.sqrt(h));
    }

    private String normalize(String value, List<String> allowed, String fallback) {
        if (value == null || value.isBlank()) return fallback;
        String v = value.toLowerCase();
        return allowed.contains(v) ? v : fallback;
    }

    public List<Route> listMine(Long userId) {
        return routeMapper.findByUserId(userId);
    }

    public Route getMine(Long userId, Long id) {
        Route route = routeMapper.findById(id, userId);
        if (route == null) throw ApiException.notFound("Route not found");
        return route;
    }
}
