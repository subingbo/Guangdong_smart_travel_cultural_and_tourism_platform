package com.travelgd.dto;

import com.travelgd.model.User;

/** Public shape of a user (never exposes password hash). */
public record UserDto(Long id, String username, String nickname, String avatarUrl, String language) {
    public static UserDto from(User u) {
        return new UserDto(u.getId(), u.getUsername(), u.getNickname(), u.getAvatarUrl(), u.getLanguage());
    }
}
