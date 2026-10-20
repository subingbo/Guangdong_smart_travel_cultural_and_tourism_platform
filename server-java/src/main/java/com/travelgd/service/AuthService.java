package com.travelgd.service;

import com.travelgd.common.ApiException;
import com.travelgd.dto.AuthResponse;
import com.travelgd.dto.LoginRequest;
import com.travelgd.dto.RegisterRequest;
import com.travelgd.dto.UserDto;
import com.travelgd.mapper.UserMapper;
import com.travelgd.model.User;
import com.travelgd.security.JwtService;
import org.springframework.security.crypto.password.PasswordEncoder;
import org.springframework.stereotype.Service;

@Service
public class AuthService {

    private final UserMapper userMapper;
    private final PasswordEncoder passwordEncoder;
    private final JwtService jwtService;

    public AuthService(UserMapper userMapper, PasswordEncoder passwordEncoder, JwtService jwtService) {
        this.userMapper = userMapper;
        this.passwordEncoder = passwordEncoder;
        this.jwtService = jwtService;
    }

    public AuthResponse register(RegisterRequest req) {
        if (userMapper.countByUsername(req.username()) > 0) {
            throw ApiException.conflict("Username already taken");
        }
        User user = new User();
        user.setUsername(req.username());
        user.setPasswordHash(passwordEncoder.encode(req.password()));
        user.setNickname(req.nickname() == null || req.nickname().isBlank() ? "Guest User" : req.nickname());
        user.setAvatarUrl("https://i.pravatar.cc/150?img=" + (int) (Math.random() * 70 + 1));
        user.setLanguage("en");
        userMapper.insert(user);
        return new AuthResponse(jwtService.issue(user.getId(), user.getUsername()), UserDto.from(user));
    }

    public AuthResponse login(LoginRequest req) {
        User user = userMapper.findByUsername(req.username());
        if (user == null || !passwordEncoder.matches(req.password(), user.getPasswordHash())) {
            throw ApiException.unauthorized("Invalid username or password");
        }
        return new AuthResponse(jwtService.issue(user.getId(), user.getUsername()), UserDto.from(user));
    }

    public User getCurrentUser(Long userId) {
        User user = userMapper.findById(userId);
        if (user == null) {
            throw ApiException.unauthorized("User no longer exists");
        }
        return user;
    }
}
