package com.travelgd.model;

import lombok.Data;

import java.time.LocalDateTime;
import java.util.List;

@Data
public class TravelBlog {
    private Long id;
    private Long userId;
    private String authorNickname;
    private String authorAvatarUrl;
    private String title;
    private String content;
    private List<String> images;
    private Integer likes;
    private Boolean likedByMe;
    private LocalDateTime createdAt;
    private LocalDateTime updatedAt;
}
