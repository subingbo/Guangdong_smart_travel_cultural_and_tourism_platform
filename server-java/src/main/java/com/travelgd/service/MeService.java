package com.travelgd.service;

import com.travelgd.common.ApiException;
import com.travelgd.dto.StatsDto;
import com.travelgd.dto.UserDto;
import com.travelgd.mapper.FavoriteMapper;
import com.travelgd.mapper.RouteMapper;
import com.travelgd.mapper.TravelBlogMapper;
import com.travelgd.mapper.UserMapper;
import com.travelgd.model.User;
import org.springframework.stereotype.Service;

import java.util.Locale;

@Service
public class MeService {

    private final AuthService authService;
    private final UserMapper userMapper;
    private final FavoriteMapper favoriteMapper;
    private final RouteMapper routeMapper;
    private final TravelBlogMapper blogMapper;

    public MeService(AuthService authService, UserMapper userMapper, FavoriteMapper favoriteMapper,
                     RouteMapper routeMapper, TravelBlogMapper blogMapper) {
        this.authService = authService;
        this.userMapper = userMapper;
        this.favoriteMapper = favoriteMapper;
        this.routeMapper = routeMapper;
        this.blogMapper = blogMapper;
    }

    public UserDto getProfile(Long userId) {
        return UserDto.from(authService.getCurrentUser(userId));
    }

    public UserDto updateProfile(Long userId, String nickname, String avatarUrl, String language) {
        User user = authService.getCurrentUser(userId);
        if (nickname != null && !nickname.isBlank()) user.setNickname(nickname);
        if (avatarUrl != null) user.setAvatarUrl(avatarUrl);
        if (language != null) {
            String lang = language.toLowerCase(Locale.ROOT);
            if (!lang.equals("en") && !lang.equals("zh")) {
                throw ApiException.badRequest("language must be 'en' or 'zh'");
            }
            user.setLanguage(lang);
        }
        userMapper.updateProfile(user);
        return UserDto.from(user);
    }

    public StatsDto getStats(Long userId) {
        return new StatsDto(
                favoriteMapper.countByUserId(userId),
                routeMapper.countByUserId(userId),
                blogMapper.countByUserId(userId));
    }
}
