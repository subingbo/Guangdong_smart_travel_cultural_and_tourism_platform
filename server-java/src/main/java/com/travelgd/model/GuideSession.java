package com.travelgd.model;

import lombok.Data;

import java.time.LocalDateTime;
import java.util.List;

@Data
public class GuideSession {
    private Long id;
    private Long userId;
    private String title;
    private List<GuideMessage> messages;
    private LocalDateTime createdAt;
    private LocalDateTime updatedAt;

    @Data
    public static class GuideMessage {
        private Long id;
        private Long sessionId;
        private String role;
        private String content;
        private LocalDateTime createdAt;
    }
}
