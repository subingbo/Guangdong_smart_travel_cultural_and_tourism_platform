package com.travelgd.model;

import lombok.Data;

import java.time.LocalDateTime;

@Data
public class ScenicImage {
    private Long id;
    private String imageUrl;
    private String titleCn;
    private String titleEn;
    private String location;
    private String tag;
    private String descriptionCn;
    private String descriptionEn;
    private Integer sortOrder;
    private LocalDateTime createdAt;
    private LocalDateTime updatedAt;
}
