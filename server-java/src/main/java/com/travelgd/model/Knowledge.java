package com.travelgd.model;

import lombok.Data;

import java.time.LocalDateTime;
import java.util.List;

@Data
public class Knowledge {
    private Long id;
    private String titleCn;
    private String titleEn;
    private String contentCn;
    private String contentEn;
    private List<String> tags;
    private LocalDateTime createdAt;
    private LocalDateTime updatedAt;
}
