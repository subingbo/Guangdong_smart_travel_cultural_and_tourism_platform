package com.travelgd.model;

import lombok.Data;

import java.time.LocalDateTime;
import java.util.List;

@Data
public class Route {
    private Long id;
    private Long userId;
    private String routeName;
    private Integer durationDays;
    private String totalDistance;
    private String mode;
    private String pace;
    private List<String> interests;
    private List<Waypoint> waypoints;
    private List<Coordinate> polyline;
    private LocalDateTime createdAt;
    private LocalDateTime updatedAt;

    @Data
    public static class Waypoint {
        private String name;
        private String nameEn;
        private Double lat;
        private Double lng;
        private String stayTime;
        private String description;
    }

    @Data
    public static class Coordinate {
        private Double lat;
        private Double lng;
    }
}
