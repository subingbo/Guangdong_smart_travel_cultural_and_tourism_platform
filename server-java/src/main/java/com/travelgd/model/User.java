package com.travelgd.model;

import lombok.Data;

import java.time.LocalDateTime;

@Data
public class User {
    private Long id;
    private String username;
    private String passwordHash;
    private String nickname;
    private String avatarUrl;
    private String language;
    private LocalDateTime createdAt;
    private LocalDateTime updatedAt;
}
