package com.travelgd.model;

import lombok.Data;

import java.time.LocalDateTime;
import java.util.List;

@Data
public class Place {
    private Long id;
    private String nameCn;
    private String nameEn;
    private Double lat;
    private Double lng;
    private List<String> tags;
    private Boolean indoor;
    private Integer duration;
    private String descriptionCn;
    private String descriptionEn;
    private LocalDateTime createdAt;
    private LocalDateTime updatedAt;
}
