package com.travelgd.service.ai;

import com.fasterxml.jackson.databind.JsonNode;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.travelgd.model.Route;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.stereotype.Component;

import java.net.URI;
import java.net.http.HttpClient;
import java.net.http.HttpRequest;
import java.net.http.HttpResponse;
import java.time.Duration;
import java.util.ArrayList;
import java.util.List;

/**
 * Optional Tencent Maps route planner. When TENCENT_MAP_API_KEY is absent the
 * caller falls back to straight lines between waypoints, matching the original
 * cloud function behaviour.
 */
@Component
public class TencentRouteClient {

    private final HttpClient http = HttpClient.newBuilder()
            .connectTimeout(Duration.ofSeconds(10))
            .build();
    private final ObjectMapper mapper = new ObjectMapper();

    private final String apiKey;
    private final String routeUrl;

    public TencentRouteClient(@Value("${app.map.tencent-key}") String apiKey,
                              @Value("${app.map.route-url}") String routeUrl) {
        this.apiKey = apiKey;
        this.routeUrl = routeUrl;
    }

    public boolean isEnabled() {
        return apiKey != null && !apiKey.isBlank();
    }

    /**
     * Fetches a driving/walking polyline through the waypoints.
     * Returns null when the key is missing or the API fails (caller falls back).
     */
    public RouteResult fetchPolyline(List<Route.Waypoint> waypoints, String mode) {
        if (!isEnabled() || waypoints == null || waypoints.size() < 2) {
            return null;
        }
        try {
            String from = waypoints.get(0).getLat() + "," + waypoints.get(0).getLng();
            String to = waypoints.get(waypoints.size() - 1).getLat() + "," + waypoints.get(waypoints.size() - 1).getLng();
            StringBuilder via = new StringBuilder();
            for (int i = 1; i < waypoints.size() - 1; i++) {
                if (via.length() > 0) via.append(";");
                via.append(waypoints.get(i).getLat()).append(",").append(waypoints.get(i).getLng());
            }
            String endpoint = switch (mode) {
                case "walking" -> routeUrl + "/walking/" ;
                case "bicycling" -> routeUrl + "/bicycling/";
                case "transit" -> routeUrl + "/transit/";
                default -> routeUrl + "/driving/";
            };
            String url = endpoint + "?from=" + from + "&to=" + to
                    + (via.length() > 0 ? "&waypoints=" + via : "") + "&key=" + apiKey;
            HttpRequest request = HttpRequest.newBuilder().uri(URI.create(url))
                    .timeout(Duration.ofSeconds(15)).GET().build();
            HttpResponse<String> response = http.send(request, HttpResponse.BodyHandlers.ofString());
            JsonNode root = mapper.readTree(response.body());
            JsonNode polyline = root.path("result").path("routes").get(0).path("polyline");
            if (!polyline.isArray() || polyline.size() < 4) {
                return null;
            }
            double[] raw = new double[polyline.size()];
            for (int i = 0; i < polyline.size(); i++) raw[i] = polyline.get(i).asDouble();
            return new RouteResult(decodePolyline(raw));
        } catch (Exception e) {
            return null;
        }
    }

    /** Decodes the compressed polyline format (first point absolute, rest deltas * 1e6). */
    private List<Route.Coordinate> decodePolyline(double[] pl) {
        if (pl.length < 2) return new ArrayList<>();
        if (Math.abs(pl[0]) > 180) {
            pl[0] = pl[0] / 1000000;
            pl[1] = pl[1] / 1000000;
        }
        for (int i = 2; i < pl.length; i++) {
            pl[i] = Math.round((pl[i - 2] + pl[i] / 1000000) * 1000000.0) / 1000000.0;
        }
        List<Route.Coordinate> points = new ArrayList<>();
        for (int i = 0; i + 1 < pl.length; i += 2) {
            Route.Coordinate c = new Route.Coordinate();
            c.setLat(pl[i]);
            c.setLng(pl[i + 1]);
            points.add(c);
        }
        return points;
    }

    public record RouteResult(List<Route.Coordinate> points) {}
}
